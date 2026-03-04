package analogdiscovery

import (
	"errors"
	"fmt"
	"log"
	"math"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	logicAnalyzerDefaultSampleRateHz = 250_000
	logicAnalyzerMinSampleRateHz     = 1_000
	logicAnalyzerMinDurationSec      = 1
	logicAnalyzerMaxDurationSec      = 10
	logicAnalyzerMaxCaptureSamples   = 20_000_000
	logicAnalyzerDefaultTimeoutSec   = 10
	logicAnalyzerMaxUserSampleRateHz = 1_000_000
	logicAnalyzerCorruptWarnMinCount = 32
	logicAnalyzerCorruptWarnMinRatio = 0.0001
	logicAnalyzerIdleSleepMin        = 25 * time.Microsecond
	logicAnalyzerIdleSleepMax        = 200 * time.Microsecond
	logicAnalyzerIdleSpinPolls       = 4
	logicAnalyzerPollsPerBuffer      = 12
	logicAnalyzerSampleBytes         = 2 // 16-bit sample format
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
	FDwfDigitalInStatusData            func(deviceHandle int32, rgBytes *byte, countOfDataBytes int) int32
	FDwfDigitalInStatusRecord          func(deviceHandle int32, pDataAvailable *int, pDataLost *int, pDataCorrupt *int) int32
	FDwfDigitalInInternalClockInfo     func(deviceHandle int32, phzFreq *float64) int32
	FDwfDigitalInDividerInfo           func(deviceHandle int32, pMax *int) int32
	FDwfDigitalInDividerSet            func(deviceHandle int32, divider int) int32
	FDwfDigitalInDividerGet            func(deviceHandle int32, pDivider *int) int32
	FDwfDigitalInSampleFormatSet       func(deviceHandle int32, bits int) int32
	FDwfDigitalInBufferSizeInfo        func(deviceHandle int32, pMax *int) int32
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

type logicAnalyzerConfig struct {
	AppliedSampleRateHz int
	BufferSizeSamples   int
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

	config, err := ad.configureLogicAnalyzer(normalizedReq)
	if err != nil {
		return captureResponse, err
	}
	if config.AppliedSampleRateHz <= 0 {
		return captureResponse, DeviceRuntimeError{Message: "invalid applied sample rate"}
	}

	captureResponse.SampleRateHz = config.AppliedSampleRateHz

	targetSamples := normalizedReq.DurationSec * config.AppliedSampleRateHz
	if targetSamples > logicAnalyzerMaxCaptureSamples {
		targetSamples = logicAnalyzerMaxCaptureSamples
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
	warningsSet := warningSetFromSlice(warnings)

	samples, triggered, loopWarnings, err := ad.collectLogicSamples(normalizedReq, config, targetSamples, triggered, triggerDeadline)
	if err != nil {
		return captureResponse, err
	}
	warningSetAddAll(warningsSet, loopWarnings)

	captureResponse.Warnings = warningSetSortedSlice(warningsSet)

	if !triggered {
		captureResponse.Warnings = appendUniqueWarning(captureResponse.Warnings, "edge trigger timeout")
		log.Printf("logic analyzer trigger timeout")
		return captureResponse, nil
	}

	if len(samples) == 0 {
		captureResponse.Warnings = appendUniqueWarning(captureResponse.Warnings, "no samples captured")
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

func (ad *AnalogDiscoveryDevice) configureLogicAnalyzer(req normalizedLogicCaptureRequest) (logicAnalyzerConfig, error) {
	if err := ad.dwfCall("FDwfDigitalInReset", FDwfDigitalInReset(ad.Handle)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if err := ad.dwfCall("FDwfDigitalInSampleFormatSet", FDwfDigitalInSampleFormatSet(ad.Handle, 16)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if err := ad.dwfCall("FDwfDigitalInAcquisitionModeSet", FDwfDigitalInAcquisitionModeSet(ad.Handle, dwfAcqModeRecord)); err != nil {
		return logicAnalyzerConfig{}, err
	}

	divider := int(math.Round(req.hardwareClock / float64(req.SampleRateHz)))
	if divider < 1 {
		divider = 1
	}
	if err := ad.dwfCall("FDwfDigitalInDividerSet", FDwfDigitalInDividerSet(ad.Handle, divider)); err != nil {
		return logicAnalyzerConfig{}, err
	}

	var appliedDivider int
	if err := ad.dwfCall("FDwfDigitalInDividerGet", FDwfDigitalInDividerGet(ad.Handle, &appliedDivider)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if appliedDivider <= 0 {
		return logicAnalyzerConfig{}, fmt.Errorf("invalid divider applied: %d", appliedDivider)
	}
	appliedRateHz := int(math.Round(req.hardwareClock / float64(appliedDivider)))
	if appliedRateHz <= 0 {
		return logicAnalyzerConfig{}, fmt.Errorf("invalid applied sample rate: %d", appliedRateHz)
	}

	var bufferMax int
	if err := ad.dwfCall("FDwfDigitalInBufferSizeInfo", FDwfDigitalInBufferSizeInfo(ad.Handle, &bufferMax)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if bufferMax <= 0 {
		return logicAnalyzerConfig{}, fmt.Errorf("invalid buffer size max: %d", bufferMax)
	}
	bufferSize := bufferMax
	if err := ad.dwfCall("FDwfDigitalInBufferSizeSet", FDwfDigitalInBufferSizeSet(ad.Handle, bufferSize)); err != nil {
		return logicAnalyzerConfig{}, err
	}

	config := logicAnalyzerConfig{
		AppliedSampleRateHz: appliedRateHz,
		BufferSizeSamples:   bufferSize,
	}

	if req.Trigger.Type == "immediate" {
		if err := ad.dwfCall("FDwfDigitalInTriggerSourceSet", FDwfDigitalInTriggerSourceSet(ad.Handle, dwfTrigSrcNone)); err != nil {
			return logicAnalyzerConfig{}, err
		}
		return config, nil
	}

	edgeRise, edgeFall, slope := digitalTriggerEdgeConfig(req.Trigger.Channel, req.Trigger.Edge)

	if err := ad.dwfCall("FDwfDigitalInTriggerSourceSet", FDwfDigitalInTriggerSourceSet(ad.Handle, dwfTrigSrcDetectorDigitalIn)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerSlopeSet", FDwfDigitalInTriggerSlopeSet(ad.Handle, slope)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerPositionSet", FDwfDigitalInTriggerPositionSet(ad.Handle, 0)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerPrefillSet", FDwfDigitalInTriggerPrefillSet(ad.Handle, 0)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerAutoTimeoutSet", FDwfDigitalInTriggerAutoTimeoutSet(ad.Handle, 0)); err != nil {
		return logicAnalyzerConfig{}, err
	}
	if err := ad.dwfCall("FDwfDigitalInTriggerSet", FDwfDigitalInTriggerSet(ad.Handle, 0, 0, edgeRise, edgeFall)); err != nil {
		return logicAnalyzerConfig{}, err
	}

	return config, nil
}

func (ad *AnalogDiscoveryDevice) collectLogicSamples(req normalizedLogicCaptureRequest, config logicAnalyzerConfig, targetSamples int, triggered bool, triggerDeadline time.Time) ([]uint16, bool, []string, error) {
	// Minimize scheduler/GC jitter while draining the small device buffer at high sample rates.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	prevGCPercent := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(prevGCPercent)

	captureStartedAt := time.Now()
	samples := make([]uint16, targetSamples)
	sampleOffset := 0
	totalLost := 0
	totalCorrupt := 0
	readOps := 0
	totalReadSamples := 0
	idleSleeps := 0
	bufferWindow := bufferWindowDuration(config.BufferSizeSamples, config.AppliedSampleRateHz)
	adaptiveIdleMax := logicAnalyzerIdleSleepMax
	adaptiveIdleMin := logicAnalyzerIdleSleepMin
	if bufferWindow > 0 {
		pollTarget := bufferWindow / logicAnalyzerPollsPerBuffer
		if pollTarget > 0 && pollTarget < adaptiveIdleMax {
			adaptiveIdleMax = pollTarget
		}
		if adaptiveIdleMax < time.Microsecond {
			adaptiveIdleMax = time.Microsecond
		}
		if adaptiveIdleMin > adaptiveIdleMax {
			adaptiveIdleMin = adaptiveIdleMax
		}
	}
	idleSleep := adaptiveIdleMin
	maxIdleSleep := time.Duration(0)
	consecutiveIdlePolls := 0

	log.Printf("logic analyzer capture loop started: requestedRate=%dHz configuredRate=%dHz targetSamples=%d bufferSize=%d",
		req.SampleRateHz, config.AppliedSampleRateHz, targetSamples, config.BufferSizeSamples)

	for sampleOffset < targetSamples {
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

		readSamplesThisPoll := 0
		for dataAvailable > 0 && sampleOffset < targetSamples {
			remaining := targetSamples - sampleOffset
			toRead := dataAvailable
			if toRead > remaining {
				toRead = remaining
			}

			bytesToRead := toRead * logicAnalyzerSampleBytes
			destBytes := unsafe.Slice((*byte)(unsafe.Pointer(&samples[sampleOffset])), bytesToRead)
			if err := ad.dwfCall("FDwfDigitalInStatusData", FDwfDigitalInStatusData(ad.Handle, &destBytes[0], bytesToRead)); err != nil {
				log.Printf("logic analyzer read error: op=FDwfDigitalInStatusData toReadSamples=%d toReadBytes=%d sampleOffset=%d targetSamples=%d configuredRate=%dHz err=%v",
					toRead, bytesToRead, sampleOffset, targetSamples, config.AppliedSampleRateHz, err)
				return nil, false, nil, err
			}

			sampleOffset += toRead
			readOps++
			totalReadSamples += toRead
			readSamplesThisPoll += toRead

			dataAvailable -= toRead
			if !triggered {
				triggered = true
				log.Printf("logic analyzer edge trigger fired")
			}
		}

		if !triggered && req.Trigger.Type == "edge" && time.Now().After(triggerDeadline) {
			warnings := logicCaptureWarnings(totalLost, totalCorrupt, sampleOffset)
			return nil, false, warnings, nil
		}

		if readSamplesThisPoll == 0 {
			// Avoid adding latency immediately after a productive poll. Short spin polls reduce
			// risk of overrunning small capture buffers at high sample rates.
			if consecutiveIdlePolls >= logicAnalyzerIdleSpinPolls {
				idleSleeps++
				if idleSleep <= time.Microsecond {
					runtime.Gosched()
				} else {
					time.Sleep(idleSleep)
				}
				if idleSleep > maxIdleSleep {
					maxIdleSleep = idleSleep
				}
				if idleSleep < adaptiveIdleMax {
					idleSleep *= 2
					if idleSleep > adaptiveIdleMax {
						idleSleep = adaptiveIdleMax
					}
				}
			}
			consecutiveIdlePolls++
		} else {
			consecutiveIdlePolls = 0
			idleSleep = adaptiveIdleMin
		}
	}

	warnings := logicCaptureWarnings(totalLost, totalCorrupt, len(samples))
	avgRead := 0
	if readOps > 0 {
		avgRead = totalReadSamples / readOps
	}
	elapsed := time.Since(captureStartedAt)
	readRate := 0.0
	if elapsed > 0 {
		readRate = float64(totalReadSamples) / elapsed.Seconds()
	}
	log.Printf("logic analyzer capture telemetry: lost=%d corrupt=%d captured=%d readOps=%d avgReadSamples=%d idleSleeps=%d maxIdleSleep=%s elapsed=%s readRate=%.0fS/s",
		totalLost, totalCorrupt, len(samples), readOps, avgRead, idleSleeps, maxIdleSleep, elapsed, readRate)

	return samples, triggered, warnings, nil
}

func bufferWindowDuration(bufferSizeSamples int, sampleRateHz int) time.Duration {
	if bufferSizeSamples <= 0 || sampleRateHz <= 0 {
		return 0
	}
	windowUs := int64(bufferSizeSamples) * 1_000_000 / int64(sampleRateHz)
	if windowUs <= 0 {
		return time.Microsecond
	}
	return time.Duration(windowUs) * time.Microsecond
}

func logicCaptureWarnings(totalLost int, totalCorrupt int, captured int) []string {
	warnings := []string{}
	capturedSafe := maxInt(captured, 1)
	if totalLost > 0 {
		lostPct := float64(totalLost) * 100 / float64(capturedSafe)
		warnings = append(warnings, fmt.Sprintf("sample data lost: %d (%.6f%%)", totalLost, lostPct))
	}

	if totalCorrupt > 0 {
		thresholdByRatio := int(float64(capturedSafe) * logicAnalyzerCorruptWarnMinRatio)
		if thresholdByRatio < logicAnalyzerCorruptWarnMinCount {
			thresholdByRatio = logicAnalyzerCorruptWarnMinCount
		}
		if totalCorrupt >= thresholdByRatio {
			corruptPct := float64(totalCorrupt) * 100 / float64(capturedSafe)
			warnings = append(warnings, fmt.Sprintf("sample data corrupt: %d (%.6f%%)", totalCorrupt, corruptPct))
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

func warningSetFromSlice(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	warningSetAddAll(set, values)
	return set
}

func warningSetAddAll(set map[string]struct{}, values []string) {
	for _, value := range values {
		set[value] = struct{}{}
	}
}

func warningSetSortedSlice(set map[string]struct{}) []string {
	if len(set) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	slices.Sort(result)
	return result
}

func appendUniqueWarning(warnings []string, warning string) []string {
	for _, existing := range warnings {
		if existing == warning {
			return warnings
		}
	}
	return append(warnings, warning)
}

func digitalTriggerEdgeConfig(channel int, edge string) (uint32, uint32, int) {
	edgeMask := uint32(1 << channel)
	switch edge {
	case "rising":
		return edgeMask, 0, dwfTrigSlopeRise
	case "falling":
		return 0, edgeMask, dwfTrigSlopeFall
	default:
		return edgeMask, edgeMask, dwfTrigSlopeEither
	}
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
