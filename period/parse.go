// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"fmt"
	"strconv"
	"strings"
)

// MustParse is as per Parse except that it panics if the string cannot be parsed.
// This is intended for setup code; don't use it for user inputs.
func MustParse(value string) Period {
	d, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return d
}

// Parse parses strings that specify periods using ISO-8601 rules.
//
// In addition, a plus or minus sign can precede the period, e.g. "-P10D"
//
// The zero value can be represented in several ways: all of the following
// are equivalent: "P0Y", "P0M", "P0W", "P0D", "PT0H", PT0M", PT0S", and "P0".
// The canonical zero is "P0D".
func Parse(period string) (Period, error) {
	if period == "" {
		return Period{}, fmt.Errorf("cannot parse a blank string as a period")
	}

	if period == "P0" {
		return Period{}, nil
	}

	pcopy := period
	negate := false
	if pcopy[0] == '-' {
		negate = true
		pcopy = pcopy[1:]
	} else if pcopy[0] == '+' {
		pcopy = pcopy[1:]
	}

	if pcopy[0] != 'P' {
		return Period{}, fmt.Errorf("expected 'P' period mark at the start: %s", period)
	}
	pcopy = pcopy[1:]

	result := Period{}

	st := parseState{period, pcopy, false, nil}
	t := strings.IndexByte(pcopy, 'T')
	if t >= 0 {
		st.pcopy = pcopy[t+1:]

		result.hours, st = parseField(st, 'H')
		if st.err != nil {
			return Period{}, fmt.Errorf("expected a number before the 'H' marker: %s", period)
		}

		result.minutes, st = parseField(st, 'M')
		if st.err != nil {
			return Period{}, fmt.Errorf("expected a number before the 'M' marker: %s", period)
		}

		result.seconds, st = parseField(st, 'S')
		if st.err != nil {
			return Period{}, fmt.Errorf("expected a number before the 'S' marker: %s", period)
		}

		st.pcopy = pcopy[:t]
	}

	result.years, st = parseField(st, 'Y')
	if st.err != nil {
		return Period{}, fmt.Errorf("expected a number before the 'Y' marker: %s", period)
	}

	result.months, st = parseField(st, 'M')
	if st.err != nil {
		return Period{}, fmt.Errorf("expected a number before the 'M' marker: %s", period)
	}

	weeks, st := parseField(st, 'W')
	if st.err != nil {
		return Period{}, fmt.Errorf("expected a number before the 'W' marker: %s", period)
	}

	days, st := parseField(st, 'D')
	if st.err != nil {
		return Period{}, fmt.Errorf("expected a number before the 'D' marker: %s", period)
	}

	result.days = weeks*7 + days
	//fmt.Printf("%#v\n", st)

	if !st.ok {
		return Period{}, fmt.Errorf("expected 'Y', 'M', 'W', 'D', 'H', 'M', or 'S' marker: %s", period)
	}
	if negate {
		return result.Negate(), nil
	}
	return result, nil
}

type parseState struct {
	period, pcopy string
	ok            bool
	err           error
}

func parseField(st parseState, mark byte) (int16, parseState) {
	//fmt.Printf("%c %#v\n", mark, st)
	r := int16(0)
	m := strings.IndexByte(st.pcopy, mark)
	if m > 0 {
		r, st.err = parseDecimalFixedPoint(st.pcopy[:m], st.period)
		if st.err != nil {
			return 0, st
		}
		st.pcopy = st.pcopy[m+1:]
		st.ok = true
	}
	return r, st
}

// Fixed-point three decimal places
func parseDecimalFixedPoint(s, original string) (int16, error) {
	//was := s
	dec := strings.IndexByte(s, '.')
	if dec < 0 {
		dec = strings.IndexByte(s, ',')
	}

	if dec >= 0 {
		dp := len(s) - dec
		if dp > 1 {
			s = s[:dec] + s[dec+1:dec+2]
		} else {
			s = s[:dec] + s[dec+1:] + "0"
		}
	} else {
		s = s + "0"
	}

	n, e := strconv.ParseInt(s, 10, 16)
	return int16(n), e
}
