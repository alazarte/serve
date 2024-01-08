package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type handler struct {
	renderIndex bool
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filepath := r.URL.Path
	log.Println("From:", r.RemoteAddr, "URI:", r.RequestURI)

	subdomain := ""
	pieces := strings.Split(r.Host, ".")
	fmt.Println(pieces)
	if len(pieces) > 1 {
		subdomain = pieces[0]
	}

	renderRoot := false

	if filepath[len(filepath)-1] == '/' {
		renderRoot = true
	}
	log.Println("renderRoot:", renderRoot, "Subdomain:", subdomain)

	if subdomain == "public" && renderRoot {
		log.Println("Handling public files:", filepath)
		renderIndexPage("public", w, r)
		return
	}

	if renderRoot {
		filepath += "index.html"
	}

	handleServe(filepath, w, r)
}

func handleServe(filepath string, w http.ResponseWriter, r *http.Request) {
	// avoid leading slash
	filepath = filepath[1:]
	log.Println("Without leading slash:", filepath)

	handleFile(filepath, w, r)
}

func handleFile(filepath string, w http.ResponseWriter, r *http.Request) {
	info, err := os.Stat(filepath)

	if err != nil || os.IsNotExist(err) {
		log.Printf("Couldn't stat=%s err=%s", filepath, err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}
