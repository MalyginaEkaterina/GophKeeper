package storage

import (
	"context"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
	"sync"
)

var _ UserStorage = (*CachedUserStorage)(nil)

type userData struct {
	id         server.UserID
	hashedPass string
}

type CachedUserStorage struct {
	userCount int
	users     map[string]userData
	mutex     sync.RWMutex
}

func NewCachedFileUserStorage() *CachedUserStorage {
	return &CachedUserStorage{users: make(map[string]userData)}
}

func (c *CachedUserStorage) AddUser(_ context.Context, login string, hashedPass string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.users[login]; ok {
		return ErrAlreadyExists
	}
	c.userCount++
	c.users[login] = userData{id: server.UserID(c.userCount), hashedPass: hashedPass}
	return nil
}

func (c *CachedUserStorage) GetUser(_ context.Context, login string) (server.UserID, string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	user, ok := c.users[login]
	if !ok {
		return 0, "", ErrNotFound
	}
	return user.id, user.hashedPass, nil
}

func (c *CachedUserStorage) Close() {
}
