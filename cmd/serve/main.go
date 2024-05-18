package main

import (
	"os"
	"flag"
	"encoding/json"
	"serve/internal/serve"
)

var (
	configFilepath string
)

func init() {
	flag.StringVar(&configFilepath, "config", "config.json", "Config filepath")
	flag.Parse()
}

func main() {
	contents, err := os.ReadFile(configFilepath)
	if err != nil {
		panic(err)
	}

	config := serve.HandlerConfig{}
	if err := json.Unmarshal(contents, &config); err != nil {
		panic(err)
	}

	serve.Listen(config)
}
