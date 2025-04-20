package bot

import (
	"errors"
	"sync"
	"time"

	"github.com/go-botx/bot/models"
	"github.com/google/uuid"
)

type ncbmStoredMessage struct {
	message   *models.NotificationCallbackRequest
	expiresAt int64 // Unix timestamp in milliseconds

}

type ncbManager struct {
	storeTime   int64 // Diration in milliseconds
	mutex       sync.RWMutex
	waiters     map[uuid.UUID]chan *models.NotificationCallbackRequest
	messages    map[uuid.UUID]ncbmStoredMessage
	closeChan   chan struct{}
	cleanupDone chan struct{}
}

func newNCBManager(storeTime time.Duration) *ncbManager {
	ncbm := &ncbManager{
		storeTime:   storeTime.Milliseconds(),
		waiters:     make(map[uuid.UUID]chan *models.NotificationCallbackRequest),
		messages:    make(map[uuid.UUID]ncbmStoredMessage),
		closeChan:   make(chan struct{}),
		cleanupDone: make(chan struct{}),
	}
	go ncbm.periodicCleanup()
	return ncbm
}

func (ncbm *ncbManager) stop() {
	close(ncbm.closeChan)
	<-ncbm.cleanupDone
}

func (ncbm *ncbManager) periodicCleanup() {
	ticker := time.NewTicker(time.Duration(ncbm.storeTime/2) * time.Millisecond)
	defer ticker.Stop()
	defer close(ncbm.cleanupDone)

	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixMilli()
			ncbm.mutex.Lock()
			for syncID, msg := range ncbm.messages {
				if now > msg.expiresAt {
					delete(ncbm.messages, syncID)
				}
			}
			ncbm.mutex.Unlock()
		case <-ncbm.closeChan:
			return
		}
	}
}

func (ncbm *ncbManager) storeCallback(message models.NotificationCallbackRequest) {
	ncbm.mutex.Lock()
	defer ncbm.mutex.Unlock()

	if ch, exists := ncbm.waiters[message.SyncId]; exists {
		ch <- &message
		delete(ncbm.waiters, message.SyncId)
		return
	}

	ncbm.messages[message.SyncId] = ncbmStoredMessage{
		message:   &message,
		expiresAt: time.Now().UnixMilli() + ncbm.storeTime,
	}
}

func (ncbm *ncbManager) awaitCallback(syncId uuid.UUID) (*models.NotificationCallbackRequest, error) {
	ncbm.mutex.Lock()

	if storedMsg, exists := ncbm.messages[syncId]; exists {
		delete(ncbm.messages, syncId)
		ncbm.mutex.Unlock()
		return storedMsg.message, storedMsg.message.GetError()
	}

	ch := make(chan *models.NotificationCallbackRequest, 1)
	ncbm.waiters[syncId] = ch
	ncbm.mutex.Unlock()

	select {
	case msg := <-ch:
		return msg, msg.GetError()
	case <-time.After(time.Duration(ncbm.storeTime) * time.Millisecond):
		ncbm.mutex.Lock()
		delete(ncbm.waiters, syncId)
		ncbm.mutex.Unlock()
		return nil, errors.New("timeout waiting for callback")
	}
}
