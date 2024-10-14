package camera

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/blackjack/webcam"
)

type WebcamServer struct {
	cam           *webcam.Webcam
	frame         []byte
	mutex         sync.Mutex
	clients       map[chan []byte]bool
	clientsMutex  sync.Mutex
	isStreaming   bool
	streamingLock sync.Mutex
}

func NewWebcamServer(devicePath string) (*WebcamServer, error) {
	cam, err := webcam.Open(devicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open webcam: %v", err)
	}

	formats := cam.GetSupportedFormats()
	var selectedFormat webcam.PixelFormat
	for f, s := range formats {
		fmt.Printf("Supported format: %s\n", s)
		if s == "Motion-JPEG" {
			selectedFormat = f
			break
		}
	}

	if selectedFormat == 0 {
		return nil, fmt.Errorf("Motion-JPEG format not supported")
	}

	_, _, _, err = cam.SetImageFormat(selectedFormat, 640, 480)
	if err != nil {
		return nil, fmt.Errorf("failed to set image format: %v", err)
	}

	return &WebcamServer{
		cam:     cam,
		clients: make(map[chan []byte]bool),
	}, nil
}

func (ws *WebcamServer) StartStreaming() error {
	ws.streamingLock.Lock()
	defer ws.streamingLock.Unlock()

	if ws.isStreaming {
		return nil // Already streaming
	}

	err := ws.cam.StartStreaming()
	if err != nil {
		return fmt.Errorf("failed to start streaming: %v", err)
	}

	ws.isStreaming = true
	go ws.captureFrames()
	return nil
}

func (ws *WebcamServer) StopStreaming() {
	ws.streamingLock.Lock()
	defer ws.streamingLock.Unlock()

	if !ws.isStreaming {
		return // Not streaming
	}

	ws.cam.StopStreaming()
	ws.isStreaming = false
}

func (ws *WebcamServer) Close() {
	ws.StopStreaming()
	ws.cam.Close()
}

func (ws *WebcamServer) captureFrames() {
	for ws.isStreaming {
		err := ws.cam.WaitForFrame(uint32(5))
		if err != nil {
			log.Printf("Error waiting for frame: %v", err)
			continue
		}

		frameData, err := ws.cam.ReadFrame()
		if err != nil {
			log.Printf("Error reading frame: %v", err)
			continue
		}

		ws.mutex.Lock()
		ws.frame = frameData
		ws.mutex.Unlock()

		ws.clientsMutex.Lock()
		for clientChan := range ws.clients {
			select {
			case clientChan <- frameData:
			default:
				// If the client is not ready, skip this frame for them
			}
		}
		ws.clientsMutex.Unlock()
	}
}

func (ws *WebcamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if r.Method == "OPTIONS" {
		return
	}

	clientChan := make(chan []byte, 10)
	ws.clientsMutex.Lock()
	ws.clients[clientChan] = true
	ws.clientsMutex.Unlock()

	defer func() {
		ws.clientsMutex.Lock()
		delete(ws.clients, clientChan)
		ws.clientsMutex.Unlock()
		close(clientChan)
	}()

	for {
		select {
		case frameData := <-clientChan:
			w.Write([]byte("--frame\r\nContent-Type: image/jpeg\r\n\r\n"))
			w.Write(frameData)
			w.Write([]byte("\r\n"))
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}