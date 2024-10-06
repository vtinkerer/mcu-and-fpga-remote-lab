package analogdiscoverysubprocess

import (
	"fmt"
)


func StartSubprocess(useMock bool) (stdin chan<- string , stdout <-chan string, stderr <-chan string, err error) {
	var process ProcessRunner

	if useMock {
		fmt.Println("Using mock process")
		process = newMockProcess()
	} else  {
		fmt.Println("Using real process")
		process = newRealProcess()
	}

	if err := process.Start(); err != nil {
		return nil, nil, nil, fmt.Errorf("Process start error: %w", err)
	}

	resultChan := make(chan string)
	inputChan := make(chan string)
	errorChan := make(chan string)

	go process.ReadOutput(resultChan)
	go func(input <-chan string, process ProcessRunner) {
		for result := range input {
			process.WriteInput(result)
		}
	}(inputChan, process)
	go process.ReadError(errorChan)

	return inputChan, resultChan, errorChan, nil
}