package models

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CommandType string

const (
	CommandTypeUser   = CommandType("user")
	CommandTypeSystem = CommandType("system")
)

type CommandRequest struct {
	SyncId       uuid.UUID         `json:"sync_id"`
	SourceSyncId uuid.UUID         `json:"source_sync_id"`
	Command      CommandData       `json:"command"`
	Attachments  json.RawMessage   `json:"attachments"`
	From         CommandFrom       `json:"from"`
	AsyncFiles   []json.RawMessage `json:"async_files"`
	BotId        uuid.UUID         `json:"bot_id"`
	ProtoVersion int               `json:"proto_version"`
	Entities     []json.RawMessage `json:"entities"`
}

type CommandData struct {
	Body        string          `json:"body"`
	CommandType CommandType     `json:"command_type"`
	Data        json.RawMessage `json:"data"`
	Metadata    json.RawMessage `json:"metadata"`
}

type CommandFrom struct {
	UserHUID    uuid.UUID `json:"user_huid"`
	GroupChatId uuid.UUID `json:"group_chat_id"`
	ChatType    ChatType  `json:"chat_type"`
	///	ad_login [String] (Default: null) - логин юзера который отправил команду
	/// ad_domain [String] (Default: null) - домен юзера который отправил команду
	/// username [String] (Default: null) - имя юзера который отправил команду
	IsAdmin bool `json:"is_admin"`
	/// is_creator [Boolean] (Default: null) - является ли юзер создателем чата
	/// manufacturer [String] (Default: null) - имя бренда производителя
	/// device [String] (Default: null) - название девайса
	/// device_software [String] (Default: null) - ОС девайса
	/// device_meta [Object] (Default: null)
	///    pushes [Boolean] - разрешение приложению на отправку пушей
	///    timezone [String] - таймзона пользователя
	///    permissions [Object] - различные разрешения приложения (использование микрофона, камеры и т.д.)
	/// platform [String] (Default: null) - название клиентской платформы (web|android|ios|desktop)
	/// platform_package_id [String] (Default: null) - идентификатор пакета с данными приложения и устройства
	/// app_version [String] (Default: null) - версия приложения Express
	Locale string `json:"locale"`
	/// host [String] - имя хоста с которого пришла команда
}

type CommandResponse struct {
	StatusCode    int
	Success       bool
	StatusMessage string
}

func (cr *CommandResponse) MarshalJSON() ([]byte, error) {
	if cr.Success {
		return json.Marshal(map[string]bool{})
	}
	return json.Marshal(map[string]any{"reason": "bot_disabled", "error_data": map[string]string{"status_message": cr.StatusMessage}})
}

func CommandResponseSuccess() *CommandResponse {
	return &CommandResponse{
		Success:    true,
		StatusCode: fiber.StatusAccepted,
	}
}

func CommandResponseFailure(err error) *CommandResponse {
	return &CommandResponse{
		StatusCode:    fiber.StatusServiceUnavailable,
		Success:       false,
		StatusMessage: err.Error(),
	}
}
