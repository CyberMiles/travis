package utils

import (
	"fmt"
	"math"
	"time"
)

func FormatUnixTime(ts int64) string {
	return time.Unix(ts, 0).Format(time.RFC3339)
}

func GetTimeBefore(ts int64, hours int) (string, error) {
	d, err := time.ParseDuration(fmt.Sprintf("-%dh", hours))
	if err != nil {
		return "", err
	}

	return time.Unix(ts, 0).Add(d).UTC().Format(time.RFC3339), nil
}

func Diff(t string) (int64, error) {
	dt, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return 0, err
	}

	d := time.Now().Sub(dt)
	return int64(math.Ceil(d.Hours())), nil
}

func GetNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func GetTimeBeforeNow(hours int) (string, error) {
	d, err := time.ParseDuration(fmt.Sprintf("-%dh", hours))
	if err != nil {
		return "", err
	}

	return time.Now().Add(d).UTC().Format(time.RFC3339), nil
}
