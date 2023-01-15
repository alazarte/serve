package main

import (
	"flag"
	"fmt"
	"os"
	"serve/internal/logger"
	"serve/internal/routes"
)

var (
	configFilepath  = flag.String("config", "/etc/serve.json", "config filepath")
	verboseRequests = flag.Bool("verboseRequest", false, "dump requests to debug log")
)

func init() {
	flag.Parse()
}

func main() {
	// get config
	config, err := routes.FromJson(*configFilepath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// create route
	r := routes.New(logger.New(os.Stdout, logger.Info))
	// configure handlers?
	r.ConfigHandlers(config.Handlers)

	cerr := r.ListenTLS(config.Pem, config.Sk)

	for {
		fmt.Printf("server error: [err=%s]", <-cerr)
	}
}
