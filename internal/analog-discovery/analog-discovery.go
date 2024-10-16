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
var FDwfAnalogInConfigure func(deviceHandle int32, fReconfigure int, fStart int)
var FDwfAnalogInRecordLengthSet func(deviceHandle int32, length float32)
var FDwfAnalogInFrequencySet func(deviceHandle int32, frequency float32)
var FDwfAnalogInBufferSizeSet func(deviceHandle int32, bufferSize int)
var FDwfAnalogInAcquisitionModeSet func(deviceHandle int32, mode int)
var FDwfAnalogInSamplingDelaySet func(deviceHandle int32, sec float32)
var FDwfAnalogInSamplingSlopeSet func(deviceHandle int32, slope int)
var FDwfAnalogInSamplingSourceSet func(deviceHandle int32, trigSrc int)
var FDwfAnalogInChannelEnableSet func(deviceHandle int32, idxChannel int, isEnabled bool)
var FDwfAnalogInChannelRangeSet func(deviceHandle int32, idxChannel int, volts float32)
var FDwfAnalogInChannelOffsetSet func(deviceHandle int32, idxChannel int, voltsOffset float32)
var FDwfAnalogInChannelAttenuationSet func(devicehandle int32, idxChannel int, xAttenuation float32)
var FDwfAnalogInTriggerPositionSet func(deviceHandle int32, secPosition float32)
var FDwfAnalogInTriggerAutoTimeoutSet func(deviceHandle int32, secTimeout float32)
var FDwfAnalogInTriggerHoldOffSet func(deviceHandle int32, secOffset float32)
var FDwfAnalogInTriggerTypeSet func(deviceHandle int32, trigType int)
var FDwfAnalogInTriggerChannelSet func(deviceHandle int32, idxChannel int)
var FDwfAnalogInTriggerConditionSet func(deviceHanlde int32, trigCond int)
var FDwfAnalogInTriggerLevelSet func(deviceHandle int32, voltsLevel float32)
var FDwfAnalogInTriggerHysteresisSet func(deviceHandle int32, voltsLevel float32)
var FDwfAnalogInTriggerLengthConditionSet func(deviceHandle int32, trigLen int)
var FDwfAnalogInTriggerLengthSet func(deviceHandle int32, trigLen float32)

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
	purego.RegisterLibFunc(&FDwfAnalogInConfigure, dwf, "FDwfAnalogInConfigure")
	purego.RegisterLibFunc(&FDwfAnalogInRecordLengthSet, dwf, "FDwfAnalogInRecordLengthSet")
	purego.RegisterLibFunc(&FDwfAnalogInFrequencySet, dwf, "FDwfAnalogInFrequencySet")
	purego.RegisterLibFunc(&FDwfAnalogInBufferSizeSet, dwf, "FDwfAnalogInBufferSizeSet")
	purego.RegisterLibFunc(&FDwfAnalogInAcquisitionModeSet, dwf, "FDwfAnalogInAcquisitionModeSet")
	purego.RegisterLibFunc(&FDwfAnalogInSamplingDelaySet, dwf, "FDwfAnalogInSamplingDelaySet")
	purego.RegisterLibFunc(&FDwfAnalogInSamplingSlopeSet, dwf, "FDwfAnalogInSamplingSlopeSet")
	purego.RegisterLibFunc(&FDwfAnalogInSamplingSourceSet, dwf, "FDwfAnalogInSamplingSourceSet")
	purego.RegisterLibFunc(&FDwfAnalogInChannelEnableSet, dwf, "FDwfAnalogInChannelEnableSet")
	purego.RegisterLibFunc(&FDwfAnalogInChannelRangeSet, dwf, "FDwfAnalogInChannelRangeSet")
	purego.RegisterLibFunc(&FDwfAnalogInChannelOffsetSet, dwf, "FDwfAnalogInChannelOffsetSet")
	purego.RegisterLibFunc(&FDwfAnalogInChannelAttenuationSet, dwf, "FDwfAnalogInChannelAttenuationSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerPositionSet, dwf, "FDwfAnalogInTriggerPositionSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerAutoTimeoutSet, dwf, "FDwfAnalogInTriggerAutoTimeoutSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerHoldOffSet, dwf, "FDwfAnalogInTriggerHoldOffSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerTypeSet, dwf, "FDwfAnalogInTriggerTypeSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerChannelSet, dwf, "FDwfAnalogInTriggerChannelSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerConditionSet, dwf, "FDwfAnalogInTriggerConditionSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerLevelSet, dwf, "FDwfAnalogInTriggerLevelSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerHysteresisSet, dwf, "FDwfAnalogInTriggerHysteresisSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerLengthConditionSet, dwf, "FDwfAnalogInTriggerLengthConditionSet")
	purego.RegisterLibFunc(&FDwfAnalogInTriggerLengthSet, dwf, "FDwfAnalogInTriggerLengthSet")
}

type AnalogDiscoveryDevice struct {
	Handle int32
}

// get function by name
func GetFuncNumByName(name string) (int16, error) {
	var funcNum int16

	switch name {
	case "funcDC":
		funcNum = 0
	case "funcSine":
		funcNum = 1
	case "funcSquare":
		funcNum = 2
	case "funcTriangle":
		funcNum = 3
	case "funcRampUp":
		funcNum = 4
	case "funcRampDown":
		funcNum = 5
	case "funcNoise":
		funcNum = 6
	case "funcPulse":
		funcNum = 7
	case "funcTrapezium":
		funcNum = 8
	case "funcSinePower":
		funcNum = 9
	case "funcSineNA":
		funcNum = 10
	case "funcCustomPattern":
		funcNum = 28
	case "funcPlayPattern":
		funcNum = 29
	case "funcCustom":
		funcNum = 30
	case "funcPlay":
		funcNum = 31
	default:
		funcNum = -1
		return funcNum, fmt.Errorf("error: %s", "no such func!")
	}
	return funcNum, nil

}

// get id of state
func GetInstrumentState(name string) (int, error) {
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
func GetAnalogOutNodeCarrier(name string) (int, error) {
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

// get analog in acquisition mode by name
func GetAcquisitionModeByName(modeName string) (int, error) {
	var m int
	switch modeName {
	case "acqmodeSingle":
		m = 0
	case "acqmodeScanShift":
		m = 1
	case "acqmodeScanScreen":
		m = 2
	case "acqmodeRecord":
		m = 3
	case "acqmodeSingle1":
		m = 5
	default:
		m = -1
		return m, fmt.Errorf("error: %s", "no such analog in acquisition mode!")
	}
	return m, nil
}

// get analog in sampling slope by name
func GetSamplingSlopeByName(sampleName string) (int, error) {
	var s int
	switch sampleName {
	case "DwfTriggerSlopeRise":
		s = 0
	case "DwfTriggerSlopeFall":
		s = 1
	case "DwfTriggerSlopeEither":
		s = 2
	default:
		s = -1
		return s, fmt.Errorf("error: %s", "no such trigger slope!")
	}
	return s, nil
}

// get analog in trigger source by name
func GetTrigSrcByName(trigSrc string) (int, error) {
	var t int
	switch trigSrc {
	case "trigsrcNone":
		t = 0
	case "trigsrcPC":
		t = 1
	case "trigsrcDetectorAnalogIn":
		t = 2
	case "trigsrcDetectorDigitalIn":
		t = 3
	case "trigsrcAnalogIn":
		t = 4
	case "trigsrcDigitalIn":
		t = 5
	case "trigsrcDigitalOut":
		t = 6
	case "trigsrcAnalogOut1":
		t = 7
	case "trigsrcAnalogOut2":
		t = 8
	case "trigsrcAnalogOut3":
		t = 9
	case "trigsrcAnalogOut4":
		t = 10
	case "trigsrcExternal1":
		t = 11
	case "trigsrcExternal2":
		t = 12
	case "trigsrcExternal3":
		t = 13
	case "trigsrcExternal4":
		t = 14
	case "trigsrcHigh":
		t = 15
	case "trigsrcLow":
		t = 16
	case "trigsrcClock":
		t = 17
	default:
		t = -1
		return t, fmt.Errorf("error: %s", "no such trigger source!")
	}
	return t, nil
}

// get analog in trigger type by name
func GetTrigTypeByName(trigType string) (int, error) {
	var t int
	switch trigType {
	case "trigtypeEdge":
		t = 0
	case "trigtypePulse":
		t = 1
	case "trigtypeTransition":
		t = 2
	case "trigtypeWindow":
		t = 3
	default:
		t = -1
		return t, fmt.Errorf("error: %s", "no such trigger type!")
	}
	return t, nil
}

// get analog in trigger length by name
func GetTrigLenByName(trigLen string) (int, error) {
	var tl int
	switch trigLen {
	case "triglenLess":
		tl = 0
	case "triglenTimeout":
		tl = 1
	case "triglenMore":
		tl = 2
	default:
		tl = -1
		return tl, fmt.Errorf("error: %s", "no such trigger length!")
	}
	return tl, nil
}

// config analog out with state
func (ad *AnalogDiscoveryDevice) ConfigAnalogOut(idxChannel int, fStart string) {
	var startId int
	startId, _ = GetInstrumentState(fStart)
	FDwfAnalogOutConfigure(ad.Handle, idxChannel, startId)
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

	ad.EnableAnalogOutChannel(idxChannel, analogNode, true)
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
func (ad *AnalogDiscoveryDevice) EnableAnalogOutChannel(indexCh int, nodeName string, isEnabled bool) error {
	a, _ := GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeEnableSet(ad.Handle, indexCh, a, isEnabled)
	return nil
}

// set function of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutFunction(indexCh int, nodeName string, funcName string) error {
	f, _ := GetFuncNumByName(funcName)
	a, _ := GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeFunctionSet(ad.Handle, indexCh, a, f)
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
	var a int
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeFrequencySet(ad.Handle, indexCh, a, frequencyValue)
	return nil
}

// set amplitude of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutAmplitude(indexCh int, nodeName string, amplitudeValue float32) error {
	var a int
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeAmplitudeSet(ad.Handle, indexCh, a, amplitudeValue)
	return nil
}

// set offset of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutOffset(indexCh int, nodeName string, offset float32) error {
	var a int
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeOffsetSet(ad.Handle, indexCh, a, offset)
	return nil
}

// set symmetry of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutSymmetry(indexCh int, nodeName string, percSymmetry float32) error {
	var a int
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodeSymmetrySet(ad.Handle, indexCh, a, percSymmetry)
	return nil
}

// set phase of analog out
func (ad *AnalogDiscoveryDevice) SetAnalogOutPhase(indexCh int, nodeName string, degreePhase float32) error {
	var a int
	a, _ = GetAnalogOutNodeCarrier(nodeName)
	FDwfAnalogOutNodePhaseSet(ad.Handle, indexCh, a, degreePhase)
	return nil
}

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

// set length of analog in record
func (ad *AnalogDiscoveryDevice) SetAnalogInRecordLength(length float32) error {
	if length <= 0.0 || length > 60.0 {
		return fmt.Errorf("incorrect or too long record length")
	}
	FDwfAnalogInRecordLengthSet(ad.Handle, length)
	return nil
}

// set frequency of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInFrequency(frequency float32) error {
	// if frequency is below 0 Hz or higher than 25 MHz
	if frequency <= 0 || frequency > 25000000.0 {
		return fmt.Errorf("incorrect or too high frequency")
	}
	FDwfAnalogInFrequencySet(ad.Handle, frequency)
	return nil
}

// set buffer size of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInBufferSize(bufferSize int) error {
	if bufferSize <= 0 {
		return fmt.Errorf("buffer size is incorrect")
	}
	FDwfAnalogInBufferSizeSet(ad.Handle, bufferSize)
	return nil
}

// set acquisition mode of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInAcquisitionMode(modeName string) error {
	var mode int
	mode, _ = GetAcquisitionModeByName(modeName)
	if mode == -1 {
		return fmt.Errorf("acquisition mode is incorrect")
	}
	FDwfAnalogInAcquisitionModeSet(ad.Handle, mode)
	return nil
}

// set trigger source of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerSource(trigSrc string) error {
	var trig int
	trig, _ = GetTrigSrcByName(trigSrc)
	if trig == -1 {
		return fmt.Errorf("trigger source is incorrect")
	}
	FDwfAnalogInSamplingSourceSet(ad.Handle, trig)
	return nil
}

// set sampling slope of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInSamplingSlope(slopeName string) error {
	var slope int
	slope, _ = GetSamplingSlopeByName(slopeName)
	if slope == -1 {
		return fmt.Errorf("trigger slope is incorrect")
	}
	FDwfAnalogInSamplingSlopeSet(ad.Handle, slope)
	return nil
}

// set sampling delay of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInSamplingDelay(sec float32) error {
	if sec <= 0.0 || sec > 30.0 {
		return fmt.Errorf("sampling delay is incorrect")
	}
	FDwfAnalogInSamplingDelaySet(ad.Handle, sec)
	return nil
}

// enable / disable specified analog in channel
func (ad *AnalogDiscoveryDevice) EnableAnalogInChannel(indexCh int, isEnabled bool) {
	FDwfAnalogInChannelEnableSet(ad.Handle, indexCh, isEnabled)
}

// set channel range of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInChannelRange(indexCh int, volts float32) error {
	if volts < -5.5 || volts > 5.5 {
		return fmt.Errorf("channel range is incorrect")
	}
	FDwfAnalogInChannelRangeSet(ad.Handle, indexCh, volts)
	return nil
}

// set channel offset of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInChannelOffset(indexCh int, voltsOffset float32) error {
	if voltsOffset < -5.5 || voltsOffset > 5.5 {
		return fmt.Errorf("volts offset is incorrect")
	}
	FDwfAnalogInChannelOffsetSet(ad.Handle, indexCh, voltsOffset)
	return nil
}

// set channel attenuation of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInChannelAttenuation(indexCh int, xAttenuation float32) error {
	if xAttenuation < 0.0 {
		return fmt.Errorf("attenuation is incorrect")
	}
	FDwfAnalogInChannelAttenuationSet(ad.Handle, indexCh, xAttenuation)
	return nil
}

// set trigger position of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerPosition(secPosition float32) error {
	if secPosition < 0.0 {
		return fmt.Errorf("trigger position is incorrect")
	}
	FDwfAnalogInTriggerPositionSet(ad.Handle, secPosition)
	return nil
}

// set trigger auto timeout of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerAutoTimeout(secTimeout float32) error {
	if secTimeout < 0.0 {
		return fmt.Errorf("trigger auto timeout is incorrect")
	}
	FDwfAnalogInTriggerAutoTimeoutSet(ad.Handle, secTimeout)
	return nil
}

// set trigger hold offset of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerHoldOffset(secHoldOff float32) error {
	if secHoldOff < 0.0 {
		return fmt.Errorf("trigger auto timeout is incorrect")
	}
	FDwfAnalogInTriggerHoldOffSet(ad.Handle, secHoldOff)
	return nil
}

// set trigger type of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerType(trigType string) error {
	var trig int
	trig, _ = GetTrigTypeByName(trigType)
	if trig == -1 {
		return fmt.Errorf("trigger type is incorrect")
	}
	FDwfAnalogInTriggerTypeSet(ad.Handle, trig)
	return nil
}

// set trigger channel of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerChannel(idxChannel int) {
	FDwfAnalogInTriggerChannelSet(ad.Handle, idxChannel)
}

// func SetAnalogInTriggerFilter

// set trigger condition of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerCondition(trigCond string) error {
	var cond int
	cond, _ = GetSamplingSlopeByName(trigCond)
	if cond == -1 {
		return fmt.Errorf("trigger condition is incorrect")
	}
	FDwfAnalogInTriggerConditionSet(ad.Handle, cond)
	return nil
}

// set trigger hysteresis of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerHysteresis(volts float32) error {
	if volts < -5.5 || volts > 5.5 {
		return fmt.Errorf("trigger hysteresis is incorrect")
	}
	FDwfAnalogInTriggerHysteresisSet(ad.Handle, volts)
	return nil
}

// set trigger level of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerLevel(volts float32) error {
	if volts < -5.5 || volts > 5.5 {
		return fmt.Errorf("trigger level is incorrect")
	}
	FDwfAnalogInTriggerLevelSet(ad.Handle, volts)
	return nil
}

// set trigger length condition of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerLengthCondition(trigLen string) error {
	var lengthCond int
	lengthCond, _ = GetTrigLenByName(trigLen)
	if lengthCond == -1 {
		return fmt.Errorf("trigger length condition is incorrect")
	}
	FDwfAnalogInTriggerLengthConditionSet(ad.Handle, lengthCond)
	return nil
}

// set trigger length of analog in
func (ad *AnalogDiscoveryDevice) SetAnalogInTriggerLength(secLength float32) error {
	if secLength < 0.0 {
		return fmt.Errorf("trigger length is incorrect")
	}
	FDwfAnalogInTriggerLengthSet(ad.Handle, secLength)
	return nil
}

// close the device
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
