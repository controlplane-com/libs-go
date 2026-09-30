package metering

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetentionCutoffIsMonthAligned(t *testing.T) {
	now := time.Date(2026, 9, 28, 14, 30, 0, 0, time.UTC)
	require.Equal(t, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), RetentionCutoff(now, 3))
	require.Equal(t, time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC), InvoiceRetentionCutoff(now))
}

func TestRetentionCutoffNormalizesToUtc(t *testing.T) {
	now := time.Date(2026, 3, 1, 1, 0, 0, 0, time.FixedZone("CEST", 2*3600))
	require.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), RetentionCutoff(now, 1))
}
