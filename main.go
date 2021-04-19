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
	skFilepath   string
	pemFilepath  string
	htmlFilepath string
	publicPath   string
	port         string
	postURL      *url.URL
	client       = http.Client{}
	allowHost    = []string{"alazarte.com"}
)

func init() {
	var urlString string

	flag.StringVar(&skFilepath, "sk", "", "secret key filepath")
	flag.StringVar(&htmlFilepath, "html", "", "where html files are located")
	flag.StringVar(&pemFilepath, "pem", "", "certificate filepath")
	flag.StringVar(&publicPath, "public", "/tmp/public", "path to get public files")
	flag.StringVar(&port, "port", "443", "port to listen for http connections")
	flag.StringVar(&urlString, "post", "", "url to post to, to upload files")
	flag.Parse()

	if skFilepath == "" || pemFilepath == "" {
		fmt.Println("missing either pem or sk filepath")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if htmlFilepath == "" {
		fmt.Println("html source path is missing")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if urlString == "" {
		fmt.Println("missing url to post to")
		flag.PrintDefaults()
		os.Exit(1)
	}
	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Println("error parsing url:", err)
		os.Exit(1)
	}
	postURL = u
	if _, err := os.Stat(publicPath); err != nil {
		if err := os.Mkdir(publicPath, 0755); err != nil {
			fmt.Println(err)
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s", r.Method, r.URL, r.RemoteAddr)
	if r.Method == http.MethodPost {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = postURL
		r2.RequestURI = ""

		res, err := client.Do(r2)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("error reading response from API", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(res.StatusCode)
		if _, err := w.Write(b); err != nil {
			log.Println(err)
		}
		return
	}
	if path.Ext(r.URL.Path) == ".css" {
		w.Header().Set("content-type", "text/css; charset=utf-8")
	}
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
	f, err := os.ReadFile("html" + r.URL.Path)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		f = []byte(http.StatusText(http.StatusNotFound))
	}
	if _, err := w.Write(f); err != nil {
		log.Println(err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	log.Println("redirecting information:")
	fmt.Printf("%#v\n", r)

	found := false
	for _, v := range allowHost {
		if r.Host == v {
			found = true
			break
		}
	}
	if !found {
		log.Printf("%s not found in list of hosts: %v\n", r.Host, allowHost)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	target := "https://" + r.Host + r.URL.Path
	// TODO handle query params?

	log.Printf("redirecting to %s", target)
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

func main() {
	log.Printf("listening on port 80 to redirect...")
	go http.ListenAndServe(":80", http.HandlerFunc(redirect))

	http.HandleFunc("/", handler)
	http.HandleFunc("/public/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL, r.RemoteAddr)
		file := strings.TrimPrefix(r.URL.Path, "/public/")
		http.ServeFile(w, r, fmt.Sprintf("%s/%s", publicPath, file))
	})

	log.Printf("going to serve on port 443, public path is %s", publicPath)
	log.Fatal(http.ListenAndServeTLS(":443", pemFilepath, skFilepath, nil))
}
