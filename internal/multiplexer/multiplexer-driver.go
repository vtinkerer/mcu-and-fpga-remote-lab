package multiplexer

import (
	"digitrans-lab-go/internal/gpio"
	"fmt"
)

type driverMultiplexer struct {
	a1PinNumber int
	a2PinNumber int

	a1Val int
	a2Val int
}

func newDriverMultiplexer(a1PinNumber int, a2PinNumber int) *driverMultiplexer {
	driver := &driverMultiplexer{
		a1PinNumber: a1PinNumber,
		a2PinNumber: a2PinNumber,
	}

	driver.SelectInputChannel(1)

	return driver
}

func (d *driverMultiplexer) setA1(val int) {
	d.a1Val = val
	gpio.WritePin(d.a1PinNumber, val)
}

func (d *driverMultiplexer) setA2(val int) {
	d.a2Val = val
	gpio.WritePin(d.a2PinNumber, val)
}

func (d *driverMultiplexer) SelectInputChannel(channel int) error {
	if channel < 1 || channel > 4 {
		return fmt.Errorf("invalid channel: %d", channel)
	}

	channel = channel - 1

	d.setA1(channel & 1)
	d.setA2(channel & 2)

	return nil
}

func (d *driverMultiplexer) GetInputChannel() (int, error) {
	if d.a1Val == 0 && d.a2Val == 0 {
		return 1, nil
	}
	if d.a1Val == 0 && d.a2Val == 1 {
		return 2, nil
	}
	if d.a1Val == 1 && d.a2Val == 0 {
		return 3, nil
	}
	if d.a1Val == 1 && d.a2Val == 1 {
		return 4, nil
	}

	return 0, fmt.Errorf("invalid channel: %d", d.a1Val, d.a2Val)
}
