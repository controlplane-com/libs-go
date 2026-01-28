package time_utils

import (
	"math"
	"time"
)

var (
	zero = time.Time{}
)

func newDirectionalTime(t time.Time, forward bool) directionalTime {
	if forward {
		return directionalTime{Time: t, forward: 1}
	}
	return directionalTime{Time: t, forward: -1}
}

type directionalTime struct {
	time.Time
	forward int
}

func (d directionalTime) lt(o directionalTime) bool {
	return d.number() < o.number()
}

func (d directionalTime) lte(o directionalTime) bool {
	return d.number() <= o.number()
}

func (d directionalTime) gt(o directionalTime) bool {
	return d.number() > o.number()
}

func (d directionalTime) gte(o directionalTime) bool {
	return d.number() >= o.number()
}

func (d directionalTime) number() float64 {
	if d.Time != zero {
		return float64(d.UnixNano())
	}
	return math.Inf(d.forward)
}

func TimeBetweenInclusive(now time.Time, startTime time.Time, endTime time.Time) bool {
	now = now.UTC()
	startTime = startTime.UTC()
	endTime = endTime.UTC()
	return greaterThanEqualStart(now, startTime) && lessThanEqualEnd(now, endTime)
}

func TimeBetweenExclusive(now time.Time, startTime time.Time, endTime time.Time) bool {
	now = now.UTC()
	startTime = startTime.UTC()
	endTime = endTime.UTC()
	return greaterThanStart(now, startTime) && lessThanEnd(now, endTime)
}

func Compare(a time.Time, b time.Time) bool {
	return a.UTC() == b.UTC()
}

func PtrCompare(a *time.Time, b *time.Time) bool {
	if a == nil {
		a = Zero()
	}
	if b == nil {
		b = Zero()
	}
	return *a == *b
}

func IsZero(a *time.Time) bool {
	if a == nil {
		return true
	}
	return *a == zero
}

func Zero() *time.Time {
	z := zero
	return &z
}

func PtrTo(t time.Time) *time.Time {
	if t == zero {
		return nil
	}
	return &t
}

func Copy(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return PtrTo(t.UTC())
}

func lessThanEnd(toTest time.Time, endTime time.Time) bool {
	if endTime == zero {
		return true
	}
	return toTest.Before(endTime)
}

func lessThanEqualEnd(toTest time.Time, endTime time.Time) bool {
	if endTime == zero {
		return true
	}
	return !toTest.After(endTime)
}

func lessThanStart(toTest time.Time, startTime time.Time) bool {
	return toTest.Before(startTime)
}

func lessThanEqualStart(toTest time.Time, startTime time.Time) bool {
	return !toTest.After(startTime)
}

func greaterThanEnd(toTest time.Time, endTime time.Time) bool {
	if endTime == zero {
		return false
	}
	return toTest.After(endTime)
}

func greaterThanEqualEnd(toTest time.Time, endTime time.Time) bool {
	if endTime == zero {
		return toTest == zero
	}
	return !toTest.Before(endTime)
}

func greaterThanStart(toTest time.Time, startTime time.Time) bool {
	return toTest.After(startTime)
}

func greaterThanEqualStart(toTest time.Time, startTime time.Time) bool {
	return !toTest.Before(startTime)
}
