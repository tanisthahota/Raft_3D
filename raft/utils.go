package raft

import "encoding/json"

// MustMarshal marshals data to JSON, ignoring errors
func MustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}