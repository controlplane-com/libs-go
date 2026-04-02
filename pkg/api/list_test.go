package api

import (
	"context"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
)

func TestListResponse_NextLink(t *testing.T) {
	tests := []struct {
		name     string
		links    []base.Link
		expected string
	}{
		{"empty links", nil, ""},
		{"no next link", []base.Link{{Rel: "self", Href: "/self"}}, ""},
		{"has next link", []base.Link{{Rel: "next", Href: "/next?cursor=abc"}}, "/next?cursor=abc"},
		{"multiple links with next", []base.Link{{Rel: "self", Href: "/self"}, {Rel: "next", Href: "/next"}}, "/next"},
		{"multiple links without next", []base.Link{{Rel: "self", Href: "/self"}, {Rel: "prev", Href: "/prev"}}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ListResponse[any]{Links: tt.links}
			if got := r.NextLink(); got != tt.expected {
				t.Errorf("NextLink() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestListResponse_HasNext(t *testing.T) {
	tests := []struct {
		name     string
		links    []base.Link
		expected bool
	}{
		{"with next link", []base.Link{{Rel: "next", Href: "/next"}}, true},
		{"without next link", []base.Link{{Rel: "self", Href: "/self"}}, false},
		{"empty links", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ListResponse[any]{Links: tt.links}
			if got := r.HasNext(); got != tt.expected {
				t.Errorf("HasNext() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestQueryResponse_NextLink(t *testing.T) {
	tests := []struct {
		name     string
		links    []base.Link
		expected string
	}{
		{"empty links", nil, ""},
		{"no next link", []base.Link{{Rel: "self", Href: "/self"}}, ""},
		{"has next link", []base.Link{{Rel: "next", Href: "/next?cursor=abc"}}, "/next?cursor=abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &QueryResponse[any]{Links: tt.links}
			if got := r.NextLink(); got != tt.expected {
				t.Errorf("NextLink() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestQueryResponse_HasNext(t *testing.T) {
	tests := []struct {
		name     string
		links    []base.Link
		expected bool
	}{
		{"with next link", []base.Link{{Rel: "next", Href: "/next"}}, true},
		{"without next link", []base.Link{{Rel: "self", Href: "/self"}}, false},
		{"empty links", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &QueryResponse[any]{Links: tt.links}
			if got := r.HasNext(); got != tt.expected {
				t.Errorf("HasNext() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildPath(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		params   map[string]string
		expected string
	}{
		{"nil params", "/api/resource", nil, "/api/resource"},
		{"empty params", "/api/resource", map[string]string{}, "/api/resource"},
		{"single param", "/api/resource", map[string]string{"limit": "10"}, "/api/resource?limit=10"},
		{"empty value param", "/api/resource", map[string]string{"limit": ""}, "/api/resource"},
		{"mixed params", "/api/resource", map[string]string{"limit": "10", "empty": ""}, "/api/resource?limit=10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPath(tt.base, tt.params)
			if got != tt.expected {
				t.Errorf("buildPath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		name     string
		fullURL  string
		baseURL  string
		expected string
	}{
		{"full URL with base", "https://api.cpln.io/org/test", "https://api.cpln.io", "/org/test"},
		{"full URL with query", "https://api.cpln.io/org/test?next=abc", "https://api.cpln.io", "/org/test?next=abc"},
		{"relative path", "/org/test", "https://api.cpln.io", "/org/test"},
		{"relative path with query", "/org/test?next=abc", "https://api.cpln.io", "/org/test?next=abc"},
		{"different base", "https://other.api.io/org/test", "https://api.cpln.io", "/org/test"},
		{"path only", "/path/to/resource", "", "/path/to/resource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPath(tt.fullURL, tt.baseURL)
			if got != tt.expected {
				t.Errorf("extractPath(%q, %q) = %q, want %q", tt.fullURL, tt.baseURL, got, tt.expected)
			}
		})
	}
}

func TestWithLimit(t *testing.T) {
	opts := &listOptions{}
	WithLimit(50)(opts)
	if opts.limit != 50 {
		t.Errorf("limit = %d, want %d", opts.limit, 50)
	}
}

func TestListResponse_NextPage_NoClient(t *testing.T) {
	r := &ListResponse[any]{Links: []base.Link{{Rel: "next", Href: "/next"}}}
	_, err := r.NextPage(context.Background())
	if err == nil {
		t.Error("NextPage should return error when client is nil")
	}
	expectedMsg := "cannot call NextPage: response was not created by the API client"
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestQueryResponse_NextPage_NoClient(t *testing.T) {
	r := &QueryResponse[any]{Links: []base.Link{{Rel: "next", Href: "/next"}}}
	_, err := r.NextPage(context.Background())
	if err == nil {
		t.Error("NextPage should return error when client is nil")
	}
	expectedMsg := "cannot call NextPage: response was not created by the API client"
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestListResponse_NextPage_NoMorePages(t *testing.T) {
	c := &Client{baseURL: "https://api.cpln.io"}
	r := &ListResponse[any]{client: c} // No links = no more pages
	next, err := r.NextPage(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if next != nil {
		t.Error("NextPage should return nil when no more pages")
	}
}

func TestQueryResponse_NextPage_NoMorePages(t *testing.T) {
	c := &Client{baseURL: "https://api.cpln.io"}
	r := &QueryResponse[any]{client: c} // No links = no more pages
	next, err := r.NextPage(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if next != nil {
		t.Error("NextPage should return nil when no more pages")
	}
}

func TestListResponse_ClientField(t *testing.T) {
	c := &Client{baseURL: "https://api.cpln.io"}
	r := &ListResponse[any]{client: c}
	if r.client != c {
		t.Error("client field should be set")
	}
}

func TestQueryResponse_ClientField(t *testing.T) {
	c := &Client{baseURL: "https://api.cpln.io"}
	r := &QueryResponse[any]{client: c}
	if r.client != c {
		t.Error("client field should be set")
	}
}
