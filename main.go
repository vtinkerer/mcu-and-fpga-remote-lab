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

var outputPins = []int{0, 1, 2, 3}
var outputChannels = []int{0, 1}
var wavegenFunctions = []string{"sine", "rampup", "triangle", "pulse"}

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

	device, err := analogdiscovery.CreateDevice()
	if err != nil {
		log.Fatalf("Error creating Analog Discovery device: %v", err)
	}

	for _, outputPin := range outputPins {
		device.SetPinMode(outputPin, true)
	}

	r.POST("/api/firmware/fpga", handleFirmware(*cfg, true))
	r.POST("/api/firmware/mcu", handleFirmware(*cfg, false))
	r.POST("/api/write-pin", handleWritePin(device))
	r.POST("/api/wavegen/write-channel", handleWavegenEnableChannel(device))
	r.POST("/api/wavegen/write-function", handleWavegenFunctionSet(device))
	r.POST("/api/wavegen/write-amplitude", handleWavegenAmplitudeSet(device))
	r.POST("/api/wavegen/write-frequency", handleWavegenFrequencySet(device))
	//r.POST("/api/wavegen/duty-cycle")
	r.POST("/api/scope/get-scope-data", handleScopeGetData(device))
	r.POST("/api/wavegen/write-config", handleWavegenRun(device))

	r.Any("/api/stream", cam.ServeHTTP)
	log.Fatal(r.Run(":" + cfg.PORT))
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
		file, err := c.FormFile("file")
		if err != nil {
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

		fmt.Println("Firmware file uploaded:", file.Filename, " to ", fp, " for ", postfix)

		if isFPGA {
			fmt.Println("Flashing FPGA")
			device := fpga.CreateFPGA(cfg.TDI, cfg.TDO, cfg.TCK, cfg.TMS)
			err = device.Flash(fp)
		} else {
			fmt.Println("Flashing STM32")
			err = stm32flash.Flash(fp, cfg.RESET_PIN, cfg.BOOT0_PIN)
		}

		if err != nil {
			fmt.Println("Error flashing device:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Firmware flashed successfully"})
	}
}

type WritePinRequest struct {
	Pin   int `json:"pin"`
	State int `json:"state"`
}

type WriteWavegenAmplitudeRequest struct {
	Channel   int     `json:"channel"`
	Amplitude float64 `json:"amplitude"`
}

type WriteWavegenFunctionRequest struct {
	Channel  int    `json:"channel"`
	Function string `json:"function"`
}

type WriteWavegenFrequencyRequest struct {
	Channel   int     `json:"channel"`
	Frequency float64 `json:"frequency"`
}

type WriteWavegenChannelEnableRequest struct {
	Channel   int `json:"channel"`
	IsEnabled int `json:"isEnabled"`
}

type WriteWavegenRunRequest struct {
	Channel int `json:"channel"`
	IsStart int `json:"isStart"`
}

type GetScopeDataRequest struct {
	Channel        int `json:"channel"`
	IsFirstCapture int `json:"isFirstCapture"`
}

func isPinAllowed(pin int) bool {
	for _, allowedPin := range outputPins {
		if pin == allowedPin {
			return true
		}
	}
	return false
}

func isChannelAllowed(channel int) bool {
	for _, allowedChannel := range outputChannels {
		if channel == allowedChannel {
			return true
		}
	}
	return false
}

func isFunctionAllowed(function string) bool {
	for _, allowedFunction := range wavegenFunctions {
		if function == allowedFunction {
			return true
		}
	}
	return false
}

func handleScopeGetData(device *analogdiscovery.AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {

		var scopeReq GetScopeDataRequest
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&scopeReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		voltages, times, _ := device.ReadScopeValues(scopeReq.Channel, scopeReq.IsFirstCapture)
		if len(voltages) <= 0 || len(times) <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No scope values"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"channel": scopeReq.Channel, "voltages": voltages, "times": times})
	}
}

func handleWritePin(device *analogdiscovery.AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var pinReq WritePinRequest
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&pinReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !isPinAllowed(pinReq.Pin) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid pin, only %v are allowed", outputPins)})
			return
		}

		if pinReq.State != 0 && pinReq.State != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state, only 0 or 1 are allowed"})
			return
		}
		device.SetPinState(pinReq.Pin, pinReq.State == 1)

		c.JSON(http.StatusOK, gin.H{"message": "Pin set successfully"})
	}
}

func handleWavegenAmplitudeSet(device *analogdiscovery.AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var wavegenAmplitude WriteWavegenAmplitudeRequest
		decoder := json.NewDecoder(c.Request.Body)

		if err := decoder.Decode(&wavegenAmplitude); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !isChannelAllowed(wavegenAmplitude.Channel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid channel, only %v are allowed", outputChannels)})
			return
		}

		if wavegenAmplitude.Amplitude < -5.0 || wavegenAmplitude.Amplitude > 5.0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amplitude, only values between -5.5 V and 5.5 are allowed"})
			return
		}

		device.SetAnalogOutAmplitude(wavegenAmplitude.Channel, "AnalogOutNodeCarrier", wavegenAmplitude.Amplitude)

		c.JSON(http.StatusOK, gin.H{"message": "Analog out set successfully"})

	}
}

func handleWavegenFunctionSet(device *analogdiscovery.AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var wavegenFunction WriteWavegenFunctionRequest
		decoder := json.NewDecoder(c.Request.Body)

		if err := decoder.Decode(&wavegenFunction); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !isChannelAllowed(wavegenFunction.Channel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid channel, only %v are allowed", outputChannels)})
			return
		}

		if !isFunctionAllowed(wavegenFunction.Function) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid function, only %v are allowed", wavegenFunction)})
			return
		}

		device.SetAnalogOutNodeFunction(wavegenFunction.Channel, "AnalogOutNodeCarrier", wavegenFunction.Function)

		c.JSON(http.StatusOK, gin.H{"message": "Analog out function set successfully"})

	}
}

func handleWavegenFrequencySet(device *analogdiscovery.AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var wavegenFrequency WriteWavegenFrequencyRequest
		decoder := json.NewDecoder(c.Request.Body)

		if err := decoder.Decode(&wavegenFrequency); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !isChannelAllowed(wavegenFrequency.Channel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid channel, only %v are allowed", outputChannels)})
			return
		}

		if wavegenFrequency.Frequency <= 0 || wavegenFrequency.Frequency > 200000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid frequency, only values between 0 Hz and 200 000 Hz are allowed"})
			return
		}

		device.SetAnalogOutFrequency(wavegenFrequency.Channel, "AnalogOutNodeCarrier", wavegenFrequency.Frequency)

		c.JSON(http.StatusOK, gin.H{"message": "Analog out frequency set successfully"})

	}
}

func handleWavegenEnableChannel(device *analogdiscovery.AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var wavegenEnableChannel WriteWavegenChannelEnableRequest
		decoder := json.NewDecoder(c.Request.Body)

		if err := decoder.Decode(&wavegenEnableChannel); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !isChannelAllowed(wavegenEnableChannel.Channel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid channel, only %v are allowed", outputChannels)})
			return
		}

		device.EnableAnalogOutChannel(wavegenEnableChannel.Channel, "AnalogOutNodeCarrier", wavegenEnableChannel.IsEnabled)

		c.JSON(http.StatusOK, gin.H{"message": "Analog out channel enabled/disabled successfully"})

	}
}

func handleWavegenRun(device *analogdiscovery.AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var wavegenRun WriteWavegenRunRequest
		decoder := json.NewDecoder(c.Request.Body)

		if err := decoder.Decode(&wavegenRun); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !isChannelAllowed(wavegenRun.Channel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid channel, only %v are allowed", outputChannels)})
			return
		}

		device.GenerateWaveform(wavegenRun.Channel, "AnalogOutNodeCarrier", wavegenRun.IsStart)

		c.JSON(http.StatusOK, gin.H{"message": "Analog out channel generator started/stopped successfully"})

	}
}
