package main

import (
	"log"
	"flag"
	"net/http"
)

var (
	certFile string
	keyFile  string
)

func init() {
	flag.StringVar(&certFile, "cert", "", "Certificate filepath")
	flag.StringVar(&keyFile, "key", "", "Key filepath")
	flag.Parse()
}

func main() {
	h := handler{
		renderIndex: false,
	}
	log.Println(certFile, keyFile)
	if certFile != "" && keyFile != "" {
		log.Println("Listening https")
		log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, h))
	} else {
		log.Fatal(http.ListenAndServe(":80", h))
	}
}
