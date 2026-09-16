package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var columns = []SearchColumn{
	{Key: "name", Label: "Name", Sortable: true},
	{Key: "created", Label: "Created", Sortable: true, DefaultSort: true, DefaultSortDirection: "desc"},
	{Key: "accountId", Label: "Account", Sortable: true, DbColumn: "a.account_id"},
	{Key: "description", Label: "Description", Sortable: false},
}

func TestGetDefaultSortColumn(t *testing.T) {
	assert.Equal(t, "created", GetDefaultSortColumn(columns))
	assert.Equal(t, "name", GetDefaultSortColumn(columns[:1]), "first column when none is marked default")
	assert.Equal(t, "created", GetDefaultSortColumn(nil), "hard-coded fallback with no columns")
}

func TestValidateSortColumn(t *testing.T) {
	assert.Equal(t, "name", ValidateSortColumn("name", columns))
	assert.Equal(t, "created", ValidateSortColumn("", columns), "empty falls back to default")
	assert.Equal(t, "created", ValidateSortColumn("description", columns), "unsortable falls back to default")
	assert.Equal(t, "created", ValidateSortColumn("nope; drop table", columns), "unknown falls back to default")
}

func TestValidateSortDirection(t *testing.T) {
	assert.Equal(t, "asc", ValidateSortDirection("ASC"))
	assert.Equal(t, "desc", ValidateSortDirection("desc"))
	assert.Equal(t, "desc", ValidateSortDirection(""))
	assert.Equal(t, "desc", ValidateSortDirection("sideways"))
}

func TestBuildOrderClause(t *testing.T) {
	assert.Equal(t, "created DESC", BuildOrderClause("", "", columns))
	assert.Equal(t, "name ASC", BuildOrderClause("name", "asc", columns))
	assert.Equal(t, "a.account_id DESC", BuildOrderClause("accountId", "desc", columns), "explicit DbColumn wins")
	assert.Equal(t, "created ASC", BuildOrderClause("description", "asc", columns), "unsortable column falls back but keeps direction")

	camel := []SearchColumn{{Key: "lastModified", Sortable: true, DefaultSort: true}}
	assert.Equal(t, "last_modified DESC", BuildOrderClause("lastModified", "", camel), "camelCase key maps to snake_case")
}
