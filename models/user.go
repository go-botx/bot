package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserHUID        uuid.UUID `json:"user_huid"`
	ADLogin         string    `json:"ad_login"`
	ADDomain        string    `json:"ad_domain"`
	Name            string    `json:"name"`
	Company         string    `json:"company"`
	CompanyPosition string    `json:"company_position"`
	Department      string    `json:"department"`
	EMails          []string  `json:"emails"`
	OtherId         string    `json:"other_id"`
	UserKind        string    `json:"user_kind"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CTSId           uuid.UUID `json:"cts_id"`
	Description     string    `json:"description"`
	IPPhone         string    `json:"ip_phone"`
	Manager         string    `json:"manager"`
	Office          string    `json:"office"`
	PublicName      string    `json:"public_name"`
	RTSId           uuid.UUID `json:"rts_id"`

	//TODO: other_ip_phone
	//TODO: other_phone

}
