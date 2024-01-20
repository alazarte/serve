package main

import (
	"log"
	"fmt"
	"os"
	"flag"
	"net/http"
	"encoding/json"
)

const (
	httpPort = ":80"
	httpsPort = ":443"
)

var (
	configFilepath string
)

func init() {
	flag.StringVar(&configFilepath, "config", "config.json", "Config filepath")
	flag.Parse()
}

func redirectToTls(w http.ResponseWriter, r *http.Request) {
	destination := fmt.Sprintf("https://%s:443%s", r.Host, r.RequestURI)
	http.Redirect(w, r, destination, http.StatusMovedPermanently)
}

func main() {
	contents, err := os.ReadFile(configFilepath)
	if err != nil {
		panic(err)
	}

	handler := HandlerConfig{}
	if err := json.Unmarshal(contents, &handler); err != nil {
		panic(err)
	}

	if handler.CertFile != "" && handler.KeyFile != "" {
		log.Printf("cert=%s key=%s port=%s", handler.CertFile, handler.KeyFile, httpsPort)

		go func() {
			if err := http.ListenAndServe(httpPort, http.HandlerFunc(redirectToTls)); err != nil {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}()

		log.Fatal(http.ListenAndServeTLS(httpsPort, handler.CertFile, handler.KeyFile, handler))
	} else {
		log.Printf("port=%s", httpPort)
		log.Fatal(http.ListenAndServe(httpPort, handler))
	}
}
