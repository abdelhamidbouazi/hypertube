package utils

import (
	"fmt"
	"io"
	"os"
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

func CheckFileExits(filePath string) error {
	_, err := os.Stat(filePath)
	return err
}

func WaitForFile(filePath string) error {
	for {
		if err := CheckFileExits(filePath); err == nil {
			time.Sleep(100 * time.Millisecond)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func GetLanguageLabel(code string) string {
	languageLabels := map[string]string{
		"en": "English",
		"fr": "French",
		"es": "Spanish",
		"ar": "Arabic",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"ru": "Russian",
		"ja": "Japanese",
		"ko": "Korean",
		"zh": "Chinese",
		"hi": "Hindi",
		"nl": "Dutch",
		"pl": "Polish",
		"tr": "Turkish",
	}
	if label, ok := languageLabels[code]; ok {
		return label
	}
	return code
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func ParseBandwidth(bitrate string) uint32 {
	bitrate = strings.TrimSpace(bitrate)
	bitrate = strings.ToLower(bitrate)

	if strings.HasSuffix(bitrate, "k") {
		value := strings.TrimSuffix(bitrate, "k")
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return uint32(v * 1000)
		}
	} else if strings.HasSuffix(bitrate, "m") {
		value := strings.TrimSuffix(bitrate, "m")
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			return uint32(v * 1000000)
		}
	} else {
		if v, err := strconv.ParseFloat(bitrate, 64); err == nil {
			return uint32(v)
		}
	}

	return 0
}
