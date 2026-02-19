package analogdiscovery

import (
	"fmt"
	"log"
	"sync"
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
var FDwfAnalogOutNodeEnableSet func(deviceHandle int32, idxChannel int, analogNode int, isEnabled int) int32
var FDwfAnalogOutNodeFunctionSet func(deviceHandle int32, idxChannel int, analogNode int, function uint16) int32
var FDwfAnalogOutNodeFrequencySet func(deviceHandle int32, idxChannel int, analogNode int, frequency float64) int32
var FDwfAnalogOutNodeAmplitudeSet func(deviceHandle int32, idxChannel int, analogNode int, amplitude float64) int32
var FDwfAnalogOutNodeOffsetSet func(deviceHandle int32, idxChannel int, analogNode int, offset float64) int32
var FDwfAnalogOutNodeSymmetrySet func(deviceHandle int32, idxChannel int, analogNode int, percSymmetry float64) int32
var FDwfAnalogOutConfigure func(deviceHandle int32, idxChannel int, fStart int)
var FDwfAnalogOutNodeSymmetryGet func(deviceHandle int32, idxChannel int, analogNode int, pSymmetry *float64) int32
var FDwfAnalogOutNodeEnableGet func(deviceHandle int32, idxChannel int, analogNode int, isEnabled *int) int32
var FDwfAnalogOutNodeFunctionGet func(deviceHandle int32, idxChannel int, analogNode int, funcName *uint16) int32
var FDwfAnalogOutNodeAmplitudeGet func(deviceHandle int32, idxChannel int, analogNode int, pAmplitude *float64) int32
var FDwfAnalogOutNodeFrequencyGet func(deviceHandle int32, idxChannel int, analogNode int, pFrequency *float64) int32
var FDwfAnalogOutNodeOffsetGet func(deviceHandle int32, idxChannel int, analogNode int, pOffset *float64) int32
var FDwfAnalogInConfigure func(deviceHandle int32, fReconfigure int, fStart int)
var FDwfAnalogInFrequencySet func(deviceHandle int32, frequency float64) int32
var FDwfAnalogInBufferSizeSet func(deviceHandle int32, bufferSize int) int32
var FDwfAnalogInChannelEnableSet func(deviceHandle int32, idxChannel int, isEnabled int) int32
var FDwfAnalogInChannelRangeSet func(deviceHandle int32, idxChannel int, volts float64) int32
var FDwfAnalogInStatus func(deviceHandle int32, readData int, pSTS *int) int32
var FDwfAnalogInStatusData func(deviceHandle int32, idxChannel int, rgdVolts *float64, cdData int) int32
var FDwfDigitalIOConfigure func(deviceHandle int32) int32

// initializing Analog Discovery library
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
	purego.RegisterLibFunc(&FDwfAnalogOutNodeSymmetrySet, dwf, "FDwfAnalogOutNodeSymmetrySet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeOffsetSet, dwf, "FDwfAnalogOutNodeOffsetSet")
	purego.RegisterLibFunc(&FDwfAnalogOutConfigure, dwf, "FDwfAnalogOutConfigure")
	purego.RegisterLibFunc(&FDwfAnalogInConfigure, dwf, "FDwfAnalogInConfigure")
	purego.RegisterLibFunc(&FDwfAnalogInFrequencySet, dwf, "FDwfAnalogInFrequencySet")
	purego.RegisterLibFunc(&FDwfAnalogInBufferSizeSet, dwf, "FDwfAnalogInBufferSizeSet")
	purego.RegisterLibFunc(&FDwfAnalogInChannelEnableSet, dwf, "FDwfAnalogInChannelEnableSet")
	purego.RegisterLibFunc(&FDwfAnalogInChannelRangeSet, dwf, "FDwfAnalogInChannelRangeSet")
	purego.RegisterLibFunc(&FDwfAnalogInStatus, dwf, "FDwfAnalogInStatus")
	purego.RegisterLibFunc(&FDwfAnalogInStatusData, dwf, "FDwfAnalogInStatusData")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeEnableGet, dwf, "FDwfAnalogOutNodeEnableGet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeSymmetryGet, dwf, "FDwfAnalogOutNodeSymmetryGet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeFrequencyGet, dwf, "FDwfAnalogOutNodeFrequencyGet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeAmplitudeGet, dwf, "FDwfAnalogOutNodeAmplitudeGet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeFunctionGet, dwf, "FDwfAnalogOutNodeFunctionGet")
	purego.RegisterLibFunc(&FDwfAnalogOutNodeOffsetGet, dwf, "FDwfAnalogOutNodeOffsetGet")
	purego.RegisterLibFunc(&FDwfDigitalIOConfigure, dwf, "FDwfDigitalIOConfigure")
}

type AnalogDiscoveryDevice struct {
	Handle  int32
	mu_gpio sync.Mutex
}

// get function by name
func GetFuncNumByName(name string) (uint16, error) {
	var funcNum uint16
	switch name {
	case "sine":
		funcNum = 1
	case "triangle":
		funcNum = 3
	case "rampup":
		funcNum = 4
	case "pulse":
		funcNum = 7
	default:
		funcNum = 0
		return funcNum, fmt.Errorf("error: %s", "no such func!")
	}
	return funcNum, nil

}

// get id of state
func GetInstrumentStateByName(name string) (int, error) {
	var state int
	switch name {
	case "stop":
		state = 0
	case "start":
		state = 1
	case "apply":
		state = 3
	default:
		state = -1
		return state, fmt.Errorf("error: %s", "no such state!")
	}
	return state, nil
}

// get analog out node carrier by name
func GetAnalogOutNodeCarrierByName(name string) (int, error) {
	var a int

	switch name {
	case "AnalogOutNodeCarrier":
		a = 0
	case "AnalogOutNodeFM":
		a = 1
	case "AnalogOutNodeAM":
		a = 2
	default:
		a = -1
		return a, fmt.Errorf("error: %s", "no such analog out node!")
	}
	return a, nil
}

// config analog out with state
func (ad *AnalogDiscoveryDevice) ConfigAnalogOut(idxChannel int, fStart int) error {
	FDwfAnalogOutConfigure(ad.Handle, idxChannel, fStart)
	return nil
}

// function that generates waveform based on transferred params
func (ad *AnalogDiscoveryDevice) GenerateWaveform(idxChannel int, analogNode string,
	fStart int) error {

	fmt.Println("Trying to generate/stop waveform for channel ", idxChannel)

	if ad.Handle == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("analog discovery handle is 0! %w", err)
		}
	}
	var a int
	a, _ = GetAnalogOutNodeCarrierByName(analogNode)

	if a == -1 {
		return fmt.Errorf("analog out node is incorrect")
	}

	var isEnabled int
	if FDwfAnalogOutNodeEnableGet(ad.Handle, idxChannel, a, &isEnabled) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error getting analog output enable: %w", err)
		}
	}

	if isEnabled == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("this channel was not enabled: %w", err)
		}
	}

	var amplitude float64
	if FDwfAnalogOutNodeAmplitudeGet(ad.Handle, idxChannel, a, &amplitude) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error getting analog output amplitude: %w", err)
		}
	}

	if amplitude < -5 || amplitude > 5 {
		if err := checkError(); err != nil {
			return fmt.Errorf("incorrect amplitude value: %w", err)
		}
	}
	fmt.Println(amplitude)

	var frequency float64
	if FDwfAnalogOutNodeFrequencyGet(ad.Handle, idxChannel, a, &frequency) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error getting analog output frequency: %w", err)
		}
	}

	if frequency <= 0.0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("incorrect frequency value: %w", err)
		}
	}

	var funcName uint16
	if FDwfAnalogOutNodeFunctionGet(ad.Handle, idxChannel, a, &funcName) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error getting analog output function: %w", err)
		}
	}

	if funcName == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("incorrect function value: %w", err)
		}
	}

	// if 'sine' - adjust offset to prevent negative voltage valaues
	if funcName == 1 {
		ad.SetAnalogOutOffset(idxChannel, analogNode, 0.5*amplitude)
		ad.SetAnalogOutAmplitude(idxChannel, analogNode, 0.5*amplitude)
	} else {
		ad.SetAnalogOutOffset(idxChannel, analogNode, 0.5*amplitude)
		ad.SetAnalogOutAmplitude(idxChannel, analogNode, 0.5*amplitude)
	}

	var symmetry float64
	if FDwfAnalogOutNodeSymmetryGet(ad.Handle, idxChannel, a, &symmetry) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error getting analog output symmetry: %w", err)
		}
	}

	if symmetry < 0.0 || symmetry > 100.0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("incorrect symmetry value: %w", err)
		}
	}

	ad.ConfigAnalogOut(idxChannel, fStart)

	if fStart == 0 {
		fmt.Println("Done !")
		//ad.Close()
	}

	return nil
}

// read values from oscilloscope
func (ad *AnalogDiscoveryDevice) ReadScopeValues(channel int, isFirstCapture int) ([]float64, []int64, error) {

	var samplingFrequency float64 = 1e06

	// before first capture
	if isFirstCapture == 1 {
		ad.SetAnalogInBufferSize(600)
		ad.SetAnalogInFrequency(samplingFrequency)
		ad.SetAnalogInChannelRange(channel, 10)
		time.Sleep(1000000000) // 1 second for stabilizing
		ad.ConfigAnalogInStart()
	}

	var sts int
	var i int64
	i = 0
	var rgdSamples [600]float64
	var timeValues [600]int64

	for i < 600 {
		FDwfAnalogInStatus(ad.Handle, 1, &sts)

		var freq int64 = int64(samplingFrequency)
		// convert to microseconds
		timeValues[i] = i * 1e06 / freq

		i++
	}

	FDwfAnalogInStatusData(ad.Handle, channel, &rgdSamples[0], 600)
	i = 0
	for i < 600 {
		i++
	}
	return rgdSamples[:], timeValues[:], nil
}

// check if there is error
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

// connect to the Analog Discovery device
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

// close the connection to device
func (ad *AnalogDiscoveryDevice) Close() {
	if ad.Handle != 0 {
		FDwfDeviceClose(ad.Handle)
	}
	ad.Handle = 0
}

// ----- ANALOG OUT (WAVEFORM GENERATOR) -----

// enable / disable specified analog out channel
func (ad *AnalogDiscoveryDevice) EnableAnalogOutChannel(indexCh int, nodeName string, isEnabled int) error {
	var enabled *int
	a, _ := GetAnalogOutNodeCarrierByName(nodeName)
	if a == -1 {
		return fmt.Errorf("no such analog out node")
	}
	if FDwfAnalogOutNodeEnableGet(ad.Handle, indexCh, a, enabled) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error getting analog output node enable: %w", err)
		}
	}
	if FDwfAnalogOutNodeEnableSet(ad.Handle, indexCh, a, isEnabled) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog output node enable: %w", err)
		}
	}
	return nil
}

// set function of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutNodeFunction(indexCh int, nodeName string, funcName string) error {
	f, _ := GetFuncNumByName(funcName)
	a, _ := GetAnalogOutNodeCarrierByName(nodeName)
	if a == -1 || f == 0 {
		return fmt.Errorf("no such analog out node or function")
	}

	if FDwfAnalogOutNodeFunctionSet(ad.Handle, indexCh, a, f) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog output node function: %w", err)
		}
	}
	var mmode uint16
	FDwfAnalogOutNodeFunctionGet(ad.Handle, indexCh, a, &mmode)
	fmt.Println("mode")
	fmt.Println(mmode)
	return nil
}

func (ad *AnalogDiscoveryDevice) SetAnalogOutOffset(indexCh int, nodeName string, offset float64) error {

	a, _ := GetAnalogOutNodeCarrierByName(nodeName)
	if a == -1 {
		return fmt.Errorf("no such analog out node")
	}
	if offset <= 0 || offset > 3.0 {
		return fmt.Errorf("incorrect or too high offset")
	}
	if FDwfAnalogOutNodeOffsetSet(ad.Handle, indexCh, a, offset) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog output offset: %w", err)
		}
	}
	return nil
}

// set symmetry of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutSymmetry(indexCh int, nodeName string, percSymmetry float64) error {
	var a int
	a, _ = GetAnalogOutNodeCarrierByName(nodeName)
	if a == -1 {
		return fmt.Errorf("analog out node is incorrect")
	}
	if FDwfAnalogOutNodeSymmetrySet(ad.Handle, indexCh, a, percSymmetry) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog output symmetry: %w", err)
		}
	}
	return nil
}

// set frequency of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutFrequency(indexCh int, nodeName string, frequencyValue float64) error {
	var a int
	a, _ = GetAnalogOutNodeCarrierByName(nodeName)
	if a == -1 {
		return fmt.Errorf("analog out node is incorrect")
	}
	if FDwfAnalogOutNodeFrequencySet(ad.Handle, indexCh, a, frequencyValue) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog output function: %w", err)
		}
	}
	return nil
}

// set amplitude of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutAmplitude(indexCh int, nodeName string, amplitudeValue float64) error {
	var a int
	a, _ = GetAnalogOutNodeCarrierByName(nodeName)
	if a == -1 {
		return fmt.Errorf("analog out node is incorrect")
	}
	fmt.Println(amplitudeValue)
	FDwfAnalogOutNodeAmplitudeSet(ad.Handle, indexCh, a, amplitudeValue)
	if FDwfAnalogOutNodeAmplitudeSet(ad.Handle, indexCh, a, amplitudeValue) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog output amplitude: %w", err)
		}
	}
	var frequency float64
	FDwfAnalogOutNodeAmplitudeGet(ad.Handle, indexCh, a, &frequency)
	fmt.Println("real value")
	fmt.Println(frequency)

	return nil
}

// ----- ANALOG IN (OSCILLOSCOPE) -----

// reconfig analog in - start
func (ad *AnalogDiscoveryDevice) ReConfigAnalogInStart() {
	FDwfAnalogInConfigure(ad.Handle, 1, 1)
}

// reconfig analog in - stop
func (ad *AnalogDiscoveryDevice) ReConfigAnalogInStop() {
	FDwfAnalogInConfigure(ad.Handle, 1, 0)
}

// config analog in - start
func (ad *AnalogDiscoveryDevice) ConfigAnalogInStart() {
	FDwfAnalogInConfigure(ad.Handle, 0, 1)
}

// config analog in - stop
func (ad *AnalogDiscoveryDevice) ConfigAnalogInStop() {
	FDwfAnalogInConfigure(ad.Handle, 0, 0)
}

// set frequency of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInFrequency(frequency float64) error {
	// if frequency is below 0 Hz or higher than 25 MHz
	if frequency <= 0 || frequency > 25000000.0 {
		return fmt.Errorf("incorrect or too high frequency")
	}
	if FDwfAnalogInFrequencySet(ad.Handle, frequency) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog input frequency: %w", err)
		}
	}
	return nil
}

// set buffer size of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInBufferSize(bufferSize int) error {
	if bufferSize <= 0 {
		return fmt.Errorf("buffer size is incorrect")
	}
	if FDwfAnalogInBufferSizeSet(ad.Handle, bufferSize) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog input buffer size: %w", err)
		}
	}
	return nil
}

// enable / disable specified analog in channel
func (ad *AnalogDiscoveryDevice) EnableAnalogInChannel(indexCh int, isEnabled int) {
	FDwfAnalogInChannelEnableSet(ad.Handle, indexCh, isEnabled)
}

// set channel range of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInChannelRange(indexCh int, volts float64) error {
	if volts >= 10.0 {
		return fmt.Errorf("channel range is incorrect")
	}
	if FDwfAnalogInChannelRangeSet(ad.Handle, indexCh, volts) == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting analog input channel range: %w", err)
		}
	}
	return nil
}

// ----- DIGITAL IO -----

// set digital input pin mode
func (ad *AnalogDiscoveryDevice) SetPinMode(pin int, mode bool) error {
	var mask uint16

	ad.mu_gpio.Lock()
	defer ad.mu_gpio.Unlock()

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

	fmt.Printf("SetPinMode Mask: %016b\n", mask)

	outputEnableResult := FDwfDigitalIOOutputEnableSet(ad.Handle, mask)
	fmt.Printf("SetPinMode output enable result: %d\n", outputEnableResult)

	if outputEnableResult == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting digital IO output enable: %w", err)
		}
	}

	return nil
}

// set digital input pin state
func (ad *AnalogDiscoveryDevice) SetPinState(pin int, value bool) error {
	var mask uint16
	var _pinModeMask uint16 // TEST LINE

	ad.mu_gpio.Lock()
	defer ad.mu_gpio.Unlock()

	if FDwfDigitalIOOutputEnableGet(ad.Handle, &_pinModeMask) == 0 { // TEST LINE
		if err := checkError(); err != nil { // TEST LINE
			return fmt.Errorf("error getting digital IO output enable: %w", err) // TEST LINE
		} // TEST LINE
	} // TEST LINE
	fmt.Printf("Pin mode mask: %016b\n", _pinModeMask) // TEST LINE

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

	fmt.Printf("Mask: %016b\n", mask)

	outputSetResult := FDwfDigitalIOOutputSet(ad.Handle, mask)
	fmt.Printf("Output set result: %d\n", outputSetResult)

	if outputSetResult == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error setting digital IO output: %w", err)
		}
	}

	ioConfigureResult := FDwfDigitalIOConfigure(ad.Handle)
	fmt.Printf("IO configure result: %d\n", ioConfigureResult)

	if ioConfigureResult == 0 {
		if err := checkError(); err != nil {
			return fmt.Errorf("error configuring digital IO: %w", err)
		}
	}

	fmt.Printf("Successfully set digital IO output\n")

	return nil
}
