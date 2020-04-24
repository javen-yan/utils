package server

import "encoding/json"

const (
	ErrServer = "0001"
)

type HttpRes struct {
	Code string      `json:"code"`
	Msg  string      `json:"message,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func (httpRes *HttpRes) ToJson() []byte {
	res, _ := json.Marshal(httpRes)
	return res
}
