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

	driver.selectInputChannel(1)

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

func (d *driverMultiplexer) selectInputChannel(channel int) error {
	if channel < 1 || channel > 4 {
		return fmt.Errorf("invalid channel: %d", channel)
	}

	channel = channel - 1

	a1Val := ((channel - 1) >> 0) & 1
	a2Val := ((channel - 1) >> 1) & 1

	fmt.Printf("setting a1Val: %d, a2Val: %d\n", a1Val, a2Val)

	d.setA1(a1Val)
	d.setA2(a2Val)

	return nil
}

func (d *driverMultiplexer) getInputChannel() (int, error) {
	fmt.Printf("getting a1Val: %d, a2Val: %d\n", d.a1Val, d.a2Val)

	if d.a2Val == 0 && d.a1Val == 0 {
		return 1, nil
	}
	if d.a2Val == 0 && d.a1Val == 1 {
		return 2, nil
	}
	if d.a2Val == 1 && d.a1Val == 0 {
		return 3, nil
	}
	if d.a2Val == 1 && d.a1Val == 1 {
		return 4, nil
	}

	return 0, fmt.Errorf("invalid channel: %d", d.a1Val, d.a2Val)
}
