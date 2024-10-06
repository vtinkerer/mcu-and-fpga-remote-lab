package analogdiscovery

import (
	"fmt"
	"log"

	"github.com/ebitengine/purego"
)

var dwf uintptr

var FDwfEnum func(device_type_must_be_zero int32, count *int32)
var FDwfDeviceConfigOpen func(index int32, auto int32, deviceHandle *int32)
var FDwfEnumDeviceType func (index_minus_one int32, device_id *int32, device_rev *int32)
var FDwfGetLastError func (error_number *int32)
var FDwfGetLastErrorMsg func (error_message *byte)
var FDwfDeviceClose func (deviceHandle int32)
var FDwfDigitalIOOutputEnableGet func (deviceHandle int32, mask *uint16) int32
var FDwfDigitalIOOutputEnableSet func (deviceHandle int32, mask uint16) int32
var FDwfDigitalIOOutputGet func (deviceHandle int32, value *uint16) int32
var FDwfDigitalIOOutputSet func (deviceHandle int32, value uint16) int32

func initDL() {
	fmt.Println("Initializing Analog Discovery dwf")

	var err error
	dwf, err = purego.Dlopen("libdwf.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		log.Fatalf("Error loading libdwf.so")
	}

	purego.RegisterLibFunc(&FDwfEnum, dwf, "FDwfEnum")
	purego.RegisterLibFunc(&FDwfDeviceConfigOpen, dwf, "FDwfDeviceConfigOpen")
	purego.RegisterLibFunc(&FDwfEnumDeviceType, dwf, "FDwfEnumDeviceType")
	purego.RegisterLibFunc(&FDwfGetLastError, dwf, "FDwfGetLastError")
	purego.RegisterLibFunc(&FDwfGetLastErrorMsg, dwf, "FDwfGetLastErrorMsg")
	purego.RegisterLibFunc(&FDwfDeviceClose, dwf, "FDwfDeviceClose")
	purego.RegisterLibFunc(&FDwfDigitalIOOutputEnableGet, dwf, "FDwfDigitalIOOutputEnableGet")
	purego.RegisterLibFunc(&FDwfDigitalIOOutputEnableSet, dwf, "FDwfDigitalIOOutputEnableSet")
	purego.RegisterLibFunc(&FDwfDigitalIOOutputGet, dwf, "FDwfDigitalIOOutputGet")
	purego.RegisterLibFunc(&FDwfDigitalIOOutputSet, dwf, "FDwfDigitalIOOutputSet")
}

type AnalogDiscoveryDevice struct {
	Handle int32
}

func checkError() error {
	errMsg := make([]byte, 512)
	FDwfGetLastErrorMsg(&errMsg[0])
	asciiString := string(errMsg)
	fmt.Println("Error: ", asciiString)
	if asciiString != "" {
		return fmt.Errorf("Error: %s", asciiString)
	}
	return nil
}

func CreateDevice () (*AnalogDiscoveryDevice, error) {
	initDL()

	deviceType := int32(0)
	var deviceCount int32 
	
	if FDwfEnum(deviceType, &deviceCount); deviceCount <= 0 {
		return nil, fmt.Errorf("No Analog Discovery devices found")
	}
	fmt.Println("Device count: ", deviceCount)
	
	var deviceHandle int32

	index := int32(0)
	for deviceHandle == 0 && index < deviceCount { 
		FDwfDeviceConfigOpen(index, 0, &deviceHandle)
		index++
	}  
	
	if deviceHandle != int32(0) {
		var deviceId int32
		var deviceRev int32
		if FDwfEnumDeviceType(index - 1, &deviceId, &deviceRev); deviceId == int32(3) {
			fmt.Println("Found Analog Discovery 2")
		} else {
			fmt.Println("Found Analog Discovery, but not an Analog Discovery 2")
		}
	}

	if deviceHandle == int32(0) {
		var err_nr int32
		if FDwfGetLastError(&err_nr); err_nr != int32(0) {
			err := checkError()
			if err != nil {
				return nil, err
			}
		}
	}

	return &AnalogDiscoveryDevice{Handle: deviceHandle}, nil
}

func (ad *AnalogDiscoveryDevice) Close() {
	if ad.Handle != 0 {
		FDwfDeviceClose(ad.Handle)
	}
	ad.Handle = 0
}

func (ad *AnalogDiscoveryDevice) SetPinMode(pin int, mode bool) error {
	var mask uint16
	
	if FDwfDigitalIOOutputEnableGet(ad.Handle, &mask) ==  0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("Error getting digital IO output enable: %w", err)
		}
	}

	if mode {
		mask |= 1 << uint(pin)
	} else {
		mask &= ^(1 << uint(pin))
	}
	
	if FDwfDigitalIOOutputEnableSet(ad.Handle, mask) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("Error setting digital IO output enable: %w", err)
		}
	}
	
	return nil
}

func (ad *AnalogDiscoveryDevice) SetPinState(pin int, value bool) error {
	var mask uint16

	if FDwfDigitalIOOutputGet(ad.Handle, &mask) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("Error getting digital IO output enable: %w", err)
		}
	}

	if value {
		mask |= 1 << uint(pin)
	} else {
		mask &= ^(1 << uint(pin))
	}

	if FDwfDigitalIOOutputSet(ad.Handle, mask) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("Error setting digital IO output: %w", err)
		}
	}

	return nil
}