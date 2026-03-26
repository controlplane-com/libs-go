/* auto-generated */

package query

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type QueryContext map[string]any

type QueryFetch string

const (
	QueryFetchLinks QueryFetch = "links"
	QueryFetchItems QueryFetch = "items"
)

type Query struct {
	Kind    base.Kind     `json:"kind,omitempty"`
	Context *QueryContext `json:"context,omitempty"`
	Fetch   QueryFetch    `json:"fetch,omitempty"`
	Spec    *Spec         `json:"spec,omitempty"`
}

type QueryResultKind string

const (
	QueryResultKindQueryresult QueryResultKind = "queryresult"
)

type QueryResult struct {
	Kind     QueryResultKind `json:"kind,omitempty"`
	ItemKind base.Kind       `json:"itemKind,omitempty"`
	Items    []any           `json:"items"`
	Links    []base.Link     `json:"links"`
	Query    Query           `json:"query,omitempty"`
}

type SpecMatch string

const (
	SpecMatchAll  SpecMatch = "all"
	SpecMatchAny  SpecMatch = "any"
	SpecMatchNone SpecMatch = "none"
)

type SpecSortOrder string

const (
	SpecSortOrderAsc  SpecSortOrder = "asc"
	SpecSortOrderDesc SpecSortOrder = "desc"
)

type SpecSort struct {
	By    string        `json:"by"`
	Order SpecSortOrder `json:"order,omitempty"`
}

type Spec struct {
	Match SpecMatch `json:"match,omitempty"`
	Terms []Term    `json:"terms,omitempty"`
	Sort  *SpecSort `json:"sort,omitempty"`
}

type TermOp string

const (
	TermOpEq        TermOp = "="
	TermOpGt        TermOp = ">"
	TermOpGte       TermOp = ">="
	TermOpLt        TermOp = "<"
	TermOpLte       TermOp = "<="
	TermOpNe        TermOp = "!="
	TermOpMatch     TermOp = "~"
	TermOpRegex     TermOp = "=~"
	TermOpExists    TermOp = "exists"
	TermOpNotExists TermOp = "!exists"
	TermOpContains  TermOp = "contains"
)

type Term struct {
	Op       TermOp `json:"op,omitempty"`
	Property string `json:"property,omitempty"`
	Rel      string `json:"rel,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Value    any/* TODO: [object Object]*/ `json:"value,omitempty"`
}
