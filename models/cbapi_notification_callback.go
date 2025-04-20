package models

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type NotificationCallbackRequest struct {
	SyncId    uuid.UUID       `json:"sync_id"`
	Status    string          `json:"status"`
	Result    json.RawMessage `json:"result,omitempty"`
	Reason    string          `json:"reason,omitempty"`
	Errors    []string        `json:"errors,omitempty"`
	ErrorData json.RawMessage `json:"error_data,omitempty"`
}

func (ncbr *NotificationCallbackRequest) GetError() error {
	if ncbr.Status != "ok" {
		return fmt.Errorf("%s", ncbr.Reason)
	}
	return nil
}
