package main

import (
	analogdiscovery "digitrans-lab-go/internal/analog-discovery"
	"digitrans-lab-go/internal/camera"
	"digitrans-lab-go/internal/config"
	stm32flash "digitrans-lab-go/internal/stm32-flash"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
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

	http.HandleFunc("/firmware", handleFirmware(cfg.BOOT0_PIN, cfg.RESET_PIN))
	http.HandleFunc("/write-pin", handleWritePin(*device))
	http.Handle("/stream", cam)

	log.Printf("Server started on http://localhost:%s", cfg.PORT)
	log.Fatal(http.ListenAndServe(":"+cfg.PORT, nil))
}

const (
	maxUploadSize = 10 * (10 << 20) // 100 MB
	uploadPath    = "./uploads"
)

func handleFirmware(boot0pin, resetpin int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Limit the size of the request body
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}

		// Get the file from the request
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()


		fp := filepath.Join(uploadPath, header.Filename)
		// Create the file on the server
		dst, err := os.Create(fp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the created file on the server
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := stm32flash.Flash(fp, resetpin, boot0pin); err != nil {
			fmt.Println("Error flashing STM32:", err)
			http.Error(w, "Error flashing STM32", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Firmware flashed successfully")
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
func handleWritePin(device analogdiscovery.AnalogDiscoveryDevice) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var pinReq WritePinRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&pinReq); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if !isPinAllowed(pinReq.Pin) {
			http.Error(w, fmt.Sprintf("Invalid pin number, only the following are allowed: ", allowedPins) , http.StatusBadRequest)
			return
		}

		if (pinReq.State != 0 && pinReq.State != 1) {
			http.Error(w, "Invalid state, only 0 or 1 are allowed", http.StatusBadRequest)
			return
		}

		device.SetPinState(pinReq.Pin, pinReq.State == 1)

		w.WriteHeader(http.StatusOK)
	}
}