package utils

import (
	"fmt"
	"time"
)

func GetNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func GetTimeBefore(hours int) (string, error) {
	d, err := time.ParseDuration(fmt.Sprintf("-%dh", hours))
	if err != nil {
		return "", err
	}

	return time.Now().Add(d).UTC().Format(time.RFC3339), nil
}
