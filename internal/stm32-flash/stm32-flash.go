package stm32flash

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var flashMutex sync.Mutex

func Flash(filePath string) error {
	if !flashMutex.TryLock() {
		return fmt.Errorf("flash is already in progress")
	}
	defer flashMutex.Unlock()

	result, err := runCommand("st-flash", "--format", "ihex", "write", filePath)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	if strings.Contains(result, "ERROR") || strings.Contains(result, "Failed") {
		return fmt.Errorf("flash failed: %s", result)
	}
	fmt.Println("Flash successful")
	fmt.Println("Resetting device")

	return nil
}


func Reset() error {
	result, err := runCommand("st-flash", "reset")
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	if strings.Contains(result, "ERROR") || strings.Contains(result, "Failed") {
		return fmt.Errorf("reset failed: %s", result)
	}
	return nil
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	fmt.Println("Running command:", cmd.String())
	result, err := cmd.CombinedOutput()
	strResult := string(result)
	fmt.Println("Command result:", strResult)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}
	return strResult, nil
}