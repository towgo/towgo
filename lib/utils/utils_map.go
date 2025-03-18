package utils

import (
	"encoding/json"
	"errors"
)

// MapPossibleItemByKey tries to find the possible key-value pair for given key ignoring cases and symbols.
//
// Note that this function might be of low performance.
func MapPossibleItemByKey(data map[string]interface{}, key string) (foundKey string, foundValue interface{}) {
	if len(data) == 0 {
		return
	}
	if v, ok := data[key]; ok {
		return key, v
	}
	// Loop checking.
	for k, v := range data {
		if EqualFoldWithoutChars(k, key) {
			return k, v
		}
	}
	return "", nil
}

func MapToStruct(data interface{}, structPtr interface{}) error {
	var (
		err error
	)
	nodeMap, ok := data.(map[string]interface{})
	if !ok {
		return errors.New("data is not a map")
	}
	marshal, err := json.Marshal(nodeMap)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshal, structPtr)
	if err != nil {
		return err
	}
	return err
}
