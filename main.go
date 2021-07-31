package main

import (
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

var (
	errLogger   *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger

	postUrl *url.URL

	skFilepath   = flag.String("sk", "privkey.pem", "secret key filepath")
	htmlFilepath = flag.String("html", "", "where html files are located")
	pemFilepath  = flag.String("pem", "fullcert.pem", "certificate filepath")
	publicPath   = flag.String("public", "/tmp/public", "path to get public files")
	post         = flag.String("post", "", "to configure location for git.sr.ht/~alazarte/uploader")

	debugFile = flag.String("debug", "", "filepath to print debug logs to, default is io.Discard")
)

func init() {
	flag.Parse()

	if *skFilepath == "" || *pemFilepath == "" {
		fmt.Println("missing either pem or sk filepath")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *htmlFilepath == "" {
		fmt.Println("html source path is missing")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *post != "" {
		u, err := url.Parse(*post)
		if err != nil {
			fmt.Println("error parsing url:", err)
			os.Exit(1)
		}
		postUrl = u
	}
	if _, err := os.Stat(*publicPath); err != nil {
		if err := os.Mkdir(*publicPath, 0755); err != nil {
			fmt.Println(err)
		}
	}
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

func init() {
	var debugout io.Writer
	debugout = io.Discard
	if *debugFile != "" {
		f, err := os.OpenFile(*debugFile, os.O_WRONLY|os.O_APPEND, 0644)
		// TODO create file if not exists?
		if err != nil {
			fmt.Println("couldn't open file for debug logs:", *debugFile)
			os.Exit(2)
		}
		debugout = f
	}
	errLogger = log.New(os.Stderr, "[error] ", log.LstdFlags)
	infoLogger = log.New(os.Stdout, "[info] ", log.LstdFlags)
	debugLogger = log.New(debugout, "[debug] ", log.LstdFlags)
}

func main() {
	cerr := make(chan error)
	go func() {
		cerr <- http.ListenAndServe(":80", http.HandlerFunc(redirect))
	}()

	m := mux{
		handlers: make(map[string]func(w http.ResponseWriter, r *http.Request)),
	}
	r := routes.Routes{
		ErrLogger:   errLogger,
		DebugLogger: debugLogger,
	}
	m.handlers["alazarte.com"] = r.HandleRoot(*htmlFilepath)
	m.handlers["192.168.1.2"] = r.HandleRoot(*htmlFilepath)
	m.handlers["public.alazarte.com"] = r.HandlePublicFiles(*publicPath)
	m.handlers["www.alazarte.com"] = r.HandleRoot(*htmlFilepath)
	m.handlers["api.alazarte.com"] = r.HandleApi(postUrl)

	server := &http.Server{Addr: ":443", Handler: m, ErrorLog: errLogger}

	go func() {
		cerr <- server.ListenAndServeTLS(*pemFilepath, *skFilepath)
	}()
	for {
		errLogger.Println(<-cerr)
	}
}
