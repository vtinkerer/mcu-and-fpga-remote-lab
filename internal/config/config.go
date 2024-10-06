package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	IS_ANALOG_DISCOVERY_MOCKED bool
	PORT string
	BOOT0_PIN int
	RESET_PIN int
}

func LoadConfig() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{}

	IS_ANALOG_DISCOVERY_MOCKED, err :=  strconv.ParseBool(os.Getenv("IS_ANALOG_DISCOVERY_MOCKED"))
	if err != nil {
		return nil, fmt.Errorf("Error parsing IS_ANALOG_DISCOVERY_MOCKED: %w", err)
	}
	config.IS_ANALOG_DISCOVERY_MOCKED = IS_ANALOG_DISCOVERY_MOCKED

	
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

	return config, nil
}