package analogdiscoverysubprocess

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// RealProcess struct
type realProcess struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stdin  io.WriteCloser
	stderr io.ReadCloser
}

func getPythonProcessPath() string {
	exe, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(exe, "subprocess", "main.py")
}


func newRealProcess() *realProcess {
	path := getPythonProcessPath()

	fmt.Println("Path: ", path)

	cmd := exec.Command("python3", getPythonProcessPath())
	return &realProcess{cmd: cmd}
}

func (rp *realProcess) Start() error {
	var err error
	rp.stdout, err = rp.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}
	rp.stdin, err = rp.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error creating stdin pipe: %v", err)
	}
	rp.stderr, err = rp.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %v", err)
	}

	fmt.Println("Starting process", rp)
	return rp.cmd.Start()
}

func (rp *realProcess) ReadOutput(resultChan chan<- string) {
	scanner := bufio.NewScanner(rp.stdout)
	for scanner.Scan() {
		resultChan <- string(scanner.Bytes())
	}
}

func (rp *realProcess) ReadError(resultChan chan<- string) {
	scanner := bufio.NewScanner(rp.stderr)
	for scanner.Scan() {
		resultChan <- string(scanner.Bytes())
	}
}

func (rp *realProcess) Stop() error {
	return rp.cmd.Process.Kill()
}

func (rp *realProcess) WriteInput(input string) error {
	fmt.Println("Writing input: ", input)
	_, err := fmt.Fprintln(rp.stdin, input)
	return err
}