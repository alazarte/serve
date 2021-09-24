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
	"serve/internal/routes"
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
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Path     string   `json:"path"`
	SubPaths []KeyVal `json:"subPaths"`
	Headers  []KeyVal `json:"headers"`
}

var (
	TypeRoot   = "root"
	TypePublic = "public"
)

var (
	errLogger   *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger

	postUrl *url.URL

	config Config

	debugFile      = flag.String("debug", "", "filepath to print debug logs to, default is io.Discard")
	configFilepath = flag.String("config", "/etc/serve.json", "config filepath")
)

func init() {
	flag.Parse()

	if _, err := os.Stat(*configFilepath); err != nil {
		panic(err)
	}
	f, err := os.ReadFile(*configFilepath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(f, &config); err != nil {
		panic(err)
	}
	if config.Sk == "" || config.Pem == "" {
		panic("missing either pem or sk filepath")
	}
	if config.Post != "" {
		u, err := url.Parse(config.Post)
		if err != nil {
			panic(fmt.Sprintf("error parsing url: %s", err))
		}
		postUrl = u
	}
	if len(config.Handlers) == 0 {
		panic("no handlers defined")
	}
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
	errLogger = log.New(os.Stderr, "[error] ", log.LstdFlags)
	infoLogger = log.New(os.Stdout, "[info] ", log.LstdFlags)
	debugLogger = log.New(debugout, "[debug] ", log.LstdFlags)
}

type mux struct {
	handlers map[string]func(w http.ResponseWriter, r *http.Request)
}

func (h mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		errLogger.Println("httputil.DumpRequest() = err:", err)
	}
	debugLogger.Printf("redirect dump: %q", dump)

	if _, ok := h.handlers[r.Host]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.handlers[r.Host](w, r)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		errLogger.Println("httputil.DumpRequest() = err:", err)
	}
	debugLogger.Printf("redirect dump: %q", dump)

	target := fmt.Sprintf("https://%s%s", r.Host, r.URL.Path)
	infoLogger.Println("redirecting to:", target)
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

func main() {
	m := mux{
		handlers: make(map[string]func(w http.ResponseWriter, r *http.Request)),
	}
	r := routes.Routes{
		InfoLogger:  infoLogger,
		ErrLogger:   errLogger,
		DebugLogger: debugLogger,
	}

	for _, h := range config.Handlers {
		switch h.Type {
		case TypeRoot:
			customPaths := make(map[string]func(w http.ResponseWriter, r *http.Request))
			extraHeaders := make(map[string]string)
			for _, p := range h.SubPaths {
				customPaths[p.Name] = r.HandleApi(p.Value)
			}
			for _, h := range h.Headers {
				extraHeaders[h.Name] = h.Value
			}
			m.handlers[h.Name] = r.HandleRoot(h.Path, extraHeaders, customPaths)
		case TypePublic:
			m.handlers[h.Name] = r.HandlePublicFiles(h.Path)
		default:
			errLogger.Println("handler type not recognized: %s", h.Type)
		}
	}

	server := &http.Server{Addr: ":443", Handler: m, ErrorLog: errLogger}

	cerr := make(chan error)
	go func() {
		cerr <- http.ListenAndServe(":80", http.HandlerFunc(redirect))
	}()

	go func() {
		cerr <- server.ListenAndServeTLS(config.Pem, config.Sk)
	}()
	for {
		errLogger.Println(<-cerr)
	}
}
