package main

import (
	"log"
	"fmt"
	"flag"
	"net/http"
)

const (
	httpPort = ":80"
	httpsPort = ":443"
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

	if certFile != "" && keyFile != "" {
		log.Printf("cert=%s key=%s port=%s", certFile, keyFile, httpsPort)

		go func() {
			if err := http.ListenAndServe(httpPort, http.HandlerFunc(redirectToTls)); err != nil {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}()

		log.Fatal(http.ListenAndServeTLS(httpsPort, certFile, keyFile, h))
	} else {
		log.Println("port=%s", httpPort)
		log.Fatal(http.ListenAndServe(httpPort, h))
	}
}
