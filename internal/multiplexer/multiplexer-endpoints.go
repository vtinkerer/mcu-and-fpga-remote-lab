package multiplexer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


func HandleSelectInputChannel(device *MultiplexerModule) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request struct {
			Multiplexer int `json:"multiplexer" binding:"required"`
			Channel     int `json:"channel" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := device.selectInputChannel(request.Multiplexer, request.Channel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Input channel selected"})
	}
}

func HandleGetInputChannel(device *MultiplexerModule) func(c *gin.Context) {
	return func(c *gin.Context) {
		channel1, err := device.getInputChannel(1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		channel2, err := device.getInputChannel(2)	
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"channel1": channel1, "channel2": channel2})
	}
}
