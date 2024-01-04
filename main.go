package main

import (
	"log"
	"net/http"
)

func main() {
	h := handler{
		renderIndex: false,
	}
	log.Fatal(http.ListenAndServe(":8080", h))
}
