package currentsession

import (
	"sync"
	"time"
)

type currentSession struct {
	isActive bool
	SessionEndTime time.Time
	Token string
}

var (
	instance *currentSession
	once sync.Once
)

func GetCurrentSession() *currentSession {
	once.Do(func() {
		instance = &currentSession{isActive: false}
	})
	return instance
}

// Returns true if the session was active before the reset
func (c *currentSession) Reset() bool {
	ret := true

	if !c.isActive {
		ret = false
	}
	c.isActive = false

	return ret
}

func (c *currentSession) ValidateToken(token string) bool {
	return c.isActive && c.Token == token && time.Now().Before(c.SessionEndTime)
}
// Returns true if the session was overwritten
func (c *currentSession) Set(token string, sessionEndTime time.Time) bool {

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