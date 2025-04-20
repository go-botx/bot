package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type SyncIdResponse struct {
	clientResponseJson
	SyncId uuid.UUID `json:"sync_id"`
}

func (r *SyncIdResponse) UnmarshalResult() error {
	err := json.Unmarshal(r.Result, r)
	return err
}
