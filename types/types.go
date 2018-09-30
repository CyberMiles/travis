package types

import (
	"encoding/json"
	"github.com/CyberMiles/travis/utils"
	"reflect"
)

type Hashable interface {
	Hash() []byte
}

func Hash(data interface{}, excludedFields []string) []byte {
	fields := make(map[string]interface{})
	v := reflect.ValueOf(data).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i)
		name := fieldInfo.Name
		if !utils.Contains(excludedFields, name) {
			fields[name] = v.Field(i).Interface()
		}
	}

	bs, err := json.Marshal(fields)
	if err != nil {
		panic(err)
	}

	return bs
}
