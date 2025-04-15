package gpio

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)


func WritePin(pinNumber int, value int) error {
	if value != 0 {
		if err := runCommand("pinctrl", "set", strconv.Itoa(pinNumber), "op", "dh"); err != nil {
			return fmt.Errorf("failed to set up boot0 pin: %w", err)
		}
		return nil
	}

	if err := runCommand("pinctrl", "set", strconv.Itoa(pinNumber), "op", "dl"); err != nil {
		return fmt.Errorf("failed to set up boot0 pin: %w", err)
	}
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}