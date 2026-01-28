package time_utils

import (
	"errors"
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/pipeline"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/slices"
	"sort"
	"time"
)

type StepAlignedTimeSegment struct {
	TimeSegment
	TimeStep
}

type TimeSegmentPair struct {
	First  TimeSegment
	Second TimeSegment
}

type TimeSegment struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (t TimeSegment) Duration() time.Duration {
	return t.End.Sub(t.Start)
}

func (t TimeSegment) Intersecting(other TimeSegment, includeBoundaries bool) bool {
	if includeBoundaries {
		if t.End.Before(other.Start) {
			return false
		}
		if t.Start.After(other.End) {
			return false
		}
	} else {
		if !t.End.After(other.Start) {
			return false
		}
		if !t.Start.Before(other.End) {
			return false
		}
	}
	return true
}

func (t TimeSegment) UTC() TimeSegment {
	return TimeSegment{Start: t.Start.UTC(), End: t.End.UTC()}
}

func (t TimeSegment) Format(layout string) string {
	return fmt.Sprintf("%s - %s", t.Start.Format(layout), t.End.Format(layout))
}

func (t TimeSegment) Align(timeStep TimeStep) StepAlignedTimeSegment {
	return AlignTimesWithTimeStep(t.Start, t.End, timeStep)
}

func (t TimeSegment) IsAlignedWithTimeStep(timeStep TimeStep) bool {
	aligned := t.Align(timeStep)
	return aligned.Start == t.Start && aligned.End == t.End
}

func ExtractBoundariesFromTimeSegments(timeSegments []TimeSegment, extraBoundaries ...time.Time) ([]time.Time, error) {
	slices.SortFunc(timeSegments, func(a TimeSegment, b TimeSegment) int { return int(a.Start.Sub(b.Start).Milliseconds()) })
	if err := AssertAdjacentTimeSegmentsAreNonOverlapping(timeSegments); err != nil {
		return nil, err
	}
	set := mapset.NewThreadUnsafeSet[int64]()
	for _, s := range timeSegments {
		set.Add(s.Start.Unix())
		set.Add(s.End.Unix())
	}
	for _, b := range extraBoundaries {
		set.Add(b.Unix())
	}
	return pipeline.MustMap(set.ToSlice(), func(i int64) time.Time { return time.Unix(i, 0) }), nil
}

// Decompose attempts to break this TimeSegment into smaller TimeSegments along the given time boundaries. The output
// TimeSegments will be aligned with one of the TimeSteps given by decomposeInto. If one or more of the boundary times
// are not aligned with at least one of the given TimeSteps, the decomposition will fail
func (t TimeSegment) Decompose(boundaries []time.Time, decomposeInto []TimeStep, greedy bool) ([]StepAlignedTimeSegment, error) {
	var smallerTimeSteps []TimeStep
	lenTargetTimeSteps := len(decomposeInto)
	if lenTargetTimeSteps == 0 {
		return nil, errors.New("decomposition failed. An empty slice of TimeSteps was given")
	}
	if lenTargetTimeSteps > 1 {
		sort.Slice(decomposeInto, func(i int, j int) bool { return decomposeInto[i].int() >= decomposeInto[j].int() })
		smallerTimeSteps = decomposeInto[1:]
	}
	if len(boundaries) > 0 {
		slices.SortFunc(boundaries, func(a time.Time, b time.Time) int { return int(a.Sub(b).Milliseconds()) })
	}

	t.Start = t.Start.UTC()
	t.End = t.End.UTC()
	for i, b := range boundaries {
		boundaries[i] = b.UTC()
	}

	//Don't include boundaries outside this time segment
	var clippedBoundaries []time.Time
	for _, b := range boundaries {
		if t.Contains(b, false) {
			clippedBoundaries = append(clippedBoundaries, b)
		}
	}

	return t.decompose(clippedBoundaries, decomposeInto[0], smallerTimeSteps, greedy)
}

func (t TimeSegment) decompose(boundaries []time.Time, targetTimeStep TimeStep, smallerTimeSteps []TimeStep, greedy bool) ([]StepAlignedTimeSegment, error) {
	var alignedTimeSegments []StepAlignedTimeSegment
	var nextStartTime time.Time
	nextBoundary := t.Start
	var boundaryIndex int

	for nextBoundary.Before(t.End) {
		nextStartTime = nextBoundary
		nextBoundary, boundaryIndex = t.findNextBoundaryTime(boundaries, boundaryIndex, nextStartTime)
		nextTimeSegment := TimeSegment{
			Start: nextStartTime,
			End:   nextBoundary,
		}

		if nextTimeSegment.IsAlignedWithTimeStep(targetTimeStep) {
			alignedTimeSegments = append(alignedTimeSegments, StepAlignedTimeSegment{
				TimeSegment: nextTimeSegment,
				TimeStep:    targetTimeStep,
			})
			continue
		}

		if len(smallerTimeSteps) == 0 {
			return nil, errors.New(fmt.Sprintf("decomposition failed. Unable to decompose time segment %s into TimeStep '%s', and no smaller TimeSteps were given", t.Format(time.RFC3339), targetTimeStep))
		}

		timeStepBoundaries := nextTimeSegment.ListTimeStepBoundaries(targetTimeStep)
		countTimeStepBoundaries := len(timeStepBoundaries)

		if countTimeStepBoundaries <= 1 {
			newSegments, err := nextTimeSegment.decompose(getNextBoundaries(boundaries, boundaryIndex, nextTimeSegment), smallerTimeSteps[0], smallerTimeSteps[1:], greedy)
			if err != nil {
				return nil, err
			}
			alignedTimeSegments = append(alignedTimeSegments, newSegments...)
			continue
		}

		if countTimeStepBoundaries > 1 {
			leftAtStart := TimeSegment{Start: nextTimeSegment.Start, End: timeStepBoundaries[0]}
			leftAtEnd := TimeSegment{Start: timeStepBoundaries[countTimeStepBoundaries-1], End: nextTimeSegment.End}

			if leftAtStart.Duration() > 0 {
				startSegments, err := leftAtStart.decompose(getNextBoundaries(boundaries, boundaryIndex, leftAtStart), smallerTimeSteps[0], smallerTimeSteps[1:], greedy)
				if err != nil {
					return nil, err
				}
				alignedTimeSegments = append(alignedTimeSegments, startSegments...)
			}

			if greedy {
				alignedTimeSegments = append(alignedTimeSegments, StepAlignedTimeSegment{
					TimeSegment: TimeSegment{Start: leftAtStart.End, End: leftAtEnd.Start},
					TimeStep:    targetTimeStep,
				})
			} else {
				for i := 1; i < countTimeStepBoundaries; i++ {
					alignedTimeSegments = append(alignedTimeSegments, StepAlignedTimeSegment{
						TimeSegment: TimeSegment{Start: timeStepBoundaries[i-1], End: timeStepBoundaries[i]},
						TimeStep:    targetTimeStep,
					})
				}
			}

			if leftAtEnd.Duration() > 0 {
				endSegments, err := leftAtEnd.decompose(getNextBoundaries(boundaries, boundaryIndex, leftAtEnd), smallerTimeSteps[0], smallerTimeSteps[1:], greedy)
				if err != nil {
					return nil, err
				}
				alignedTimeSegments = append(alignedTimeSegments, endSegments...)
			}
		}
	}
	return alignedTimeSegments, nil
}

func (t TimeSegment) findNextBoundaryTime(boundaryTimes []time.Time, i int, startTime time.Time) (time.Time, int) {
	lenBoundaryTimes := len(boundaryTimes)
	for ; i < lenBoundaryTimes; i++ {
		if !boundaryTimes[i].Before(startTime) {
			return boundaryTimes[i], i + 1
		}
	}
	return t.End, lenBoundaryTimes
}

func getNextBoundaries(boundaries []time.Time, boundaryIndex int, nextTimeSegment TimeSegment) []time.Time {
	lenBoundaries := len(boundaries)
	if lenBoundaries == 0 || boundaryIndex < 0 {
		return nil
	}
	return pipeline.MustFilter(boundaries[boundaryIndex:], func(b time.Time) bool {
		return nextTimeSegment.Contains(b, false)
	})
}

func (t TimeSegment) ListTimeStepBoundaries(timeStep TimeStep) []time.Time {
	var boundaries []time.Time
	_ = t.ExecuteOncePerBoundary(func(t time.Time) error {
		boundaries = append(boundaries, t)
		return nil
	}, timeStep)
	return boundaries
}

func (t TimeSegment) CountTimeStepBoundaries(timeStep TimeStep) int {
	var boundaryCount int
	_ = t.ExecuteOncePerBoundary(func(_ time.Time) error {
		boundaryCount++
		return nil
	}, timeStep)
	return boundaryCount
}

func (t TimeSegment) ExecuteOncePerBoundary(f func(t time.Time) error, timeStep TimeStep) error {
	start := AlignTimeWithStepStart(t.Start, timeStep)
	if start == t.Start {
		if err := f(start); err != nil {
			return err
		}
	}
	for {
		start = timeStep.Advance(start, 1)
		if start.After(t.End) {
			break
		}
		if err := f(start); err != nil {
			return err
		}
	}
	return nil
}

func (t TimeSegment) ExecuteOncePerTimeStep(f func(t TimeSegment) error, timeStep TimeStep) error {
	aligned := t.Align(timeStep)
	start := aligned.Start
	for !start.After(t.End) {
		end := timeStep.Advance(start, 1)
		if err := f(TimeSegment{Start: start, End: end}); err != nil {
			return err
		}
		start = end
	}
	return nil
}

func (t TimeSegment) ExecuteOncePerDuration(function func(t TimeSegment) error, d time.Duration) error {
	for current := t.Start; current.Before(t.End); current = current.Add(d) {
		segmentEndTime := current.Add(d)
		if segmentEndTime.After(t.End) {
			segmentEndTime = t.End
		}
		s, err := NewTimeSegment(current, segmentEndTime)
		if err != nil {
			return err
		}
		if err := function(s); err != nil {
			return err
		}
	}
	return nil
}

func (t TimeSegment) SplitByDuration(d time.Duration) ([]TimeSegment, error) {
	var segments []TimeSegment
	err := t.ExecuteOncePerDuration(func(thisSegment TimeSegment) error {
		segments = append(segments, thisSegment)
		return nil
	}, d)
	if err != nil {
		return nil, err
	}
	return segments, nil
}

// SplitByTimeStep returns a slice of segments aligned with time step boundaries. NOTE: The time covered by the returned TimeSegments will extend beyond the original TimeSegment if it wasn't already boundary aligned
func (t TimeSegment) SplitByTimeStep(timeStep TimeStep) ([]StepAlignedTimeSegment, error) {
	var segments []StepAlignedTimeSegment
	_ = t.ExecuteOncePerTimeStep(func(t TimeSegment) error {
		segments = append(segments, StepAlignedTimeSegment{TimeSegment: t, TimeStep: timeStep})
		return nil
	}, timeStep)
	return segments, nil
}

func (t TimeSegment) Clip(boundary TimeSegment) TimeSegment {
	var clippedStart, clippedEnd time.Time
	clippedStart = t.Start
	clippedEnd = t.End
	if clippedStart.Before(boundary.Start) || clippedStart.IsZero() {
		clippedStart = boundary.Start
	}
	if clippedEnd.After(boundary.End) || clippedEnd.IsZero() {
		clippedEnd = boundary.End
	}
	return TimeSegment{Start: clippedStart, End: clippedEnd}
}

func (t TimeSegment) Contains(toTest time.Time, includeEdges bool) bool {
	return t.contains(toTest, false, includeEdges)
}

func (t TimeSegment) contains(toTest time.Time, isEndTime bool, includeEdges bool) bool {
	toTestDirectional := newDirectionalTime(toTest, isEndTime)
	start := newDirectionalTime(t.Start, false)
	end := newDirectionalTime(t.End, true)
	if includeEdges {
		return start.lte(toTestDirectional) && toTestDirectional.lte(end)
	}
	return start.lt(toTestDirectional) && toTestDirectional.lt(end)
}

func (t TimeSegment) Overlaps(other TimeSegment, includeEdges bool) bool {
	start := newDirectionalTime(t.Start, false)
	end := newDirectionalTime(t.End, true)
	otherStart := newDirectionalTime(other.Start, false)
	otherEnd := newDirectionalTime(other.End, true)
	if includeEdges {
		return end.gte(otherStart) && start.lte(otherEnd)
	}
	return end.gt(otherStart) && start.lt(otherEnd)
}

func (t TimeSegment) Intersect(other TimeSegment) (TimeSegment, bool) {
	containsStart := t.contains(other.Start, false, false)
	containsEnd := t.contains(other.End, true, false)
	if containsStart && !containsEnd {
		return TimeSegment{Start: other.Start, End: t.End}, true
	}

	if containsStart && containsEnd {
		return other, true
	}

	if !containsStart && containsEnd {
		return TimeSegment{Start: t.Start, End: other.End}, true
	}

	otherContainsMyStart := other.contains(t.Start, false, false)
	otherContainsMyEnd := other.contains(t.End, true, false)
	if otherContainsMyStart && otherContainsMyEnd {
		return t, true
	}
	return TimeSegment{}, false
}

func (t TimeSegment) ContainsTimeSegment(other TimeSegment, includeEdges bool) bool {
	containsStart := t.contains(other.Start, false, includeEdges)
	containsEnd := t.contains(other.End, true, includeEdges)
	return containsStart && containsEnd
}

// AssertIsComposedOf returns an error if this TimeSegment is not "composed of" the given subSegments
//
// A TimeSegment is "composed of" subSegments if:
//   - The subSegments are continuous
//   - The subSegments are non-overlapping
//   - The subSegments completely cover the TimeSegment
func (t TimeSegment) AssertIsComposedOf(subSegments []TimeSegment) error {
	lenSubSegments := len(subSegments)
	if lenSubSegments == 0 {
		return errors.New("no subSegments given")
	}
	slices.SortFunc(subSegments, func(a TimeSegment, b TimeSegment) int { return int(a.Start.Sub(b.Start).Milliseconds()) })
	if err := AssertTimeSegmentsAreContinuous(subSegments); err != nil {
		return err
	}
	first := subSegments[0]
	last := subSegments[lenSubSegments-1]
	if !first.contains(t.Start, false, true) {
		return errors.New(fmt.Sprintf("there is a gap between Start (%s) and the first subSegment (%s)", t.Start.Format(time.RFC3339), first.Format(time.RFC3339)))
	}
	if !last.contains(t.End, true, true) {
		return errors.New(fmt.Sprintf("there is a gap between Start (%s) and the last subSegment (%s)", t.Start.Format(time.RFC3339), last.Format(time.RFC3339)))
	}
	return nil
}

func (t TimeSegment) IsContainedBy(parent TimeSegment, includeEdges bool) bool {
	return parent.ContainsTimeSegment(t, includeEdges)
}

var InvalidTimeSegmentErr = errors.New("invalid start/end times. End cannot be before start")

func NewTimeSegment(start time.Time, end time.Time) (TimeSegment, error) {
	s := newDirectionalTime(start, false)
	e := newDirectionalTime(end, true)
	if e.lt(s) {
		return TimeSegment{}, InvalidTimeSegmentErr
	}
	return TimeSegment{
		Start: start,
		End:   end,
	}, nil
}
