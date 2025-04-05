package potentiometer

import "fmt"

const maxResistance = 10e3
const step = maxResistance / 255

type Potentiometer struct {
	driver *driverMAX5395
	tapSelected uint8
}

func NewPotentiometer() (*Potentiometer, error) {
	// Find the right config here
	driver, err := New("1", addrGND)
	if err != nil {
		return nil, err
	}

	return &Potentiometer{
		driver: driver,
		tapSelected: 0,
	}, nil
}

func (p *Potentiometer) SetResistance(resistance float64) (float64, error) {
	tap := calculateClosestTapForResistance(resistance)
	p.tapSelected = tap
	if err := p.driver.SetWiper(tap); err != nil {
		fmt.Println("error setting wiper: ", err)
		return 0, err
	}
	return calculateResistanceForTap(tap), nil
}

func (p *Potentiometer) GetResistance() float64 {
	return calculateResistanceForTap(p.tapSelected)
}

func calculateClosestTapForResistance(resistance float64) uint8 {
	taps := float64(resistance) / float64(step)
	fmt.Println("calculated taps: ", taps, "for resistance: ", resistance)
	if taps < 0 || taps > 255 {
		fmt.Println("taps out of range")
		return 0
	}
	return uint8(taps)
}

func calculateResistanceForTap(tap uint8) float64 {
	return float64(tap) * float64(step)
}
