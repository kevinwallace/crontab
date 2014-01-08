package crontab

import (
	"time"
)

type field struct {
	min, max int
	name     string
}

var (
	minuteField  = field{0, 59, "minute"}
	hourField    = field{0, 23, "hour"}
	dayField     = field{1, 31, "day"}
	monthField   = field{1, 12, "month"}
	weekdayField = field{0, 6, "weekday"}
)

type valueSpec interface {
	// wildcard determines whether the range matches all values for the given field.
	wildcard(field) bool
	// matches determines whether a given value of a given field falls within this range.
	matches(int) bool
}

type rangeSpec struct {
	// First and last values for this range.
	start, end int
	// Distance between values.
	step int
}

func (r rangeSpec) wildcard(f field) bool {
	return r.step == 1 && r.start == f.min && r.end == f.max
}

func (r rangeSpec) matches(i int) bool {
	return r.start <= i && i <= r.end && (i-r.start)%r.step == 0
}

// valid determines whether the range is valid for the specified field.
func (r rangeSpec) valid(f field) bool {
	return f.min <= r.start && r.start <= r.end && r.end <= f.max
}

type listSpec []rangeSpec

func (l listSpec) wildcard(f field) bool {
	return len(l) == 1 && l[0].wildcard(f)
}

func (l listSpec) matches(i int) bool {
	for _, r := range l {
		if r.matches(i) {
			return true
		}
	}
	return false
}

// Schedule is a set of constraints on the minute/hour/day/month/weekday of a date.
type Schedule struct {
	minute, hour, day, month, weekday listSpec
}

// dayMatches determines wheter the day and weekday fields match the given date.
// If either is unrestricted, both fields must match for the date to match.
// Otherwise, if either matches, the date matches.
func (s Schedule) dayMatches(t time.Time) bool {
	dayWildcard := s.day.wildcard(dayField)
	weekdayWildcard := s.weekday.wildcard(weekdayField)

	dayMatches := s.day.matches(t.Day())
	weekdayMatches := s.weekday.matches(int(t.Weekday()))
	if dayWildcard || weekdayWildcard {
		return dayMatches && weekdayMatches
	}
	return dayMatches || weekdayMatches
}

// Next calculates the next time at which this schedule is active.
// If no such time exists, the zero time is returned.
func (s Schedule) Next(t time.Time) time.Time {
	// Time after which further searching is pointless if we haven't found a match yet.
	// 8 years in the future accounts for the longest possible gap between two leap days.
	horizon := t.AddDate(8, 0, 0)

	// Truncate to the current minute, which is the smallest resolution we support,
	// then increment the minute to get the first candidate time.
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	t = t.Add(1 * time.Minute)

wrap:
	for t.Before(horizon) {
		// For each field, truncate then increment until we find a value that matches,
		// then move on to the next field.
		// If the field we're incrementing wraps, start this process over again from the first field.
		// TODO: We can calculate the next matching value, and advance directly to it.

		for !s.month.matches(int(t.Month())) {
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
			t = t.AddDate(0, 1, 0)
		}

		for !s.dayMatches(t) {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			t = t.AddDate(0, 0, 1)
			if t.Day() == 1 {
				continue wrap
			}
		}

		for !s.hour.matches(t.Hour()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
			t = t.Add(1 * time.Hour)
			if t.Hour() == 0 {
				continue wrap
			}
		}

		for !s.minute.matches(t.Minute()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
			t = t.Add(1 * time.Minute)
			if t.Minute() == 0 {
				continue wrap
			}
		}

		return t
	}

	// No next time was found.
	return time.Time{}
}

// Entry is a single line in a crontab.
type Entry struct {
	Schedule Schedule
	Command  string
}
