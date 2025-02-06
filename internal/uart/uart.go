package uart

import (
	"fmt"
	"log"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type UART struct  {
	port serial.Port
	ReadChannel chan []byte
}

func logPortsInfo() {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		fmt.Println("No serial ports found!")
	}
	for _, port := range ports {
		fmt.Printf("Found port: %s\n", port.Name)
		if port.IsUSB {
			fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
			fmt.Printf("   USB serial %s\n", port.SerialNumber)
		}
	}
}

func NewUART() (*UART, error) {
	logPortsInfo()

	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open("/dev/ttyACM0", mode)
	if err != nil {
		return nil, err
	}

	return &UART{port: port}, nil
}

func (u *UART) StartListening() {
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := u.port.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Read %d bytes: %s\n", n, buf[:n])
			u.ReadChannel <- buf[:n]
		}
	}()
}

func (u *UART) Write(data []byte) {
	n, err := u.port.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote %d bytes: %s\n", n, data)
}

func (u *UART) Close() {
	fmt.Println("Closing UART port")
	u.port.Close()
}