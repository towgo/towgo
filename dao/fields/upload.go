package fields

import "encoding/json"

type Upload []interface{}

func (m *Upload) FromDB(data []byte) error {
	if data != nil {
		var result []interface{}
		json.Unmarshal(data, &result)
		*m = result
	}
	return nil
}

func (m *Upload) ToDB() ([]byte, error) {
	if *m == nil {
		return nil, nil
	}
	strByte, err := json.Marshal(m)
	if err != nil {
		return nil, nil
	}
	str := string(strByte)
	return []byte(str), nil
}
