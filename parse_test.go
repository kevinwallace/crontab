package crontab

import (
	"reflect"
	"testing"
)

func TestParseEntry(t *testing.T) {
	test := func(line string, expected Entry) {
		actual, err := ParseEntry(line)
		if err != nil {
			t.Errorf("Error parsing %v: %s", line, err)
			return
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("ParseEntry(%v) was %#v, expected %#v", line, actual, expected)
		}
	}
	testBad := func(line string) {
		actual, err := ParseEntry(line)
		if err == nil {
			t.Errorf("Expected error when parsing %v, but got %v", line, actual)
		}
	}

	test("0 1 2 3 4 /bin/echo foo", Entry{
		Schedule{
			[]rangeSpec{{0, 0, 1}},
			[]rangeSpec{{1, 1, 1}},
			[]rangeSpec{{2, 2, 1}},
			[]rangeSpec{{3, 3, 1}},
			[]rangeSpec{{4, 4, 1}},
		},
		"/bin/echo foo"})
	test("*/5 ? 2-10/2 jan-5 7-wed/2,thu", Entry{
		Schedule{
			[]rangeSpec{{0, 59, 5}},
			[]rangeSpec{{0, 23, 1}},
			[]rangeSpec{{2, 10, 2}},
			[]rangeSpec{{1, 5, 1}},
			[]rangeSpec{{0, 3, 2}, {4, 4, 1}},
		},
		""})
	test("@daily lol  ", Entry{
		Schedule{
			[]rangeSpec{{0, 0, 1}},
			[]rangeSpec{{0, 0, 1}},
			[]rangeSpec{{1, 31, 1}},
			[]rangeSpec{{1, 12, 1}},
			[]rangeSpec{{0, 6, 1}},
		},
		"lol  "})

	testBad("lol")
	testBad("@daily,")
	testBad("0 1 2 3 4/5/6")
	testBad("0 1 2 3 4/5-6")
	testBad("0 1 2 3 4-5-6")
	testBad("0 1 2 3 4/?")
}

func TestParseCrontab(t *testing.T) {
	test := func(s string, expected ...Entry) {
		actual, err := ParseCrontab(s)
		if err != nil {
			t.Errorf("Error parsing crontab:\n%s\n%s", s, err)
			return
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Crontab:\n%s\nExpected: %v\nActual: %v", s, expected, actual)
		}
	}

	testBad := func(s string) {
		actual, err := ParseCrontab(s)
		if err == nil {
			t.Errorf("Expected error parsing crontab:\n%s\nActual: %v", s, actual)
		}
	}

	test("# this is a comment")
	test(
		"0 1 2 3 4 # this is not a comment",
		MustParseEntry("0 1 2 3 4 # this is not a comment"))
	test(
		"0 1 2 3 4 a\n"+
			"# this is a comment\n"+
			"\n"+
			" \n"+
			"     # another comment\n"+
			"1 2 3 4 5 b\n",
		MustParseEntry("0 1 2 3 4 a"),
		MustParseEntry("1 2 3 4 5 b"))

	testBad(
		"0 1 2 3 4 this line is fine\n" +
			"this line is bogus\n" +
			"0 1 2 3 4 this line is also fine\n")
}
