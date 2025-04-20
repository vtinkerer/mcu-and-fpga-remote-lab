package pcbswitch

import (
	"time"

	"digitrans-lab-go/internal/gpio"
)

type PCBSwitch struct {
	pin int
}

func NewPCBSwitch(pin int) *PCBSwitch {
	return &PCBSwitch{
		pin: pin,
	}
}

func (s *PCBSwitch) PowerOn() error {
	return gpio.WritePin(s.pin, 1)
}

func (s *PCBSwitch) PowerOff() error {
	return gpio.WritePin(s.pin, 0)
}

func (s *PCBSwitch) Reset() error {

	s.PowerOff()
	time.Sleep(1 * time.Second)
	s.PowerOn()

	return nil
}