package fpga

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type FPGA struct {
	TDI int
	TMS int
	TCK int
	TDO int
}

func CreateFPGA(TDI, TDO, TCK, TMS int) *FPGA {
	return &FPGA{
		TDI: TDI,
		TMS: TMS,
		TCK: TCK,
		TDO: TDO,
	}
}

var flashMutex sync.Mutex

func (fpga *FPGA) Flash(svfFilePath string) error {
	if !flashMutex.TryLock() {
		fmt.Println("Failed to lock flash mutex")
		return fmt.Errorf("Already flashing")
	}
	fmt.Println("Locked flash mutex")
	defer func() {
		fmt.Println("Unlocking flash mutex")
		flashMutex.Unlock()
	}()

	return fpga.runUrjtag(svfFilePath)
}

func (fpga *FPGA) runUrjtag(svfFilePath string) error {
	// Start the urjtag process
	cmd := exec.Command("/home/pi/urjtag-2021.03/src/apps/jtag/jtag")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("Failed to get stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Failed to get stderr pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Failed to start urjtag: %w", err)
	}

	resultChan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("stdout:", line)
			if strings.Contains(line, "Scanned device output matched expected TDO values") {
				resultChan <- nil
				return
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("stderr:", line)
			if !strings.HasPrefix(line, "warning:") {
				resultChan <- fmt.Errorf("urjtag stderr: %s", line)
				return
			}
		}
	}()

	// Send commands to urjtag
	commands := []string{
		fmt.Sprintf("cable gpio tdi=%d tdo=%d tck=%d tms=%d", fpga.TDI, fpga.TDO, fpga.TCK, fpga.TMS),
		"detect",
		"idcode",
		"include /home/pi/EP4CE10E22.bsdl",
		"svf " + svfFilePath + " progress",
		"quit",
	}

	for _, cmd := range commands {
		fmt.Println("Sending command:", cmd)
		io.WriteString(stdin, cmd+"\n")
	}

	result := <-resultChan

	cmd.Process.Kill()
	cmd.Wait()

	return result
}