package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT string
	BOOT0_PIN int
	RESET_PIN int
	TDI int
	TMS int
	TCK int
	TDO int
}

func LoadConfig() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{}
	
	PORT, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing PORT: %d %w", PORT, err)
	}
	config.PORT = os.Getenv("PORT")


	BOOT0_PIN, err := strconv.Atoi(os.Getenv("BOOT0_PIN"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing BOOT0_PIN: %w", err)
	}
	config.BOOT0_PIN = BOOT0_PIN


	RESET_PIN, err := strconv.Atoi(os.Getenv("NRST_PIN"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing NRST_PIN: %w", err)
	}
	config.RESET_PIN = RESET_PIN


	TDI, err := strconv.Atoi(os.Getenv("TDI"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing TDI: %w", err)
	}
	config.TDI = TDI

	TMS, err := strconv.Atoi(os.Getenv("TMS"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing TMS: %w", err)
	}
	config.TMS = TMS

	TCK, err := strconv.Atoi(os.Getenv("TCK"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing TCK: %w", err)
	}
	config.TCK = TCK

	TDO, err := strconv.Atoi(os.Getenv("TDO"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing TDO: %w", err)
	}
	config.TDO = TDO

	return config, nil
}