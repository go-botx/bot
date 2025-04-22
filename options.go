package botx

import (
	"errors"
	"io"

	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Option func(b *Bot) error

func WithStatusHandler(handler StatusCallbackHandler) Option {
	return func(b *Bot) error {
		if handler != nil {
			b.statusCallbackHandler = handler
			return nil
		}
		return errors.New("StatusCallbackHandler is nil")
	}
}

func WithCommandHandler(handler CommandCallbackHandler) Option {
	return func(b *Bot) error {
		if handler != nil {
			b.commandCallbackHandler = handler
			return nil
		}
		return errors.New("CommandCallbackHandler is nil")
	}
}

func WithRecoverUnauthorized() Option {
	return func(b *Bot) error {
		b.recoverUnauthorized = true
		return nil
	}
}

func WithDebugHTTPClient(writer ...io.Writer) Option {
	return func(b *Bot) error {
		b.debugHTTPClient = true
		if len(writer) > 0 && writer[0] != nil {
			b.debugHTTPClientWriter = writer[0]
		}
		return nil
	}
}

func WithDebugHTTPService(loggerConfig ...logger.Config) Option {
	return func(b *Bot) error {
		b.debugHTTPService = true
		b.debugHTTPServiceLoggerConfig = logger.ConfigDefault
		if len(loggerConfig) > 0 {
			b.debugHTTPServiceLoggerConfig = loggerConfig[0]
		}
		return nil
	}
}
