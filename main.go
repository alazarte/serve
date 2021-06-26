package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

var (
	errLogger  *log.Logger
	infoLogger *log.Logger

	postURL *url.URL

	skFilepath   = flag.String("sk", "", "secret key filepath")
	htmlFilepath = flag.String("html", "", "where html files are located")
	pemFilepath  = flag.String("pem", "", "certificate filepath")
	publicPath   = flag.String("public", "/tmp/public", "path to get public files")
	port         = flag.String("port", "443", "port to listen for http connections")
	urlString    = flag.String("post", "", "url to post to, to upload files")

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
	if *urlString != "" {
		u, err := url.Parse(*urlString)
		if err != nil {
			fmt.Println("error parsing url:", err)
			os.Exit(1)
		}
		postURL = u
	}
	if _, err := os.Stat(*publicPath); err != nil {
		if err := os.Mkdir(*publicPath, 0755); err != nil {
			fmt.Println(err)
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	infoLogger.Printf("%s %s %s", r.Method, r.URL, r.RemoteAddr)

	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		errLogger.Println("invalid method:", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		if postURL == nil {
			errLogger.Println("missing url to post to...")
			return
		}
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = postURL
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
		errLogger.Println(err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	infoLogger.Printf("redirecting information: %#v\n", r)

	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		errLogger.Println("invalid method:", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	target := "https://" + r.Host + r.URL.Path

	infoLogger.Printf("redirecting to %s", target)
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

func init() {
	errLogger = log.New(os.Stderr, "[error] ", log.LstdFlags)
	infoLogger = log.New(os.Stdout, "[info] ", log.LstdFlags)
}

func main() {
	infoLogger.Printf("listening on port 80 to redirect...")
	go http.ListenAndServe(":80", http.HandlerFunc(redirect))

	http.HandleFunc("/", handler)
	http.HandleFunc("/public/", func(w http.ResponseWriter, r *http.Request) {
		infoLogger.Printf("%s %s %s", r.Method, r.URL, r.RemoteAddr)
		file := strings.TrimPrefix(r.URL.Path, "/public/")
		http.ServeFile(w, r, fmt.Sprintf("%s/%s", publicPath, file))
	})

	infoLogger.Printf("going to serve on port 443, public path is %s", publicPath)
	errLogger.Fatal(http.ListenAndServeTLS(":443", *pemFilepath, *skFilepath, nil))
}
