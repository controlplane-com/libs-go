package api

import (
	"context"
	"fmt"
	"iter"

	"github.com/controlplane-com/libs-go/pkg/schema/image"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
)

// ImageService handles operations on images.
type ImageService struct {
	client *Client
}

// List returns an iterator over all images in the organization.
func (s *ImageService) List(ctx context.Context, org string) iter.Seq2[image.Image, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/image", org)
	return listIterator[image.Image](ctx, s.client, path)
}

// ListAll returns all images in the organization.
func (s *ImageService) ListAll(ctx context.Context, org string) ([]image.Image, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/image", org)
	return listAll[image.Image](ctx, s.client, path)
}

// ListPage returns a single page of images.
func (s *ImageService) ListPage(ctx context.Context, org, cursor string) (*ListResponse[image.Image], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/image", org)
	if cursor != "" {
		path = buildPath(path, map[string]string{"next": cursor})
	}
	return listPageWithClient[image.Image](ctx, s.client, path)
}

// Get returns an image by name.
func (s *ImageService) Get(ctx context.Context, org, name string) (*image.Image, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/image/%s", org, name)
	var result image.Image
	if err := s.client.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes an image by name.
func (s *ImageService) Delete(ctx context.Context, org, name string) error {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/image/%s", org, name)
	return s.client.delete(ctx, path)
}

// Query returns an iterator over images matching the query.
func (s *ImageService) Query(ctx context.Context, org string, q *query.Query) iter.Seq2[image.Image, error] {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/image/-query", org)
	return queryIterator[image.Image](ctx, s.client, path, q)
}

// QueryAll returns all images matching the query.
func (s *ImageService) QueryAll(ctx context.Context, org string, q *query.Query) ([]image.Image, error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/image/-query", org)
	return queryAll[image.Image](ctx, s.client, path, q)
}

// QueryPage returns a single page of images matching the query.
func (s *ImageService) QueryPage(ctx context.Context, org string, q *query.Query) (*QueryResponse[image.Image], error) {
	org = s.client.resolveOrg(org)
	path := fmt.Sprintf("/org/%s/image/-query", org)
	return queryPageWithClient[image.Image](ctx, s.client, path, q)
}
