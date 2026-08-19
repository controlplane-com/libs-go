package schedule

import (
	"context"
	"testing"
	"time"
)

func mustParseLastDay(t *testing.T, spec string) *lastDayOfMonthSchedule {
	t.Helper()
	sched, ok, err := parseLastDaySchedule(spec)
	if err != nil {
		t.Fatalf("parseLastDaySchedule(%q) unexpected error: %v", spec, err)
	}
	if !ok {
		t.Fatalf("parseLastDaySchedule(%q) did not recognise the L token", spec)
	}
	ld, ok := sched.(*lastDayOfMonthSchedule)
	if !ok {
		t.Fatalf("parseLastDaySchedule(%q) returned %T, want *lastDayOfMonthSchedule", spec, sched)
	}
	return ld
}

func TestLastDayOfMonthNext(t *testing.T) {
	utc := time.UTC
	sched := mustParseLastDay(t, "0 0 L * *")

	cases := []struct {
		name string
		from time.Time
		want time.Time
	}{
		{
			name: "mid-january -> 31 jan",
			from: time.Date(2026, time.January, 15, 12, 0, 0, 0, utc),
			want: time.Date(2026, time.January, 31, 0, 0, 0, 0, utc),
		},
		{
			name: "non-leap february -> 28 feb",
			from: time.Date(2026, time.February, 1, 0, 0, 0, 0, utc),
			want: time.Date(2026, time.February, 28, 0, 0, 0, 0, utc),
		},
		{
			name: "leap february -> 29 feb",
			from: time.Date(2024, time.February, 1, 0, 0, 0, 0, utc),
			want: time.Date(2024, time.February, 29, 0, 0, 0, 0, utc),
		},
		{
			name: "30-day month -> 30 apr",
			from: time.Date(2026, time.April, 10, 0, 0, 0, 0, utc),
			want: time.Date(2026, time.April, 30, 0, 0, 0, 0, utc),
		},
		{
			name: "exactly on last-day fire rolls to next month",
			from: time.Date(2026, time.January, 31, 0, 0, 0, 0, utc),
			want: time.Date(2026, time.February, 28, 0, 0, 0, 0, utc),
		},
		{
			name: "december crosses year boundary",
			from: time.Date(2026, time.December, 31, 0, 0, 1, 0, utc),
			want: time.Date(2027, time.January, 31, 0, 0, 0, 0, utc),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sched.Next(tc.from)
			if !got.Equal(tc.want) {
				t.Fatalf("Next(%s) = %s, want %s", tc.from, got, tc.want)
			}
		})
	}
}

func TestLastDayHonoursHourMinuteAndMonthList(t *testing.T) {
	utc := time.UTC
	// 23:30 on the last day of March, June, September, December only.
	sched := mustParseLastDay(t, "30 23 L 3,6,9,12 *")

	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, utc)
	got := sched.Next(from)
	want := time.Date(2026, time.June, 30, 23, 30, 0, 0, utc)
	if !got.Equal(want) {
		t.Fatalf("Next(%s) = %s, want %s (should skip disallowed months)", from, got, want)
	}
}

func TestParseLastDayRejectsUnsupported(t *testing.T) {
	bad := []string{
		"* 0 L * *",    // wildcard minute not supported with L
		"0 * L * *",    // wildcard hour not supported with L
		"0 0 L * 1",    // day-of-week must be wildcard
		"0 0 L,15 * *", // partial L expression
		"0 0 L 13 *",   // month out of range
	}
	for _, spec := range bad {
		t.Run(spec, func(t *testing.T) {
			if _, _, err := parseLastDaySchedule(spec); err == nil {
				t.Fatalf("parseLastDaySchedule(%q) expected error, got nil", spec)
			}
		})
	}
}

func TestParseLastDayIgnoresStandardSpecs(t *testing.T) {
	standard := []string{"0 0 1 * *", "*/5 * * * *", "0 0 * * MON"}
	for _, spec := range standard {
		t.Run(spec, func(t *testing.T) {
			sched, ok, err := parseLastDaySchedule(spec)
			if err != nil {
				t.Fatalf("parseLastDaySchedule(%q) unexpected error: %v", spec, err)
			}
			if ok || sched != nil {
				t.Fatalf("parseLastDaySchedule(%q) should defer to the standard parser", spec)
			}
		})
	}
}

// TestSchedulerAcceptsLastDayViaParse exercises the parse() integration point so
// a last-day spec flows through Add/Next like any other schedule.
func TestSchedulerAcceptsLastDayViaParse(t *testing.T) {
	ClearRejectedSchedulesCache()
	s := NewScheduler[*testItem](func(_ context.Context, _ *testItem) error { return nil }, DefaultConfig())
	if err := s.ValidateSchedule("0 0 L * *"); err != nil {
		t.Fatalf("ValidateSchedule(last-day) failed: %v", err)
	}
	sched, err := s.parse("0 0 L * *")
	if err != nil {
		t.Fatalf("parse(last-day) failed: %v", err)
	}
	next := sched.Next(time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC))
	want := time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("scheduler parse Next = %s, want %s", next, want)
	}
}
