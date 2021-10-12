package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
)

type logger interface {
	Infof(string, ...interface{})
	Errf(string, ...interface{})
	Debugf(string, ...interface{})
}

type mux struct {
	handlers map[string]func(w http.ResponseWriter, r *http.Request)
	verbose  bool
	logger   logger
}

type routes struct {
	mux    mux
	logger logger
}

func New(logger logger) routes {
	return routes{
		logger: logger,
		mux: mux{
			logger:   logger,
			handlers: make(map[string]func(w http.ResponseWriter, r *http.Request)),
		},
	}
}

func (ro *routes) HandleApi(name, surl string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			ro.logger.Errf("Invalid method: [method=%s]", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// TODO to accept multiple clients, should compare to
		// previously known hosts and return the origin
		w.Header().Set("Access-Control-Allow-Origin", "https://alazarte.com")
		url, err := url.Parse(surl)
		if err != nil {
			ro.logger.Errf("Failed to parse target as url: [url=%s, err=%s]", url)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		client := http.Client{}
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = url
		r2.RequestURI = ""
		res, err := client.Do(r2)
		if err != nil {
			ro.logger.Errf("%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			ro.logger.Errf("Failed reading body from API response: [err=%s]", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(res.StatusCode)
		if _, err := w.Write(b); err != nil {
			ro.logger.Errf("%s", err)
		}
		return
	}
}

func (ro *routes) HandlePublicFiles(name, path string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Infof("Serving file: [url=%s%s, from=%s]", path, r.URL.Path, r.RemoteAddr)
		http.ServeFile(w, r, fmt.Sprintf("%s%s", path, r.URL.Path))
	}
}

func (ro *routes) HandleRoot(name, root string, extraHeaders map[string]string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			ro.logger.Errf("HandleRoot: Invalid method: [method=%s]", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}
		switch path.Ext(r.URL.Path) {
		case ".css":
			w.Header().Set("content-type", "text/css; charset=utf-8")
			fallthrough
		case ".html":
			f, err := os.Open(path.Join(root, r.URL.Path))
			if err != nil {
				ro.logger.Errf("HandleRoot: Failed to read html file: [err=%s]", err)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			for k, v := range extraHeaders {
				w.Header().Set(k, v)
			}
			if _, err := io.Copy(w, f); err != nil {
				ro.logger.Errf("HandleRoot: Failed to write file to ResponseWriter: [err=%s]", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			ro.logger.Infof("Serving file: [url=%s, from=%s]", r.URL.Path, r.RemoteAddr)
		default:
			ro.logger.Infof("Bad request: [path=%s]", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func (m mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.verbose {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			m.logger.Debugf("Failed to get dump of request: [err=%q]", err)
		}
		m.logger.Debugf("Request dump: [dump=%q]", dump)
	}

	if _, ok := m.handlers[r.Host]; !ok {
		m.logger.Errf("Failed to handle host: [host=%s]", r.Host)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m.handlers[r.Host](w, r)

}

// TODO log this
func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("https://%s%s", r.Host, r.URL.Path), http.StatusTemporaryRedirect)
}

func (ro *routes) ListenTLS(pem, sk string) chan error {
	cerr := make(chan error)

	server := &http.Server{
		Addr:     ":443",
		Handler:  ro.mux,
		ErrorLog: log.New(os.Stderr, "[server error] ", log.LstdFlags),
	}

	go func() {
		cerr <- http.ListenAndServe(":80", http.HandlerFunc(redirect))
	}()

	go func() {
		cerr <- server.ListenAndServeTLS(pem, sk)
	}()

	return cerr
}
