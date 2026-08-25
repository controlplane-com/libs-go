package bucket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/controlplane-com/libs-go/pkg/cache"
	cplnErrors "github.com/controlplane-com/libs-go/pkg/errors"
	"github.com/controlplane-com/libs-go/pkg/logging"
)

// OrgResolver answers whether an org that is unknown locally exists upstream.
type OrgResolver interface {
	// ResolveOrg returns (false, nil) only when the org definitively does not exist.
	// Any other failure must be reported as an error so the caller can retry later.
	ResolveOrg(ctx context.Context, name string) (bool, error)
}

// CachingOption configures optional CachingRepository behavior.
type CachingOption func(*CachingRepository)

// WithOrgResolver makes GetOrg fall back to resolver for orgs that are in neither the
// cache nor the database, creating them locally when the resolver confirms they exist.
func WithOrgResolver(resolver OrgResolver) CachingOption {
	return func(r *CachingRepository) {
		r.orgResolver = resolver
	}
}

// CachingRepository is a wrapper that adds CollectionCache-based caching to any PartitionedRepository implementation.
type CachingRepository struct {
	PartitionedRepository
	collectionCache cache.CollectionCache[Org]
	orgCacheTtl     time.Duration
	orgResolver     OrgResolver
}

// NewCachingRepository wraps an existing PartitionedRepository with caching behavior for org collection operations.
// If collectionCache is nil, the wrapper behaves as a transparent pass-through.
func NewCachingRepository(inner PartitionedRepository, collectionCache cache.CollectionCache[Org], orgCacheTtl time.Duration, options ...CachingOption) PartitionedRepository {
	r := &CachingRepository{
		PartitionedRepository: inner,
		collectionCache:       collectionCache,
		orgCacheTtl:           orgCacheTtl,
	}
	for _, option := range options {
		option(r)
	}
	return r
}

// CreateOrg checks cache for existence to short-circuit, delegates creation, then updates cache.
func (r *CachingRepository) CreateOrg(orgName string) error {
	if r.collectionCache != nil {
		exists, err := r.collectionCache.Exists(context.Background(), orgName)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
	}
	if err := r.PartitionedRepository.CreateOrg(orgName); err != nil {
		return err
	}
	if r.collectionCache != nil {
		if org, err := r.PartitionedRepository.GetOrg(orgName); err == nil {
			_ = r.collectionCache.Set(context.Background(), org)
		}
	}
	return nil
}

// GetOrg attempts cache, falls back to delegate, and then writes the single org to cache.
func (r *CachingRepository) GetOrg(name string) (*Org, error) {
	if r.collectionCache != nil {
		if org, err := r.collectionCache.Get(context.Background(), name); err == nil {
			return org, nil
		}
	}
	org, err := r.PartitionedRepository.GetOrg(name)
	if err != nil {
		if org, err = r.resolveMissingOrg(name, err); err != nil {
			return nil, err
		}
	}
	if r.collectionCache != nil {
		_ = r.collectionCache.SetMany(context.Background(), []*Org{org})
	}
	return org, nil
}

// resolveMissingOrg gives an org the local store has never seen a single chance to be
// resolved upstream. getErr is returned unchanged unless the org is confirmed to exist,
// so an org that genuinely does not exist keeps failing exactly as it did before.
func (r *CachingRepository) resolveMissingOrg(name string, getErr error) (*Org, error) {
	if r.orgResolver == nil || !isNotFound(getErr) {
		return nil, getErr
	}
	exists, err := r.orgResolver.ResolveOrg(context.Background(), name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve unknown org %s: %w", name, err)
	}
	if !exists {
		return nil, getErr
	}
	logging.Logger().Sugar().Infof("org %s was missing locally but exists upstream, adding it", name)
	if err := r.PartitionedRepository.CreateOrg(name); err != nil {
		return nil, fmt.Errorf("failed to create resolved org %s: %w", name, err)
	}
	return r.PartitionedRepository.GetOrg(name)
}

// isNotFound reports whether err means "no such org" rather than a failed lookup.
func isNotFound(err error) bool {
	if cplnErrors.IsDomainError(err, cplnErrors.ErrKindNotFound) {
		return true
	}
	var code cplnErrors.ErrorCode
	if errors.As(err, &code) {
		return code.Code() == http.StatusNotFound
	}
	return false
}

// ListOrgs serves from cache when valid; otherwise loads from delegate, fills cache, marks valid, then filters.
func (r *CachingRepository) ListOrgs(pattern string) ([]*Org, error) {
	if r.collectionCache == nil {
		return r.PartitionedRepository.ListOrgs(pattern)
	}
	if valid, err := r.collectionCache.IsValid(context.Background()); err == nil && valid {
		if orgs, err := r.collectionCache.ListAll(context.Background()); err == nil {
			return r.filterOrgList(orgs, pattern)
		}
	}
	// Load fresh full list from the inner repo (unfiltered)
	orgs, err := r.PartitionedRepository.ListOrgs("")
	if err != nil {
		return nil, err
	}
	// Refresh cache and mark as valid
	_ = r.collectionCache.Clear(context.Background())
	_ = r.collectionCache.SetMany(context.Background(), orgs)
	_ = r.collectionCache.MarkValid(context.Background(), r.orgCacheTtl)
	return r.filterOrgList(orgs, pattern)
}

// FilterOrgs delegates to inner by default.
func (r *CachingRepository) FilterOrgs(orgs []*Org, filter string) ([]*Org, error) {
	return r.PartitionedRepository.FilterOrgs(orgs, filter)
}

// ListBucketOrgs delegates (no collection-level cache here because it's already part of the full orgs cache)
func (r *CachingRepository) ListBucketOrgs(bucket *Bucket) ([]*Org, error) {
	return r.PartitionedRepository.ListBucketOrgs(bucket)
}

// ListOrgsByName reuses ListOrgs with a compiled regex pattern.
func (r *CachingRepository) ListOrgsByName(names []string) ([]*Org, error) {
	pattern := "/^(" + stringJoin(names, "|") + ")$/"
	return r.ListOrgs(pattern)
}

// MoveOrgToBucket clears cache before delegating to ensure consistency.
func (r *CachingRepository) MoveOrgToBucket(org *Org, bucket *Bucket) error {
	if r.collectionCache != nil {
		_ = r.collectionCache.Clear(context.Background())
	}
	return r.PartitionedRepository.MoveOrgToBucket(org, bucket)
}

// Helpers copied from Postgres repo to keep filtering logic consistent
func (r *CachingRepository) filterOrgList(orgs []*Org, pattern string) ([]*Org, error) {
	if pattern == "" {
		return orgs, nil
	}
	var filteredOrgs []*Org
	lenPattern := len(pattern)
	if lenPattern >= 2 && pattern[0] == '/' && pattern[lenPattern-1] == '/' {
		regex, err := regexp.CompilePOSIX(pattern[1 : lenPattern-1])
		if err != nil {
			return nil, err
		}
		for _, org := range orgs {
			if regex.MatchString(org.Name) {
				filteredOrgs = append(filteredOrgs, org)
			}
		}
		return filteredOrgs, nil
	}
	for _, org := range orgs {
		if org.Name == pattern {
			return []*Org{org}, nil
		}
	}
	return nil, nil
}
