package time_utils

import "time"

func MustParseTimeSegment(format string, s string, e string) TimeSegment {
	start, err := time.Parse(format, s)
	if err != nil {
		panic(err)
	}
	end, err := time.Parse(format, e)
	if err != nil {
		panic(err)
	}
	segment, err := NewTimeSegment(start, end)
	if err != nil {
		panic(err)
	}
	return segment
}

func MustParseTime(format string, s string) time.Time {
	start, err := time.Parse(format, s)
	if err != nil {
		panic(err)
	}
	return start
}
