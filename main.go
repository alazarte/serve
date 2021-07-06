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
	"path"
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

	infoFile  = flag.String("info", "", "filepath to print info logs to, default is stdout")
	errFile   = flag.String("err", "", "filepath to print error logs to, default is stderr")
	debugFile = flag.String("debug", "", "filepath to print debug logs to, default is io.Discard")

	client = http.Client{}
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
		errLogger.Println("handler not implemented for:", r.Host)
		w.WriteHeader(http.StatusInternalServerError)
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
	var infout, errout, debugout io.Writer
	if *infoFile != "" {
		f, err := os.OpenFile(*infoFile, os.O_WRONLY|os.O_APPEND, 0644)
		// TODO create file if not exists?
		if err != nil {
			fmt.Println("couldn't open file for info logs:", *infoFile)
			os.Exit(2)
		}
		infout = f
	} else {
		infout = os.Stdout
	}
	if *errFile != "" {
		f, err := os.OpenFile(*errFile, os.O_WRONLY|os.O_APPEND, 0644)
		// TODO create file if not exists?
		if err != nil {
			fmt.Println("couldn't open file for err logs:", *errFile)
			os.Exit(2)
		}
		errout = f
	} else {
		errout = os.Stderr
	}
	if *debugFile != "" {
		f, err := os.OpenFile(*debugFile, os.O_WRONLY|os.O_APPEND, 0644)
		// TODO create file if not exists?
		if err != nil {
			fmt.Println("couldn't open file for debug logs:", *debugFile)
			os.Exit(2)
		}
		debugout = f
	} else {
		debugout = io.Discard
	}
	errLogger = log.New(errout, "[error] ", log.LstdFlags)
	infoLogger = log.New(infout, "[info] ", log.LstdFlags)
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
	m.handlers["alazarte.com"] = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errLogger.Println("invalid method:", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if path.Ext(r.URL.Path) == ".css" {
			w.Header().Set("content-type", "text/css; charset=utf-8")
		}
		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}
		f, err := os.ReadFile(path.Join(*htmlFilepath, r.URL.Path))
		if err != nil {
			errLogger.Println(err)
			w.WriteHeader(http.StatusNotFound)
			f = []byte(http.StatusText(http.StatusNotFound))
		}
		if _, err := w.Write(f); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errLogger.Println(err)
		}
	}
	m.handlers["192.168.1.2"] = m.handlers["alazarte.com"]
	m.handlers["public.alazarte.com"] = func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, fmt.Sprintf("%s%s", *publicPath, r.URL.Path))
	}
	m.handlers["api.alazarte.com"] = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errLogger.Println("invalid method:", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if postUrl == nil {
			errLogger.Println("missing url to post to...")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = postUrl
		r2.RequestURI = ""
		res, err := client.Do(r2)
		if err != nil {
			errLogger.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			errLogger.Println("error reading response from API", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(res.StatusCode)
		if _, err := w.Write(b); err != nil {
			errLogger.Println(err)
		}
		return
	}

	server := &http.Server{Addr: ":443", Handler: m, ErrorLog: errLogger}

	go func() {
		cerr <- server.ListenAndServeTLS(*pemFilepath, *skFilepath)
	}()
	for {
		errLogger.Println(<-cerr)
	}
}
