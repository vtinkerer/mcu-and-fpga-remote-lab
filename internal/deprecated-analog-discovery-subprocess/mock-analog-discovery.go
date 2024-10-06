package analogdiscoverysubprocess

import (
	"fmt"
	"time"
)

// MockProcess struct
type mockProcess struct {
	stopChan chan struct{}
}

func newMockProcess() *mockProcess {
	return &mockProcess{stopChan: make(chan struct{})}
}

func (mp *mockProcess) Start() error {
	return nil
}

func (mp *mockProcess) ReadError(errorChan chan<- string) {
}

func (mp *mockProcess) ReadOutput(resultChan chan<- string) {
	for {
		select {
		case <-mp.stopChan:
			return
		default:
			resultChan <- fmt.Sprintf("Mock command at %v", time.Now())
			time.Sleep(300 * time.Millisecond)
		}
	}
}

func (mp *mockProcess) Stop() error {
	close(mp.stopChan)
	return nil
}

func (mp *mockProcess) WriteInput(input string) error {
	return nil
}