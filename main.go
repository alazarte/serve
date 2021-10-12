package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"serve/internal/routes"
	"serve/internal/utils"
)

type Config struct {
	Pem      string    `json:"pem"`
	Sk       string    `json:"sk"`
	Debug    string    `json:"debug"`
	Handlers []Handler `json:"handlers"`
}

type KeyVal struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Handler struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Path    string   `json:"path"`
	Port    string   `json:"port"`
	Headers []KeyVal `json:"headers"`
}

var (
	logger utils.Logger

	postUrl *url.URL

	config Config

	debugFile       = flag.String("debug", "", "filepath to print debug logs to, default is io.Discard")
	configFilepath  = flag.String("config", "/etc/serve.json", "config filepath")
	verboseRequests = flag.Bool("verboseRequest", false, "dump requests to debug log")

	TypeRoot   = "root"
	TypePublic = "public"
	TypeApi    = "api"
)

func init() {
	flag.Parse()

	var debugout io.Writer
	switch config.Debug {
	case "":
		debugout = io.Discard
	default:
		f, err := os.OpenFile(config.Debug, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(fmt.Sprintln("couldn't open file for debug logs:", config.Debug))
		}
		debugout = f
	}

	errLogger := log.New(os.Stderr, "[error] ", log.LstdFlags)
	infoLogger := log.New(os.Stdout, "[info] ", log.LstdFlags)
	debugLogger := log.New(debugout, "[debug] ", log.LstdFlags)
	logger = func(t utils.LogType, s string, a ...interface{}) {
		switch t {
		case utils.Info:
			infoLogger.Printf(s, a...)
		case utils.Error:
			errLogger.Printf(s, a...)
		default:
			debugLogger.Printf(s, a...)
		}
	}

	f, err := os.ReadFile(*configFilepath)
	if err != nil {
		logger.Errf("Couldn't open config file: [file=%s, err=%s]", *configFilepath, err)
		os.Exit(1)
	}
	if err := json.Unmarshal(f, &config); err != nil {
		logger.Errf("Couldn't parse config file as json: [file=%s, err=%s]", *configFilepath, err)
		os.Exit(1)
	}
	if _, err := os.Stat(config.Sk); err != nil {
		logger.Errf("Couldn't open sk file: [sk=%s, err=%s]", config.Sk, err)
		os.Exit(1)
	}
	if _, err := os.Stat(config.Pem); err != nil {
		logger.Errf("Couldn't open pem file: [pem=%s, err=%s]", config.Pem, err)
		os.Exit(1)
	}
	if len(config.Handlers) == 0 {
		logger.Errf("No handlers defined, nothing to do...")
		os.Exit(1)
	}
}

func main() {
	r := routes.New(logger)

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
		case TypeApi:
			r.HandleApi(h.Name, h.Path)
		default:
			logger.Errf("Main: Handler type not recognized: [type=%s]", h.Type)
		}
	}

	cerr := r.ListenTLS(config.Pem, config.Sk)

	for {
		logger.Errf("server error: [err=%s]", <-cerr)
	}
}
