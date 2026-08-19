package time_utils

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	OneDay  = time.Hour * 24
	OneWeek = OneDay * 7
)

func AssertAdjacentTimeSegmentsAreNonIntersecting(segments []TimeSegment) error {
	if intersectingPairs := FindAdjacentIntersectingTimeSegments(segments, true); intersectingPairs != nil {
		var summary []string
		for _, i := range intersectingPairs {
			summary = append(summary, fmt.Sprintf("The time segments: [%s] and [%s] are intersecting", i.First.Format(time.RFC3339), i.Second.Format(time.RFC3339)))
		}
		return errors.New(fmt.Sprintf("One or more intersecting adjacent time segment pairs found. The following pairs are intersecting:\n%s", strings.Join(summary, "\n")))
	}
	return nil
}

func AssertAdjacentTimeSegmentsAreNonOverlapping(segments []TimeSegment) error {
	if overlappingPairs := FindAdjacentIntersectingTimeSegments(segments, false); overlappingPairs != nil {
		var summary []string
		for _, i := range overlappingPairs {
			summary = append(summary, fmt.Sprintf("[%s] and [%s]", i.First.Format(time.RFC3339), i.Second.Format(time.RFC3339)))
		}
		return errors.New(fmt.Sprintf("One or more overlapping adjacent time segment pairs found. The following pairs are overlapping:\n%s", strings.Join(summary, "\n")))
	}
	return nil
}

func FindAdjacentIntersectingTimeSegments(segments []TimeSegment, includeBoundaries bool) []TimeSegmentPair {
	lenSegments := len(segments)
	var adjacentIntersectingPairs []TimeSegmentPair
	for i := 1; i < lenSegments; i++ {
		p := TimeSegmentPair{
			First:  segments[i-1],
			Second: segments[i],
		}
		if p.First.Intersecting(p.Second, includeBoundaries) {
			adjacentIntersectingPairs = append(adjacentIntersectingPairs, p)
		}
	}
	return adjacentIntersectingPairs
}

func AssertTimeSegmentsAreContinuous(segments []TimeSegment) error {
	if discontinuities := FindDiscontinuities(segments); discontinuities != nil {
		var summary []string
		for _, d := range discontinuities {
			summary = append(summary, fmt.Sprintf("[%s] and [%s]", d.First.Format(time.RFC3339), d.Second.Format(time.RFC3339)))
		}
		return errors.New(fmt.Sprintf("One or more discontinuities found in the given sub segments. There are gaps between segment pairs:\n%s", strings.Join(summary, "\n")))
	}
	return nil
}

func FindDiscontinuities(segments []TimeSegment) []TimeSegmentPair {
	lenSegments := len(segments)
	if lenSegments < 2 {
		return nil
	}
	var discontinuities []TimeSegmentPair
	for i := 1; i < lenSegments; i++ {
		if segments[i-1].End != segments[i].Start {
			discontinuities = append(discontinuities, TimeSegmentPair{First: segments[i-1], Second: segments[i]})
		}
	}
	return discontinuities
}

type TimeStep string

const (
	TimeStepHour  TimeStep = "hour"
	TimeStepDay   TimeStep = "day"
	TimeStepWeek  TimeStep = "week"
	TimeStepMonth TimeStep = "month"
)

func (s TimeStep) int() int {
	switch s {
	case TimeStepHour:
		return 1
	case TimeStepDay:
		return 2
	case TimeStepWeek:
		return 3
	case TimeStepMonth:
		return 4
	default:
		return int(math.Inf(1))
	}
}

// IsFinerThan reports whether this step covers a shorter span than the other. An unrecognized
// step is treated as coarser than every known step.
func (s TimeStep) IsFinerThan(other TimeStep) bool {
	return s.int() < other.int()
}

func (s TimeStep) Advance(t time.Time, steps int) time.Time {
	aligned := AlignTimeWithStepStart(t, s)
	stepCount := time.Duration(steps)
	switch s {
	case TimeStepHour:
		return aligned.Add(time.Hour * stepCount)
	case TimeStepDay:
		return aligned.Add(OneDay * stepCount)
	case TimeStepWeek:
		return aligned.Add(OneWeek * stepCount)
	case TimeStepMonth:
		return AddMonths(aligned, steps)
	}
	return t
}

func TimeIsAlignedWithTimeStep(time time.Time, step TimeStep) bool {
	time = time.UTC()
	return AlignTimeWithStepStart(time, step) == time
}

func AlignTimeWithStepStart(startTime time.Time, step TimeStep) time.Time {
	switch step {
	case TimeStepHour:
		return StartOfTheHour(startTime)
	case TimeStepDay:
		return StartOfTheDay(startTime)
	case TimeStepMonth:
		return FirstDayOfTheMonth(startTime)
	case TimeStepWeek:
		return Weekday(startTime, time.Monday)
	default:
		return startTime
	}
}

func AlignTimeWithStepEnd(endTime time.Time, step TimeStep) time.Time {
	switch step {
	case TimeStepHour:
		e := StartOfTheHour(endTime)
		if e != endTime {
			e = e.Add(time.Hour)
		}
		return e
	case TimeStepDay:
		e := StartOfTheDay(endTime)
		if e != endTime {
			e = e.Add(OneDay)
		}
		return e
	case TimeStepMonth:
		e := FirstDayOfTheMonth(endTime)
		if e != endTime {
			e = AddMonths(e, 1)
		}
		return e
	case TimeStepWeek:
		e := Weekday(endTime, time.Monday)
		if e != endTime {
			e = e.Add(OneDay * 7)
		}
		return e
	default:
		return endTime
	}
}

func AlignTimesWithTimeStep(startTime time.Time, endTime time.Time, step TimeStep) StepAlignedTimeSegment {
	return StepAlignedTimeSegment{TimeSegment: TimeSegment{Start: AlignTimeWithStepStart(startTime, step), End: AlignTimeWithStepEnd(endTime, step)}, TimeStep: step}
}
