package stm32flash

import (
	"digitrans-lab-go/internal/config"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleSTM32Reset(cfg config.Config) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := Reset(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("Error resetting STM32: %v", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "STM32 has been reset"})
	}
}