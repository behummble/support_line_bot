package encoding

import (
	"encoding/json"
)

func ToJSON(data interface{}) ([]byte, error) {
	res, err := json.Marshal(data)
	return res, err
}

func FromJSON(data []byte, to interface{}) (interface{}, error) {
	err := json.Unmarshal(data, &to)
	return to, err
}