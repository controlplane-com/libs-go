package bucket

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
)

var identifierRegex = regexp.MustCompilePOSIX(`^[a-zA-Z_][a-zA-Z0-9_-]{0,62}$`)

// IsPostgresIdentifier checks if a string is a valid PostgreSQL identifier
func IsPostgresIdentifier(identifier string) bool {
	return identifierRegex.MatchString(identifier)
}

// ModulatedHash converts the input string to a number in the half-closed interval (0, space]
func ModulatedHash(input string, space int) int {
	hash := sha256.Sum256([]byte(input))
	hashInt := binary.BigEndian.Uint64(hash[:8])
	hashInt = hashInt % uint64(space)
	if hashInt == 0 {
		hashInt = uint64(space)
	}
	return int(hashInt)
}

// SchemaName sanitizes a schema name by replacing hyphens with underscores
func SchemaName(schema string) string {
	return strings.ReplaceAll(schema, "-", "_")
}

// TableName returns a fully qualified table name with schema
func TableName(schema string, tableName string) string {
	return fmt.Sprintf(`"%s".%s`, strings.ReplaceAll(schema, "-", "_"), tableName)
}

// IndexName generates an index name with schema, table, and suffix
func IndexName(schema string, tableName string, suffix string) string {
	return fmt.Sprintf(`%s_%s_idx_%s`, strings.ReplaceAll(schema, "-", "_"), tableName, suffix)
}
