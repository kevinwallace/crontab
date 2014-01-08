package crontab

import (
	"math/rand"
	"testing"
	"time"
)

func TestNext(t *testing.T) {
	p := func(s string) time.Time {
		result, err := time.Parse("2006-01-02 15:04", s)
		if err != nil {
			panic(err)
		}
		return result
	}

	test := func(line string, start, next time.Time) bool {
		entry, err := ParseEntry(line)
		if err != nil {
			t.Fatalf("Error when parsing line %v: %s", line, err)
		}
		actualNext := entry.Schedule.Next(start)
		if next != actualNext {
			t.Errorf("ParseEntry(%q).Schedule.Next(%v) was %v, expected %v", line, start, actualNext, next)
			return false
		}
		return true
	}

	// Test that for times in the range [start, end), the next scheduled time is end.
	// Test start and end-episilon, in particular.
	testRange := func(line string, start, end time.Time) {
		if !test(line, start, end) || !test(line, end.Add(time.Duration(-1)), end) {
			return
		}

		duration := end.Sub(start)
		for i := 0; i < 100; i++ {
			t := start.Add(time.Duration(rand.Int63n(int64(duration))))
			if !test(line, t, end) {
				return
			}
		}
	}

	// schedule that never fires
	test("0 0 31 2 *", p("2000-01-01 00:00"), time.Time{})

	// @predefined schedules
	testRange("@yearly", p("2000-01-01 00:00"), p("2001-01-01 00:00"))
	testRange("@monthly", p("2000-01-01 00:00"), p("2000-02-01 00:00"))
	testRange("@weekly", p("2000-01-02 00:00"), p("2000-01-09 00:00"))
	testRange("@daily", p("2000-01-01 00:00"), p("2000-01-02 00:00"))
	testRange("@hourly", p("2000-01-01 00:00"), p("2000-01-01 01:00"))

	// wildcard-with-step schedules
	testRange("*/5 * * * *", p("2000-01-01 00:00"), p("2000-01-01 00:05"))
	testRange("*/5 * * * *", p("2000-01-01 00:55"), p("2000-01-01 01:00"))

	// range schedules
	testRange("0-10 * * * * a", p("2000-01-01 00:00"), p("2000-01-01 00:01"))
	testRange("0-10 * * * * b", p("2000-01-01 00:09"), p("2000-01-01 00:10"))
	testRange("0-10 * * * * c", p("2000-01-01 00:10"), p("2000-01-01 01:00"))

	// range-with-step schedules
	testRange("0-10/5 * * * *", p("2000-01-01 00:00"), p("2000-01-01 00:05"))
	testRange("0-10/5 * * * *", p("2000-01-01 00:05"), p("2000-01-01 00:10"))
	testRange("0-10/5 * * * *", p("2000-01-01 00:10"), p("2000-01-01 01:00"))

	// lists
	testRange("0,5,25 * * * *", p("2000-01-01 00:00"), p("2000-01-01 00:05"))
	testRange("0,5,25 * * * *", p("2000-01-01 00:05"), p("2000-01-01 00:25"))
	testRange("0,5,25 * * * *", p("2000-01-01 00:25"), p("2000-01-01 01:00"))

	// both day and weekday restricted
	testRange("0 0 13 * 5", p("2000-01-13 00:00"), p("2000-01-14 00:00"))
	testRange("0 0 13 * 5", p("2000-01-14 00:00"), p("2000-01-21 00:00"))
	testRange("0 0 13 * 5", p("2000-01-21 00:00"), p("2000-01-28 00:00"))
	testRange("0 0 13 * 5", p("2000-01-28 00:00"), p("2000-02-04 00:00"))
	testRange("0 0 13 * 5", p("2000-02-04 00:00"), p("2000-02-11 00:00"))
	testRange("0 0 13 * 5", p("2000-02-11 00:00"), p("2000-02-13 00:00"))
}
