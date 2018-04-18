package lib

import (
	"errors"
	"strconv"
)

func ConvertTime(s string) (int, error) {
	secondsPerUnit := map[string]int{
		"s": 1,
		"m": 60,
		"h": 3600,
		"d": 86400,
		"w": 604800,
	}

	lastChar := len(s) - 1

	if seconds, err := strconv.Atoi(s); err == nil {
		return seconds, nil
	}

	if seconds, err := strconv.Atoi(s[:lastChar]); err == nil {
		return seconds * secondsPerUnit[s[lastChar:]], nil
	}

	return 0, errors.New("Could not convert string to number of seconds")
}
