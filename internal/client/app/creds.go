package app

import (
	"sync"
	"time"
)

const needAuthAfter = 20 * time.Minute

type Creds struct {
	username string
	password string
	token    string
	expiry   time.Time
	mutex    sync.Mutex
}

func (c *Creds) getUsername() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.username
}

func (c *Creds) setUsername(username string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.username = username
}

func (c *Creds) getPassword() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.password
}

func (c *Creds) getToken() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.token
}

func (c *Creds) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.password = ""
	c.token = ""
}

func (c *Creds) set(password, token string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.password = password
	c.token = token
	c.expiry = time.Now().Add(needAuthAfter)
}

func (c *Creds) isExpired() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return time.Now().After(c.expiry)
}
