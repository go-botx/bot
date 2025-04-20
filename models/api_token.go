package models

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type TokenRequest struct {
	ClientRequest `json:"-"`
}

func NewTokenRequest(botId uuid.UUID, tokenSignature string) *TokenRequest {
	return &TokenRequest{
		ClientRequest: newGetRequest(fmt.Sprintf("/api/v2/botx/bots/%s/token?signature=%s", botId, tokenSignature)),
	}
}

func (r *TokenRequest) GetResponse(callFunc ClientApiCallFunc) (resp *TokenResponse, err error) {

	clientResponse, err := parseClientResponseJson(callFunc(r))
	if err != nil {
		return
	}
	resp = &TokenResponse{clientResponseJson: clientResponse}
	err = resp.UnmarshalResult()
	return
}

type TokenResponse struct {
	clientResponseJson
	token string
}

func (r *TokenResponse) UnmarshalResult() error {
	return json.Unmarshal(r.Result, &r.token)
}

func (r *TokenResponse) Token() string {
	return r.token
}
