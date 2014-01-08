package crontab

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/kevinwallace/fieldsn"
)

var predefinedLabels = map[string]Schedule{
	"@yearly":   MustParseSchedule([]string{"0", "0", "1", "1", "*"}),
	"@annually": MustParseSchedule([]string{"0", "0", "1", "1", "*"}),
	"@monthly":  MustParseSchedule([]string{"0", "0", "1", "*", "*"}),
	"@weekly":   MustParseSchedule([]string{"0", "0", "*", "*", "0"}),
	"@daily":    MustParseSchedule([]string{"0", "0", "*", "*", "*"}),
	"@midnight": MustParseSchedule([]string{"0", "0", "*", "*", "*"}),
	"@hourly":   MustParseSchedule([]string{"0", "*", "*", "*", "*"}),
}

var monthSubstitutions = map[string]int{
	"jan": 1,
	"feb": 2,
	"mar": 3,
	"apr": 4,
	"may": 5,
	"jun": 6,
	"jul": 7,
	"aug": 8,
	"sep": 9,
	"oct": 10,
	"nov": 11,
	"dec": 12,
}

var weekdaySubstitutions = map[string]int{
	"sun": 0,
	"mon": 1,
	"tue": 2,
	"wed": 3,
	"thu": 4,
	"fri": 5,
	"sat": 6,
	"7":   0,
}

func parseRangeSpec(s string, field field, substitutions map[string]int) (rangeSpec, error) {
	var start, end, step int

	slashParts := strings.SplitN(s, "/", 2)
	if len(slashParts) == 2 {
		parsedStep, err := strconv.Atoi(slashParts[1])
		if err != nil {
			return rangeSpec{}, fmt.Errorf("invalid range (can't parse part after slash): %s", err)
		}
		step = parsedStep
	} else {
		step = 1
	}

	if slashParts[0] == "*" || slashParts[0] == "?" {
		start = field.min
		end = field.max
	} else {
		dashParts := strings.SplitN(slashParts[0], "-", 2)
		if substitution, ok := substitutions[dashParts[0]]; ok {
			start = substitution
		} else {
			parsedStart, err := strconv.Atoi(dashParts[0])
			if err != nil {
				return rangeSpec{}, fmt.Errorf("invalid range (can't parse start value): %s", err)
			}
			start = parsedStart
		}

		if len(dashParts) > 1 {
			if substitution, ok := substitutions[dashParts[1]]; ok {
				end = substitution
			} else {
				parsedEnd, err := strconv.Atoi(dashParts[1])
				if err != nil {
					return rangeSpec{}, fmt.Errorf("invalid range (can't parse end value): %s", err)
				}
				end = parsedEnd
			}
		} else {
			end = start
		}
	}

	r := rangeSpec{start, end, step}

	if !r.valid(field) {
		return rangeSpec{}, fmt.Errorf("%s must be between %d and %d", field.name, field.min, field.max)
	}

	return r, nil
}

func parseListSpec(s string, field field, substitutions map[string]int) (listSpec, error) {
	var rangeSpecs listSpec
	for _, rangeString := range strings.Split(s, ",") {
		rangeSpec, err := parseRangeSpec(rangeString, field, substitutions)
		if err != nil {
			return nil, err
		}
		rangeSpecs = append(rangeSpecs, rangeSpec)
	}
	return rangeSpecs, nil
}

// ParseSchedule parses a 5-element array containing the first 5 colunms of a crontab line.
func ParseSchedule(fields []string) (s Schedule, err error) {
	if len(fields) != 5 {
		err = fmt.Errorf("wrong number of fields; expected 5")
		return
	}
	var minute, hour, day, month, weekday listSpec

	minute, err = parseListSpec(fields[0], minuteField, nil)
	if err != nil {
		return
	}
	hour, err = parseListSpec(fields[1], hourField, nil)
	if err != nil {
		return
	}
	day, err = parseListSpec(fields[2], dayField, nil)
	if err != nil {
		return
	}
	month, err = parseListSpec(fields[3], monthField, monthSubstitutions)
	if err != nil {
		return
	}
	weekday, err = parseListSpec(fields[4], weekdayField, weekdaySubstitutions)
	if err != nil {
		return
	}

	s.minute = minute
	s.hour = hour
	s.day = day
	s.month = month
	s.weekday = weekday
	return
}

// MustParseSchedule wraps ParseScheduling, panicing on error.
func MustParseSchedule(fields []string) Schedule {
	s, err := ParseSchedule(fields)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseEntry parses a single line in a crontab.
func ParseEntry(line string) (Entry, error) {
	var schedule Schedule
	var command string
	if line[0] == '@' {
		fields := fieldsn.FieldsN(line, 2)
		label := fields[0]
		predefinedSchedule, ok := predefinedLabels[label]
		if !ok {
			return Entry{}, fmt.Errorf("unknown label %s", label)
		}
		schedule = predefinedSchedule
		if len(fields) > 1 {
			command = fields[1]
		}
	} else {
		fields := fieldsn.FieldsN(line, 6)
		parsedSchedule, err := ParseSchedule(fields[0:5])
		if err != nil {
			return Entry{}, err
		}
		schedule = parsedSchedule
		if len(fields) > 5 {
			command = fields[5]
		}
	}
	return Entry{schedule, command}, nil
}

// MustParseEntry wraps ParseEntry, panicing on error.
func MustParseEntry(line string) Entry {
	e, err := ParseEntry(line)
	if err != nil {
		panic(err)
	}
	return e
}

// ParseCrontab parses the contents of a crontab file.
func ParseCrontab(s string) ([]Entry, error) {
	var entries []Entry
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimLeftFunc(line, unicode.IsSpace)
		if line == "" || line[0] == '#' {
			continue
		}
		entry, err := ParseEntry(line)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
