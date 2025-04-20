package models

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Credentials struct {
	Id        uuid.UUID `json:"id"`
	CTSHost   string    `json:"cts_host"`
	SecretKey string    `json:"secret_key"`
}

func ParseCredentials(accountString string) (c Credentials, err error) {
	parts := strings.Split(accountString, "@")
	if len(parts) != 3 {
		err = fmt.Errorf("bot credentials must be in 'cts_host@secret_key@bot_id' form")
		return
	}
	id, err := uuid.Parse(parts[2])
	c = Credentials{
		CTSHost:   parts[0],
		SecretKey: parts[1],
		Id:        id,
	}
	return
}
