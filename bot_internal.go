package bot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-botx/go-botx/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func new(botCredentials string, options ...Option) (bot *Bot, err error) {
	bot = nil
	account, err := models.ParseCredentials(botCredentials)
	if err != nil {
		return
	}
	bot = &Bot{
		account:                account,
		userIdChatIdCache:      newKVCache[uuid.UUID, uuid.UUID](0),
		emailUserIdCache:       newKVCache[string, uuid.UUID](time.Minute * 30),
		ncbManager:             newNCBManager(time.Second * 30),
		statusCallbackHandler:  nil,
		commandCallbackHandler: nil,
		fiberApp:               nil,
		debugHTTPClient:        false,
		debugHTTPClientWriter:  nil,
		debugHTTPService:       false,
		jwtLeeway:              time.Duration(time.Minute * 5),
	}

	bot.initTokenSignature()
	bot.initJWTValidatingKey()

	for _, option := range options {
		err = option(bot)
		if err != nil {
			return
		}
	}

	return
}

func (b *Bot) initFiberApp() {
	if b.fiberApp != nil {
		return
	}

	jwtValidMethodNames := []string{
		jwt.SigningMethodHS256.Name,
	}
	b.jwtParser = jwt.NewParser(jwt.WithLeeway(b.jwtLeeway), jwt.WithValidMethods(jwtValidMethodNames))

	b.fiberApp = fiber.New()
	if b.debugHTTPService {
		b.fiberApp.Use(logger.New(b.debugHTTPServiceLoggerConfig))
	}
	recoverConfig := recover.ConfigDefault
	if b.debugHTTPService {
		recoverConfig.EnableStackTrace = true
	}
	b.fiberApp.Use(recover.New(recoverConfig))
	b.fiberApp.Use(b.authenticateCallback)
	b.fiberApp.Get("/status", b.handleStatusCallback)
	b.fiberApp.Post("/command", b.handleCommandCallback)
	b.fiberApp.Post("/notification/callback", b.handleNotificationCallback)
}

func (b *Bot) initTokenSignature() {
	hasher := hmac.New(sha256.New, []byte(b.account.SecretKey))
	hasher.Write([]byte(b.account.Id.String()))
	b.tokenSig = strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
}

func (b *Bot) initJWTValidatingKey() {
	b.jwtValidatingKey = []byte(b.account.SecretKey)
}

func (b *Bot) getToken() (token string) {
	b.tokenRWMtx.RLock()
	token = b.token
	b.tokenRWMtx.RUnlock()
	if token != "" {
		return token
	}

	b.tokenRWMtx.Lock()
	defer b.tokenRWMtx.Unlock()
	if b.token != "" {
		return b.token
	}

	resp, err := models.NewTokenRequest(b.account.Id, b.tokenSig).GetResponse(b.callApi)
	if err != nil {
		panic(fmt.Errorf("failed get token: %w", err))
	}
	b.token = resp.Token()
	return b.token
}

func (b *Bot) callApi(req models.ClientRequest) (statusCode int, resp []byte, err error) {
	a := fiber.AcquireAgent()
	a.NoDefaultUserAgentHeader = true
	a.Request().SetRequestURI("https://" + b.account.CTSHost + req.RelativeReference())
	a.Request().Header.SetMethod(req.Method())
	if req.HasContentType() {
		a.Request().Header.SetContentType(req.ContentType())
	}
	if req.NeedAuthorization() {
		a.Set(fiber.HeaderAuthorization, ("Bearer " + b.getToken()))
	}
	if req.HasBody() {
		a.Request().SetBody(req.Body())
	}

	if b.debugHTTPClient {
		if b.debugHTTPClientWriter != nil {
			a.Debug(b.debugHTTPClientWriter)
		} else {
			a.Debug()
		}
	}

	err = a.Parse()
	if err != nil {
		return -1, nil, err
	}

	statusCode, body, errs := a.Bytes()
	if len(errs) > 0 {
		return statusCode, nil, errs[0]
	}
	return statusCode, body, nil
}

func (b *Bot) handleNotificationCallback(c *fiber.Ctx) error {
	var ncbr models.NotificationCallbackRequest
	err := c.BodyParser(&ncbr)
	if err != nil {
		return err
	}
	b.ncbManager.storeCallback(ncbr)
	return c.SendStatus(fiber.StatusAccepted)
}

func (b *Bot) handleStatusCallback(c *fiber.Ctx) error {
	var statusReq models.StatusRequest
	err := c.QueryParser(&statusReq)
	if err != nil {
		return err
	}
	if statusReq.BotId != b.account.Id {
		return errors.New("status requested for different bot id")
	}
	var response *models.StatusResponse = nil
	if b.statusCallbackHandler != nil {
		response = b.statusCallbackHandler(b, &statusReq)
	}
	if response == nil {
		response = models.NewStatusResponse(false, "")
	}
	return c.Status(fiber.StatusOK).JSON(&response, fiber.MIMEApplicationJSONCharsetUTF8)
}

func (b *Bot) handleCommandCallback(c *fiber.Ctx) error {
	var commandReq models.CommandRequest
	err := c.BodyParser(&commandReq)
	if err != nil {
		return err
	}
	if commandReq.BotId != b.account.Id {
		return errors.New("command requested for different bot id")
	}
	var response *models.CommandResponse = nil
	if b.commandCallbackHandler != nil {
		response = models.CommandResponseSuccess()
		dataReady := make(chan bool)
		go func() {
			var b = b
			var commandReq = commandReq
			dataReady <- true
			b.commandCallbackHandler(b, &commandReq)
		}()
		<-dataReady
	} else {
		response = models.CommandResponseFailure(errors.New("callback not implemented"))
	}
	return c.Status(response.StatusCode).JSON(response)
}

func (b *Bot) authenticateCallback(c *fiber.Ctx) error {
	authString := c.Get(fiber.HeaderAuthorization, "")
	if authString == "" {
		return b.sendAuthenticateCallbackError(c, fiber.StatusUnauthorized, "expected 'Authorization: Bearer ...' header")
	}
	authStringParts := strings.SplitN(authString, " ", 2)
	if len(authStringParts) < 2 || strings.ToLower(authStringParts[0]) != "bearer" {
		return b.sendAuthenticateCallbackError(c, fiber.StatusUnprocessableEntity, "wrong Authorization header")
	}
	tokenString := strings.TrimSpace(authStringParts[1])
	token, err := b.jwtParser.Parse(tokenString, b.jwtKeyFunc)
	if err != nil {
		return b.sendAuthenticateCallbackError(c, fiber.StatusUnprocessableEntity, err.Error())
	}
	if !token.Valid {
		return b.sendAuthenticateCallbackError(c, fiber.StatusUnprocessableEntity, "token was not validated")
	}
	return c.Next()
}

func (b *Bot) jwtKeyFunc(t *jwt.Token) (interface{}, error) {
	audienceList, err := t.Claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT audience: %w", err)
	}
	audienceListLen := len(audienceList)
	if audienceListLen != 1 {
		return nil, fmt.Errorf("failed to get JWT audience: number of audiences is %d, expected 1", audienceListLen)
	}
	tokenBotId, err := uuid.Parse(audienceList[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT audience '%s' as Bot Id: %w", audienceList[0], err)
	}
	if b.account.Id != tokenBotId {
		return nil, fmt.Errorf("token is issued for bot with id '%s', but '%s' expected", audienceList[0], b.account.Id)
	}
	return b.jwtValidatingKey, nil
}

func (b *Bot) sendAuthenticateCallbackError(c *fiber.Ctx, statusCode int, reason string) error {
	return c.Status(fiber.StatusForbidden).JSON(&map[string]any{"status": "error", "reason": reason}, fiber.MIMEApplicationJSONCharsetUTF8)
}
