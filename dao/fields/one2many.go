package fields

import (
	"encoding/json"
	"strings"
)

type One2Many []interface{}

func (m *One2Many) FromDB(data []byte) error {
	if data != nil {
		arr := strings.Split(string(data), ",")
		var result []interface{}
		for _, str := range arr {
			result = append(result, str)
		}
		*m = result
	}
	return nil
}

func (m *One2Many) ToDB() ([]byte, error) {
	if *m == nil {
		return nil, nil
	}
	strByte, err := json.Marshal(m)
	if err != nil {
		return nil, nil
	}
	str := string(strByte)
	return []byte(str[1 : len(str)-1]), nil
}
