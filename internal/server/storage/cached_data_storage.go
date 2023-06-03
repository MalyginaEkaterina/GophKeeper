package storage

import (
	"context"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/MalyginaEkaterina/GophKeeper/internal/server"
	"sync"
)

var _ DataStorage = (*CachedDataStorage)(nil)

type CachedDataStorage struct {
	data  map[server.UserID]map[string]*pb.Value
	mutex sync.RWMutex
}

func NewCachedFileDataStorage() *CachedDataStorage {
	return &CachedDataStorage{data: make(map[server.UserID]map[string]*pb.Value)}
}

func (c *CachedDataStorage) GetByKey(_ context.Context, userID server.UserID, key string) (*pb.Value, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	usersData, ok := c.data[userID]
	if !ok {
		return nil, ErrNotFound
	}
	data, ok := usersData[key]
	if !ok {
		return nil, ErrNotFound
	}
	return data, nil
}

func (c *CachedDataStorage) Put(_ context.Context, userID server.UserID, key string, value *pb.Value) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	usersData, ok := c.data[userID]
	if !ok {
		c.data[userID] = map[string]*pb.Value{key: value}
		return nil
	}
	data, ok := usersData[key]
	if ok {
		if data.Version >= value.Version {
			return ErrConflict
		}
	}
	usersData[key] = value
	c.data[userID] = usersData
	return nil
}

func (c *CachedDataStorage) GetAllKeysByUser(_ context.Context, userID server.UserID) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	usersData, ok := c.data[userID]
	if !ok {
		return []string{}, nil
	}
	var keys []string
	for k := range usersData {
		keys = append(keys, k)
	}
	return keys, nil
}

func (c *CachedDataStorage) GetAllByUser(_ context.Context, userID server.UserID) (map[string]*pb.Value, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	usersData, ok := c.data[userID]
	if !ok {
		return map[string]*pb.Value{}, nil
	}
	return usersData, nil
}
