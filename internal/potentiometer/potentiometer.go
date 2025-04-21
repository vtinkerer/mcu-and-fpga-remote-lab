package potentiometer

import (
	"fmt"
	"math"
)

const maxResistance = 10e3
const step = maxResistance / 255

type Potentiometer struct {
	driver *driverMAX5395
	tapSelected uint8
}

func NewPotentiometer() (*Potentiometer, error) {
	// Find the right config here
	driver, err := newDriver("/dev/i2c-1", addrGND)
	if err != nil {
		return nil, err
	}

	return &Potentiometer{
		driver: driver,
		tapSelected: 0,
	}, nil
}

func (p *Potentiometer) SetResistancePercentage(percentage int) (int, error) {
	// Clamp the percentage to the range 0-100
	if percentage > 100 {
		percentage = 100
	}
	if percentage < 0 {
		percentage = 0
	}

	tap := calculateClosestTapForResistancePercentage(percentage)
	p.tapSelected = tap
	if err := p.driver.setWiper(tap); err != nil {
		fmt.Println("error setting wiper: ", err)
		return 0, err
	}
	return calculateResistancePercentageForTap(tap), nil
}

func (p *Potentiometer) GetResistancePercentage() int {
	wiper, err := p.driver.getWiper()
	if err == nil {
		fmt.Println("wiper: ", wiper)
	} else {
		fmt.Println("error getting wiper: ", err)
	}
	return calculateResistancePercentageForTap(p.tapSelected)
}

func calculateClosestTapForResistancePercentage(percentage int) uint8 {
	taps := uint8(255 * percentage / 100)
	fmt.Println("calculated taps: ", taps, "for percentage: ", percentage)
	return taps
}

func calculateResistancePercentageForTap(tap uint8) int {
	return int(math.Round(float64(tap) * 100 / 255))
}
