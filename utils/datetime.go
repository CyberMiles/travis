package utils

import (
	"fmt"
	"math"
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

func DiffMinutes(t string) (int64, error) {
	dt, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return 0, err
	}

	d := time.Now().Sub(dt)
	return int64(math.Ceil(d.Hours())), nil
}
