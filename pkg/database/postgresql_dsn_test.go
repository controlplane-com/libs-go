package database

import (
	"strings"
	"testing"
	"time"
)

// TestBuildDSNStatementTimeout pins the statement_timeout DSN wiring (ports
// connection.ts afterCreate `SET statement_timeout`): set → `options='-c
// statement_timeout=<ms>'`; unset (0) → no options (opt-in, other services
// unaffected).
func TestBuildDSNStatementTimeout(t *testing.T) {
	base := &ConnectionConfiguration{Host: "h", Port: "5432", User: "u", Password: "p", Database: "d"}

	dsn := buildDSN(base)
	if strings.Contains(dsn, "statement_timeout") || strings.Contains(dsn, "options") {
		t.Fatalf("unset StatementTimeout must not add options; got %q", dsn)
	}

	withTimeout := &ConnectionConfiguration{Host: "h", Port: "5432", User: "u", Password: "p", Database: "d", StatementTimeout: 30 * time.Second}
	dsn = buildDSN(withTimeout)
	if !strings.Contains(dsn, "options='-c statement_timeout=30000'") {
		t.Fatalf("30s StatementTimeout must add `options='-c statement_timeout=30000'`; got %q", dsn)
	}
}
