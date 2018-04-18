package utils

import (
	"github.com/satori/go.uuid"
)

func GetUUID() []byte {
	return uuid.Must(uuid.NewV4()).Bytes()
}

func RemoveFromSlice(slice []interface{}, i int) []interface{} {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}
