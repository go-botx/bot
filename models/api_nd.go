package models

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type NDRequest struct {
	ClientRequest `json:"-"`
	ChatId        uuid.UUID
	Recipients    uuid.UUIDs
	Notification  NDNotification
	File          *NDFile
	StealthMode   bool
	SendPush      bool
	ForceDND      bool
}

func NewNDRequest(chatId uuid.UUID, body string, options ...NDRequestOption) (*NDRequest, error) {
	ndr := &NDRequest{
		ChatId: chatId,
		Notification: NDNotification{
			Body: body,
		},
	}
	for _, opt := range options {
		err := opt(ndr)
		if err != nil {
			return nil, err
		}
	}
	ndr.ClientRequest = newPostRequest("/api/v4/botx/notifications/direct").WithAuth().SetContentTypeJSONUTF8().SetBodyJSON(ndr)
	return ndr, nil
}

func (nd *NDRequest) GetResponse(callFunc ClientApiCallFunc) (resp *SyncIdResponse, err error) {
	clientResponse, err := parseClientResponseJson(callFunc(nd))
	if err != nil {
		return
	}
	resp = &SyncIdResponse{clientResponseJson: clientResponse}
	err = resp.UnmarshalResult()
	return
}

func (nd *NDRequest) MarshalJSON() (data []byte, err error) {
	m := map[string]any{}
	m["group_chat_id"] = nd.ChatId
	m["notification"] = nd.Notification
	if len(nd.Recipients) > 0 {
		m["recipients"] = []string(nd.Recipients.Strings())
	}
	if nd.File != nil {
		mFile := map[string]string{}
		mFile["file_name"] = nd.File.FileName
		mFile["data"] = nd.File.Data
		m["file"] = mFile
	}
	if !nd.SendPush || nd.StealthMode || nd.ForceDND {
		mOpts := map[string]any{}
		if nd.StealthMode {
			mOpts["stealth_mode"] = true
		}
		if !nd.SendPush || nd.ForceDND {
			mNOpts := map[string]bool{}
			if !nd.SendPush {
				mNOpts["send"] = nd.SendPush
			}
			if nd.ForceDND {
				mNOpts["force_dnd"] = nd.ForceDND
			}
		}
		m["opts"] = mOpts
	}
	return json.Marshal(m)
}

type NDNotification struct {
	Status   string          `json:"status,omitempty"`
	Body     string          `json:"body"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Opts     struct {
		SilentResponse    bool `json:"silent_response,omitempty"`
		ButtonsAutoAdjust bool `json:"buttons_auto_adjust,omitempty"`
	} `json:"opts,omitempty"`
	Keyboard NDButtons   `json:"keyboard,omitempty"`
	Bubble   NDButtons   `json:"bubble,omitempty"`
	Mentions []NDMention `json:"mentions,omitempty"`
}

type NDFile struct {
	FileName string
	Data     string
}

type NDButtonRow []NDButton

type NDButtons []NDButtonRow

type NDMentionType string

const (
	NDMentionTypeAll     NDMentionType = NDMentionType("all")
	NDMentionTypeChannel NDMentionType = NDMentionType("channel")
	NDMentionTypeChat    NDMentionType = NDMentionType("chat")
	NDMentionTypeContact NDMentionType = NDMentionType("contact")
	NDMentionTypeUser    NDMentionType = NDMentionType("user")
)

type NDMention struct {
	MentionType NDMentionType  `json:"mention_type"`
	MentionId   uuid.UUID      `json:"mention_id"`
	MentionData *NDMentionData `json:"mention_data,omitempty"`
}

type NDMentionData struct {
	ChatId   uuid.UUID `json:"group_chat_id,omitempty"`
	UserHUID uuid.UUID `json:"user_huidm,omitempty"`
	Name     string    `json:"name,omitempty"`
}

type NDRequestOption func(ndr *NDRequest) error

func WithNDBubbleRow(buttons ...NDButton) NDRequestOption {
	return func(ndr *NDRequest) error {
		ndr.Notification.Bubble = append(ndr.Notification.Bubble, NDButtonRow(buttons))
		return nil
	}
}

func WithNDMetadata(metadata any) NDRequestOption {
	data, err := json.Marshal(metadata)
	return func(ndr *NDRequest) error {
		ndr.Notification.Metadata = data
		return err
	}
}

func WithNDMention(mentionId uuid.UUID, mentionType NDMentionType, subjectName string, subjectUUID uuid.UUID) NDRequestOption {
	return func(ndr *NDRequest) error {
		m := NDMention{
			MentionId:   mentionId,
			MentionType: mentionType,
		}
		switch mentionType {
		case NDMentionTypeAll:

		case NDMentionTypeChannel, NDMentionTypeChat:
			m.MentionData.Name = subjectName
			m.MentionData.ChatId = subjectUUID
		case NDMentionTypeContact, NDMentionTypeUser:
			m.MentionData.Name = subjectName
			m.MentionData.UserHUID = subjectUUID
		default:
			return fmt.Errorf("unknown mention type '%s'", mentionType)
		}
		ndr.Notification.Mentions = append(ndr.Notification.Mentions, m)
		return nil
	}
}
