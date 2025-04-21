package potentiometer

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostPotentiometerSetResistanceRequest struct {
	Percentage int `json:"percentage"`
}
// Can be used throughout the session to try to set the resistance and get the real (actual) resistance in response
func HandlePotentiometerSetResistancePercentage(device *Potentiometer) func(c *gin.Context) {
	return func(c *gin.Context) {

		var req PostPotentiometerSetResistanceRequest
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		percentage, err := device.SetResistancePercentage(req.Percentage)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to set resistance"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"percentage": percentage})
	}
}

// Can be used at the very beginning of the session to get the initial resistance
func HandlePotentiometerGetResistancePercentage(device *Potentiometer) func(c *gin.Context) {
	return func(c *gin.Context) {
		percentage := device.GetResistancePercentage()
		c.JSON(http.StatusOK, gin.H{"percentage": percentage})
	}
}

