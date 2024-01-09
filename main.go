package main

import (
	"log"
	"fmt"
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

func redirectToTls(w http.ResponseWriter, r *http.Request) {
	destination := fmt.Sprintf("https://%s:443%s", r.Host, r.RequestURI)
	http.Redirect(w, r, destination, http.StatusMovedPermanently)
}

func main() {
	h := handler{
		renderIndex: false,
	}
	log.Println(certFile, keyFile)
	if certFile != "" && keyFile != "" {
		log.Println("Listening https")
		go func() {
			if err := http.ListenAndServe(":80", http.HandlerFunc(redirectToTls)); err != nil {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}()

		log.Fatal(http.ListenAndServeTLS(":443", certFile, keyFile, h))
	} else {
		log.Fatal(http.ListenAndServe(":80", h))
	}
}
