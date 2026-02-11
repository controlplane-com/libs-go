package api

import (
	"context"
	"encoding/json"
	"iter"
	"net/url"
	"strings"

	"github.com/controlplane-com/types-go/pkg/base"
	"github.com/controlplane-com/types-go/pkg/query"
)

// ListResponse represents a paginated list response from the API.
type ListResponse[T any] struct {
	Kind     string      `json:"kind"`
	ItemKind string      `json:"itemKind"`
	Items    []T         `json:"items"`
	Links    []base.Link `json:"links"`
}

// NextLink returns the URL for the next page, or empty string if no more pages.
func (r *ListResponse[T]) NextLink() string {
	for _, link := range r.Links {
		if link.Rel == "next" {
			return link.Href
		}
	}
	return ""
}

// HasNext returns true if there are more pages available.
func (r *ListResponse[T]) HasNext() bool {
	return r.NextLink() != ""
}

// listOptions configures list operations.
type listOptions struct {
	limit int
}

// ListOption is a functional option for list operations.
type ListOption func(*listOptions)

// WithLimit sets the maximum number of items to return per page.
func WithLimit(limit int) ListOption {
	return func(o *listOptions) {
		o.limit = limit
	}
}

// listPage fetches a single page from the given path.
func listPage[T any](ctx context.Context, c *Client, path string) (*ListResponse[T], error) {
	var resp ListResponse[T]
	if err := c.get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// listPageRaw fetches a single page and unmarshals items into the specific type.
func listPageRaw[T any](ctx context.Context, c *Client, path string) (*ListResponse[T], error) {
	body, err := c.doRaw(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ListResponse[T]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// listIterator returns an iterator that yields items from all pages.
func listIterator[T any](ctx context.Context, c *Client, path string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		currentPath := path
		for {
			page, err := listPageRaw[T](ctx, c, currentPath)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}

			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}

			nextLink := page.NextLink()
			if nextLink == "" {
				return
			}

			// Extract path from the next link URL
			currentPath = extractPath(nextLink, c.baseURL)
		}
	}
}

// listAll fetches all items from all pages and returns them as a slice.
func listAll[T any](ctx context.Context, c *Client, path string) ([]T, error) {
	var items []T
	for item, err := range listIterator[T](ctx, c, path) {
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// extractPath extracts the path from a full URL, relative to the base URL.
func extractPath(fullURL, baseURL string) string {
	if strings.HasPrefix(fullURL, baseURL) {
		return fullURL[len(baseURL):]
	}
	// Try to parse as URL and extract path + query
	if u, err := url.Parse(fullURL); err == nil {
		if u.RawQuery != "" {
			return u.Path + "?" + u.RawQuery
		}
		return u.Path
	}
	return fullURL
}

// buildPath constructs an API path with optional query parameters.
func buildPath(basePath string, params map[string]string) string {
	if len(params) == 0 {
		return basePath
	}

	values := url.Values{}
	for key, value := range params {
		if value != "" {
			values.Set(key, value)
		}
	}

	if len(values) == 0 {
		return basePath
	}

	return basePath + "?" + values.Encode()
}

// QueryResponse represents a query result with typed items.
type QueryResponse[T any] struct {
	Kind     query.QueryResultKind `json:"kind,omitempty"`
	ItemKind base.Kind             `json:"itemKind,omitempty"`
	Items    []T                   `json:"items"`
	Links    []base.Link           `json:"links"`
	Query    query.Query           `json:"query,omitempty"`
}

// NextLink returns the URL for the next page, or empty string if no more pages.
func (r *QueryResponse[T]) NextLink() string {
	for _, link := range r.Links {
		if link.Rel == "next" {
			return link.Href
		}
	}
	return ""
}

// HasNext returns true if there are more pages available.
func (r *QueryResponse[T]) HasNext() bool {
	return r.NextLink() != ""
}

// queryPage posts a query and returns a single page of results.
func queryPage[T any](ctx context.Context, c *Client, path string, q *query.Query) (*QueryResponse[T], error) {
	body, err := c.doRaw(ctx, "POST", path, q)
	if err != nil {
		return nil, err
	}

	var resp QueryResponse[T]
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// queryIterator returns an iterator that yields items from all query result pages.
func queryIterator[T any](ctx context.Context, c *Client, path string, q *query.Query) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		// First page uses POST with query
		page, err := queryPage[T](ctx, c, path, q)
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}

		for _, item := range page.Items {
			if !yield(item, nil) {
				return
			}
		}

		// Subsequent pages use the next link
		nextLink := page.NextLink()
		for nextLink != "" {
			currentPath := extractPath(nextLink, c.baseURL)
			listPage, err := listPageRaw[T](ctx, c, currentPath)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}

			for _, item := range listPage.Items {
				if !yield(item, nil) {
					return
				}
			}

			nextLink = listPage.NextLink()
		}
	}
}

// queryAll executes a query and returns all items from all pages as a slice.
func queryAll[T any](ctx context.Context, c *Client, path string, q *query.Query) ([]T, error) {
	var items []T
	for item, err := range queryIterator[T](ctx, c, path, q) {
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
