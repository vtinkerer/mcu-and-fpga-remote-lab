package main

import (
	analogdiscovery "digitrans-lab-go/internal/analog-discovery"
	"digitrans-lab-go/internal/camera"
	"digitrans-lab-go/internal/config"
	currentsession "digitrans-lab-go/internal/current-session"
	"digitrans-lab-go/internal/fpga"
	stm32flash "digitrans-lab-go/internal/stm32-flash"
	"digitrans-lab-go/internal/uart"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
    u        	*uart.UART
	session  	*currentsession.CurrentSession
    wsUpgrader  websocket.Upgrader
    wsConn 		*websocket.Conn
    wsConnMu   	sync.Mutex
}

func NewServer() *Server {
	u := uart.NewUART()
	u.Open()
    return &Server{
        u: u,
		session: currentsession.GetCurrentSession(),
        wsUpgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // Adjust based on your security needs
            },
        },
		wsConn: nil,
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

	// go func() {
	// 	buf := make([]byte, 1024)
	// 	for {
	// 		val, err  := u.Read(buf)
	// 		if err != nil {
	// 			fmt.Println("Error reading from UART: ", err)
	// 		} else {
	// 			fmt.Println("Read value: ", string(buf[:val]))
	// 		}
	// 	}
	// }()

	r.POST("/api/firmware/fpga", handleFirmware(*cfg, true, server))
	r.POST("/api/firmware/mcu", handleFirmware(*cfg, false, server))
	r.POST("/api/write-pin", analogdiscovery.HandleWritePin(device))
	r.POST("/api/wavegen/write-channel", analogdiscovery.HandleWavegenEnableChannel(device))
	r.POST("/api/wavegen/write-function", analogdiscovery.HandleWavegenFunctionSet(device))
	r.POST("/api/wavegen/write-amplitude", analogdiscovery.HandleWavegenAmplitudeSet(device))
	r.POST("/api/wavegen/write-frequency", analogdiscovery.HandleWavegenFrequencySet(device))
	r.POST("/api/wavegen/write-duty-cycle", analogdiscovery.HandleWavegenDutyCycleSet(device))
	r.POST("/api/scope/get-scope-data", analogdiscovery.HandleScopeGetData(device))
	r.POST("/api/wavegen/write-config", analogdiscovery.HandleWavegenRun(device))
	r.Any("/api/stream", cam.ServeHTTP)
	r.POST("/session", currentsession.HandleCreateSession(*cfg))
	r.GET("/session", currentsession.HandleGetSession(*cfg))
	r.DELETE("/session", currentsession.HandleDeleteSession(*cfg))
	r.GET("/ws", func (c *gin.Context) {
		server.handleWebSocket(c.Writer, c.Request)
	})

	log.Fatal(r.Run(":" + cfg.PORT))
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

        // Forward message to UART
        if err := s.u.Write(message); err != nil {
            log.Printf("UART write error: %v", err)
        }
    }
}

func (s *Server) handleUARTToWS(conn *websocket.Conn) {
    buffer := make([]byte, 1024)
    for {
        n, err := s.u.Read(buffer)
        if err != nil {
            log.Printf("UART read error: %v", err)
            return
        }

        // Check if this connection is still the active one
        s.wsConnMu.Lock()
        isActive := s.wsConn == conn
        s.wsConnMu.Unlock()

        if !isActive {
            return
        }

        // Forward UART data to WebSocket
        if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
            log.Printf("WebSocket write error: %v", err)
            return
        }
    }
}
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Check if there's already an active connection
    s.wsConnMu.Lock()
    if s.wsConn != nil {
        s.wsConnMu.Unlock()
        http.Error(w, "Only one WebSocket connection allowed at a time", http.StatusConflict)
        return
    }

    // Upgrade the connection
    conn, err := s.wsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        s.wsConnMu.Unlock()
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }

    // Store the connection
    s.wsConn = conn
    s.wsConnMu.Unlock()

    // Clean up on disconnect
    defer func() {
        s.wsConnMu.Lock()
        if s.wsConn == conn {  // Only clear if it's still the same connection
            s.wsConn = nil
        }
        s.wsConnMu.Unlock()
        conn.Close()
    }()

    log.Println("New WebSocket connection established")

    // Start message handling
    go s.handleUARTToWS(conn)
    s.handleWSToUART(conn)
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
			fmt.Println("Flashing FPGA")
			device := fpga.CreateFPGA(cfg.TDI, cfg.TDO, cfg.TCK, cfg.TMS)
			err = device.Flash(fp)
		} else {
			server.u.Close()
			defer server.u.Reset()
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