package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"serve/internal/gemini"
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
	TypeGem    = "gemini"
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

type mux struct {
	handlers map[string]func(w http.ResponseWriter, r *http.Request)
}

func (h mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if *verboseRequests {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			logger.Debugf("Failed to get dump of request: [err=%q]", err)
		}
		logger.Debugf("Request dump: [dump=%q]", dump)
	}

	if _, ok := h.handlers[r.Host]; !ok {
		logger.Errf("Failed to handle host: [host=%s]", r.Host)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.handlers[r.Host](w, r)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	if *verboseRequests {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			logger.Debugf("Failed to get dump of request: [err=%q]", err)
		}
		logger.Debugf("Request dump: [dump=%q]", dump)
	}

	target := fmt.Sprintf("https://%s%s", r.Host, r.URL.Path)
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	logger.Infof("Http redirect: [target=%s]", target)
}

func main() {
	m := mux{
		handlers: make(map[string]func(w http.ResponseWriter, r *http.Request)),
	}
	r := routes.Routes{
		Logger: logger,
	}

	cerr := make(chan error)

	for _, h := range config.Handlers {
		switch h.Type {
		case TypeRoot:
			extraHeaders := make(map[string]string)
			for _, h := range h.Headers {
				extraHeaders[http.CanonicalHeaderKey(h.Name)] = h.Value
			}
			m.handlers[h.Name] = r.HandleRoot(h.Path, extraHeaders)
		case TypePublic:
			m.handlers[h.Name] = r.HandlePublicFiles(h.Path)
		case TypeApi:
			m.handlers[h.Name] = r.HandleApi(h.Path)
		case TypeGem:
			go func() {
				cerr <- gemini.Serve(h.Port, config.Pem, config.Sk, h.Path)
			}()
		default:
			logger.Errf("Main: Handler type not recognized: [type=%s]", h.Type)
		}
	}

	server := &http.Server{
		Addr:     ":443",
		Handler:  m,
		ErrorLog: log.New(os.Stderr, "[server error] ", log.LstdFlags),
	}

	go func() {
		cerr <- http.ListenAndServe(":80", http.HandlerFunc(redirect))
	}()

	go func() {
		cerr <- server.ListenAndServeTLS(config.Pem, config.Sk)
	}()
	for {
		logger.Errf(fmt.Sprintln(<-cerr))
	}
}
