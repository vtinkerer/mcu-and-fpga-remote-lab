package currentsession

import (
	"sync"
	"time"
)

type CurrentSession struct {
	isActive bool
	SessionEndTime time.Time
	Token string
}

var (
	instance *CurrentSession
	once sync.Once
)

func (c *CurrentSession) IsActive() bool {
	return c.isActive
}

func GetCurrentSession() *CurrentSession {
	once.Do(func() {
		instance = &CurrentSession{isActive: false}
	})
	return instance
}

// Returns true if the session was active before the reset
func (c *CurrentSession) Reset() bool {
	ret := true

	if !c.isActive {
		ret = false
	}
	c.isActive = false

	return ret
}

func (c *CurrentSession) ValidateToken(token string) bool {
	return c.isActive && c.Token == token && time.Now().Before(c.SessionEndTime)
}
// Returns true if the session was overwritten
func (c *CurrentSession) Set(token string, sessionEndTime time.Time) bool {

	ret := false

	if c.isActive {
		c.Reset()
		ret = true
	}

	c.isActive = true
	c.Token = token
	c.SessionEndTime = sessionEndTime

	return ret
}