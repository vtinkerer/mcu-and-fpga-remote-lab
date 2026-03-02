package analogdiscovery

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/ebitengine/purego"
)

const (
	logicAnalyzerDefaultSampleRateHz = 250_000
	logicAnalyzerMinSampleRateHz     = 1_000
	logicAnalyzerMinDurationSec      = 1
	logicAnalyzerMaxDurationSec      = 10
	logicAnalyzerMaxCaptureSamples   = 20_000_000
	logicAnalyzerDefaultTimeoutSec   = 10
	logicAnalyzerMaxUserSampleRateHz = 2_000_000
	logicAnalyzerCorruptWarnMinCount = 32
	logicAnalyzerCorruptWarnMinRatio = 0.0001
	logicAnalyzerReadChunkSamples    = 8192
)

const (
	dwfAcqModeRecord = 3

	dwfTrigSrcNone              = 0
	dwfTrigSrcDetectorDigitalIn = 3

	dwfTrigSlopeRise   = 0
	dwfTrigSlopeFall   = 1
	dwfTrigSlopeEither = 2
)

var (
	FDwfDigitalInReset                 func(deviceHandle int32) int32
	FDwfDigitalInConfigure             func(deviceHandle int32, fReconfigure int, fStart int) int32
	FDwfDigitalInStatus                func(deviceHandle int32, fReadData int, pStatus *byte) int32
	FDwfDigitalInStatusData            func(deviceHandle int32, rgBytes *byte, sampleCount int) int32
	FDwfDigitalInStatusRecord          func(deviceHandle int32, pDataAvailable *int, pDataLost *int, pDataCorrupt *int) int32
	FDwfDigitalInInternalClockInfo     func(deviceHandle int32, phzFreq *float64) int32
	FDwfDigitalInDividerInfo           func(deviceHandle int32, pMin *int, pMax *int) int32
	FDwfDigitalInDividerSet            func(deviceHandle int32, divider int) int32
	FDwfDigitalInDividerGet            func(deviceHandle int32, pDivider *int) int32
	FDwfDigitalInSampleFormatSet       func(deviceHandle int32, bits int) int32
	FDwfDigitalInBufferSizeInfo        func(deviceHandle int32, pMin *int, pMax *int) int32
	FDwfDigitalInBufferSizeSet         func(deviceHandle int32, bufferSize int) int32
	FDwfDigitalInAcquisitionModeSet    func(deviceHandle int32, acqMode int) int32
	FDwfDigitalInTriggerSourceSet      func(deviceHandle int32, trigsrc uint8) int32
	FDwfDigitalInTriggerSlopeSet       func(deviceHandle int32, slope int) int32
	FDwfDigitalInTriggerPositionSet    func(deviceHandle int32, samples uint32) int32
	FDwfDigitalInTriggerPrefillSet     func(deviceHandle int32, samples uint32) int32
	FDwfDigitalInTriggerAutoTimeoutSet func(deviceHandle int32, secTimeout float64) int32
	FDwfDigitalInTriggerSet            func(deviceHandle int32, levelLow uint32, levelHigh uint32, edgeRise uint32, edgeFall uint32) int32
)

var ErrLogicCaptureBusy = errors.New("logic analyzer capture already in progress")

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type DeviceUnavailableError struct {
	Message string
}

func (e DeviceUnavailableError) Error() string {
	return e.Message
}

type DeviceRuntimeError struct {
	Message string
}

func (e DeviceRuntimeError) Error() string {
	return e.Message
}

type LogicCaptureTrigger struct {
	Type       string `json:"type"`
	Channel    int    `json:"channel"`
	Edge       string `json:"edge"`
	TimeoutSec int    `json:"timeoutSec"`
}

type LogicCaptureRequest struct {
	DurationSec  int                 `json:"durationSec"`
	Channels     []int               `json:"channels"`
	SampleRateHz int                 `json:"sampleRateHz"`
	Trigger      LogicCaptureTrigger `json:"trigger"`
}

type LogicCaptureTransition struct {
	TUs   int64 `json:"tUs"`
	Value int   `json:"value"`
}

type LogicCaptureChannel struct {
	Channel     int                      `json:"channel"`
	Transitions []LogicCaptureTransition `json:"transitions"`
}

type LogicCaptureResponse struct {
	DurationSec        int                   `json:"durationSec"`
	SampleRateHz       int                   `json:"sampleRateHz"`
	Triggered          bool                  `json:"triggered"`
	TriggerTimestampUs int64                 `json:"triggerTimestampUs"`
	Channels           []LogicCaptureChannel `json:"channels"`
	Warnings           []string              `json:"warnings"`
}

type normalizedLogicCaptureRequest struct {
	DurationSec  int
	Channels     []int
	SampleRateHz int
	Trigger      LogicCaptureTrigger

	requestedRate int
	hardwareClock float64
}

func registerDigitalInFunctions() {
	purego.RegisterLibFunc(&FDwfDigitalInReset, dwf, "FDwfDigitalInReset")
	purego.RegisterLibFunc(&FDwfDigitalInConfigure, dwf, "FDwfDigitalInConfigure")
	purego.RegisterLibFunc(&FDwfDigitalInStatus, dwf, "FDwfDigitalInStatus")
	purego.RegisterLibFunc(&FDwfDigitalInStatusData, dwf, "FDwfDigitalInStatusData")
	purego.RegisterLibFunc(&FDwfDigitalInStatusRecord, dwf, "FDwfDigitalInStatusRecord")
	purego.RegisterLibFunc(&FDwfDigitalInInternalClockInfo, dwf, "FDwfDigitalInInternalClockInfo")
	purego.RegisterLibFunc(&FDwfDigitalInDividerInfo, dwf, "FDwfDigitalInDividerInfo")
	purego.RegisterLibFunc(&FDwfDigitalInDividerSet, dwf, "FDwfDigitalInDividerSet")
	purego.RegisterLibFunc(&FDwfDigitalInDividerGet, dwf, "FDwfDigitalInDividerGet")
	purego.RegisterLibFunc(&FDwfDigitalInSampleFormatSet, dwf, "FDwfDigitalInSampleFormatSet")
	purego.RegisterLibFunc(&FDwfDigitalInBufferSizeInfo, dwf, "FDwfDigitalInBufferSizeInfo")
	purego.RegisterLibFunc(&FDwfDigitalInBufferSizeSet, dwf, "FDwfDigitalInBufferSizeSet")
	purego.RegisterLibFunc(&FDwfDigitalInAcquisitionModeSet, dwf, "FDwfDigitalInAcquisitionModeSet")
	purego.RegisterLibFunc(&FDwfDigitalInTriggerSourceSet, dwf, "FDwfDigitalInTriggerSourceSet")
	purego.RegisterLibFunc(&FDwfDigitalInTriggerSlopeSet, dwf, "FDwfDigitalInTriggerSlopeSet")
	purego.RegisterLibFunc(&FDwfDigitalInTriggerPositionSet, dwf, "FDwfDigitalInTriggerPositionSet")
	purego.RegisterLibFunc(&FDwfDigitalInTriggerPrefillSet, dwf, "FDwfDigitalInTriggerPrefillSet")
	purego.RegisterLibFunc(&FDwfDigitalInTriggerAutoTimeoutSet, dwf, "FDwfDigitalInTriggerAutoTimeoutSet")
	purego.RegisterLibFunc(&FDwfDigitalInTriggerSet, dwf, "FDwfDigitalInTriggerSet")
}

func (ad *AnalogDiscoveryDevice) CaptureLogicTransitions(req LogicCaptureRequest) (LogicCaptureResponse, error) {
	if !ad.mu_logicAnalyzer.TryLock() {
		return LogicCaptureResponse{}, ErrLogicCaptureBusy
	}
	defer ad.mu_logicAnalyzer.Unlock()

	if ad.Handle == 0 {
		return LogicCaptureResponse{}, DeviceUnavailableError{Message: "analog discovery device is not available"}
	}

	normalizedReq, warnings, err := ad.normalizeLogicCaptureRequest(req)
	if err != nil {
		return LogicCaptureResponse{}, err
	}

	log.Printf("logic analyzer capture request: durationSec=%d channels=%v sampleRateHz=%d triggerType=%s triggerChannel=%d triggerEdge=%s timeoutSec=%d",
		normalizedReq.DurationSec,
		normalizedReq.Channels,
		normalizedReq.SampleRateHz,
		normalizedReq.Trigger.Type,
		normalizedReq.Trigger.Channel,
		normalizedReq.Trigger.Edge,
		normalizedReq.Trigger.TimeoutSec,
	)

	response, err := ad.performLogicCaptureOnce(normalizedReq, warnings)
	if err != nil {
		return LogicCaptureResponse{}, err
	}
	return response, nil
}

func (ad *AnalogDiscoveryDevice) performLogicCaptureOnce(normalizedReq normalizedLogicCaptureRequest, warnings []string) (LogicCaptureResponse, error) {
	captureResponse := LogicCaptureResponse{
		DurationSec:        normalizedReq.DurationSec,
		SampleRateHz:       normalizedReq.SampleRateHz,
		Triggered:          false,
		TriggerTimestampUs: 0,
		Channels:           []LogicCaptureChannel{},
		Warnings:           slices.Clone(warnings),
	}

	appliedSampleRateHz, err := ad.configureLogicAnalyzer(normalizedReq)
	if err != nil {
		return captureResponse, err
	}
	if appliedSampleRateHz <= 0 {
		return captureResponse, DeviceRuntimeError{Message: "invalid applied sample rate"}
	}
	if appliedSampleRateHz != normalizedReq.SampleRateHz {
		warningsSet := map[string]struct{}{}
		for _, warning := range captureResponse.Warnings {
			warningsSet[warning] = struct{}{}
		}
		warningsSet["sampleRateHz adjusted by hardware divider"] = struct{}{}
		captureResponse.Warnings = slices.Collect(mapKeys(warningsSet))
		slices.Sort(captureResponse.Warnings)
	}
	captureResponse.SampleRateHz = appliedSampleRateHz

	targetSamples := normalizedReq.DurationSec * appliedSampleRateHz
	if targetSamples > logicAnalyzerMaxCaptureSamples {
		targetSamples = logicAnalyzerMaxCaptureSamples
		captureResponse.Warnings = append(captureResponse.Warnings, "capture sample budget limit reached")
		captureResponse.Warnings = deduplicateStrings(captureResponse.Warnings)
	}
	if targetSamples <= 0 {
		return captureResponse, ValidationError{Message: "invalid capture target"}
	}

	defer func() {
		_ = ad.dwfCall("FDwfDigitalInConfigure(stop)", FDwfDigitalInConfigure(ad.Handle, 0, 0))
		_ = ad.dwfCall("FDwfDigitalInReset", FDwfDigitalInReset(ad.Handle))
	}()

	if err := ad.dwfCall("FDwfDigitalInConfigure(start)", FDwfDigitalInConfigure(ad.Handle, 1, 1)); err != nil {
		return captureResponse, err
	}

	triggered := normalizedReq.Trigger.Type == "immediate"
	if triggered {
		log.Printf("logic analyzer trigger fired immediately")
	} else {
		log.Printf("logic analyzer edge trigger armed")
	}

	triggerDeadline := time.Now().Add(time.Duration(normalizedReq.Trigger.TimeoutSec) * time.Second)
	warningsSet := map[string]struct{}{}
	for _, warning := range warnings {
		warningsSet[warning] = struct{}{}
	}

	samples, triggered, loopWarnings, err := ad.collectLogicSamples(normalizedReq, targetSamples, triggered, triggerDeadline)
	if err != nil {
		return captureResponse, err
	}
	for _, warning := range loopWarnings {
		warningsSet[warning] = struct{}{}
	}

	captureResponse.Warnings = slices.Collect(mapKeys(warningsSet))
	slices.Sort(captureResponse.Warnings)

	if !triggered {
		captureResponse.Warnings = append(captureResponse.Warnings, "edge trigger timeout")
		captureResponse.Warnings = deduplicateStrings(captureResponse.Warnings)
		log.Printf("logic analyzer trigger timeout")
		return captureResponse, nil
	}

	if len(samples) == 0 {
		captureResponse.Warnings = append(captureResponse.Warnings, "no samples captured")
		captureResponse.Warnings = deduplicateStrings(captureResponse.Warnings)
		captureResponse.Triggered = true
		return captureResponse, nil
	}

	captureResponse.Triggered = true
	captureResponse.Channels = buildLogicTransitions(samples, captureResponse.SampleRateHz, normalizedReq.Channels)
	log.Printf("logic analyzer capture complete: samples=%d durationSec=%d", len(samples), normalizedReq.DurationSec)

	return captureResponse, nil
}

func (ad *AnalogDiscoveryDevice) normalizeLogicCaptureRequest(req LogicCaptureRequest) (normalizedLogicCaptureRequest, []string, error) {
	if req.DurationSec < logicAnalyzerMinDurationSec || req.DurationSec > logicAnalyzerMaxDurationSec {
		return normalizedLogicCaptureRequest{}, nil, ValidationError{Message: "durationSec must be in range 1..10"}
	}

	if len(req.Channels) == 0 {
		return normalizedLogicCaptureRequest{}, nil, ValidationError{Message: "channels must be non-empty"}
	}

	seenChannels := map[int]struct{}{}
	for _, channel := range req.Channels {
		if channel < 0 || channel > 15 {
			return normalizedLogicCaptureRequest{}, nil, ValidationError{Message: "channel values must be in range 0..15"}
		}
		if _, exists := seenChannels[channel]; exists {
			return normalizedLogicCaptureRequest{}, nil, ValidationError{Message: "channels must contain unique values"}
		}
		seenChannels[channel] = struct{}{}
	}

	triggerType := strings.ToLower(strings.TrimSpace(req.Trigger.Type))
	if triggerType != "immediate" && triggerType != "edge" {
		return normalizedLogicCaptureRequest{}, nil, ValidationError{Message: "trigger.type must be one of: immediate, edge"}
	}

	normalizedTrigger := LogicCaptureTrigger{
		Type:       triggerType,
		Channel:    req.Trigger.Channel,
		Edge:       strings.ToLower(strings.TrimSpace(req.Trigger.Edge)),
		TimeoutSec: req.Trigger.TimeoutSec,
	}

	if triggerType == "edge" {
		if normalizedTrigger.Channel < 0 || normalizedTrigger.Channel > 15 {
			return normalizedLogicCaptureRequest{}, nil, ValidationError{Message: "trigger.channel must be in range 0..15 for edge trigger"}
		}
		if normalizedTrigger.Edge != "rising" && normalizedTrigger.Edge != "falling" && normalizedTrigger.Edge != "either" {
			return normalizedLogicCaptureRequest{}, nil, ValidationError{Message: "trigger.edge must be one of: rising, falling, either"}
		}
		if normalizedTrigger.TimeoutSec <= 0 {
			normalizedTrigger.TimeoutSec = logicAnalyzerDefaultTimeoutSec
		}
	}

	if triggerType == "immediate" {
		normalizedTrigger.TimeoutSec = 0
		normalizedTrigger.Edge = ""
		normalizedTrigger.Channel = 0
	}

	requestedRate := req.SampleRateHz
	if requestedRate == 0 {
		requestedRate = logicAnalyzerDefaultSampleRateHz
	}

	internalClock, err := ad.getDigitalInClockAndDividerInfo()
	if err != nil {
		return normalizedLogicCaptureRequest{}, nil, err
	}

	maxAllowedRate := logicAnalyzerMaxCaptureSamples / req.DurationSec
	if maxAllowedRate > logicAnalyzerMaxUserSampleRateHz {
		maxAllowedRate = logicAnalyzerMaxUserSampleRateHz
	}
	hardwareSafeMaxRate := int(math.Floor(internalClock))
	if maxAllowedRate > hardwareSafeMaxRate {
		maxAllowedRate = hardwareSafeMaxRate
	}
	if maxAllowedRate < logicAnalyzerMinSampleRateHz {
		maxAllowedRate = logicAnalyzerMinSampleRateHz
	}

	effectiveRate := requestedRate
	if effectiveRate < logicAnalyzerMinSampleRateHz {
		effectiveRate = logicAnalyzerMinSampleRateHz
	}
	if effectiveRate > maxAllowedRate {
		effectiveRate = maxAllowedRate
	}

	warnings := []string{}
	if effectiveRate != requestedRate {
		warnings = append(warnings, "sampleRateHz clamped")
	}

	normalizedReq := normalizedLogicCaptureRequest{
		DurationSec:   req.DurationSec,
		Channels:      slices.Clone(req.Channels),
		SampleRateHz:  effectiveRate,
		Trigger:       normalizedTrigger,
		requestedRate: requestedRate,
		hardwareClock: internalClock,
	}

	return normalizedReq, warnings, nil
}

func (ad *AnalogDiscoveryDevice) configureLogicAnalyzer(req normalizedLogicCaptureRequest) (int, error) {
	if err := ad.dwfCall("FDwfDigitalInReset", FDwfDigitalInReset(ad.Handle)); err != nil {
		return 0, err
	}
	if err := ad.dwfCall("FDwfDigitalInSampleFormatSet", FDwfDigitalInSampleFormatSet(ad.Handle, 16)); err != nil {
		return 0, err
	}
	if err := ad.dwfCall("FDwfDigitalInAcquisitionModeSet", FDwfDigitalInAcquisitionModeSet(ad.Handle, dwfAcqModeRecord)); err != nil {
		return 0, err
	}

	divider := int(math.Round(req.hardwareClock / float64(req.SampleRateHz)))
	if divider < 1 {
		divider = 1
	}
	if err := ad.dwfCall("FDwfDigitalInDividerSet", FDwfDigitalInDividerSet(ad.Handle, divider)); err != nil {
		return 0, err
	}

	var appliedDivider int
	if err := ad.dwfCall("FDwfDigitalInDividerGet", FDwfDigitalInDividerGet(ad.Handle, &appliedDivider)); err != nil {
		return 0, err
	}
	if appliedDivider <= 0 {
		return 0, fmt.Errorf("invalid divider applied: %d", appliedDivider)
	}
	appliedRateHz := int(math.Round(req.hardwareClock / float64(appliedDivider)))
	if appliedRateHz <= 0 {
		return 0, fmt.Errorf("invalid applied sample rate: %d", appliedRateHz)
	}

	var bufferMin int
	var bufferMax int
	if err := ad.dwfCall("FDwfDigitalInBufferSizeInfo", FDwfDigitalInBufferSizeInfo(ad.Handle, &bufferMin, &bufferMax)); err != nil {
		return 0, err
	}
	bufferSize := bufferMax
	if bufferSize > bufferMax {
		bufferSize = bufferMax
	}
	if bufferSize < bufferMin {
		bufferSize = bufferMin
	}
	if err := ad.dwfCall("FDwfDigitalInBufferSizeSet", FDwfDigitalInBufferSizeSet(ad.Handle, bufferSize)); err != nil {
		return 0, err
	}

	if req.Trigger.Type == "immediate" {
		if err := ad.dwfCall("FDwfDigitalInTriggerSourceSet", FDwfDigitalInTriggerSourceSet(ad.Handle, dwfTrigSrcNone)); err != nil {
			return 0, err
		}
		return appliedRateHz, nil
	}

	edgeMask := uint32(1 << req.Trigger.Channel)
	var edgeRise uint32
	var edgeFall uint32
	var slope int

	switch req.Trigger.Edge {
	case "rising":
		edgeRise = edgeMask
		edgeFall = 0
		slope = dwfTrigSlopeRise
	case "falling":
		edgeRise = 0
		edgeFall = edgeMask
		slope = dwfTrigSlopeFall
	default:
		edgeRise = edgeMask
		edgeFall = edgeMask
		slope = dwfTrigSlopeEither
	}

	if err := ad.dwfCall("FDwfDigitalInTriggerSourceSet", FDwfDigitalInTriggerSourceSet(ad.Handle, dwfTrigSrcDetectorDigitalIn)); err != nil {
		return 0, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerSlopeSet", FDwfDigitalInTriggerSlopeSet(ad.Handle, slope)); err != nil {
		return 0, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerPositionSet", FDwfDigitalInTriggerPositionSet(ad.Handle, 0)); err != nil {
		return 0, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerPrefillSet", FDwfDigitalInTriggerPrefillSet(ad.Handle, 0)); err != nil {
		return 0, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerAutoTimeoutSet", FDwfDigitalInTriggerAutoTimeoutSet(ad.Handle, 0)); err != nil {
		return 0, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerSet", FDwfDigitalInTriggerSet(ad.Handle, 0, 0, edgeRise, edgeFall)); err != nil {
		return 0, err
	}

	return appliedRateHz, nil
}

func (ad *AnalogDiscoveryDevice) collectLogicSamples(req normalizedLogicCaptureRequest, targetSamples int, triggered bool, triggerDeadline time.Time) ([]uint16, bool, []string, error) {
	samples := make([]uint16, 0, targetSamples)
	totalLost := 0
	totalCorrupt := 0
	readBuf := make([]byte, logicAnalyzerReadChunkSamples*2)

	for len(samples) < targetSamples {
		var status byte
		if err := ad.dwfCall("FDwfDigitalInStatus", FDwfDigitalInStatus(ad.Handle, 1, &status)); err != nil {
			return nil, false, nil, err
		}

		var dataAvailable int
		var dataLost int
		var dataCorrupt int
		if err := ad.dwfCall("FDwfDigitalInStatusRecord", FDwfDigitalInStatusRecord(ad.Handle, &dataAvailable, &dataLost, &dataCorrupt)); err != nil {
			return nil, false, nil, err
		}

		if dataLost > 0 {
			totalLost += dataLost
		}
		if dataCorrupt > 0 {
			totalCorrupt += dataCorrupt
		}

		for dataAvailable > 0 && len(samples) < targetSamples {
			remaining := targetSamples - len(samples)
			toRead := dataAvailable
			if toRead > remaining {
				toRead = remaining
			}
			if toRead > logicAnalyzerReadChunkSamples {
				toRead = logicAnalyzerReadChunkSamples
			}

			if err := ad.dwfCall("FDwfDigitalInStatusData", FDwfDigitalInStatusData(ad.Handle, &readBuf[0], toRead)); err != nil {
				return nil, false, nil, err
			}

			for i := 0; i < toRead; i++ {
				samples = append(samples, binary.LittleEndian.Uint16(readBuf[i*2:(i+1)*2]))
			}

			dataAvailable -= toRead
			if !triggered {
				triggered = true
				log.Printf("logic analyzer edge trigger fired")
			}
		}

		if !triggered && req.Trigger.Type == "edge" && time.Now().After(triggerDeadline) {
			warnings := logicCaptureWarnings(totalLost, totalCorrupt, len(samples))
			return nil, false, warnings, nil
		}

		if dataAvailable <= 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}

	warnings := logicCaptureWarnings(totalLost, totalCorrupt, len(samples))
	log.Printf("logic analyzer record counters: lost=%d corrupt=%d captured=%d", totalLost, totalCorrupt, len(samples))

	return samples, triggered, warnings, nil
}

func logicCaptureWarnings(totalLost int, totalCorrupt int, captured int) []string {
	warnings := []string{}
	if totalLost > 0 {
		warnings = append(warnings, "sample data lost")
	}

	if totalCorrupt > 0 {
		thresholdByRatio := int(float64(maxInt(captured, 1)) * logicAnalyzerCorruptWarnMinRatio)
		if thresholdByRatio < logicAnalyzerCorruptWarnMinCount {
			thresholdByRatio = logicAnalyzerCorruptWarnMinCount
		}
		if totalCorrupt >= thresholdByRatio {
			warnings = append(warnings, "sample data corrupt")
		}
	}
	return warnings
}

func buildLogicTransitions(samples []uint16, sampleRateHz int, channels []int) []LogicCaptureChannel {
	result := make([]LogicCaptureChannel, 0, len(channels))

	for _, channel := range channels {
		channelTransitions := make([]LogicCaptureTransition, 0)
		firstValue := bitValue(samples[0], channel)
		channelTransitions = append(channelTransitions, LogicCaptureTransition{
			TUs:   0,
			Value: firstValue,
		})

		previousValue := firstValue
		for sampleIndex := 1; sampleIndex < len(samples); sampleIndex++ {
			currentValue := bitValue(samples[sampleIndex], channel)
			if currentValue == previousValue {
				continue
			}
			channelTransitions = append(channelTransitions, LogicCaptureTransition{
				TUs:   int64(sampleIndex) * 1_000_000 / int64(sampleRateHz),
				Value: currentValue,
			})
			previousValue = currentValue
		}

		result = append(result, LogicCaptureChannel{
			Channel:     channel,
			Transitions: channelTransitions,
		})
	}

	return result
}

func bitValue(word uint16, channel int) int {
	if ((word >> channel) & 1) == 1 {
		return 1
	}
	return 0
}

func (ad *AnalogDiscoveryDevice) getDigitalInClockAndDividerInfo() (float64, error) {
	var internalClock float64
	if err := ad.dwfCall("FDwfDigitalInInternalClockInfo", FDwfDigitalInInternalClockInfo(ad.Handle, &internalClock)); err != nil {
		return 0, err
	}
	if internalClock <= 0 {
		return 0, fmt.Errorf("invalid internal clock: %f", internalClock)
	}

	return internalClock, nil
}

func (ad *AnalogDiscoveryDevice) dwfCall(op string, result int32) error {
	if result != 0 {
		return nil
	}
	message := getLastDwfErrorMessage()
	if message == "" {
		message = "unknown DWF error"
	}
	return DeviceRuntimeError{Message: fmt.Sprintf("%s failed: %s", op, message)}
}

func getLastDwfErrorMessage() string {
	buf := make([]byte, 512)
	FDwfGetLastErrorMsg(&buf[0])
	message := string(buf)
	if idx := strings.IndexByte(message, 0); idx >= 0 {
		message = message[:idx]
	}
	return strings.TrimSpace(message)
}

func deduplicateStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func mapKeys[K comparable, V any](m map[K]V) func(func(K) bool) {
	return func(yield func(K) bool) {
		for key := range m {
			if !yield(key) {
				return
			}
		}
	}
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
