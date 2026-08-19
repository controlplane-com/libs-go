package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// lastDayOfMonthSchedule implements robfig's cron.Schedule for the "last day of
// month" case, which the standard 5-field parser cannot express (it has no `L`
// token). It fires at a fixed hour:minute on the last calendar day of every
// allowed month, honouring leap years (e.g. 28-Feb, 29-Feb, 30, 31).
//
// It is intentionally scoped to the pattern `<min> <hour> L <month> *`:
//   - minute and hour must be explicit integers,
//   - month is `*` (all months) or a comma-list of 1-12,
//   - day-of-week must be `*` (last-day ∧ weekday is ambiguous and unsupported).
type lastDayOfMonthSchedule struct {
	minute int
	hour   int
	months uint64 // bit i (1..12) set means month i is allowed
}

// maxMonthSearch bounds the month-by-month search so an (impossible, though not
// currently constructible) schedule returns the zero time rather than looping
// forever — mirroring robfig's ~5-year search-window contract that
// validateNextExecution relies on.
const maxMonthSearch = 60

// Next returns the next activation time strictly after t, in t's location
// (matching the timezone-naive behaviour of the standard parser used elsewhere
// in this package).
func (l *lastDayOfMonthSchedule) Next(t time.Time) time.Time {
	year, month := t.Year(), t.Month()
	for i := 0; i < maxMonthSearch; i++ {
		if l.months&(uint64(1)<<uint(month)) != 0 {
			lastDay := daysInMonth(year, month, t.Location())
			candidate := time.Date(year, month, lastDay, l.hour, l.minute, 0, 0, t.Location())
			if candidate.After(t) {
				return candidate
			}
		}
		if month == time.December {
			month = time.January
			year++
		} else {
			month++
		}
	}
	return time.Time{}
}

// daysInMonth returns the number of days in the given month. time.Date
// normalises day 0 of the following month to the last day of the requested
// month (and normalises month 13 to January of the next year), so this is
// correct for December and for leap-year February.
func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

// parseLastDaySchedule detects and parses a `<min> <hour> L <month> *` cron
// expression. It returns (schedule, true, nil) when the expression uses the `L`
// day-of-month token, (nil, false, nil) when it does not (so the caller falls
// back to the standard parser), and (nil, false, err) when it uses `L` but is
// otherwise malformed/unsupported.
func parseLastDaySchedule(spec string) (cron.Schedule, bool, error) {
	fields := strings.Fields(spec)
	if len(fields) != 5 {
		return nil, false, nil
	}
	dom := fields[2]
	if !strings.ContainsAny(dom, "Ll") {
		return nil, false, nil
	}
	if dom != "L" && dom != "l" {
		return nil, false, fmt.Errorf("unsupported use of 'L' in day-of-month field %q: only a bare 'L' is supported", dom)
	}
	if fields[4] != "*" {
		return nil, false, fmt.Errorf("last-day-of-month schedules require a wildcard day-of-week, got %q", fields[4])
	}

	minute, err := parseIntField(fields[0], 0, 59, "minute")
	if err != nil {
		return nil, false, err
	}
	hour, err := parseIntField(fields[1], 0, 23, "hour")
	if err != nil {
		return nil, false, err
	}
	months, err := parseMonthField(fields[3])
	if err != nil {
		return nil, false, err
	}

	return &lastDayOfMonthSchedule{minute: minute, hour: hour, months: months}, true, nil
}

// parseIntField parses a single integer cron field within [min, max]. Wildcards,
// ranges, lists and steps are intentionally not supported for minute/hour when
// combined with `L`, to keep last-day semantics unambiguous.
func parseIntField(field string, min, max int, name string) (int, error) {
	v, err := strconv.Atoi(field)
	if err != nil {
		return 0, fmt.Errorf("last-day-of-month schedules require an explicit integer %s, got %q", name, field)
	}
	if v < min || v > max {
		return 0, fmt.Errorf("%s %d out of range [%d,%d]", name, v, min, max)
	}
	return v, nil
}

// parseMonthField parses the month field as `*` (all months) or a comma-list of
// integers 1-12, returning a bitmask with bit i set for each allowed month i.
func parseMonthField(field string) (uint64, error) {
	if field == "*" {
		var mask uint64
		for m := 1; m <= 12; m++ {
			mask |= uint64(1) << uint(m)
		}
		return mask, nil
	}
	var mask uint64
	for _, part := range strings.Split(field, ",") {
		m, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return 0, fmt.Errorf("invalid month %q: only '*' or a comma-list of 1-12 is supported with 'L'", part)
		}
		if m < 1 || m > 12 {
			return 0, fmt.Errorf("month %d out of range [1,12]", m)
		}
		mask |= uint64(1) << uint(m)
	}
	return mask, nil
}
