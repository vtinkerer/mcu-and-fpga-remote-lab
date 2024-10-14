package analogdiscovery

import (
	"fmt"
	"log"
	"time"

	"github.com/ebitengine/purego"
)

var dwf uintptr

var FDwfEnum func(device_type_must_be_zero int32, count *int32)
var FDwfDeviceConfigOpen func(index int32, auto int32, deviceHandle *int32)
var FDwfEnumDeviceType func(index_minus_one int32, device_id *int32, device_rev *int32)
var FDwfGetLastError func(error_number *int32)
var FDwfGetLastErrorMsg func(error_message *byte)
var FDwfDeviceClose func(deviceHandle int32)
var FDwfDigitalIOOutputEnableGet func(deviceHandle int32, mask *uint16) int32
var FDwfDigitalIOOutputEnableSet func(deviceHandle int32, mask uint16) int32
var FDwfDigitalIOOutputGet func(deviceHandle int32, value *uint16) int32
var FDwfDigitalIOOutputSet func(deviceHandle int32, value uint16) int32
var FDwfAnalogInChannelCount func(deviceHandle int32, channels *int) int32
var FDwfAnalogOutNodeEnableSet func(deviceHandle int32, idxChannel int, analogNode int, isEnabled bool)
var FDwfAnalogOutNodeFunctionSet func(deviceHandle int32, idxChannel int, analogNode int, function int16)
var FDwfAnalogOutNodeFrequencySet func(deviceHandle int32, idxChannel int, analogNode int, frequency float32)
var FDwfAnalogOutNodeAmplitudeSet func(deviceHandle int32, idxChannel int, analogNode int, amplitude float32)
var FDwfAnalogOutNodeOffsetSet func(deviceHandle int32, idxChannel int, analogNode int, offset float32)
var FDwfAnalogOutNodeSymmetrySet func(deviceHandle int32, idxChannel int, analogNode int, percSymmetry float32)
var FDwfAnalogOutNodePhaseSet func(deviceHandle int32, idxChannel int, analogNode int, phaseDegree float32)
var FDwfAnalogOutConfigure func(deviceHandle int32, idxChannel int, fStart int)

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
	purego.RegisterLibFunc(&FDwfAnalogInChannelCount, dwf, "FDwfAnalogInChannelCount")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeEnableSet, dwf, "FDwfAnalogOutNodeEnableSet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeFunctionSet, dwf, "FDwfAnalogOutNodeFunctionSet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeFrequencySet, dwf, "FDwfAnalogOutNodeFrequencySet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeAmplitudeSet, dwf, "FDwfAnalogOutNodeAmplitudeSet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeOffsetSet, dwf, "FDwfAnalogOutNodeOffsetSet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeSymmetrySet, dwf, "FDwfAnalogOutNodeSymmetrySet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodePhaseSet, dwf, "FDwfAnalogOutNodePhaseSet")
	purego.RegisterLibFunc(&FDwfAnalogOutConfigure, dwf, "FDwfAnalogOutConfigure")

}

type FunctionType struct {
	FUNC int16
}

type AnalogOutNode struct {
	NODE int
}

type AnalogDiscoveryDevice struct {
	Handle int32
}

// get function by name
func GetFuncNumByName(name string) (FunctionType, error) {
	var f FunctionType
	if name == "funcDC" {
		f.FUNC = 0
		return f, nil
	}
	if name == "funcSine" {
		f.FUNC = 1
		return f, nil
	}
	if name == "funcSquare" {
		f.FUNC = 2
		return f, nil
	}
	if name == "funcTriangle" {
		f.FUNC = 3
		return f, nil
	}
	if name == "funcRampUp" {
		f.FUNC = 4
		return f, nil
	}
	if name == "funcRampDown" {
		f.FUNC = 5
		return f, nil
	}
	if name == "funcNoise" {
		f.FUNC = 6
		return f, nil
	}
	if name == "funcPulse" {
		f.FUNC = 7
		return f, nil
	}
	if name == "funcTrapezium" {
		f.FUNC = 8
		return f, nil
	}
	if name == "funcSinePower" {
		f.FUNC = 9
		return f, nil
	}
	if name == "funcSineNA" {
		f.FUNC = 10
		return f, nil
	}
	if name == "funcCustomPattern" {
		f.FUNC = 28
		return f, nil
	}
	if name == "funcPlayPattern" {
		f.FUNC = 29
		return f, nil
	}
	if name == "funcCustom" {
		f.FUNC = 30
		return f, nil
	}
	if name == "funcPlay" {
		f.FUNC = 31
		return f, nil
	}
	f.FUNC = -1
	return f, fmt.Errorf("error: %s", "No such function!")

}

// get id of state
func GetInstrumentState(name string) int {
	var state int
	if name == "stop" {
		state = 0
	} else if name == "start" {
		state = 1
	} else if name == "apply" {
		state = 3
	} else {
		state = -1
	}
	return state
}

// get analog out node carrier by name
func GetAnalogOutNodeCarrier(name string) (AnalogOutNode, error) {
	var a AnalogOutNode
	if name == "AnalogOutNodeCarrier" {
		a.NODE = 0
		return a, nil
	}
	if name == "AnalogOutNodeFM" {
		a.NODE = 1
		return a, nil
	}
	if name == "AnalogOutNodeAM" {
		a.NODE = 2
	}
	return a, fmt.Errorf("error: %s", "no such analog out node!")
}

// config analog out with state
func (ad *AnalogDiscoveryDevice) ConfigAnalogOut(idxChannel int, fStart string) error {

	var startId int = GetInstrumentState(fStart)
	FDwfAnalogOutConfigure(ad.Handle, idxChannel, startId)
	return nil
}

// function that generates waveform based on transferred params
func (ad *AnalogDiscoveryDevice) GenerateWaveform(idxChannel int, analogNode string,
	funcName string, frequency float32,
	amplitude float32, symmetry float32,
	offset float32, degreePhase float32,
	fStart string) error {

	fmt.Println("Trying to generate waveform for channel ", idxChannel)

	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}

	ad.EnableAnalogChannel(idxChannel, analogNode, true)
	ad.SetAnalogOutFunction(idxChannel, analogNode, funcName)
	ad.SetAnalogOutAmplitude(idxChannel, analogNode, amplitude)
	ad.SetAnalogOutFrequency(idxChannel, analogNode, frequency)
	ad.SetAnalogOutOffset(idxChannel, analogNode, offset)
	ad.SetAnalogOutSymmetry(idxChannel, analogNode, symmetry)
	ad.SetAnalogOutPhase(idxChannel, analogNode, degreePhase)
	ad.ConfigAnalogOut(idxChannel, fStart)

	time.Sleep(10 * time.Second)

	fmt.Println("Done !")

	ad.Close()
	return nil
}

func checkError() error {
	errMsg := make([]byte, 512)
	FDwfGetLastErrorMsg(&errMsg[0])
	asciiString := string(errMsg)
	fmt.Println("Error: ", asciiString)
	if asciiString != "" {
		return fmt.Errorf("error: %s", asciiString)
	}
	return nil
}

func CreateDevice() (*AnalogDiscoveryDevice, error) {
	initDL()

	deviceType := int32(0)
	var deviceCount int32

	if FDwfEnum(deviceType, &deviceCount); deviceCount <= 0 {
		return nil, fmt.Errorf("no Analog Discovery devices found")
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
		if FDwfEnumDeviceType(index-1, &deviceId, &deviceRev); deviceId == int32(3) {
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

// enable / disable specified analog channel
func (ad *AnalogDiscoveryDevice) EnableAnalogChannel(indexCh int, nodeName string, isEnabled bool) error {
	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	a, _ := GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeEnableSet(ad.Handle, indexCh, a.NODE, isEnabled)
	return nil
}

// set function of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutFunction(indexCh int, nodeName string, funcName string) error {
	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	f, _ := GetFuncNumByName(funcName)
	a, _ := GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeFunctionSet(ad.Handle, indexCh, a.NODE, f.FUNC)
	return nil
}

// get number of channels configured as input
func (ad *AnalogDiscoveryDevice) GetInputChannelsNum() (int, error) {
	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return 0, fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	var channels int
	var pChannels *int
	FDwfAnalogInChannelCount(ad.Handle, pChannels)
	channels = *pChannels
	return channels, nil
}

// set frequency of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutFrequency(indexCh int, nodeName string, frequencyValue float32) error {
	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	var a AnalogOutNode
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeFrequencySet(ad.Handle, indexCh, a.NODE, frequencyValue)
	return nil
}

// set amplitude of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutAmplitude(indexCh int, nodeName string, amplitudeValue float32) error {
	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	var a AnalogOutNode
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeAmplitudeSet(ad.Handle, indexCh, a.NODE, amplitudeValue)
	return nil
}

// set offset of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutOffset(indexCh int, nodeName string, offset float32) error {
	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	var a AnalogOutNode
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeOffsetSet(ad.Handle, indexCh, a.NODE, offset)
	return nil
}

// set symmetry of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutSymmetry(indexCh int, nodeName string, percSymmetry float32) error {
	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	var a AnalogOutNode
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeSymmetrySet(ad.Handle, indexCh, a.NODE, percSymmetry)
	return nil
}

// set phase of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutPhase(indexCh int, nodeName string, degreePhase float32) error {
	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	var a AnalogOutNode
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodePhaseSet(ad.Handle, indexCh, a.NODE, degreePhase)
	return nil
}

func (ad *AnalogDiscoveryDevice) Close() {
	if ad.Handle != 0 {
		FDwfDeviceClose(ad.Handle)
	}
	ad.Handle = 0
}

func (ad *AnalogDiscoveryDevice) SetPinMode(pin int, mode bool) error {
	var mask uint16

	if FDwfDigitalIOOutputEnableGet(ad.Handle, &mask) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error getting digital IO output enable: %w", err)
		}
	}

	if mode {
		mask |= 1 << uint(pin)
	} else {
		mask &= ^(1 << uint(pin))
	}

	if FDwfDigitalIOOutputEnableSet(ad.Handle, mask) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting digital IO output enable: %w", err)
		}
	}

	return nil
}

func (ad *AnalogDiscoveryDevice) SetPinState(pin int, value bool) error {
	var mask uint16

	if FDwfDigitalIOOutputGet(ad.Handle, &mask) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error getting digital IO output enable: %w", err)
		}
	}

	if value {
		mask |= 1 << uint(pin)
	} else {
		mask &= ^(1 << uint(pin))
	}

	if FDwfDigitalIOOutputSet(ad.Handle, mask) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting digital IO output: %w", err)
		}
	}

	return nil
}
