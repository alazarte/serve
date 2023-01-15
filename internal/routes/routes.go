package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	INTERNAL_SERVER_ERROR = `<h2>Internal Server Error</h2>
<p>Found a problem when handling the request</p>
`
)

type logger interface {
	Infof(string, ...interface{})
	Errf(string, ...interface{})
	Debugf(string, ...interface{})
}

type mux struct {
	handlers map[string]func(w http.ResponseWriter, r *http.Request)
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
			handlers: make(map[string]func(w http.ResponseWriter, r *http.Request)),
		},
	}
}

func (m mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, ok := m.handlers[r.Host]; !ok {
		if w != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}
	m.handlers[r.Host](w, r)
}

// TODO log this
func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("https://%s%s", r.Host, r.URL.Path), http.StatusTemporaryRedirect)
}

func (ro *routes) ConfigHandlers(handlers []Handler) {
	for _, h := range handlers {
		switch h.Type {
		case TypeRoot:
			extraHeaders := make(map[string]string)
			for _, h := range h.Headers {
				extraHeaders[http.CanonicalHeaderKey(h.Name)] = h.Value
			}
			ro.HandleRoot(h.Name, h.Path, extraHeaders)
		case TypePublic:
			ro.HandlePublicFiles(h.Name, h.Path)
		case TypeProxy:
			ro.HandleProxy(h.Name, h.Path)
		default:
			fmt.Printf("Main: Handler type not recognized: [type=%s]", h.Type)
		}
	}
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
