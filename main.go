package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"serve/internal/logger"
	"serve/internal/routes"
)

type Config struct {
	Pem      string    `json:"pem"`
	Sk       string    `json:"sk"`
	Debug    string    `json:"debug"`
	Handlers []Handler `json:"handlers"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Handler struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Path    string   `json:"path"`
	Port    string   `json:"port"`
	Headers []Header `json:"headers"`
}

var (
	config Config

	configFilepath  = flag.String("config", "/etc/serve.json", "config filepath")
	verboseRequests = flag.Bool("verboseRequest", false, "dump requests to debug log")

	TypeRoot   = "root"
	TypePublic = "public"
	TypeProxy  = "proxy"
)

func getLogOutput() io.Writer {
	var debugOut io.Writer = io.Discard

	switch config.Debug {
	case "":
		debugOut = io.Discard
	default:
		f, err := os.OpenFile(config.Debug, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(fmt.Sprintln("couldn't open file for debug logs:", config.Debug))
		}
		debugOut = f
	}

	return debugOut
}

func init() {
	flag.Parse()

	f, err := os.ReadFile(*configFilepath)
	if err != nil {
		fmt.Printf("Couldn't open config file: [file=%s, err=%s]", *configFilepath, err)
		os.Exit(1)
	}
	if err := json.Unmarshal(f, &config); err != nil {
		fmt.Printf("Couldn't parse config file as json: [file=%s, err=%s]", *configFilepath, err)
		os.Exit(1)
	}
	if _, err := os.Stat(config.Sk); err != nil {
		fmt.Printf("Couldn't open sk file: [sk=%s, err=%s]", config.Sk, err)
		os.Exit(1)
	}
	if _, err := os.Stat(config.Pem); err != nil {
		fmt.Printf("Couldn't open pem file: [pem=%s, err=%s]", config.Pem, err)
		os.Exit(1)
	}
	if len(config.Handlers) == 0 {
		fmt.Printf("No handlers defined, nothing to do...")
		os.Exit(1)
	}
}

func main() {
	l := logger.New(os.Stdout, logger.Info)
	r := routes.New(l)

	for _, h := range config.Handlers {
		switch h.Type {
		case TypeRoot:
			extraHeaders := make(map[string]string)
			for _, h := range h.Headers {
				extraHeaders[http.CanonicalHeaderKey(h.Name)] = h.Value
			}
			r.HandleRoot(h.Name, h.Path, extraHeaders)
		case TypePublic:
			r.HandlePublicFiles(h.Name, h.Path)
		case TypeProxy:
			r.HandleProxy(h.Name, h.Path)
		default:
			fmt.Printf("Main: Handler type not recognized: [type=%s]", h.Type)
		}
	}

	cerr := r.ListenTLS(config.Pem, config.Sk)

	for {
		fmt.Printf("server error: [err=%s]", <-cerr)
	}
}
