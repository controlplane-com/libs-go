package bucket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	cplnErrors "github.com/controlplane-com/libs-go/pkg/errors"
)

// fakeOrgRepo is a PartitionedRepository whose org store lives in memory. Only the org
// methods exercised by the caching wrapper are implemented.
type fakeOrgRepo struct {
	PartitionedRepository
	orgs        map[string]*Org
	getErr      error
	createErr   error
	getCalls    int
	createCalls int
}

func newFakeOrgRepo(names ...string) *fakeOrgRepo {
	repo := &fakeOrgRepo{orgs: map[string]*Org{}}
	for _, name := range names {
		repo.orgs[name] = &Org{Name: name, BucketId: 1, Bucket: &Bucket{Id: 1}}
	}
	return repo
}

func (f *fakeOrgRepo) GetOrg(name string) (*Org, error) {
	f.getCalls++
	if f.getErr != nil {
		return nil, f.getErr
	}
	org, ok := f.orgs[name]
	if !ok {
		return nil, cplnErrors.NewErrorCode(fmt.Sprintf("invalid org: %s", name), http.StatusNotFound)
	}
	return org, nil
}

func (f *fakeOrgRepo) CreateOrg(name string) error {
	f.createCalls++
	if f.createErr != nil {
		return f.createErr
	}
	f.orgs[name] = &Org{Name: name, BucketId: 1, Bucket: &Bucket{Id: 1}}
	return nil
}

// fakeOrgCache is an in-memory CollectionCache[Org].
type fakeOrgCache struct {
	mutex sync.Mutex
	items map[string]*Org
	valid bool
}

func newFakeOrgCache() *fakeOrgCache {
	return &fakeOrgCache{items: map[string]*Org{}}
}

func (f *fakeOrgCache) Exists(_ context.Context, name string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	_, ok := f.items[name]
	return ok, nil
}

func (f *fakeOrgCache) Get(_ context.Context, name string) (*Org, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	org, ok := f.items[name]
	if !ok {
		return nil, errors.New("cache miss")
	}
	return org, nil
}

func (f *fakeOrgCache) Set(_ context.Context, item *Org) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.items[item.Name] = item
	return nil
}

func (f *fakeOrgCache) SetMany(ctx context.Context, items []*Org) error {
	for _, item := range items {
		if err := f.Set(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeOrgCache) Delete(_ context.Context, name string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.items, name)
	return nil
}

func (f *fakeOrgCache) DeleteMany(ctx context.Context, names []string) error {
	for _, name := range names {
		if err := f.Delete(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeOrgCache) ListAll(_ context.Context) ([]*Org, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	orgs := make([]*Org, 0, len(f.items))
	for _, org := range f.items {
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (f *fakeOrgCache) Clear(_ context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.items = map[string]*Org{}
	f.valid = false
	return nil
}

func (f *fakeOrgCache) IsValid(_ context.Context) (bool, error) { return f.valid, nil }

func (f *fakeOrgCache) MarkValid(_ context.Context, _ time.Duration) error {
	f.valid = true
	return nil
}

func (f *fakeOrgCache) Close() error { return nil }

// fakeResolver records lookups and returns a canned answer.
type fakeResolver struct {
	exists bool
	err    error
	calls  int
	names  []string
}

func (f *fakeResolver) ResolveOrg(_ context.Context, name string) (bool, error) {
	f.calls++
	f.names = append(f.names, name)
	return f.exists, f.err
}

func TestGetOrgResolvesUnknownOrg(t *testing.T) {
	inner := newFakeOrgRepo()
	orgCache := newFakeOrgCache()
	resolver := &fakeResolver{exists: true}
	repo := NewCachingRepository(inner, orgCache, time.Minute, WithOrgResolver(resolver))

	org, err := repo.GetOrg("new-org")
	require.NoError(t, err)
	require.Equal(t, "new-org", org.Name)
	require.Equal(t, 1, resolver.calls)
	require.Equal(t, []string{"new-org"}, resolver.names)
	require.Equal(t, 1, inner.createCalls)

	// the resolved org is cached, so a second lookup asks neither the repo nor the resolver
	getCalls := inner.getCalls
	org, err = repo.GetOrg("new-org")
	require.NoError(t, err)
	require.Equal(t, "new-org", org.Name)
	require.Equal(t, getCalls, inner.getCalls)
	require.Equal(t, 1, resolver.calls)
}

func TestGetOrgKeepsNotFoundWhenResolverDeniesOrg(t *testing.T) {
	inner := newFakeOrgRepo()
	resolver := &fakeResolver{exists: false}
	repo := NewCachingRepository(inner, newFakeOrgCache(), time.Minute, WithOrgResolver(resolver))

	org, err := repo.GetOrg("missing-org")
	require.Nil(t, org)
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, cplnErrors.ToHTTPStatus(err))
	require.Equal(t, 1, resolver.calls)
	require.Equal(t, 0, inner.createCalls)
}

func TestGetOrgFailsWhenResolverFails(t *testing.T) {
	inner := newFakeOrgRepo()
	resolver := &fakeResolver{err: errors.New("data-service unreachable")}
	repo := NewCachingRepository(inner, newFakeOrgCache(), time.Minute, WithOrgResolver(resolver))

	org, err := repo.GetOrg("missing-org")
	require.Nil(t, org)
	require.ErrorContains(t, err, "data-service unreachable")
	require.Equal(t, 0, inner.createCalls)
}

func TestGetOrgDoesNotResolveNonNotFoundErrors(t *testing.T) {
	inner := newFakeOrgRepo()
	inner.getErr = errors.New("connection refused")
	resolver := &fakeResolver{exists: true}
	repo := NewCachingRepository(inner, newFakeOrgCache(), time.Minute, WithOrgResolver(resolver))

	org, err := repo.GetOrg("some-org")
	require.Nil(t, org)
	require.ErrorContains(t, err, "connection refused")
	require.Equal(t, 0, resolver.calls)
}

func TestGetOrgWithoutResolverIsUnchanged(t *testing.T) {
	inner := newFakeOrgRepo()
	repo := NewCachingRepository(inner, newFakeOrgCache(), time.Minute)

	org, err := repo.GetOrg("missing-org")
	require.Nil(t, org)
	require.Equal(t, http.StatusNotFound, cplnErrors.ToHTTPStatus(err))
	require.Equal(t, 0, inner.createCalls)
}

func TestGetOrgServesKnownOrgWithoutResolving(t *testing.T) {
	inner := newFakeOrgRepo("known-org")
	resolver := &fakeResolver{exists: true}
	repo := NewCachingRepository(inner, newFakeOrgCache(), time.Minute, WithOrgResolver(resolver))

	org, err := repo.GetOrg("known-org")
	require.NoError(t, err)
	require.Equal(t, "known-org", org.Name)
	require.Equal(t, 0, resolver.calls)
}

func TestGetOrgSurfacesCreateFailureOfResolvedOrg(t *testing.T) {
	inner := newFakeOrgRepo()
	inner.createErr = errors.New("bucket assignment failed")
	repo := NewCachingRepository(inner, newFakeOrgCache(), time.Minute, WithOrgResolver(&fakeResolver{exists: true}))

	org, err := repo.GetOrg("new-org")
	require.Nil(t, org)
	require.ErrorContains(t, err, "bucket assignment failed")
}
