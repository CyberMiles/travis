package utils

import(
	"github.com/satori/go.uuid"
)

func GetUUID() []byte {
	return uuid.Must(uuid.NewV4()).Bytes()
}
