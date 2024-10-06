package analogdiscoverysubprocess

import "fmt"

func SetPinCommand(pinNumber int, value int) string {
	return fmt.Sprintf("set=%d:%d", pinNumber, value)
}