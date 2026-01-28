package time_utils

import (
	"fmt"
	"testing"
	"time"
)

func TestFirstDayOfTheMonth(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{input: "2023-01-30T00:00:00Z", expectedOutput: "2023-01-01T00:00:00Z"},
		{input: "2023-01-01T00:00:00Z", expectedOutput: "2023-01-01T00:00:00Z"},
		{input: "2023-02-27T00:00:00Z", expectedOutput: "2023-02-01T00:00:00Z"},
		{input: "2023-03-30T00:00:00Z", expectedOutput: "2023-03-01T00:00:00Z"},
		{input: "2023-04-30T00:00:00Z", expectedOutput: "2023-04-01T00:00:00Z"},
		{input: "2023-05-30T00:00:00Z", expectedOutput: "2023-05-01T00:00:00Z"},
		{input: "2023-06-30T00:00:00Z", expectedOutput: "2023-06-01T00:00:00Z"},
		{input: "2023-07-30T00:00:00Z", expectedOutput: "2023-07-01T00:00:00Z"},
		{input: "2023-08-30T00:00:00Z", expectedOutput: "2023-08-01T00:00:00Z"},
		{input: "2023-09-30T00:00:00Z", expectedOutput: "2023-09-01T00:00:00Z"},
		{input: "2023-10-30T00:00:00Z", expectedOutput: "2023-10-01T00:00:00Z"},
		{input: "2023-11-30T00:00:00Z", expectedOutput: "2023-11-01T00:00:00Z"},
		{input: "2023-12-30T00:00:00Z", expectedOutput: "2023-12-01T00:00:00Z"},
	}
	for _, test := range tests {
		input, err := time.Parse(time.RFC3339, test.input)
		if err != nil {
			panic(err)
		}
		expectedOutput, err := time.Parse(time.RFC3339, test.expectedOutput)
		if err != nil {
			panic(err)
		}
		t.Run(test.input, func(t *testing.T) {
			if actualOutput := FirstDayOfTheMonth(input); actualOutput != expectedOutput {
				fmt.Printf("Expected %s but received %s", test.expectedOutput, actualOutput)
				t.Fail()
			}
		})
	}
}

func TestAddMonths(t *testing.T) {
	tests := []struct {
		input          string
		delta          int
		expectedOutput string
	}{
		{input: "2023-05-30T00:00:00Z", delta: 1, expectedOutput: "2023-06-30T00:00:00Z"},
	}
	for _, test := range tests {
		input, err := time.Parse(time.RFC3339, test.input)
		if err != nil {
			panic(err)
		}
		expectedOutput, err := time.Parse(time.RFC3339, test.expectedOutput)
		if err != nil {
			panic(err)
		}
		t.Run(test.input, func(t *testing.T) {
			if actualOutput := AddMonths(input, test.delta); actualOutput != expectedOutput {
				fmt.Printf("Expected %s but received %s", test.expectedOutput, actualOutput)
				t.FailNow()
			}
		})
	}
}

func TestLastDayOfTheMonth(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{input: "2023-01-01T00:00:00Z", expectedOutput: "2023-01-31T00:00:00Z"},
		{input: "2023-01-31T00:00:00Z", expectedOutput: "2023-01-31T00:00:00Z"},
		{input: "2023-02-01T00:00:00Z", expectedOutput: "2023-02-28T00:00:00Z"},
		{input: "2023-03-01T00:00:00Z", expectedOutput: "2023-03-31T00:00:00Z"},
		{input: "2023-04-01T00:00:00Z", expectedOutput: "2023-04-30T00:00:00Z"},
		{input: "2023-05-01T00:00:00Z", expectedOutput: "2023-05-31T00:00:00Z"},
		{input: "2023-06-01T00:00:00Z", expectedOutput: "2023-06-30T00:00:00Z"},
		{input: "2023-07-01T00:00:00Z", expectedOutput: "2023-07-31T00:00:00Z"},
		{input: "2023-08-01T00:00:00Z", expectedOutput: "2023-08-31T00:00:00Z"},
		{input: "2023-09-01T00:00:00Z", expectedOutput: "2023-09-30T00:00:00Z"},
		{input: "2023-10-01T00:00:00Z", expectedOutput: "2023-10-31T00:00:00Z"},
		{input: "2023-11-01T00:00:00Z", expectedOutput: "2023-11-30T00:00:00Z"},
		{input: "2023-12-01T00:00:00Z", expectedOutput: "2023-12-31T00:00:00Z"},
	}
	for _, test := range tests {
		input, err := time.Parse(time.RFC3339, test.input)
		if err != nil {
			panic(err)
		}
		expectedOutput, err := time.Parse(time.RFC3339, test.expectedOutput)
		if err != nil {
			panic(err)
		}
		t.Run(test.input, func(t *testing.T) {
			if actualOutput := LastDayOfTheMonth(input); actualOutput != expectedOutput {
				fmt.Printf("Expected %s but received %s", test.expectedOutput, actualOutput)
				t.FailNow()
			}
		})
	}
}

func TestWeekday(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
		weekday        time.Weekday
	}{
		{input: "2023-07-02T00:00:00Z", weekday: time.Monday, expectedOutput: "2023-06-26T00:00:00Z"},
		{input: "2023-07-02T00:00:00Z", weekday: time.Tuesday, expectedOutput: "2023-06-27T00:00:00Z"},
		{input: "2023-07-02T00:00:00Z", weekday: time.Wednesday, expectedOutput: "2023-06-28T00:00:00Z"},
		{input: "2023-07-02T00:00:00Z", weekday: time.Thursday, expectedOutput: "2023-06-29T00:00:00Z"},
		{input: "2023-07-02T00:00:00Z", weekday: time.Friday, expectedOutput: "2023-06-30T00:00:00Z"},
		{input: "2023-07-02T00:00:00Z", weekday: time.Saturday, expectedOutput: "2023-07-01T00:00:00Z"},
		{input: "2023-07-02T00:00:00Z", weekday: time.Sunday, expectedOutput: "2023-07-02T00:00:00Z"},

		{input: "2023-07-03T00:00:00Z", weekday: time.Monday, expectedOutput: "2023-07-03T00:00:00Z"},
		{input: "2023-07-03T00:00:00Z", weekday: time.Tuesday, expectedOutput: "2023-07-04T00:00:00Z"},
		{input: "2023-07-03T00:00:00Z", weekday: time.Wednesday, expectedOutput: "2023-07-05T00:00:00Z"},
		{input: "2023-07-03T00:00:00Z", weekday: time.Thursday, expectedOutput: "2023-07-06T00:00:00Z"},
		{input: "2023-07-03T00:00:00Z", weekday: time.Friday, expectedOutput: "2023-07-07T00:00:00Z"},
		{input: "2023-07-03T00:00:00Z", weekday: time.Saturday, expectedOutput: "2023-07-08T00:00:00Z"},
		{input: "2023-07-03T00:00:00Z", weekday: time.Sunday, expectedOutput: "2023-07-09T00:00:00Z"},

		{input: "2023-07-04T00:00:00Z", weekday: time.Monday, expectedOutput: "2023-07-03T00:00:00Z"},
		{input: "2023-07-04T00:00:00Z", weekday: time.Tuesday, expectedOutput: "2023-07-04T00:00:00Z"},
		{input: "2023-07-04T00:00:00Z", weekday: time.Wednesday, expectedOutput: "2023-07-05T00:00:00Z"},
		{input: "2023-07-04T00:00:00Z", weekday: time.Thursday, expectedOutput: "2023-07-06T00:00:00Z"},
		{input: "2023-07-04T00:00:00Z", weekday: time.Friday, expectedOutput: "2023-07-07T00:00:00Z"},
		{input: "2023-07-04T00:00:00Z", weekday: time.Saturday, expectedOutput: "2023-07-08T00:00:00Z"},
		{input: "2023-07-04T00:00:00Z", weekday: time.Sunday, expectedOutput: "2023-07-09T00:00:00Z"},

		{input: "2023-07-05T00:00:00Z", weekday: time.Monday, expectedOutput: "2023-07-03T00:00:00Z"},
		{input: "2023-07-05T00:00:00Z", weekday: time.Tuesday, expectedOutput: "2023-07-04T00:00:00Z"},
		{input: "2023-07-05T00:00:00Z", weekday: time.Wednesday, expectedOutput: "2023-07-05T00:00:00Z"},
		{input: "2023-07-05T00:00:00Z", weekday: time.Thursday, expectedOutput: "2023-07-06T00:00:00Z"},
		{input: "2023-07-05T00:00:00Z", weekday: time.Friday, expectedOutput: "2023-07-07T00:00:00Z"},
		{input: "2023-07-05T00:00:00Z", weekday: time.Saturday, expectedOutput: "2023-07-08T00:00:00Z"},
		{input: "2023-07-05T00:00:00Z", weekday: time.Sunday, expectedOutput: "2023-07-09T00:00:00Z"},

		{input: "2023-07-06T00:00:00Z", weekday: time.Monday, expectedOutput: "2023-07-03T00:00:00Z"},
		{input: "2023-07-06T00:00:00Z", weekday: time.Tuesday, expectedOutput: "2023-07-04T00:00:00Z"},
		{input: "2023-07-06T00:00:00Z", weekday: time.Wednesday, expectedOutput: "2023-07-05T00:00:00Z"},
		{input: "2023-07-06T00:00:00Z", weekday: time.Thursday, expectedOutput: "2023-07-06T00:00:00Z"},
		{input: "2023-07-06T00:00:00Z", weekday: time.Friday, expectedOutput: "2023-07-07T00:00:00Z"},
		{input: "2023-07-06T00:00:00Z", weekday: time.Saturday, expectedOutput: "2023-07-08T00:00:00Z"},
		{input: "2023-07-06T00:00:00Z", weekday: time.Sunday, expectedOutput: "2023-07-09T00:00:00Z"},

		{input: "2023-07-07T00:00:00Z", weekday: time.Monday, expectedOutput: "2023-07-03T00:00:00Z"},
		{input: "2023-07-07T00:00:00Z", weekday: time.Tuesday, expectedOutput: "2023-07-04T00:00:00Z"},
		{input: "2023-07-07T00:00:00Z", weekday: time.Wednesday, expectedOutput: "2023-07-05T00:00:00Z"},
		{input: "2023-07-07T00:00:00Z", weekday: time.Thursday, expectedOutput: "2023-07-06T00:00:00Z"},
		{input: "2023-07-07T00:00:00Z", weekday: time.Friday, expectedOutput: "2023-07-07T00:00:00Z"},
		{input: "2023-07-07T00:00:00Z", weekday: time.Saturday, expectedOutput: "2023-07-08T00:00:00Z"},
		{input: "2023-07-07T00:00:00Z", weekday: time.Sunday, expectedOutput: "2023-07-09T00:00:00Z"},

		{input: "2023-07-08T00:00:00Z", weekday: time.Monday, expectedOutput: "2023-07-03T00:00:00Z"},
		{input: "2023-07-08T00:00:00Z", weekday: time.Tuesday, expectedOutput: "2023-07-04T00:00:00Z"},
		{input: "2023-07-08T00:00:00Z", weekday: time.Wednesday, expectedOutput: "2023-07-05T00:00:00Z"},
		{input: "2023-07-08T00:00:00Z", weekday: time.Thursday, expectedOutput: "2023-07-06T00:00:00Z"},
		{input: "2023-07-08T00:00:00Z", weekday: time.Friday, expectedOutput: "2023-07-07T00:00:00Z"},
		{input: "2023-07-08T00:00:00Z", weekday: time.Saturday, expectedOutput: "2023-07-08T00:00:00Z"},
		{input: "2023-07-08T00:00:00Z", weekday: time.Sunday, expectedOutput: "2023-07-09T00:00:00Z"},

		{input: "2023-07-09T00:00:00Z", weekday: time.Monday, expectedOutput: "2023-07-03T00:00:00Z"},
		{input: "2023-07-09T00:00:00Z", weekday: time.Tuesday, expectedOutput: "2023-07-04T00:00:00Z"},
		{input: "2023-07-09T00:00:00Z", weekday: time.Wednesday, expectedOutput: "2023-07-05T00:00:00Z"},
		{input: "2023-07-09T00:00:00Z", weekday: time.Thursday, expectedOutput: "2023-07-06T00:00:00Z"},
		{input: "2023-07-09T00:00:00Z", weekday: time.Friday, expectedOutput: "2023-07-07T00:00:00Z"},
		{input: "2023-07-09T00:00:00Z", weekday: time.Saturday, expectedOutput: "2023-07-08T00:00:00Z"},
		{input: "2023-07-09T00:00:00Z", weekday: time.Sunday, expectedOutput: "2023-07-09T00:00:00Z"},
	}
	for _, test := range tests {
		input, err := time.Parse(time.RFC3339, test.input)
		if err != nil {
			panic(err)
		}
		expectedOutput, err := time.Parse(time.RFC3339, test.expectedOutput)
		if err != nil {
			panic(err)
		}
		t.Run(test.input, func(t *testing.T) {
			if actualOutput := Weekday(input, test.weekday); actualOutput != expectedOutput {
				fmt.Printf("Expected %s but received %s", test.expectedOutput, actualOutput)
				t.FailNow()
			}
		})
	}
}
