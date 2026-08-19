package data_service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testItem struct {
	Name string `json:"name"`
}

func listServer(t *testing.T, pages map[string]string) (*httptest.Server, *DataServiceClient) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, ok := pages[r.URL.RequestURI()]
		if !ok {
			t.Errorf("unexpected request %s", r.URL.RequestURI())
			w.WriteHeader(404)
			return
		}
		if page == "ERROR" {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(page))
	}))
	t.Cleanup(server.Close)
	return server, NewClient(server.URL, "token", "test")
}

func page(items []string, next string) string {
	links := []map[string]string{}
	if next != "" {
		links = append(links, map[string]string{"rel": "next", "href": next})
	}
	list := map[string]any{"items": func() []testItem {
		var out []testItem
		for _, n := range items {
			out = append(out, testItem{Name: n})
		}
		return out
	}(), "links": links}
	b, _ := json.Marshal(list)
	return string(b)
}

// An empty page with a next link is a continuation, not the end — stopping
// there silently truncates the listing and, for orphan detection, makes every
// item beyond it look untracked.
func TestListIteratorContinuesPastEmptyPage(t *testing.T) {
	_, client := listServer(t, map[string]string{
		"/org":      page([]string{"a"}, "/org?pt=1"),
		"/org?pt=1": page(nil, "/org?pt=2"),
		"/org?pt=2": page([]string{"b", "c"}, ""),
	})
	items, err := NewListIterator[testItem](client, "/org").List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || items[0].Name != "a" || items[1].Name != "b" || items[2].Name != "c" {
		t.Fatalf("expected [a b c], got %v", items)
	}
}

func TestListIteratorSurfacesMidPaginationError(t *testing.T) {
	_, client := listServer(t, map[string]string{
		"/org":           page([]string{"a"}, "/org?pt=broken"),
		"/org?pt=broken": "ERROR",
	})
	items, err := NewListIterator[testItem](client, "/org").List()
	if err == nil {
		t.Fatalf("expected an error for the failing second page, got items %v", items)
	}
}

func TestListIteratorRejectsNonAdvancingPagination(t *testing.T) {
	_, client := listServer(t, map[string]string{
		"/org":      page([]string{"a"}, "/org?pt=1"),
		"/org?pt=1": page(nil, "/org?pt=1"),
	})
	_, err := NewListIterator[testItem](client, "/org").List()
	if err == nil {
		t.Fatal("expected an error when the next link stops advancing")
	}
}

func TestListIteratorEmptyList(t *testing.T) {
	_, client := listServer(t, map[string]string{"/org": page(nil, "")})
	items, err := NewListIterator[testItem](client, "/org").List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no items, got %v", items)
	}
}
