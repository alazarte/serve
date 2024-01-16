package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type handler struct {
	renderIndex bool
	htmlFilesRoot string
	publicFilesRoot string
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filepath := r.URL.Path
	log.Println("From:", r.RemoteAddr, "URI:", r.RequestURI)

	renderRoot := false

	if filepath[len(filepath)-1] == '/' {
		renderRoot = true
	}

	// TODO don't hardcode hostnames, replace with array of hosts in config and
	// choose handler from it
	switch r.Host {
	case "public.alazarte.com":
		handlePublicFiles(h.publicFilesRoot, renderRoot, filepath, w, r)
		return
	case "cal.alazarte.com":
		log.Println("TBD Handling calendar")
		return
	case "localhost:8080":
		fallthrough
	case "alazarte.com":
		if renderRoot {
			filepath += "index.html"
		}
		handleFile(h.htmlFilesRoot, filepath, w, r)
		return
	default:
		log.Println("Can't handle that host:", r.Host)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}
}

func handleFile(pathPrefix string, filepath string, w http.ResponseWriter, r *http.Request) {
	// avoid leading slash
	filepath = filepath[1:]
	filepath = fmt.Sprintf("%s%s", pathPrefix, filepath)

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
