package storage

import (
	"context"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
	"sync"
)

var _ UserStorage = (*CachedStorage)(nil)
var _ DataStorage = (*CachedStorage)(nil)

type user struct {
	id         server.UserID
	hashedPass string
}

type userData struct {
	data    map[string]*pb.Value
	version int32
}

type CachedStorage struct {
	userCount int
	users     map[string]user
	userMutex sync.RWMutex

	data  map[server.UserID]userData
	mutex sync.RWMutex
}

func NewCachedStorage() *CachedStorage {
	return &CachedStorage{users: make(map[string]user), data: make(map[server.UserID]userData)}
}

func (c *CachedStorage) AddUser(_ context.Context, login string, hashedPass string) error {
	c.userMutex.Lock()
	defer c.userMutex.Unlock()
	if _, ok := c.users[login]; ok {
		return ErrAlreadyExists
	}
	c.userCount++
	c.users[login] = user{id: server.UserID(c.userCount), hashedPass: hashedPass}
	return nil
}

func (c *CachedStorage) GetUser(_ context.Context, login string) (server.UserID, string, error) {
	c.userMutex.RLock()
	defer c.userMutex.RUnlock()
	user, ok := c.users[login]
	if !ok {
		return 0, "", ErrNotFound
	}
	return user.id, user.hashedPass, nil
}

func (c *CachedStorage) GetByKey(_ context.Context, userID server.UserID, key string) (*pb.Value, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	usersData, ok := c.data[userID]
	if !ok {
		return nil, ErrNotFound
	}
	data, ok := usersData.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return data, nil
}

func (c *CachedStorage) Put(_ context.Context, userID server.UserID, key string, value *pb.Value) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	usersData, ok := c.data[userID]
	if !ok {
		c.data[userID] = userData{data: map[string]*pb.Value{key: value}, version: 1}
		return nil
	}
	data, ok := usersData.data[key]
	if ok {
		if data.Version >= value.Version {
			return ErrConflict
		}
	}
	usersData.data[key] = value
	c.data[userID] = userData{data: usersData.data, version: usersData.version + 1}
	return nil
}

func (c *CachedStorage) GetAllKeysByUser(_ context.Context, userID server.UserID) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	usersData, ok := c.data[userID]
	if !ok {
		return []string{}, nil
	}
	var keys []string
	for k := range usersData.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (c *CachedStorage) GetAllByUser(_ context.Context, userID server.UserID, version int32) (map[string]*pb.Value, int32, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	usersData, ok := c.data[userID]
	if !ok {
		return nil, 0, ErrNotFound
	}
	if usersData.version > version {
		return usersData.data, usersData.version, nil
	} else {
		return nil, 0, ErrNotFound
	}
}

func (c *CachedStorage) Close() {
}
