package main

import (
	"context"
	analogdiscovery "digitrans-lab-go/internal/analog-discovery"
	"digitrans-lab-go/internal/camera"
	"digitrans-lab-go/internal/config"
	currentsession "digitrans-lab-go/internal/current-session"
	"digitrans-lab-go/internal/fpga"
	stm32flash "digitrans-lab-go/internal/stm32-flash"
	"digitrans-lab-go/internal/timer"
	"digitrans-lab-go/internal/uart"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
    u        	*uart.UART
    wsUpgrader  websocket.Upgrader
    wsConn 		*websocket.Conn
    wsConnMu   	sync.Mutex
	timer 		*timer.Timer
	deviceType  string
}

func NewServer() *Server {
	u := uart.NewUART()
	u.Open()
    return &Server{
        u: u,
        wsUpgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // Adjust based on your security needs
            },
        },
		wsConn: nil,
		timer: timer.NewTimer(10 * time.Second, func() {}),
    }
}

func main() {
	r := gin.Default()

	server := NewServer()

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

	for _, outputPin := range analogdiscovery.OutputPins {
		device.SetPinMode(outputPin, true)
	}

	clientAuthQueryRoutes := r.Group("")
	{
		clientAuthQueryRoutes.Use(ClientAuthQueryMiddleware())

		clientAuthQueryRoutes.Any("/api/stream", cam.ServeHTTP)
	}

	clientAuthRoutes := r.Group("")
	{
		clientAuthRoutes.Use(ClientAuthMiddleware())

		clientAuthRoutes.POST("/api/firmware/fpga", handleFirmware(*cfg, true, server))
		clientAuthRoutes.POST("/api/firmware/mcu", handleFirmware(*cfg, false, server))
		clientAuthRoutes.POST("/api/write-pin", analogdiscovery.HandleWritePin(device))
		clientAuthRoutes.POST("/api/wavegen/write-channel", analogdiscovery.HandleWavegenEnableChannel(device))
		clientAuthRoutes.POST("/api/wavegen/write-function", analogdiscovery.HandleWavegenFunctionSet(device))
		clientAuthRoutes.POST("/api/wavegen/write-amplitude", analogdiscovery.HandleWavegenAmplitudeSet(device))
		clientAuthRoutes.POST("/api/wavegen/write-frequency", analogdiscovery.HandleWavegenFrequencySet(device))
		clientAuthRoutes.POST("/api/wavegen/write-duty-cycle", analogdiscovery.HandleWavegenDutyCycleSet(device))
		clientAuthRoutes.POST("/api/scope/get-scope-data", analogdiscovery.HandleScopeGetData(device))
		clientAuthRoutes.POST("/api/wavegen/write-config", analogdiscovery.HandleWavegenRun(device))
		clientAuthRoutes.GET("/api/my-session", func(c *gin.Context) {
			cs := currentsession.GetCurrentSession()
			c.JSON(http.StatusOK, gin.H{"sessionEndTime": cs.SessionEndTime, "deviceType": server.deviceType})
		})
		clientAuthRoutes.GET("/ws", func (c *gin.Context) {
			server.handleWebSocket(c.Writer, c.Request)
		})
	}

	backendAuthRoutes := r.Group("")
	{
		backendAuthRoutes.Use(BackendAuthMiddleware(*cfg))

		backendAuthRoutes.POST("/api/session", currentsession.HandleCreateSession(*cfg, func() {
			secondsRemaining := currentsession.GetCurrentSession().SessionEndTime.Sub(time.Now()).Seconds()
			fmt.Println("Session created, starting timer for ", secondsRemaining, " seconds")
			server.timer.SetDuration(time.Duration(secondsRemaining) * time.Second)
			server.timer.Start(server.diconnectWebSocket)
		}, func() {
			server.diconnectWebSocket()
		}))
		backendAuthRoutes.GET("/api/session", currentsession.HandleGetSession(*cfg))
		backendAuthRoutes.DELETE("/api/session", currentsession.HandleDeleteSession(*cfg, func() {
			server.timer.Stop()
			server.diconnectWebSocket()
		}))
	}

	// Check which device is connected:
	err = server.CheckDeviceType(cfg)

	if (err == nil) {
		log.Println("Found device: ", server.deviceType)
		log.Fatal(r.Run(":" + cfg.PORT))
	} else {
		log.Fatal("Couldn't detect any device connected: ", err)
	}
}

func (s *Server) CheckDeviceType(cfg *config.Config) error {
	err := flashMCU(*cfg, filepath.Join("./example-firmware", "mcu.hex"), s)
	if (err == nil) {
		s.deviceType = "mcu"
		return nil
	}

	err = flashFPGA(*cfg, filepath.Join("./example-firmware", "fpga.svg"))
	if (err == nil) {
		s.deviceType = "fpga"
		return nil
	}

	return fmt.Errorf("No device detected: %w", err)
}

func ClientAuthQueryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if (!currentsession.GetCurrentSession().ValidateTokenHttpQuery(c)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

func ClientAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if (!currentsession.GetCurrentSession().ValidateTokenHttpHeader(c)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

func BackendAuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if (c.Request.Header.Get("Authorization") != cfg.MASTER_SERVER_API_SECRET) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

func (s *Server) diconnectWebSocket() {
	message := WsMessage{
		Type: "disconnect",
	}
	json, err := json.Marshal(message)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}
	if s.wsConn != nil {
		s.wsConn.WriteMessage(websocket.TextMessage, json)
		s.wsConn.Close()
	}
}

type WsMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
// for FPGA and MCU program files
const (
	maxUploadSize = 10 * (10 << 20) // 100 MB
	uploadPath    = "./uploads"
)
func (s *Server) handleWSToUART(conn *websocket.Conn) {
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Printf("WebSocket read error: %v", err)
            return
        }

		var wsMessage WsMessage
		err = json.Unmarshal(message, &wsMessage)
		if err != nil {
			log.Printf("Incoming from WS JSON unmarshal error: %v", err)
		}

        // Forward message to UART
        if err := s.u.Write([]byte(wsMessage.Text)); err != nil {
            log.Printf("UART write error: %v", err)
        }
    }
}
func (s *Server) handleUARTToWS(conn *websocket.Conn, ctx context.Context) {
    buffer := make([]byte, 1024)
	logCounter := 50
    for {
		select {
		case <-ctx.Done():
			fmt.Println("Exit UART-to-WS loop because of context cancellation")
			return
		default:
			n, err := s.u.Read(buffer)
			if err != nil {
				log.Printf("UART read error: %v", err)
				return
			}
			// Better to sleep than to keep using CPU
			if (n == 0) {
				logCounter-- 
				if logCounter == 0 {
					logCounter = 50
					fmt.Println("No data from UART (Read 0 bytes)")
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}

			fmt.Println("Data from UART: ", string(buffer[:n]))

			message := WsMessage{
				Type: "uart",
				Text: string(buffer[:n]),
			}

			json, err := json.Marshal(message)
			if err != nil {
				log.Printf("JSON marshal error: %v", err)
				return
			}

			// Forward UART data to WebSocket
			if err := conn.WriteMessage(websocket.TextMessage, json); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
    }
}
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Check if there's already an active connection
	fmt.Println("Trying to lock wsConnMu")
    s.wsConnMu.Lock()
	fmt.Println("Handle WebSocket called")
    if s.wsConn != nil {
		log.Println("Only one WebSocket connection allowed at a time")
		http.Error(w, "Only one WebSocket connection allowed at a time", http.StatusConflict)
        s.wsConnMu.Unlock()
        return
    }
	fmt.Println("Checked that there is no active connection")

    // Upgrade the connection
    conn, err := s.wsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        s.wsConnMu.Unlock()
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }
	fmt.Println("Upgraded WebSocket connection")

    // Store the connection
    s.wsConn = conn
    s.wsConnMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

    // Clean up on disconnect
    defer func() {
		fmt.Println("Trying to get lock to handle disconnect (clean up)")
        s.wsConnMu.Lock()
		fmt.Println("Got lock to handle disconnect (clean up)")
		s.timer.Stop()
		currentsession.GetCurrentSession().Reset()
        s.wsConn = nil
		cancel()
		conn.Close()
		fmt.Println("Disconnected WebSocket (clean up)")
        s.wsConnMu.Unlock()
		fmt.Println("Released lock to handle disconnect (clean up)")
    }()

    fmt.Println("New WebSocket connection established")

    // Start message handling
    go s.handleUARTToWS(conn, ctx)
    s.handleWSToUART(conn)
}

func flashFPGA(cfg config.Config, fp string) error {
	fmt.Println("Flashing FPGA")
	device := fpga.CreateFPGA(cfg.TDI, cfg.TDO, cfg.TCK, cfg.TMS)
	err := device.Flash(fp)
	return err
}

func flashMCU(cfg config.Config, fp string, server *Server) error {
	server.u.Close()
	defer server.u.Reset()
	fmt.Println("Flashing STM32")
	err := stm32flash.Flash(fp, cfg.RESET_PIN, cfg.BOOT0_PIN)
	return err
}

// handler for programming FPGA and MCU
func handleFirmware(cfg config.Config, isFPGA bool, server *Server) func(c *gin.Context) {
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
			err = flashFPGA(cfg, fp)
		} else {
			err = flashMCU(cfg, fp, server)
		}

		if err != nil {
			fmt.Println("Error flashing device:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Firmware flashed successfully"})
	}
}