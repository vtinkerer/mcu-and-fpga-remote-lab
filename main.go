package main

import (
	analogdiscovery "digitrans-lab-go/internal/analog-discovery"
	"digitrans-lab-go/internal/camera"
	"digitrans-lab-go/internal/config"
	"digitrans-lab-go/internal/fpga"
	stm32flash "digitrans-lab-go/internal/stm32-flash"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config:", err)
	}

	cam, err := camera.NewWebcamServer("/dev/video0")
	if err != nil {
		log.Fatal(err)
	}
	defer cam.Close()

	err = cam.StartStreaming()
	if err != nil {
		log.Fatal(err)
	}

	device, err := analogdiscovery.CreateDevice()
	if err != nil {
		log.Fatalf("Error creating Analog Discovery device: %v", err)
	}

	device.SetPinMode(0, true)
	device.SetPinMode(1, true)
	device.SetPinMode(2, true)

	r.POST("/api/firmware/fpga", handleFirmware(*cfg, true))
	r.POST("/api/firmware/mcu", handleFirmware(*cfg, false))
	r.POST("/api/write-pin", handleWritePin(*device))
	r.Any("/api/stream", cam.ServeHTTP)

	log.Printf("Server started on http://localhost:%s", cfg.PORT)
	log.Fatal(http.ListenAndServe(":"+cfg.PORT, nil))
}

const (
	maxUploadSize = 10 * (10 << 20) // 100 MB
	uploadPath    = "./uploads"
)

func handleFirmware(cfg config.Config, isFPGA bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Limit the size of the request body
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

		if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
			return
		}

		// Get the file from the request
		file, err := c.FormFile("file"); if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
				
		postfix := "MCU"
		if isFPGA {
			postfix = "FPGA"
		}
		fp := filepath.Join(uploadPath, postfix)

		// Create the file on the server
		c.SaveUploadedFile(file, fp)

		if isFPGA {
			fmt.Println("Flashing FPGA")
			device := fpga.CreateFPGA(cfg.TDI, cfg.TDO, cfg.TCK, cfg.TMS)
			err = device.Flash(fp)
		} else {
			fmt.Println("Flashing STM32")
			err = stm32flash.Flash(fp, cfg.RESET_PIN, cfg.BOOT0_PIN);
		}

		if err != nil {
			fmt.Println("Error flashing device:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error flashing device"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Firmware flashed successfully"})
	}
}

type WritePinRequest struct {
	Pin int `json:"pin"`
	State int `json:"state"`
}
var allowedPins = []int{0, 1, 2}
func isPinAllowed(pin int) bool {
	for _, allowedPin := range allowedPins {
		if pin == allowedPin {
			return true
		}
	}
	return false
}
func handleWritePin(device analogdiscovery.AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var pinReq WritePinRequest
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&pinReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !isPinAllowed(pinReq.Pin) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid pin, only %v are allowed", allowedPins)})
			return
		}

		if (pinReq.State != 0 && pinReq.State != 1) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state, only 0 or 1 are allowed"})
			return
		}

		device.SetPinState(pinReq.Pin, pinReq.State == 1)

		c.JSON(http.StatusOK, gin.H{"message": "Pin state set successfully"})
	}
}