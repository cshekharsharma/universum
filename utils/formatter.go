package utils

import (
	"strconv"
	"strings"
	"time"
)

func ParseDurationToSeconds(input string) (int64, error) {
	if strings.HasSuffix(input, "s") || strings.HasSuffix(input, "m") || strings.HasSuffix(input, "h") {
		duration, err := time.ParseDuration(input)
		if err != nil {
			return 0, err
		}
		return int64(duration.Seconds()), nil
	}

	// If no suffix is provided, assume the input is in seconds
	seconds, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0, err
	}
	return seconds, nil
}

func ParseSizeToBytes(input string) (int64, error) {
	// Define multipliers for data sizes
	multipliers := map[string]int64{
		"B": 1,
		"K": 1024,
		"M": 1024 * 1024,
		"G": 1024 * 1024 * 1024,
		"T": 1024 * 1024 * 1024 * 1024,
	}

	// Iterate through the known suffixes
	for suffix, multiplier := range multipliers {
		if strings.HasSuffix(input, suffix) {
			valueStr := strings.TrimSuffix(input, suffix)
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return 0, err
			}
			return int64(value * float64(multiplier)), nil
		}
	}

	// If no suffix is provided, assume the input is in bytes
	bytes, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0, err
	}
	return bytes, nil
}
