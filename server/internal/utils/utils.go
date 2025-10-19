package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseDuration(s string) (time.Duration, error) {
	var tm time.Duration

	r, err := regexp.Compile("([0-9]+)(y|mo|w|d|h|m|s){1}")
	if err != nil {
		return tm, err
	}

	if len(r.FindString(s)) != len(s) {
		return tm, fmt.Errorf("could not parse duration")
	}

	var dur int

	switch {
	// year
	case strings.Contains(s, "y"):
		_d, err := strconv.Atoi(s[:strings.Index(s, "y")])
		dur = _d * 365 * 24 * 60 * 60
		if err != nil {
			return tm, fmt.Errorf("could not parse duration")
		}
	// month
	case strings.Contains(s, "mo"):
		_d, err := strconv.Atoi(s[:strings.Index(s, "mo")])
		dur = _d * 30 * 24 * 60 * 60
		if err != nil {
			return tm, fmt.Errorf("could not parse duration")
		}
	// weeks
	case strings.Contains(s, "w"):
		_d, err := strconv.Atoi(s[:strings.Index(s, "w")])
		dur = _d * 7 * 24 * 60 * 60
		if err != nil {
			return tm, fmt.Errorf("could not parse duration")
		}
	// days
	case strings.Contains(s, "d"):
		_d, err := strconv.Atoi(s[:strings.Index(s, "d")])
		dur = _d * 24 * 60 * 60
		if err != nil {
			return tm, fmt.Errorf("could not parse duration")
		}
	// hours
	case strings.Contains(s, "h"):
		_d, err := strconv.Atoi(s[:strings.Index(s, "h")])
		dur = _d * 60 * 60
		if err != nil {
			return tm, fmt.Errorf("could not parse duration")
		}
	// minutes
	case strings.Contains(s, "m"):
		_d, err := strconv.Atoi(s[:strings.Index(s, "m")])
		dur = _d * 60
		if err != nil {
			return tm, fmt.Errorf("could not parse duration")
		}
	// seconds
	case strings.Contains(s, "s"):
		_d, err := strconv.Atoi(s[:strings.Index(s, "s")])
		dur = _d
		if err != nil {
			return tm, fmt.Errorf("could not parse duration")
		}

	default:
		return tm, fmt.Errorf("could not parse duration")
	}

	return time.Second * time.Duration(dur), nil
}
