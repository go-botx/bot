package models

import (
	"encoding/json"
)

type FindUsersByMailRequest struct {
	ClientRequest `json:"-"`
	Emails        []string `json:"emails"`
}

func NewFindUsersByMailsRequest(emails []string) *FindUsersByMailRequest {
	return &FindUsersByMailRequest{
		ClientRequest: newPostRequest("/api/v3/botx/users/by_email").
			WithAuth().SetContentTypeJSONUTF8().SetBodyJSON(FindUsersByMailRequest{
			Emails: emails,
		}),
	}
}

func (r *FindUsersByMailRequest) GetResponse(callFunc ClientApiCallFunc) (resp *FindUsersByMailResponse, err error) {
	clientResponse, err := parseClientResponseJson(callFunc(r))
	if err != nil {
		return
	}
	resp = &FindUsersByMailResponse{clientResponseJson: clientResponse}
	err = resp.UnmarshalResult()
	return
}

type FindUsersByMailResponse struct {
	clientResponseJson
	Users []User
}

func (r *FindUsersByMailResponse) UnmarshalResult() error {
	return json.Unmarshal(r.Result, &r.Users)
}
