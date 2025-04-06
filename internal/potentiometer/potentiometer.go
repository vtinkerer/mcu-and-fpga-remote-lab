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

func (p *Potentiometer) SetResistancePercentage(percentage uint8) (uint8, error) {
	// Clamp the percentage to the range 0-100
	if percentage > 100 {
		percentage = 100
	}
	if percentage < 0 {
		percentage = 0
	}

	tap := calculateClosestTapForResistancePercentage(percentage)
	p.tapSelected = tap
	if err := p.driver.SetWiper(tap); err != nil {
		fmt.Println("error setting wiper: ", err)
		return 0, err
	}
	return calculateResistancePercentageForTap(tap), nil
}

func (p *Potentiometer) GetResistancePercentage() uint8 {
	return calculateResistancePercentageForTap(p.tapSelected)
}

func calculateClosestTapForResistancePercentage(percentage uint8) uint8 {
	taps := uint8(255 * percentage / 100)
	fmt.Println("calculated taps: ", taps, "for percentage: ", percentage)
	return taps
}

func calculateResistancePercentageForTap(tap uint8) uint8 {
	return uint8(tap * 100 / 255)
}
