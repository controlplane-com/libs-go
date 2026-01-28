package time_utils

import (
	"time"
)

func FirstDayOfTheMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func LastDayOfTheMonth(t time.Time) time.Time {
	firstOfNextMonth := AddMonths(FirstDayOfTheMonth(t), 1)
	return firstOfNextMonth.Add(-time.Hour * 24)
}

func AddMonths(t time.Time, delta int) time.Time {
	return time.Date(t.Year(), t.Month()+time.Month(delta), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

func getWeekdayStartingOnMonday(weekday time.Weekday) time.Weekday {
	switch weekday {
	case 0:
		return 6
	case 1:
		return 0
	case 2:
		return 1
	case 3:
		return 2
	case 4:
		return 3
	case 5:
		return 4
	case 6:
		return 5
	default:
		panic("invalid week day")
	}
}

func Weekday(t time.Time, weekday time.Weekday) time.Time {
	var dayDelta time.Weekday
	currentWeekday := getWeekdayStartingOnMonday(t.Weekday())
	targetWeekday := getWeekdayStartingOnMonday(weekday)
	dayDelta = targetWeekday - currentWeekday
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, int(dayDelta))
}

func StartOfTheDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func StartOfTheHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}
