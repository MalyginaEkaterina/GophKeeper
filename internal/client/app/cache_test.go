package app

import (
	"github.com/MalyginaEkaterina/GophKeeper/internal/client"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestCacheFlush(t *testing.T) {
	t.Run("Test flush into file after put", func(t *testing.T) {
		cacheFile := path.Join(os.TempDir(), "test_cache_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(cacheFile)
		putsFile := path.Join(os.TempDir(), "test_puts_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(putsFile)
		cfg := client.Config{CacheFilePath: cacheFile, PutsFilePath: putsFile}
		c := newCache(cfg)

		v := &pb.Value{Data: []byte("test data"), Metadata: "test metadata", Version: 1}
		const key = "key1"
		const pass = "test123"
		c.put(key, v)
		err := c.flushIntoFile(pass)
		require.NoError(t, err)
		require.False(t, c.cacheUpdated)

		c2 := newCache(cfg)
		err = c2.fillFromFile(pass)
		require.NoError(t, err)
		v2 := c2.getByKey(key)
		assert.NotNil(t, v2)
		assert.Equal(t, v.Data, v2.Data)
		assert.Equal(t, v.Metadata, v2.Metadata)
		assert.Equal(t, v.Version, v2.Version)
	})
	t.Run("Test flush into file after set cache", func(t *testing.T) {
		cacheFile := path.Join(os.TempDir(), "test_cache_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(cacheFile)
		putsFile := path.Join(os.TempDir(), "test_puts_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(putsFile)
		cfg := client.Config{CacheFilePath: cacheFile, PutsFilePath: putsFile}
		c := newCache(cfg)
		keys := []string{"key1", "key2"}
		values := map[string]*pb.Value{
			keys[0]: {Data: []byte("test data"), Metadata: "", Version: 1},
			keys[1]: {Data: []byte("test data 2"), Metadata: "", Version: 1},
		}
		c.setCache(values, 1)
		const pass = "test123"
		err := c.flushIntoFile(pass)
		require.NoError(t, err)
		require.False(t, c.cacheUpdated)

		c2 := newCache(cfg)
		err = c2.fillFromFile(pass)
		require.NoError(t, err)
		keys2 := c2.getKeys()
		assert.ElementsMatch(t, keys2, keys)
	})
}

func TestPutsFlush(t *testing.T) {
	t.Run("Test flush when puts updated", func(t *testing.T) {
		cacheFile := path.Join(os.TempDir(), "test_cache_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(cacheFile)
		putsFile := path.Join(os.TempDir(), "test_puts_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(putsFile)
		cfg := client.Config{CacheFilePath: cacheFile, PutsFilePath: putsFile}
		c := newCache(cfg)
		c.put("key", &pb.Value{Data: []byte("test data"), Metadata: "test metadata", Version: 1})

		pr := &pb.PutReq{Key: "key1", Value: &pb.Value{Data: []byte("test data 1"), Metadata: "", Version: 1}}
		c.appendPutRequest(pr)

		const pass = "test123"
		err := c.flushIntoFile(pass)
		require.NoError(t, err)
		require.False(t, c.putRequestsUpdated)

		c2 := newCache(cfg)
		err = c2.fillFromFile(pass)
		require.NoError(t, err)
		require.True(t, c2.hasPutRequests())
		pr2, is := c2.nextPutRequest()
		require.True(t, is)
		assert.Equal(t, pr.Key, pr2.Key)
		assert.Equal(t, pr.Value.Data, pr2.Value.Data)
	})
}

func TestPut(t *testing.T) {
	t.Run("Test next/pop put request", func(t *testing.T) {
		cacheFile := path.Join(os.TempDir(), "test_cache_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(cacheFile)
		putsFile := path.Join(os.TempDir(), "test_puts_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(putsFile)
		cfg := client.Config{CacheFilePath: cacheFile, PutsFilePath: putsFile}
		c := newCache(cfg)
		pr1 := &pb.PutReq{Key: "key1", Value: &pb.Value{Data: []byte("test data 1"), Metadata: "", Version: 1}}
		pr2 := &pb.PutReq{Key: "key2", Value: &pb.Value{Data: []byte("test data 2"), Metadata: "", Version: 1}}
		c.appendPutRequest(pr1)
		c.appendPutRequest(pr2)
		cpr1, ok := c.nextPutRequest()
		require.True(t, ok)
		assert.Equal(t, pr1.Key, cpr1.Key)
		assert.Equal(t, pr1.Value.Data, cpr1.Value.Data)
		c.popPutRequest()
		cpr2, ok := c.nextPutRequest()
		require.True(t, ok)
		assert.Equal(t, pr2.Key, cpr2.Key)
		assert.Equal(t, pr2.Value.Data, cpr2.Value.Data)
		c.popPutRequest()
		_, ok = c.nextPutRequest()
		require.False(t, ok)
	})
}

func TestCacheRead(t *testing.T) {
	t.Run("Test read with wrong password", func(t *testing.T) {
		cacheFile := path.Join(os.TempDir(), "test_cache_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(cacheFile)
		putsFile := path.Join(os.TempDir(), "test_puts_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(putsFile)
		cfg := client.Config{CacheFilePath: cacheFile, PutsFilePath: putsFile}
		c := newCache(cfg)
		v := &pb.Value{Data: []byte("test data"), Metadata: "test metadata", Version: 1}
		c.put("key", v)
		err := c.flushIntoFile("pass123")
		require.NoError(t, err)
		c2 := newCache(cfg)
		err = c2.fillFromFile("test123")
		assert.ErrorIs(t, err, errDecryption)
	})
	t.Run("Test read with not existed file", func(t *testing.T) {
		cacheFile := path.Join(os.TempDir(), "test_cache_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(cacheFile)
		putsFile := path.Join(os.TempDir(), "test_puts_file_name.data"+strconv.Itoa(rand.Int()))
		defer os.Remove(putsFile)
		cfg := client.Config{CacheFilePath: cacheFile, PutsFilePath: putsFile}
		c2 := newCache(cfg)
		err := c2.fillFromFile("pass123")
		require.ErrorIs(t, err, errNeedFirstLogin)
	})
}
