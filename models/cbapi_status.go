package models

import (
	"github.com/google/uuid"
)

type StatusRequest struct {
	BotId    uuid.UUID `query:"bot_id"`
	UserHUID uuid.UUID `query:"user_huid"`
	ADLogin  string    `query:"ad_login"`
	ADDomain string    `query:"ad_domain"`
	IsAdmin  bool      `query:"is_admin"`
	ChatType ChatType  `query:"chat_type"`
}

type StatusResponse struct {
	Status string               `json:"status,omitempty"`
	Result StatusResponseResult `json:"result,omitempty"`
}

type StatusResponseResult struct {
	Enabled       bool                    `json:"enabled"`
	StatusMessage string                  `json:"status_message,omitempty"`
	Commands      []StatusResponseCommand `json:"commands,omitempty"`
}

type StatusResponseCommand struct {
	Description string `json:"description"`
	Body        string `json:"body"`
	Name        string `json:"name"`
}

func NewStatusResponse(enabled bool, statusMessage string, commands ...StatusResponseCommand) *StatusResponse {
	sr := &StatusResponse{
		Status: "ok",
		Result: StatusResponseResult{
			Enabled:       enabled,
			StatusMessage: statusMessage,
			Commands:      make([]StatusResponseCommand, 0),
		},
	}
	sr.Result.Commands = append(sr.Result.Commands, commands...)
	return sr
}
