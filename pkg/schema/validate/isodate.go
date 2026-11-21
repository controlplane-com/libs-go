package validate

// isodate.go ports Joi's ISO-8601 gate (`Joi.date().iso()` → common.js isoDate).
//
// The Joi pattern uses two negative lookaheads and a backreference, none of which RE2
// has, so it is split into the four separator combinations plus the two lookaheads as
// explicit checks. `.iso()` used to be `return r.clone()` — a silent no-op that accepted
// "2020/01/02", "Jan 2 2020" and epoch-millis strings on every field carrying it.

import (
	"fmt"
	"regexp"
	"strings"
)

// isoPattern is Joi's isoDate with the date separator (%[1]s) and the time separator
// (%[2]s) substituted in, standing in for the `(-?)`/`(:?)` captures and their
// backreferences.
const isoPattern = `^(?:[-+]\d{2})?\d{4}` +
	`(?:%[1]s(?:(?:0[1-9]|1[0-2])(?:%[1]s(?:[12]\d|0[1-9]|3[01]))?` +
	`|W(?:[0-4]\d|5[0-2])(?:-?[1-7])?` +
	`|(?:00[1-9]|0[1-9]\d|[12]\d{2}|3(?:[0-5]\d|6[1-6])))` +
	`(?:[T\s](?:(?:(?:[01]\d|2[0-3])(?:%[2]s[0-5]\d)?|24:?00)(?:[.,]\d+)?)` +
	`(?:%[2]s[0-5]\d(?:[.,]\d+)?)?` +
	`(?:Z|[+-](?:[01]\d|2[0-3])(?::?[0-5]\d)?)?)?)?$`

var isoVariants = func() []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, dsep := range []string{"", "-"} {
		for _, tsep := range []string{"", ":"} {
			out = append(out, regexp.MustCompile(fmt.Sprintf(isoPattern, dsep, tsep)))
		}
	}
	return out
}()

// isoBasicZ is the `[T][\d]+Z$` half of Joi's second lookahead: a basic-format time
// ending in Z is rejected, so "20200102T030405Z" is NOT an ISO date to Joi. That is a
// Joi quirk, kept because the port's contract is the TS service's behaviour.
var isoBasicZ = regexp.MustCompile(`T\d+Z$`)

// IsISODate is Joi's isoDate test, for a check made outside a schema walk.
func IsISODate(s string) bool { return isISODate(s) }

// ISODate is Joi's string isoDate() with convert: the value as JS Date.toISOString() renders it; false when not an ISO date.
func ISODate(s string) (string, bool) {
	if !isISODate(s) {
		return "", false
	}
	t, ok := parseDateString(s, true)
	if !ok {
		return "", false
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z"), true
}

func isISODate(s string) bool {
	i := 0
	if len(s) >= 3 && (s[0] == '-' || s[0] == '+') && isDigits(s[1:3]) {
		i = 3
	}
	if len(s) < i+4 || !isDigits(s[i:i+4]) {
		return false
	}
	// (?!\d{2}\b) after the year: "202001" is not a date, "20200102" is
	if j := i + 4; len(s) >= j+2 && isDigits(s[j:j+2]) && (len(s) == j+2 || !isWordByte(s[j+2])) {
		return false
	}
	// (?![T]$|[T][\d]+Z$) after the date part
	if strings.HasSuffix(s, "T") || isoBasicZ.MatchString(s) {
		return false
	}
	for _, re := range isoVariants {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

func isWordByte(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
