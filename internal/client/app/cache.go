package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/MalyginaEkaterina/GophKeeper/internal/client"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"sync"
)

var (
	errDecryption = errors.New("file decryption error")
)

// Cache is struct for working with local cache.
// When flags cacheUpdated or/and putRequestsUpdated are true,
// flushIntoFile method saves encrypted protobuf data from Cache into files.
// Method fillFromFile reads files with encrypted protobuf data,
// tries to decipher it and if successes puts this data into Cache struct.
type Cache struct {
	cache              map[string]*pb.Value
	version            int32
	cacheUpdated       bool
	putRequests        []*pb.PutReq
	putRequestsUpdated bool
	config             client.Config
	mutex              sync.Mutex
}

func newCache(cfg client.Config) *Cache {
	return &Cache{cache: make(map[string]*pb.Value), config: cfg}
}

func (c *Cache) put(key string, value *pb.Value) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache[key] = value
	c.cacheUpdated = true
}

func (c *Cache) setCache(newCache map[string]*pb.Value, version int32) {
	if newCache != nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.cache = newCache
		c.version = version
		c.cacheUpdated = true
	}
}

func (c *Cache) getVersion() int32 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.version
}

func (c *Cache) getByKey(key string) *pb.Value {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.cache[key]
}

func (c *Cache) getKeys() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var keys []string
	for k := range c.cache {
		keys = append(keys, k)
	}
	return keys
}

func (c *Cache) getNextVersionForKey(key string) int32 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if v, ok := c.cache[key]; ok {
		return v.Version + 1
	}
	return 1
}

func (c *Cache) appendPutRequest(req *pb.PutReq) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.putRequests = append(c.putRequests, req)
	c.putRequestsUpdated = true
}

func (c *Cache) setPutRequests(puts []*pb.PutReq) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.putRequests = puts
	c.putRequestsUpdated = true
}

func (c *Cache) hasPutRequests() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.putRequests) > 0
}

func (c *Cache) nextPutRequest() (*pb.PutReq, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.putRequests) == 0 {
		return nil, false
	}
	return c.putRequests[0], true
}

func (c *Cache) popPutRequest() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.putRequests = c.putRequests[1:]
	c.putRequestsUpdated = true
}

func (c *Cache) flushIntoFile(password string) error {
	aesGcm, err := getAesGcm(password)
	if err != nil {
		return fmt.Errorf("get aes gcm error: %w", err)
	}
	nonce := make([]byte, aesGcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return fmt.Errorf("get nonce error: %w", err)
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.cacheUpdated {
		cache := &pb.Cache{Data: c.cache}
		binCache, err := proto.Marshal(cache)
		if err != nil {
			return fmt.Errorf("write cache error: %w", err)
		}
		cipherCache := aesGcm.Seal(nil, nonce, binCache, nil)
		err = writeIntoFile(c.config.CacheFilePath, nonce, cipherCache)
		if err != nil {
			return fmt.Errorf("write cache error: %w", err)
		}
		c.cacheUpdated = false
	}
	if c.putRequestsUpdated {
		putReqs := &pb.PutRequests{List: c.putRequests}
		binPuts, err := proto.Marshal(putReqs)
		if err != nil {
			return fmt.Errorf("write put requests error: %w", err)
		}
		cipherPuts := aesGcm.Seal(nil, nonce, binPuts, nil)
		err = writeIntoFile(c.config.PutsFilePath, nonce, cipherPuts)
		if err != nil {
			return fmt.Errorf("write put requests error: %w", err)
		}
		c.putRequestsUpdated = false
	}
	return nil
}

func writeIntoFile(filePath string, nonce []byte, data []byte) error {
	tmpPath := filePath + ".tmp"
	dataWithNonce := append(nonce, data...)
	err := os.WriteFile(tmpPath, dataWithNonce, 0777)
	if err != nil {
		return err
	}
	err = os.Rename(tmpPath, filePath)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) writeUsernameIntoFile(username string) error {
	tmpPath := c.config.LoggedUserFilePath + ".tmp"
	err := os.WriteFile(tmpPath, []byte(username), 0777)
	if err != nil {
		return err
	}
	err = os.Rename(tmpPath, c.config.LoggedUserFilePath)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) readUsernameFromFile() (string, error) {
	data, err := os.ReadFile(c.config.LoggedUserFilePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readFromFile(filePath string, nonceSize int) ([]byte, []byte, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	if len(fileData) < nonceSize {
		return nil, nil, errors.New("file is truncated")
	}
	return fileData[:nonceSize], fileData[nonceSize:], nil
}

func getAesGcm(password string) (cipher.AEAD, error) {
	h := sha256.New()
	h.Write([]byte(password))
	key := h.Sum(nil)
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(aesBlock)
}

func (c *Cache) fillFromFile(password string) error {
	aesGcm, err := getAesGcm(password)
	if err != nil {
		return fmt.Errorf("get aes gcm error: %w", err)
	}
	nonce, cipherCache, err := readFromFile(c.config.CacheFilePath, aesGcm.NonceSize())
	if errors.Is(err, os.ErrNotExist) {
		return errNeedFirstLogin
	} else if err != nil {
		return fmt.Errorf("read cache error: %w", err)
	}
	binCache, err := aesGcm.Open(nil, nonce, cipherCache, nil)
	if err != nil {
		return errDecryption
	}
	var cache pb.Cache
	err = proto.Unmarshal(binCache, &cache)
	if err != nil {
		return fmt.Errorf("read cache  error: %w", err)
	}
	c.setCache(cache.Data, 0)

	nonce, cipherPuts, err := readFromFile(c.config.PutsFilePath, aesGcm.NonceSize())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("read put requests error: %w", err)
	}
	binPuts, err := aesGcm.Open(nil, nonce, cipherPuts, nil)
	if err != nil {
		return errDecryption
	}
	var puts pb.PutRequests
	err = proto.Unmarshal(binPuts, &puts)
	if err != nil {
		return fmt.Errorf("read put requests error: %w", err)
	}
	c.setPutRequests(puts.List)
	return nil
}
