// Package search holds the column metadata and sort helpers shared by services that expose
// paged, sortable search endpoints.
package search

// SearchColumn defines metadata for a searchable/sortable column
type SearchColumn struct {
	Key                  string `json:"key"`                            // Column key (frontend identifier)
	Label                string `json:"label"`                          // Display name
	Sortable             bool   `json:"sortable"`                       // Can sort by this column
	DefaultSort          bool   `json:"defaultSort"`                    // Is this the default sort column
	DefaultSortDirection string `json:"defaultSortDirection,omitempty"` // "asc" or "desc" for default sort column
	DbColumn             string `json:"-"`                              // Override DB column name when different from snake_case(Key)
}

// ColumnsResponse returns available columns for an entity type
type ColumnsResponse struct {
	Columns []SearchColumn `json:"columns"`
}

// SearchRequest is the paging and sorting shape every search request shares. Embed it in a
// service's request type and add the entity-specific filters alongside.
type SearchRequest struct {
	Query         string `json:"query"`
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	SortColumn    string `json:"sortColumn,omitempty"`
	SortDirection string `json:"sortDirection,omitempty"` // "asc" or "desc"
}
