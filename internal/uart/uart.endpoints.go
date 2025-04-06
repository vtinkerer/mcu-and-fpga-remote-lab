package uart

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PostUartChangeSpeedRequest struct {
	Speed int `json:"speed"`
}
// Can be used throughout the session to try to set the resistance and get the real (actual) resistance in response
func HandleUartChangeSpeed(u *UART) func(c *gin.Context) {
	return func(c *gin.Context) {

		var req PostUartChangeSpeedRequest
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err := u.ChangeSpeed(req.Speed)
		if err != nil {
			fmt.Println("error changing speed: ", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to change speed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"speed": req.Speed})
	}
}


