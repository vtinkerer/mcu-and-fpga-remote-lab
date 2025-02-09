package uart

import (
	"fmt"
	"sync"
	"time"

	"go.bug.st/serial"
)

type UART struct {
    port     serial.Port
    mu       sync.Mutex
    isActive bool
}

func NewUART() *UART {
    return &UART{}
}

func openSerialPort() (serial.Port, error) {
	// Create new UART but with our existing ReadChannel
	mode := &serial.Mode{
		BaudRate: 115200,
	}

	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to list ports: %w", err)
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("no serial ports found")
	}

	port, err := serial.Open(ports[0], mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open port %s: %w", ports[0], err)
	}

	return port, nil
}

func (u *UART) Open() error {
    u.mu.Lock()
    defer u.mu.Unlock()

    if u.isActive {
        return nil
    }

    port, err := openSerialPort()
    if err != nil {
        return err
    }

    u.port = port
    u.isActive = true
    return nil
}

func (u *UART) Close() error {
    u.mu.Lock()
    defer u.mu.Unlock()

    if !u.isActive {
        return nil
    }

    err := u.port.Close()
    if err != nil {
        return err
    }

    u.port = nil
    u.isActive = false
    return nil
}

func (u *UART) Reset() error {
    if err := u.Close(); err != nil {
        return err
    }
    return u.Open()
}

func (u *UART) Read(buffer []byte) (int, error) {
    u.mu.Lock()
    defer u.mu.Unlock()

    
    if !u.isActive {
        return 0, nil
    }
    u.port.SetReadTimeout(time.Millisecond * 100)
    return u.port.Read(buffer)
}

func (u *UART) Write(data []byte) error {
    u.mu.Lock()
    defer u.mu.Unlock()

    if !u.isActive {
        return nil
    }
    _, err := u.port.Write(data)
    return err
}