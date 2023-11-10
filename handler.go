package main

import (
	"log"
	"net/http"
	"strings"
)

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	toSwitch := parts[0]
	if r.URL.Path[0] == '/' {
		toSwitch = parts[1]
	}

	log.Println("Switching:", toSwitch)

	switch toSwitch {
	case "public":
		handlePublicFiles(w, r)
	default:
		handleServeFiles(w, r)
	}
}
