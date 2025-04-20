package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ChatType string

const (
	ChatTypeChat      = ChatType("chat")
	ChatTypeGroupChat = ChatType("group_chat")
	ChatTypeChannel   = ChatType("channel")
)

type CreateChatRequest struct {
	ClientRequest `json:"-"`
	Name          string     `json:"name"`
	ChatType      ChatType   `json:"chat_type"`
	Members       uuid.UUIDs `json:"members,omitempty"`
}

func NewCreateUserChatRequest(userId uuid.UUID) *CreateChatRequest {
	return &CreateChatRequest{
		ClientRequest: newPostRequest("/api/v3/botx/chats/create").
			SetContentTypeJSONUTF8().
			SetBodyJSON(CreateChatRequest{
				ChatType: ChatTypeChat,
				Members:  uuid.UUIDs{userId},
				Name:     "Personal chat",
			}).WithAuth(),
	}
}

func (r *CreateChatRequest) GetResponse(callFunc ClientApiCallFunc) (resp *CreateChatResponse, err error) {

	clientResponse, err := parseClientResponseJson(callFunc(r))
	if err != nil {
		return
	}
	resp = &CreateChatResponse{clientResponseJson: clientResponse}
	err = resp.UnmarshalResult()
	return
}

type CreateChatResponse struct {
	clientResponseJson

	ChatId uuid.UUID `json:"chat_id"`
}

func (r *CreateChatResponse) UnmarshalResult() error {
	err := json.Unmarshal(r.Result, r)
	return err
}
