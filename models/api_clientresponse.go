package models

import (
	"encoding/json"
	"fmt"
)

type ClientResponse interface {
	GetReason() string
	SetStatusCode(statusCode int)
	GetStatusCode() int
	GetError() error
}

type clientResponseJson struct {
	ClientResponse
	StatusCode int             `json:"-"`
	Status     string          `json:"status,omitempty"`
	Reason     string          `json:"reason,omitempty"`
	ErrorData  any             `json:"error_data,omitempty"`
	Errors     *[]string       `jsob:"errors,omitempty"`
	Result     json.RawMessage `jsob:"errors,result"`
}

func (rb *clientResponseJson) GetReason() (reason string) {
	return rb.Reason
}

func (rb *clientResponseJson) SetStatusCode(statusCode int) {
	rb.StatusCode = statusCode
}

func (rb *clientResponseJson) GetStatusCode() (statusCode int) {
	return rb.StatusCode
}

func (rb *clientResponseJson) GetError() (err error) {
	if rb.Status != "ok" {
		return fmt.Errorf("%s", rb.Reason)
	}
	return nil
}

func parseClientResponseJson(code int, body []byte, errIn error) (clientResponse clientResponseJson, err error) {
	if errIn != nil {
		err = errIn
		return
	}
	err = json.Unmarshal(body, &clientResponse)
	if err != nil {
		return
	}
	err = clientResponse.GetError()
	if err != nil {
		return
	}
	clientResponse.SetStatusCode(code)
	return clientResponse, nil
}
