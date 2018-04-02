package utils

import "time"

func GetNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}