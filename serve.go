package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type HandlerConfig struct {
	HtmlRoot string `json:"htmlFilesRoot"`
	PublicRoot string `json:"publicFilesRoot"`
	KeyFile string `json:"key"`
	CertFile string `json:"cert"`
	Hosts map[string]struct{
		Root string `json:"root"`
		HandlerType string `json:"handlerType"`
	} `json:"hosts"`
}

const (
	PublicHandler = "public"
	HTMLHandler = "html"
)

func (h HandlerConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filepath := r.URL.Path
	log.Println("From:", r.RemoteAddr, "URI:", r.RequestURI)

	renderRoot := false

	if filepath[len(filepath)-1] == '/' {
		renderRoot = true
	}

	host := h.Hosts[r.Host]
	switch host.HandlerType {
	case PublicHandler:
		handlePublicFiles(host.Root, renderRoot, filepath, w, r)
	case HTMLHandler:
		if renderRoot {
			filepath += "index.html"
		}
		handleFile(host.Root, filepath, w, r)
	default:
		log.Println("Can't handle that host:", r.Host)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
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
