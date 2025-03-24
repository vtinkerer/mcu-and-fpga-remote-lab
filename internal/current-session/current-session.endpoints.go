package currentsession

import (
	"digitrans-lab-go/internal/config"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func validateAuthorization(c *gin.Context, cfg config.Config) bool {
	return c.Request.Header.Get("Authorization") == cfg.MASTER_SERVER_API_SECRET
}

type CreateSessionRequest struct {
	Token string `json:"token"`
	SessionEndTime string `json:"sessionEndTime"`
}
func HandleCreateSession(cfg config.Config, createdCb func(), overwrittenCb func()) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request CreateSessionRequest
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !validateAuthorization(c, cfg) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		parsedTime, err := time.Parse(time.RFC3339, request.SessionEndTime);
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}

		session := GetCurrentSession()
		isOverwritten := session.Set(request.Token, parsedTime)
		createdCb()

		if isOverwritten {
			overwrittenCb()
			c.JSON(http.StatusCreated, gin.H{"message": "Session overwritten"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Successfully created"})
	}
}

func HandleGetSession(cfg config.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		if !validateAuthorization(c, cfg) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		session := GetCurrentSession()
		var res gin.H
		if !session.isActive {
			res = gin.H{"token": nil, "sessionEndTime": nil}
		} else {
			res = gin.H{"token": session.Token, "sessionEndTime": session.SessionEndTime}
		}
		c.JSON(http.StatusOK, res)
	}
}

func HandleDeleteSession(cfg config.Config, cb func()) func(c *gin.Context) {
	return func(c *gin.Context) {
		if !validateAuthorization(c, cfg) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		session := GetCurrentSession()
		isReallyReset := session.Reset()

		cb()

		if isReallyReset {
			c.JSON(http.StatusOK, gin.H{"message": "Session was reset"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Session has already been not active"})
	}
}

func UserGetSessionMetadata() func(c *gin.Context) {
	return func(c *gin.Context) {
		session := GetCurrentSession()
		c.JSON(http.StatusOK, gin.H{"sessionEndTime": session.SessionEndTime})
	}
}