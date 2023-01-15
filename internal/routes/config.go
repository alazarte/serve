package routes

import (
	"encoding/json"
	"fmt"
	"os"
)

func FromJson(jsonFilepath string) (Config, error) {
	var config Config

	f, err := os.ReadFile(jsonFilepath)
	if err != nil {
		return config, fmt.Errorf("Couldn't open config file: [file=%s, err=%s]", jsonFilepath, err)
	}

	if err := json.Unmarshal(f, &config); err != nil {
		return config, fmt.Errorf("Couldn't parse config file as json: [file=%s, err=%s]", jsonFilepath, err)
	}

	if _, err := os.Stat(config.Sk); err != nil {
		return config, fmt.Errorf("Couldn't open sk file: [sk=%s, err=%s]", config.Sk, err)
	}

	if _, err := os.Stat(config.Pem); err != nil {
		return config, fmt.Errorf("Couldn't open pem file: [pem=%s, err=%s]", config.Pem, err)
	}

	if len(config.Handlers) == 0 {
		return config, fmt.Errorf("No handlers defined, nothing to do...")
	}

	return config, nil
}
