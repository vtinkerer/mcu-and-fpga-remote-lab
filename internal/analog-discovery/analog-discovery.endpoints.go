package analogdiscovery

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var OutputPins = []int{0, 1, 2, 3}
var outputChannels = []int{0, 1}
var wavegenFunctions = []string{"sine", "rampup", "triangle", "pulse"}

// check if given pin is allowed
func isPinAllowed(pin int) bool {
	for _, allowedPin := range OutputPins {
		if pin == allowedPin {
			return true
		}
	}
	return false
}

// check if given channel is allowed
func isChannelAllowed(channel int) bool {
	for _, allowedChannel := range outputChannels {
		if channel == allowedChannel {
			return true
		}
	}
	return false
}

// check if given function is allowed
func isFunctionAllowed(function string) bool {
	for _, allowedFunction := range wavegenFunctions {
		if function == allowedFunction {
			return true
		}
	}
	return false
}

// handler for oscilloscope feature
type GetScopeDataRequest struct {
	Channel        int `json:"channel"`
	IsFirstCapture int `json:"isFirstCapture"`
}
func HandleScopeGetData(device *AnalogDiscoveryDevice) func(c *gin.Context) {
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

		if scopeReq.IsFirstCapture == 0 {
			c.JSON(http.StatusOK, gin.H{"channel": scopeReq.Channel, "voltages": voltages, "times": 0})
		} else {
			c.JSON(http.StatusOK, gin.H{"channel": scopeReq.Channel, "voltages": voltages, "times": times})
		}

	}
}

// handler for digital inputs pin set
type WritePinRequest struct {
	Pin   int `json:"pin"`
	State int `json:"state"`
}
func HandleWritePin(device *AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var pinReq WritePinRequest
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(&pinReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Frontend sends 1, 2, 3, 4, but we need to convert it to 0, 1, 2, 3
		pin := pinReq.Pin - 1

		if !isPinAllowed(pin) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid pin, only %v are allowed", OutputPins)})
			return
		}

		if pinReq.State != 0 && pinReq.State != 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state, only 0 or 1 are allowed"})
			return
		}
		device.SetPinState(pin, pinReq.State == 1)

		c.JSON(http.StatusOK, gin.H{"message": "Pin set successfully"})
	}
}

// handler for setting wavegen amplitude
type WriteWavegenAmplitudeRequest struct {
	Channel   int     `json:"channel"`
	Amplitude float64 `json:"amplitude"`
}
func HandleWavegenAmplitudeSet(device *AnalogDiscoveryDevice) func(c *gin.Context) {
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

// handler for setting wavegen duty cycle (symmetry)
type WriteWavegenDutyCycleRequest struct {
	Channel   int     `json:"channel"`
	DutyCycle float64 `json:"dutyCycle"`
}
func HandleWavegenDutyCycleSet(device *AnalogDiscoveryDevice) func(c *gin.Context) {
	return func(c *gin.Context) {
		var wavegenDutyCycle WriteWavegenDutyCycleRequest
		decoder := json.NewDecoder(c.Request.Body)

		if err := decoder.Decode(&wavegenDutyCycle); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if !isChannelAllowed(wavegenDutyCycle.Channel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid channel, only %v are allowed", outputChannels)})
			return
		}

		if wavegenDutyCycle.DutyCycle < 0.0 || wavegenDutyCycle.DutyCycle > 100.0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duty cycle, only values between 0% and 100% are allowed"})
			return
		}

		device.SetAnalogOutSymmetry(wavegenDutyCycle.Channel, "AnalogOutNodeCarrier", wavegenDutyCycle.DutyCycle)

		c.JSON(http.StatusOK, gin.H{"message": "Analog out set successfully"})

	}
}

// handler for setting wavegen function type
type WriteWavegenFunctionRequest struct {
	Channel  int    `json:"channel"`
	Function string `json:"function"`
}
func HandleWavegenFunctionSet(device *AnalogDiscoveryDevice) func(c *gin.Context) {
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

// handler for setting wavegen frequency
type WriteWavegenFrequencyRequest struct {
	Channel   int     `json:"channel"`
	Frequency float64 `json:"frequency"`
}
func HandleWavegenFrequencySet(device *AnalogDiscoveryDevice) func(c *gin.Context) {
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

// handler for enabling wavegen channel
type WriteWavegenChannelEnableRequest struct {
	Channel   int `json:"channel"`
	IsEnabled int `json:"isEnabled"`
}
func HandleWavegenEnableChannel(device *AnalogDiscoveryDevice) func(c *gin.Context) {
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

// handler for running waveform generator
type WriteWavegenRunRequest struct {
	Channel int `json:"channel"`
	IsStart int `json:"isStart"`
}
func HandleWavegenRun(device *AnalogDiscoveryDevice) func(c *gin.Context) {
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
