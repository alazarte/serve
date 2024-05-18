package serve

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func (h HandlerConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filepath := r.URL.Path
	log.Println("From:", r.RemoteAddr, "Path:", r.URL.Path)

	renderRoot := false
	if filepath[len(filepath)-1] == '/' {
		renderRoot = true
	}

	host := h.Hosts[r.Host]
	switch host.HandlerType {
	case RedirectHandler:
		redirectToURL(w, r, fmt.Sprintf("%s%s", host.URL, filepath))
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

func redirectToURL(w http.ResponseWriter, r *http.Request, url string) {
	log.Println("Redirecting to:", url)
	// TODO can be some other method
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func redirectToTls(w http.ResponseWriter, r *http.Request) {
	destination := fmt.Sprintf("https://%s:443%s", r.Host, r.RequestURI)
	http.Redirect(w, r, destination, http.StatusMovedPermanently)
}

func Listen(config HandlerConfig) {
	if config.CertFile != "" && config.KeyFile != "" {
		log.Printf("cert=%s key=%s port=%s", config.CertFile, config.KeyFile, httpsPort)

		go func() {
			if err := http.ListenAndServe(httpPort, http.HandlerFunc(redirectToTls)); err != nil {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}()

		log.Fatal(http.ListenAndServeTLS(httpsPort, config.CertFile, config.KeyFile, config))
	} else {
		log.Printf("port=%s", httpPort)
		log.Fatal(http.ListenAndServe(httpPort, config))
	}
}
