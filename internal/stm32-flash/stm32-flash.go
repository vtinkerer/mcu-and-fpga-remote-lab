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

	result, err := runCommand("st-flash", "--reset", "--format", "ihex", "write", filePath)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	if strings.Contains(result, "ERROR") || strings.Contains(result, "Failed") {
		return fmt.Errorf("flash failed: %s", result)
	}
	fmt.Println("Flash successful")

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

// DEPRECATED VERSION:

// package stm32flash

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"os/exec"
// 	"strconv"
// 	"sync"
// 	"time"
// )

// var flashMutex sync.Mutex

// func Flash(filePath string, resetPin int, boot0Pin int) error {

// 	if !flashMutex.TryLock() {
// 		return fmt.Errorf("Already flashing")
// 	}
// 	defer flashMutex.Unlock()

// 	// Set up pin control
// 	if err := enterBootloader(resetPin, boot0Pin); err != nil {
// 		return fmt.Errorf("[stm32-flash] Failed to enter bootloader: %w", err)
// 	}

// 	// Flash the file using stm32flash
// 	err := runFlash(filePath, resetPin, boot0Pin, 5)

// 	if err != nil {
// 		return fmt.Errorf("[stm32-flash] Failed to flash: %w", err)
// 	}

// 	// Exit bootloader
// 	if err := Reset(resetPin, boot0Pin); err != nil {
// 		return fmt.Errorf("[stm32-flash] Failed to exit bootloader: %w", err)
// 	}

// 	fmt.Println("STM32 flashed successfully")

// 	return nil
// }

// func runFlash(filePath string, resetPin, boot0Pin int, attempts int) error {
// 	cmd := exec.Command("stm32flash", "-w", filePath, "-v", "-b", "115200", "-g", "0x0", "/dev/ttyACM0")
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	if err := cmd.Run(); err != nil {
// 		fmt.Println("Error flashing STM32:", err)
// 		if attempts == 0 {
// 			errReset := Reset(resetPin, boot0Pin)
// 			if errReset != nil {
// 				log.Println("Failed to exit bootloader during the failed bootloader exit:", errReset)
// 			}
// 			return fmt.Errorf("failed to flash after many attempts: %w", err)
// 		}
// 		return runFlash(filePath, resetPin, boot0Pin, attempts-1)
// 	}
// 	return nil
// }

// func enterBootloader(resetPin, boot0Pin int) error {
// 	// BOOT0 UP
// 	if err := runCommand("pinctrl", "set", strconv.Itoa(boot0Pin), "op", "dh"); err != nil {
// 		return fmt.Errorf("failed to set up boot0 pin: %w", err)
// 	}

// 	time.Sleep(10 * time.Millisecond)

// 	// NRST DOWN
// 	if err := runCommand("pinctrl", "set", strconv.Itoa(resetPin), "op", "dl"); err != nil {
// 		return fmt.Errorf("failed to set up reset pin: %w", err)
// 	}

// 	time.Sleep(10 * time.Millisecond)

// 	// NRST UP
// 	if err := runCommand("pinctrl", "set", strconv.Itoa(resetPin), "op", "dh"); err != nil {
// 		return fmt.Errorf("failed to set up reset pin: %w", err)
// 	}

// 	time.Sleep(100 * time.Millisecond)

// 	return nil
// }

// func Reset(resetPin, boot0Pin int) error {
// 	// BOOT0 DOWN
// 	if err := runCommand("pinctrl", "set", strconv.Itoa(boot0Pin), "op", "dl"); err != nil {
// 		return fmt.Errorf("failed to set up boot0 pin: %w", err)
// 	}

// 	time.Sleep(10 * time.Millisecond)

// 	// NRST DOWN //IF USED WITH TRANSISTOR, THEN IT'S REVERTED
// 	if err := runCommand("pinctrl", "set", strconv.Itoa(resetPin), "op", "dl"); err != nil {
// 		return fmt.Errorf("failed to set up reset pin: %w", err)
// 	}

// 	time.Sleep(10 * time.Millisecond)

// 	// NRST UP //IF USED WITH TRANSISTOR, THEN IT'S REVERTED
// 	if err := runCommand("pinctrl", "set", strconv.Itoa(resetPin), "op", "dh"); err != nil {
// 		return fmt.Errorf("failed to set up reset pin: %w", err)
// 	}

// 	time.Sleep(10 * time.Millisecond)

// 	return nil
// }

// func runCommand(name string, args ...string) error {
// 	cmd := exec.Command(name, args...)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	return cmd.Run()
// }
