package potentiometer

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostPotentiometerSetResistanceRequest struct {
	Resistance float64 `json:"resistance"`
}
// Can be used throughout the session to try to set the resistance and get the real (actual) resistance in response
func HandlePotentiometerSetResistance(device *Potentiometer) func(c *gin.Context) {
	return func(c *gin.Context) {

		var req PostPotentiometerSetResistanceRequest
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		resistance, err := device.SetResistance(req.Resistance)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to set resistance"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"resistance": resistance})
	}
}

// Can be used at the very beginning of the session to get the initial resistance
func HandlePotentiometerGetResistance(device *Potentiometer) func(c *gin.Context) {
	return func(c *gin.Context) {
		resistance := device.GetResistance()
		c.JSON(http.StatusOK, gin.H{"resistance": resistance})
	}
}

