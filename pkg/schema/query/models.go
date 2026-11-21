/* hand-maintained — canonical source of truth for the `query` schema.
 *
 * These structs (formerly emitted by joi-to-go from nodelibs/schema/src/query.ts)
 * are now the CANONICAL definition. `validate:` struct tags carry the validation
 * rules; validate.FromType reflects over these types to build the *validate.Rule
 * tree the engine runs. See nodelibs/schema/src/query.ts for the Joi source, and
 * go-libs/schema/ipSet/models.go for the tag DSL.
 *
 * Query is EMBEDDED by other resources (policy.targetQuery, gvc
 * spec.staticPlacement.locationQuery, group spec.memberQuery, …), so the rules
 * below are the ONE place the Joi defaults (op '=', match 'all', sort.order
 * 'asc', fetch 'items') and the term structure (op enum, xor, value) are
 * defined — for both the stored-query path and the `-query` request path
 * (pkg/query.Validate).
 */

package query

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type QueryContext map[string]any

type QueryFetch string

const (
	QueryFetchLinks QueryFetch = "links"
	QueryFetchItems QueryFetch = "items"
)

// EnumValues ports `Joi.string().valid('links','items')` (query.ts fetch).
func (QueryFetch) EnumValues() []string {
	return []string{string(QueryFetchLinks), string(QueryFetchItems)}
}

type Query struct {
	Kind    base.Kind     `json:"kind,omitzero"`
	Context *QueryContext `json:"context,omitzero" validate:"readOnly"`
	Fetch   QueryFetch    `json:"fetch,omitzero" validate:"default=items,optional"`
	Spec    *Spec         `json:"spec,omitzero"`
}

type QueryResultKind string

const (
	QueryResultKindQueryresult QueryResultKind = "queryresult"
)

type QueryResult struct {
	Kind     QueryResultKind `json:"kind,omitzero"`
	ItemKind base.Kind       `json:"itemKind,omitzero"`
	Items    []any           `json:"items"`
	Links    []base.Link     `json:"links"`
	Query    *Query          `json:"query,omitzero"`
}

type SpecMatch string

const (
	SpecMatchAll  SpecMatch = "all"
	SpecMatchAny  SpecMatch = "any"
	SpecMatchNone SpecMatch = "none"
)

// EnumValues ports `Joi.string().valid('all','any','none')` (query.ts match).
func (SpecMatch) EnumValues() []string {
	return []string{string(SpecMatchAll), string(SpecMatchAny), string(SpecMatchNone)}
}

type SpecSortOrder string

const (
	SpecSortOrderAsc  SpecSortOrder = "asc"
	SpecSortOrderDesc SpecSortOrder = "desc"
)

// EnumValues ports `Joi.string().valid('asc','desc')` (query.ts sort.order).
func (SpecSortOrder) EnumValues() []string {
	return []string{string(SpecSortOrderAsc), string(SpecSortOrderDesc)}
}

type SpecSort struct {
	By    string        `json:"by" validate:"required"`
	Order SpecSortOrder `json:"order,omitzero" validate:"default=asc"`
}

type Spec struct {
	Match SpecMatch `json:"match,omitzero" validate:"trim,default=all"`
	Terms []Term    `json:"terms,omitzero"`
	Sort  *SpecSort `json:"sort,omitzero"`
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

// EnumValues ports the Term.op `Joi.string().valid(...)` enum (query.ts:7-10),
// in Joi's declaration order so the "must be one of [...]" message matches.
func (TermOp) EnumValues() []string {
	return []string{
		string(TermOpEq), string(TermOpGt), string(TermOpGte),
		string(TermOpLt), string(TermOpLte), string(TermOpNe),
		string(TermOpMatch), string(TermOpRegex),
		string(TermOpExists), string(TermOpNotExists), string(TermOpContains),
	}
}

// TermValue is a scalar match value: string | number | boolean | date (RFC3339). Go has no union type; kept any-backed, referenced by name.
type TermValue any //ts:string | number | boolean | Date

type Term struct {
	Op       TermOp    `json:"op,omitzero" validate:"trim,default=="`
	Property string    `json:"property,omitzero" validate:"trim"`
	Rel      string    `json:"rel,omitzero" validate:"trim"`
	Tag      string    `json:"tag,omitzero" validate:"trim"`
	Value    TermValue `json:"value,omitzero"`
}
