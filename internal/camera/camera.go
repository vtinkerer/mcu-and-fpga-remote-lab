package camera

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/blackjack/webcam"
	"github.com/gin-gonic/gin"
)

type WebcamServer struct {
	cam           *webcam.Webcam
	frame         []byte
	mutex         sync.Mutex
	clients       map[chan []byte]bool
	clientsMutex  sync.Mutex
	isStreaming   atomic.Bool
	streamingLock sync.Mutex
}

const desiredBufferCount uint32 = 4

func NewWebcamServer(devicePath string) (*WebcamServer, error) {
	cam, err := webcam.Open(devicePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open webcam: %v", err)
	}

	// webcam defaults to a very high buffer count (256). Keep this small for low-latency live preview.
	if err := cam.SetBufferCount(desiredBufferCount); err != nil {
		return nil, fmt.Errorf("failed to set buffer count: %v", err)
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

	_, _, _, err = cam.SetImageFormat(selectedFormat, 1920, 1080)
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

	if ws.isStreaming.Load() {
		return nil // Already streaming
	}

	err := ws.cam.StartStreaming()
	if err != nil {
		return fmt.Errorf("failed to start streaming: %v", err)
	}

	ws.isStreaming.Store(true)
	go ws.captureFrames()
	return nil
}

func (ws *WebcamServer) stopStreaming() {
	ws.streamingLock.Lock()
	defer ws.streamingLock.Unlock()

	if !ws.isStreaming.Load() {
		return // Not streaming
	}

	ws.cam.StopStreaming()
	ws.isStreaming.Store(false)
}

func (ws *WebcamServer) Close() {
	ws.stopStreaming()
	ws.cam.Close()
}

func (ws *WebcamServer) captureFrames() {
	for ws.isStreaming.Load() {
		err := ws.cam.WaitForFrame(uint32(5))
		if err != nil {
			log.Printf("Error waiting for frame: %v", err)
			continue
		}

		// Drain all available frames and keep only the most recent one.
		var latestFrame []byte
		for {
			frameData, err := ws.cam.ReadFrame()
			if err != nil {
				log.Printf("Error reading frame: %v", err)
				latestFrame = nil
				break
			}
			if len(frameData) == 0 {
				break
			}

			latestFrame = append(latestFrame[:0], frameData...)
		}

		if len(latestFrame) == 0 {
			continue
		}

		ws.mutex.Lock()
		ws.frame = latestFrame
		ws.mutex.Unlock()

		ws.clientsMutex.Lock()
		for clientChan := range ws.clients {
			// Keep only the latest frame per client to avoid accumulated lag.
			select {
			case clientChan <- latestFrame:
			default:
				select {
				case <-clientChan:
				default:
				}
				select {
				case clientChan <- latestFrame:
				default:
				}
			}
		}
		ws.clientsMutex.Unlock()
	}
}

func (ws *WebcamServer) ServeHTTP(c *gin.Context) {
	c.Header("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("X-Accel-Buffering", "no")

	if c.Request.Method == "OPTIONS" {
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// Buffer size 1 keeps memory low and prevents stale-frame backlog.
	clientChan := make(chan []byte, 1)

	ws.clientsMutex.Lock()
	ws.clients[clientChan] = true
	ws.clientsMutex.Unlock()

	defer func() {
		ws.clientsMutex.Lock()
		delete(ws.clients, clientChan)
		shouldStop := len(ws.clients) == 0
		ws.clientsMutex.Unlock()

		if shouldStop {
			ws.stopStreaming()
		}
		close(clientChan)
	}()

	if err := ws.StartStreaming(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to start camera stream"})
		return
	}

	// Send the last known frame immediately when available.
	ws.mutex.Lock()
	lastFrame := append([]byte(nil), ws.frame...)
	ws.mutex.Unlock()
	if len(lastFrame) > 0 {
		select {
		case clientChan <- lastFrame:
		default:
		}
	}

	for {
		select {
		case frameData, ok := <-clientChan:
			if !ok {
				return
			}
			if _, err := c.Writer.Write([]byte("--frame\r\nContent-Type: image/jpeg\r\n\r\n")); err != nil {
				return
			}
			if _, err := c.Writer.Write(frameData); err != nil {
				return
			}
			if _, err := c.Writer.Write([]byte("\r\n")); err != nil {
				return
			}
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
