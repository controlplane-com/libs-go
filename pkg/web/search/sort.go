package search

import (
	"strings"

	"github.com/iancoleman/strcase"
)

// GetDefaultSortColumn returns the default sort column from a column definition list
func GetDefaultSortColumn(columns []SearchColumn) string {
	for _, col := range columns {
		if col.DefaultSort {
			return col.Key
		}
	}
	// Fallback if no default specified
	if len(columns) > 0 {
		return columns[0].Key
	}
	return "created"
}

// ValidateSortColumn checks if a column is valid and sortable
// Returns the validated column or the default if invalid
func ValidateSortColumn(sortColumn string, columns []SearchColumn) string {
	if sortColumn == "" {
		return GetDefaultSortColumn(columns)
	}

	for _, col := range columns {
		if col.Key == sortColumn && col.Sortable {
			return sortColumn
		}
	}

	// Invalid column requested, return default
	return GetDefaultSortColumn(columns)
}

// ValidateSortDirection ensures direction is either "asc" or "desc"
// Returns validated direction with "desc" as default
func ValidateSortDirection(direction string) string {
	direction = strings.ToLower(direction)
	if direction == "asc" || direction == "desc" {
		return direction
	}
	return "desc"
}

// mapColumnToDb converts a frontend property name to its database column name.
// It first checks if the column has an explicit DbColumn override, then falls back to snake_case conversion.
func mapColumnToDb(column string, columns []SearchColumn) string {
	for _, col := range columns {
		if col.Key == column && col.DbColumn != "" {
			return col.DbColumn
		}
	}
	return strcase.ToSnake(column)
}

// BuildOrderClause constructs a SQL ORDER BY clause
// Validates column and direction before building
// Maps frontend property names to database column names
func BuildOrderClause(sortColumn, sortDirection string, columns []SearchColumn) string {
	validColumn := ValidateSortColumn(sortColumn, columns)
	validDirection := ValidateSortDirection(sortDirection)
	dbColumn := mapColumnToDb(validColumn, columns)
	return dbColumn + " " + strings.ToUpper(validDirection)
}
