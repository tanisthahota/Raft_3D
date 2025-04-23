package api

import "encoding/json"

func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
