package bot

import (
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/go-botx/go-botx/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type StatusCallbackHandler func(b *Bot, req *models.StatusRequest) *models.StatusResponse
type CommandCallbackHandler func(b *Bot, req *models.CommandRequest)

type Bot struct {
	account          models.Credentials
	tokenSig         string
	jwtValidatingKey []byte

	token      string
	tokenRWMtx sync.RWMutex

	fiberApp *fiber.App

	recoverUnauthorized bool

	debugHTTPClient              bool
	debugHTTPClientWriter        io.Writer
	debugHTTPService             bool
	debugHTTPServiceLoggerConfig logger.Config

	// chache
	userIdChatIdCache *kvCache[uuid.UUID, uuid.UUID]
	emailUserIdCache  *kvCache[string, uuid.UUID]

	// sync callback manager
	ncbManager *ncbManager

	// handlers
	statusCallbackHandler  StatusCallbackHandler
	commandCallbackHandler CommandCallbackHandler

	// jwt options
	jwtLeeway time.Duration
	jwtParser *jwt.Parser
}

func New(botCredentials string, options ...Option) (bot *Bot, err error) {
	return new(botCredentials, options...)
}

func (b *Bot) FiberApp() *fiber.App {
	b.initFiberApp()
	return b.fiberApp
}

func (b *Bot) Id() uuid.UUID {
	return b.account.Id
}

func (b *Bot) CTSHost() string {
	return b.account.CTSHost
}

func (b *Bot) FindUserHUIDsByMails(emails []string) (users []models.User, err error) {
	result := map[uuid.UUID]map[string]bool{}
	mailsNotFound := map[string]bool{}

	for _, email := range emails {
		email = strings.ToLower(email)
		if huid, ok := b.emailUserIdCache.Get(email); ok {
			if _, ok := result[huid]; !ok {
				result[huid] = map[string]bool{}
			}
			result[huid][email] = true
		} else {
			mailsNotFound[email] = true
		}
	}
	mailsToFind := []string{}
	for email, _ := range mailsNotFound {
		mailsToFind = append(mailsToFind, email)
	}
	foundUsers, err := b.FindUsersByMails(mailsToFind)
	if err != nil {
		return nil, err
	}
	for _, u := range foundUsers {
		emails := map[string]bool{}
		for _, email := range u.EMails {
			email = strings.ToLower(email)
			emails[email] = true
			b.emailUserIdCache.Set(email, u.UserHUID)
		}
		result[u.UserHUID] = emails
	}
	for userId, mails := range result {
		mailsList := make([]string, len(mails))
		for email, _ := range mails {
			mailsList = append(mailsList, email)
		}
		users = append(users, models.User{
			UserHUID: userId,
			EMails:   mailsList,
		})
	}
	return users, nil
}

func (b *Bot) FindUsersByMails(mails []string) (users []models.User, err error) {
	resp, err := models.NewFindUsersByMailsRequest(mails).GetResponse(b.callApi)
	if err != nil {
		return
	}
	// Put found users to cache
	for _, u := range resp.Users {
		for _, email := range u.EMails {
			b.emailUserIdCache.Set(strings.ToLower(email), u.UserHUID)
		}
	}
	return resp.Users, err
}

func (b *Bot) CreateChatWithUser(user models.User) (chatId uuid.UUID, err error) {
	chatId, ok := b.userIdChatIdCache.Get(user.UserHUID)
	if !ok {
		var resp *models.CreateChatResponse
		resp, err = models.NewCreateUserChatRequest(user.UserHUID).GetResponse(b.callApi)
		if err != nil {
			return uuid.Nil, err
		}
		chatId = resp.ChatId
		// Put created chat to cache
		b.userIdChatIdCache.Set(user.UserHUID, chatId)
	}
	return chatId, err
}

func (b *Bot) SendMessageAsync(message *models.NDRequest) (syncId uuid.UUID, err error) {
	resp, err := message.GetResponse(b.callApi)
	if err != nil {
		return uuid.Nil, err
	}
	return resp.SyncId, err
}

func (b *Bot) SendMessageSync(message *models.NDRequest) (ncbr *models.NotificationCallbackRequest, err error) {
	if b.fiberApp == nil {
		return nil, errors.New("bot is not configured to handle callbacks")
	}
	syncId, err := b.SendMessageAsync(message)
	if err != nil {
		return nil, err
	}
	resp, err := b.ncbManager.awaitCallback(syncId)
	return resp, err
}
