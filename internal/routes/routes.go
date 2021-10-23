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
	"text/template"
)

const (
	DIR_TEMPLATE = `
<style>
body {
  background-color: lightgray;
}
</style>
{{range .}}
<a href='{{.}}'>{{.}}</a> <br/>
{{end}}
`
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

func listEntries(dirs []os.DirEntry) []string {
	list := []string{}
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			name = name + "/"
		}
		list = append(list, name)
	}
	return list
}

func (ro *routes) HandlePublicFiles(name, path string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Infof("Serving file: [url=%s%s, from=%s]", path, r.URL.Path, r.RemoteAddr)
		filename := fmt.Sprintf("%s%s", path, r.URL.Path)
		stat, err := os.Stat(filename)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("<h1>Not found</h1>"))
			return
		}
		if !stat.IsDir() {
			f, err := os.Open(filename)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.ServeContent(w, r, filename, stat.ModTime(), f)
			return
		}
		list, err := os.ReadDir(filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t, err := template.New("dir").Parse(DIR_TEMPLATE)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sList := listEntries(list)
		if r.URL.Path != "/" {
			sList = append([]string{".."}, sList...)
		}
		if err := t.Execute(w, sList); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (ro *routes) HandleProxy(name, surl string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		url, err := url.Parse(fmt.Sprintf("%s%s?%s", surl, r.URL.Path, r.URL.RawQuery))
		if err != nil {
			ro.logger.Errf("HandleProxy: Failed to parse target as url: [url=%s, err=%s]", url)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.URL = url
		r.RequestURI = ""
		r.Host = "alazarte.com"
		client := http.Client{}
		res, err := client.Do(r)
		if err != nil {
			ro.logger.Errf("Error proxying request: [err=%s]", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "https://alazarte.com")
		w.Header().Set("content-type", r.Header.Get("content-type"))
		if _, err := io.Copy(w, res.Body); err != nil {
			ro.logger.Errf("Error proxying request: [err=%s]", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if res.StatusCode != http.StatusOK {
			w.WriteHeader(res.StatusCode)
		}
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
