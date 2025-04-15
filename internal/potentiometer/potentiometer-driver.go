package potentiometer

import (
	"fmt"
	"sync"

	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// DeviceAddress represents the available I2C addresses for MAX5395
const (
	addrGND  = 0x28 // ADDR0 connected to GND
	addrNC   = 0x29 // ADDR0 not connected
	addrVDD  = 0x2B // ADDR0 connected to VDD
)

// Command byte values for MAX5395
const (
	cmdWiper      = 0x00 // Write wiper position
	cmdSDClr      = 0x40 // Clear shutdown
	cmdSDHWreg    = 0x48 // Open H terminal, keep wiper at register value
	cmdSDHZero    = 0x49 // Open H terminal, move wiper to zero
	cmdSDHMid     = 0x4A // Open H terminal, move wiper to midscale
	cmdSDHFull    = 0x4B // Open H terminal, move wiper to full-scale
	cmdSDLWreg    = 0x44 // Open L terminal, keep wiper at register value
	cmdSDLZero    = 0x45 // Open L terminal, move wiper to zero
	cmdSDLMid     = 0x46 // Open L terminal, move wiper to midscale
	cmdSDLFull    = 0x47 // Open L terminal, move wiper to full-scale
	cmdSDW        = 0x41 // Open W terminal, keep internal tap position
	cmdQPOff      = 0x50 // Disable charge pump (low power mode)
	cmdQPOn       = 0x51 // Enable charge pump
	cmdRST        = 0x60 // Reset to default, wiper to midscale
	
	// Read command bytes
	cmdReadWiper  = 0x00 // Read wiper position
	cmdReadConfig = 0x80 // Read configuration register
)

// MAX5395 digital potentiometer
type driverMAX5395 struct {
	dev  i2c.Dev
	mu   sync.Mutex
	addr uint16
}

// New creates a new MAX5395 device using the provided I2C bus and address
func newDriver(busName string, addr uint16) (*driverMAX5395, error) {

	_, err := host.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize host: %v", err)
	}

	fmt.Println("Host initialized")

	if _, err := driverreg.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize driver registry: %v", err)
	}

	fmt.Println("Driver registry initialized")

	// Open the I2C bus
	bus, err := i2creg.Open(busName)
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C bus %s: %v", busName, err)
	}

	// Create device
	return &driverMAX5395{
		dev:  i2c.Dev{Bus: bus, Addr: addr},
		addr: addr,
	}, nil
}

// SetWiper sets the wiper position (0-255)
// 0 = position closest to L
// 255 = position closest to H
func (d *driverMAX5395) setWiper(position uint8) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	return d.dev.Tx([]byte{cmdWiper, position}, nil)
}

// GetWiper reads the current wiper position (0-255)
func (d *driverMAX5395) getWiper() (uint8, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Send the read wiper command
	if err := d.dev.Tx([]byte{cmdReadWiper}, nil); err != nil {
		return 0, err
	}
	
	// Read the wiper value
	readBuf := make([]byte, 1)
	if err := d.dev.Tx(nil, readBuf); err != nil {
		return 0, err
	}
	
	return readBuf[0], nil
}

// Reset resets the device to its default state
// This sets the wiper to midscale (0x80), enables the charge pump,
// and clears any shutdown modes
func (d *driverMAX5395) Reset() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	return d.dev.Tx([]byte{cmdRST}, nil)
}

// ClearShutdown removes any existing shutdown condition
// This connects all potentiometer terminals and returns the wiper to
// the value stored in the wiper register
func (d *driverMAX5395) clearShutdown() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	return d.dev.Tx([]byte{cmdSDClr}, nil)
}

// ShutdownH opens the H terminal with wiper at the specified position
func (d *driverMAX5395) shutdownH(position string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	var cmd byte
	switch position {
	case "wiper": // Keep wiper at register value
		cmd = cmdSDHWreg
	case "zero": // Move wiper to zero scale
		cmd = cmdSDHZero
	case "mid": // Move wiper to midscale
		cmd = cmdSDHMid
	case "full": // Move wiper to full scale
		cmd = cmdSDHFull
	default:
		return fmt.Errorf("invalid position: %s", position)
	}
	
	return d.dev.Tx([]byte{cmd}, nil)
}

// ShutdownL opens the L terminal with wiper at the specified position
func (d *driverMAX5395) shutdownL(position string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	var cmd byte
	switch position {
	case "wiper": // Keep wiper at register value
		cmd = cmdSDLWreg
	case "zero": // Move wiper to zero scale
		cmd = cmdSDLZero
	case "mid": // Move wiper to midscale
		cmd = cmdSDLMid
	case "full": // Move wiper to full scale
		cmd = cmdSDLFull
	default:
		return fmt.Errorf("invalid position: %s", position)
	}
	
	return d.dev.Tx([]byte{cmd}, nil)
}

// ShutdownW opens the W terminal
func (d *driverMAX5395) shutdownW() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	return d.dev.Tx([]byte{cmdSDW}, nil)
}

// EnableChargePump enables the internal charge pump
// This allows operation down to 1.7V and H/W/L voltages up to 5.25V
func (d *driverMAX5395) enableChargePump() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	return d.dev.Tx([]byte{cmdQPOn}, nil)
}

// DisableChargePump disables the internal charge pump for low power operation
// This limits minimum supply voltage to 2.6V and H/W/L voltages to VDD+0.3V
func (d *driverMAX5395) disableChargePump() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	return d.dev.Tx([]byte{cmdQPOff}, nil)
}

// Configuration holds the device configuration state
type Configuration struct {
	ChargePumpEnabled bool
	HTerminalOpen     bool
	WTerminalOpen     bool
	LTerminalOpen     bool
	TapSelect         uint8 // 0: wiper reg, 1: zero, 2: midscale, 3: full-scale
}

// GetConfiguration reads the current device configuration
func (d *driverMAX5395) getConfiguration() (*Configuration, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Send the read config command
	if err := d.dev.Tx([]byte{cmdReadConfig}, nil); err != nil {
		return nil, err
	}
	
	// Read the config value
	readBuf := make([]byte, 1)
	if err := d.dev.Tx(nil, readBuf); err != nil {
		return nil, err
	}
	
	config := readBuf[0]
	return &Configuration{
		ChargePumpEnabled: (config & 0x80) != 0,
		HTerminalOpen:     (config & 0x10) != 0,
		LTerminalOpen:     (config & 0x08) != 0,
		WTerminalOpen:     (config & 0x04) != 0,
		TapSelect:         config & 0x03,
	}, nil
}
