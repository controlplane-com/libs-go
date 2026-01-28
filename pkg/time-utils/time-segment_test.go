package time_utils

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type RangeTestSuite struct {
	suite.Suite
}

func TestRanges(t *testing.T) {
	suite.Run(t, new(RangeTestSuite))
}

func (s *RangeTestSuite) TestCountSegmentBoundaries() {
	tests := []struct {
		name               string
		startTime          string
		endTime            string
		timeStep           TimeStep
		expectErr          bool
		expectedBoundaries int
	}{
		{
			"zero-days",
			"2023-06-01T00:00:00Z",
			"2023-06-01T23:59:59Z",
			TimeStepDay,
			false,
			1,
		},
		{
			"one-day",
			"2023-06-01T23:59:59Z",
			"2023-06-02T00:00:00Z",
			TimeStepDay,
			false,
			1,
		},
		{
			"two-days",
			"2023-06-01T23:59:59Z",
			"2023-06-03T00:00:00Z",
			TimeStepDay,
			false,
			2,
		},
		{
			"zero-weeks",
			"2023-06-01T00:00:00Z",
			"2023-06-02T00:00:00Z",
			TimeStepWeek,
			false,
			0,
		},
		{
			"one-week",
			"2023-06-04T23:59:59Z",
			"2023-06-05T00:00:00Z",
			TimeStepWeek,
			false,
			1,
		},
		{
			"two-weeks",
			"2023-06-04T23:59:59Z",
			"2023-06-12T00:00:00Z",
			TimeStepWeek,
			false,
			2,
		},
		{
			"invalid-dates",
			"2023-06-02T00:00:00Z",
			"2023-06-01T00:00:00Z",
			TimeStepDay,
			true,
			0,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			startTime, err := time.Parse(time.RFC3339, test.startTime)
			s.Nil(err)
			endTime, err := time.Parse(time.RFC3339, test.endTime)
			s.Nil(err)
			segment, err := NewTimeSegment(startTime, endTime)
			if test.expectErr {
				s.NotNil(err)
				return
			}
			if !test.expectErr && !s.Nil(err) {
				return
			}
			b := segment.CountTimeStepBoundaries(test.timeStep)
			s.Equal(test.expectedBoundaries, b)
		})
	}
}

func (s *RangeTestSuite) TestAlignTimesWithTimeStep() {
	tests := []struct {
		name              string
		startTime         string
		endTime           string
		timeStep          TimeStep
		expectedStartTime string
		expectedEndTime   string
	}{
		{
			name:              "hourly",
			startTime:         "2023-06-01T23:59:59Z",
			endTime:           "2023-06-01T23:59:59Z",
			timeStep:          TimeStepHour,
			expectedStartTime: "2023-06-01T23:00:00Z",
			expectedEndTime:   "2023-06-02T00:00:00Z",
		},
		{
			name:              "hourly_multi",
			startTime:         "2023-06-01T22:59:59Z",
			endTime:           "2023-06-01T23:00:01Z",
			timeStep:          TimeStepHour,
			expectedStartTime: "2023-06-01T22:00:00Z",
			expectedEndTime:   "2023-06-02T00:00:00Z",
		},
		{
			name:              "hourly_edge_end",
			startTime:         "2023-06-01T22:59:59Z",
			endTime:           "2023-06-01T23:00:00Z",
			timeStep:          TimeStepHour,
			expectedStartTime: "2023-06-01T22:00:00Z",
			expectedEndTime:   "2023-06-01T23:00:00Z",
		},
		{
			name:              "hourly_edge_start",
			startTime:         "2023-06-01T22:00:00Z",
			endTime:           "2023-06-01T22:00:01Z",
			timeStep:          TimeStepHour,
			expectedStartTime: "2023-06-01T22:00:00Z",
			expectedEndTime:   "2023-06-01T23:00:00Z",
		},
		{
			name:              "hourly_edge_both",
			startTime:         "2023-06-01T22:00:00Z",
			endTime:           "2023-06-01T23:00:00Z",
			timeStep:          TimeStepHour,
			expectedStartTime: "2023-06-01T22:00:00Z",
			expectedEndTime:   "2023-06-01T23:00:00Z",
		},
		{
			name:              "hourly_edge_both_no_change",
			startTime:         "2023-06-01T22:00:00Z",
			endTime:           "2023-06-01T22:00:00Z",
			timeStep:          TimeStepHour,
			expectedStartTime: "2023-06-01T22:00:00Z",
			expectedEndTime:   "2023-06-01T22:00:00Z",
		},
		{
			name:              "daily",
			startTime:         "2023-06-01T23:59:59Z",
			endTime:           "2023-06-01T23:59:59Z",
			timeStep:          TimeStepDay,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-06-02T00:00:00Z",
		},
		{
			name:              "daily_multi",
			startTime:         "2023-06-01T22:59:59Z",
			endTime:           "2023-06-02T23:00:01Z",
			timeStep:          TimeStepDay,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-06-03T00:00:00Z",
		},
		{
			name:              "daily_edge_end",
			startTime:         "2023-06-01T22:59:59Z",
			endTime:           "2023-06-02T00:00:00Z",
			timeStep:          TimeStepDay,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-06-02T00:00:00Z",
		},
		{
			name:              "daily_edge_start",
			startTime:         "2023-06-01T00:00:00Z",
			endTime:           "2023-06-01T22:00:01Z",
			timeStep:          TimeStepDay,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-06-02T00:00:00Z",
		},
		{
			name:              "daily_edge_both",
			startTime:         "2023-06-01T00:00:00Z",
			endTime:           "2023-06-02T00:00:00Z",
			timeStep:          TimeStepDay,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-06-02T00:00:00Z",
		},
		{
			name:              "daily_edge_both_no_change",
			startTime:         "2023-06-01T00:00:00Z",
			endTime:           "2023-06-01T00:00:00Z",
			timeStep:          TimeStepDay,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-06-01T00:00:00Z",
		},
		{
			name:              "weekly",
			startTime:         "2023-06-01T23:59:59Z",
			endTime:           "2023-06-01T23:59:59Z",
			timeStep:          TimeStepWeek,
			expectedStartTime: "2023-05-29T00:00:00Z",
			expectedEndTime:   "2023-06-05T00:00:00Z",
		},
		{
			name:              "weekly_multi",
			startTime:         "2023-06-01T22:59:59Z",
			endTime:           "2023-06-05T00:00:01Z",
			timeStep:          TimeStepWeek,
			expectedStartTime: "2023-05-29T00:00:00Z",
			expectedEndTime:   "2023-06-12T00:00:00Z",
		},
		{
			name:              "weekly_edge_end",
			startTime:         "2023-06-01T22:59:59Z",
			endTime:           "2023-06-05T00:00:00Z",
			timeStep:          TimeStepWeek,
			expectedStartTime: "2023-05-29T00:00:00Z",
			expectedEndTime:   "2023-06-05T00:00:00Z",
		},
		{
			name:              "weekly_edge_start",
			startTime:         "2023-05-29T00:00:00Z",
			endTime:           "2023-06-01T22:00:01Z",
			timeStep:          TimeStepWeek,
			expectedStartTime: "2023-05-29T00:00:00Z",
			expectedEndTime:   "2023-06-05T00:00:00Z",
		},
		{
			name:              "weekly_edge_both",
			startTime:         "2023-05-29T00:00:00Z",
			endTime:           "2023-06-05T00:00:00Z",
			timeStep:          TimeStepWeek,
			expectedStartTime: "2023-05-29T00:00:00Z",
			expectedEndTime:   "2023-06-05T00:00:00Z",
		},
		{
			name:              "weekly_edge_both_no_change",
			startTime:         "2023-05-29T00:00:00Z",
			endTime:           "2023-05-29T00:00:00Z",
			timeStep:          TimeStepWeek,
			expectedStartTime: "2023-05-29T00:00:00Z",
			expectedEndTime:   "2023-05-29T00:00:00Z",
		},
		{
			name:              "monthly",
			startTime:         "2023-06-01T23:59:59Z",
			endTime:           "2023-06-01T23:59:59Z",
			timeStep:          TimeStepMonth,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-07-01T00:00:00Z",
		},
		{
			name:              "monthly_multi",
			startTime:         "2023-06-01T22:59:59Z",
			endTime:           "2023-07-01T00:00:01Z",
			timeStep:          TimeStepMonth,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-08-01T00:00:00Z",
		},
		{
			name:              "monthly_edge_end",
			startTime:         "2023-06-01T22:59:59Z",
			endTime:           "2023-07-01T00:00:00Z",
			timeStep:          TimeStepMonth,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-07-01T00:00:00Z",
		},
		{
			name:              "monthly_edge_start",
			startTime:         "2023-06-01T00:00:00Z",
			endTime:           "2023-06-01T22:00:01Z",
			timeStep:          TimeStepMonth,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-07-01T00:00:00Z",
		},
		{
			name:              "monthly_edge_both",
			startTime:         "2023-06-01T00:00:00Z",
			endTime:           "2023-07-01T00:00:00Z",
			timeStep:          TimeStepMonth,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-07-01T00:00:00Z",
		},
		{
			name:              "monthly_edge_both_no_change",
			startTime:         "2023-06-01T00:00:00Z",
			endTime:           "2023-06-01T00:00:00Z",
			timeStep:          TimeStepMonth,
			expectedStartTime: "2023-06-01T00:00:00Z",
			expectedEndTime:   "2023-06-01T00:00:00Z",
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			startTime, _ := time.Parse(time.RFC3339, test.startTime)
			endTime, _ := time.Parse(time.RFC3339, test.endTime)
			expectedStartTime, _ := time.Parse(time.RFC3339, test.expectedStartTime)
			expectedEndTime, _ := time.Parse(time.RFC3339, test.expectedEndTime)

			aligned := AlignTimesWithTimeStep(startTime, endTime, test.timeStep)
			s.Equal(expectedStartTime, aligned.Start)
			s.Equal(expectedEndTime, aligned.End)
		})
	}
}
func (s *RangeTestSuite) TestDecomposeTimeSegment() {

	tests := []struct {
		name          string
		segment       TimeSegment
		subSegments   []TimeSegment
		decomposeInto []TimeStep
		greedy        bool
		expectError   bool
		expected      []StepAlignedTimeSegment
	}{
		{
			name:          "empty_sub_segments_month_aligned",
			segment:       MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-08-01T00:00:00Z"),
			decomposeInto: []TimeStep{TimeStepHour, TimeStepMonth},
			expectError:   false,
			greedy:        true,
			expected: []StepAlignedTimeSegment{
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-08-01T00:00:00Z"), TimeStep: TimeStepMonth},
			},
		},
		{
			name:          "empty_sub_segments_month_day_hour",
			segment:       MustParseTimeSegment(time.RFC3339, "2023-05-02T01:00:00Z", "2023-08-02T02:00:00Z"),
			decomposeInto: []TimeStep{TimeStepHour, TimeStepMonth, TimeStepDay},
			expectError:   false,
			greedy:        true,
			expected: []StepAlignedTimeSegment{
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-05-02T01:00:00Z", "2023-05-03T00:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-05-03T00:00:00Z", "2023-06-01T00:00:00Z"), TimeStep: TimeStepDay},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-08-01T00:00:00Z"), TimeStep: TimeStepMonth},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-08-01T00:00:00Z", "2023-08-02T00:00:00Z"), TimeStep: TimeStepDay},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-08-02T00:00:00Z", "2023-08-02T02:00:00Z"), TimeStep: TimeStepHour},
			},
		},
		{
			name:          "empty_sub_segments_month_day",
			segment:       MustParseTimeSegment(time.RFC3339, "2023-05-02T01:00:00Z", "2023-08-02T02:00:00Z"),
			decomposeInto: []TimeStep{TimeStepHour, TimeStepMonth},
			expectError:   false,
			greedy:        true,
			expected: []StepAlignedTimeSegment{
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-05-02T01:00:00Z", "2023-06-01T00:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-08-01T00:00:00Z"), TimeStep: TimeStepMonth},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-08-01T00:00:00Z", "2023-08-02T02:00:00Z"), TimeStep: TimeStepHour},
			},
		},
		{
			name:    "valid_month_hour",
			segment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-08-01T00:00:00Z"),
			subSegments: []TimeSegment{
				MustParseTimeSegment(time.RFC3339, "2023-01-01T00:00:00Z", "2023-06-15T00:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-15T00:00:00Z", "2024-01-01T00:00:00Z"),
			},
			decomposeInto: []TimeStep{TimeStepHour, TimeStepMonth},
			expectError:   false,
			greedy:        true,
			expected: []StepAlignedTimeSegment{
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-06-01T00:00:00Z"), TimeStep: TimeStepMonth},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-15T00:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-15T00:00:00Z", "2023-07-01T00:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-07-01T00:00:00Z", "2023-08-01T00:00:00Z"), TimeStep: TimeStepMonth},
			},
		},
		{
			name:    "valid_month_day_hour",
			segment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-08-01T00:00:00Z"),
			subSegments: []TimeSegment{
				MustParseTimeSegment(time.RFC3339, "2023-01-01T00:00:00Z", "2023-06-15T00:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-15T00:00:00Z", "2024-01-01T00:00:00Z"),
			},
			decomposeInto: []TimeStep{TimeStepHour, TimeStepDay, TimeStepMonth},
			expectError:   false,
			greedy:        true,
			expected: []StepAlignedTimeSegment{
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-06-01T00:00:00Z"), TimeStep: TimeStepMonth},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-15T00:00:00Z"), TimeStep: TimeStepDay},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-15T00:00:00Z", "2023-07-01T00:00:00Z"), TimeStep: TimeStepDay},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-07-01T00:00:00Z", "2023-08-01T00:00:00Z"), TimeStep: TimeStepMonth},
			},
		},
		{
			name:    "valid_month_day_hour_with_discontinuous_sub_segments",
			segment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-08-01T00:00:00Z"),
			subSegments: []TimeSegment{
				MustParseTimeSegment(time.RFC3339, "2023-01-01T00:00:00Z", "2023-06-15T00:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-15T01:00:00Z", "2023-06-17T13:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-30T01:00:00Z", "2024-01-01T00:00:00Z"),
			},
			decomposeInto: []TimeStep{TimeStepHour, TimeStepDay, TimeStepMonth},
			expectError:   false,
			greedy:        true,
			expected: []StepAlignedTimeSegment{
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-06-01T00:00:00Z"), TimeStep: TimeStepMonth},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-15T00:00:00Z"), TimeStep: TimeStepDay},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-15T00:00:00Z", "2023-06-15T01:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-15T01:00:00Z", "2023-06-16T00:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-16T00:00:00Z", "2023-06-17T00:00:00Z"), TimeStep: TimeStepDay},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-17T00:00:00Z", "2023-06-17T13:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-17T13:00:00Z", "2023-06-18T00:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-18T00:00:00Z", "2023-06-30T00:00:00Z"), TimeStep: TimeStepDay},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-30T00:00:00Z", "2023-06-30T01:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-30T01:00:00Z", "2023-07-01T00:00:00Z"), TimeStep: TimeStepHour},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-07-01T00:00:00Z", "2023-08-01T00:00:00Z"), TimeStep: TimeStepMonth},
			},
		},
		{
			name:    "day_aligned_sub_segments",
			segment: MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-07-01T00:00:00Z"),
			subSegments: []TimeSegment{
				MustParseTimeSegment(time.RFC3339, "2023-05-31T19:00:00-05:00", "2023-06-15T19:00:00-05:00"),
				MustParseTimeSegment(time.RFC3339, "2023-06-15T19:00:00-05:00", "2023-07-01T00:00:00Z"),
			},
			decomposeInto: []TimeStep{TimeStepHour, TimeStepDay},
			expectError:   false,
			greedy:        true,
			expected: []StepAlignedTimeSegment{
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-16T00:00:00Z"), TimeStep: TimeStepDay},
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-06-16T00:00:00Z", "2023-07-01T00:00:00Z"), TimeStep: TimeStepDay},
			},
		},
		{
			name:          "partial_month_decomposed_into_hours",
			segment:       MustParseTimeSegment(time.RFC3339, "2023-07-01T00:00:00Z", "2023-07-15T00:00:00Z"),
			subSegments:   []TimeSegment{},
			decomposeInto: []TimeStep{TimeStepHour, TimeStepDay, TimeStepMonth},
			expectError:   false,
			greedy:        true,
			expected: []StepAlignedTimeSegment{
				{TimeSegment: MustParseTimeSegment(time.RFC3339, "2023-07-01T00:00:00Z", "2023-07-15T00:00:00Z"), TimeStep: TimeStepDay},
			},
		},
		{
			name:    "overlapping_sub_segments",
			segment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-08-01T00:00:00Z"),
			subSegments: []TimeSegment{
				MustParseTimeSegment(time.RFC3339, "2023-01-01T00:00:00Z", "2023-06-15T02:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-15T01:00:00Z", "2023-06-17T13:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-17T00:00:00Z", "2024-01-01T00:00:00Z"),
			},
			decomposeInto: []TimeStep{TimeStepHour, TimeStepDay, TimeStepMonth},
			expectError:   true,
			greedy:        true,
			expected:      nil,
		},
		{
			name:    "overlapping_incorrectly_ordered_sub_segments",
			segment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-08-01T00:00:00Z"),
			subSegments: []TimeSegment{
				MustParseTimeSegment(time.RFC3339, "2023-01-01T00:00:00Z", "2023-06-15T02:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-17T00:00:00Z", "2024-01-01T00:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-15T00:00:00Z", "2023-06-15T03:00:00Z"),
			},
			decomposeInto: []TimeStep{TimeStepHour, TimeStepDay, TimeStepMonth},
			expectError:   true,
			greedy:        true,
			expected:      nil,
		},
		{
			name:    "overlapping_nested_sub_segments",
			segment: MustParseTimeSegment(time.RFC3339, "2023-05-01T00:00:00Z", "2023-08-01T00:00:00Z"),
			subSegments: []TimeSegment{
				MustParseTimeSegment(time.RFC3339, "2023-01-01T00:00:00Z", "2023-06-15T02:00:00Z"),
				MustParseTimeSegment(time.RFC3339, "2023-06-13T01:00:00Z", "2023-06-13T02:00:00Z"),
			},
			decomposeInto: []TimeStep{TimeStepHour, TimeStepDay, TimeStepMonth},
			expectError:   true,
			greedy:        true,
			expected:      nil,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			boundaries, err := ExtractBoundariesFromTimeSegments(test.subSegments)
			if test.expectError {
				s.NotNil(err)
				return
			} else {
				s.Nil(err)
			}

			segments, err := test.segment.Decompose(boundaries, test.decomposeInto, test.greedy)
			if test.expectError {
				s.NotNil(err)
			} else {
				s.Nil(err)
			}
			s.Equal(test.expected, segments)
		})
	}
}

func (s *RangeTestSuite) TestContainsTimeSegment() {
	runs := []struct {
		name           string
		t              TimeSegment
		other          TimeSegment
		inclusive      bool
		expectContains bool
	}{
		{
			name:           "fully-contained-non-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T01:00:00Z", "2023-06-01T02:00:00Z"),
			inclusive:      false,
			expectContains: true,
		},
		{
			name:           "fully-contained-start-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-01T02:00:00Z"),
			inclusive:      true,
			expectContains: true,
		},
		{
			name:           "non-contained-start-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-01T02:00:00Z"),
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "fully-contained-end-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T01:00:00Z", "2023-06-02T00:00:00Z"),
			inclusive:      true,
			expectContains: true,
		},
		{
			name:           "non-contained-end-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T01:00:00Z", "2023-06-02T00:00:00Z"),
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "fully-before",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-03T00:00:00Z", "2023-05-04T00:00:00Z"),
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "fully-after",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-03T00:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "overlap-end-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-02T00:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      true,
			expectContains: false,
		},
		{
			name:           "overlap-end-edge-non-inclusive",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-02T00:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "overlap-end",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T12:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      true,
			expectContains: false,
		},
		{
			name:           "overlap-end-non-inclusive",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T12:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "overlap-start-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-30T00:00:00Z", "2023-06-01T00:00:00Z"),
			inclusive:      true,
			expectContains: false,
		},
		{
			name:           "overlap-start-edge-non-inclusive",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-30T00:00:00Z", "2023-06-01T00:00:00Z"),
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "overlap-start",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-30T00:00:00Z", "2023-06-03T00:00:00Z"),
			inclusive:      true,
			expectContains: false,
		},
		{
			name:           "overlap-start-non-inclusive",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-30T00:00:00Z", "2023-06-03T00:00:00Z"),
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "fully-contained-nil-end",
			t:              TimeSegment{Start: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			other:          TimeSegment{Start: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			inclusive:      true,
			expectContains: true,
		},
		{
			name:           "fully-contained-nil-start",
			t:              TimeSegment{End: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			other:          TimeSegment{End: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			inclusive:      true,
			expectContains: true,
		},
		{
			name:           "non-contained-nil-start-exclusive",
			t:              TimeSegment{End: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			other:          TimeSegment{End: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "non-contained-nil-end-exclusive",
			t:              TimeSegment{Start: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			other:          TimeSegment{Start: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			inclusive:      false,
			expectContains: false,
		},
		{
			name:           "fully-contained-nil-start-and-end",
			t:              TimeSegment{},
			other:          TimeSegment{},
			inclusive:      true,
			expectContains: true,
		},
		{
			name:           "non-contained-opposing-nil-bounds",
			t:              TimeSegment{End: MustParseTime(time.RFC3339, "2023-10-01T00:00:00Z")},
			other:          TimeSegment{Start: MustParseTime(time.RFC3339, "2023-01-01T00:00:00Z")},
			inclusive:      true,
			expectContains: false,
		},
	}
	for _, r := range runs {
		s.Run(r.name, func() {
			actual := r.t.ContainsTimeSegment(r.other, r.inclusive)
			s.Equal(r.expectContains, actual)
		})
	}
}

func (s *RangeTestSuite) TestOverlapsTimeSegment() {
	runs := []struct {
		name           string
		t              TimeSegment
		other          TimeSegment
		inclusive      bool
		expectOverlaps bool
	}{
		{
			name:           "fully-contained-by-non-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T01:00:00Z", "2023-06-01T02:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "fully-contained-by-start-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T01:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "same-segment",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "fully-contained-by-end-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-01T02:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "fully-contained-non-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T01:00:00Z", "2023-06-01T02:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "fully-contained-start-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-01T02:00:00Z"),
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "overlapping-start-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-01T02:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "fully-contained-end-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T01:00:00Z", "2023-06-02T00:00:00Z"),
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "overlapping-end-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T01:00:00Z", "2023-06-02T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "fully-before",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-03T00:00:00Z", "2023-05-04T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: false,
		},
		{
			name:           "fully-after",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-03T00:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: false,
		},
		{
			name:           "overlap-end-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-02T00:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "overlap-end-edge-non-inclusive",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-02T00:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: false,
		},
		{
			name:           "overlap-end",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T12:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "overlap-end-non-inclusive",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-06-01T12:00:00Z", "2023-06-04T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "overlap-start-edge",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-30T00:00:00Z", "2023-06-01T00:00:00Z"),
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "overlap-start-edge-non-inclusive",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-30T00:00:00Z", "2023-06-01T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: false,
		},
		{
			name:           "overlap-start",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-30T00:00:00Z", "2023-06-03T00:00:00Z"),
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "overlap-start-non-inclusive",
			t:              MustParseTimeSegment(time.RFC3339, "2023-06-01T00:00:00Z", "2023-06-02T00:00:00Z"),
			other:          MustParseTimeSegment(time.RFC3339, "2023-05-30T00:00:00Z", "2023-06-03T00:00:00Z"),
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "fully-contained-nil-end",
			t:              TimeSegment{Start: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			other:          TimeSegment{Start: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "fully-contained-nil-start",
			t:              TimeSegment{End: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			other:          TimeSegment{End: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "overlapping-nil-start-exclusive",
			t:              TimeSegment{End: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			other:          TimeSegment{End: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "overlapping-nil-end-exclusive",
			t:              TimeSegment{Start: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			other:          TimeSegment{Start: MustParseTime(time.RFC3339, "2023-06-01T00:00:01Z")},
			inclusive:      false,
			expectOverlaps: true,
		},
		{
			name:           "fully-contained-nil-start-and-end",
			t:              TimeSegment{},
			other:          TimeSegment{},
			inclusive:      true,
			expectOverlaps: true,
		},
		{
			name:           "overlapping-opposing-nil-bounds",
			t:              TimeSegment{End: MustParseTime(time.RFC3339, "2023-10-01T00:00:00Z")},
			other:          TimeSegment{Start: MustParseTime(time.RFC3339, "2023-01-01T00:00:00Z")},
			inclusive:      true,
			expectOverlaps: true,
		},
	}
	for _, r := range runs {
		s.Run(r.name, func() {
			actual := r.t.Overlaps(r.other, r.inclusive)
			s.Equal(r.expectOverlaps, actual)
		})
	}
}
