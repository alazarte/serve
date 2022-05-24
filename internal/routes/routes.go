package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"text/template"
)

const (
	DIR_TEMPLATE = `<head>
<link rel="stylesheet" type="text/css" href="/style.css" />
</head>
<body>
<h1>alazarte</h1>
[<a href="https://alazarte.com">home</a>] <hr/>
<table>
{{range .}}
<tr>
<td><a href='{{.}}'>{{.}}</a></td>
</tr>
{{end}}
</table>
</body>
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

func listDirEntries(dirs []os.DirEntry) []string {
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
		ro.logger.Debugf("HandlePublicFiles: dumping request: %+v", r)
		filename := fmt.Sprintf("%s%s", path, r.URL.Path)
		stat, err := os.Stat(filename)
		if err != nil {
			if r.URL.Path == "/404.html" {
				w.Write([]byte("404 Error"))
				w.WriteHeader(http.StatusNotFound)
				return
			}
			r.URL.Path = "/404.html"
			redirect(w, r)
			return
		}
		if !stat.IsDir() {
			f, err := os.Open(filename)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.ServeContent(w, r, filename, stat.ModTime(), f)
			ro.logger.Infof("Serving file: [url=%s%s, from=%s]", path, r.URL.Path, r.RemoteAddr)
			return
		}
		list, err := os.ReadDir(filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ro.logger.Errf("os.ReadDir(%s): [err=%s]", filename, err)
			return
		}
		t, err := template.New("dir").Parse(DIR_TEMPLATE)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ro.logger.Errf("template.New(): [err=%s]", err)
			return
		}
		sList := listDirEntries(list)
		if r.URL.Path != "/" {
			sList = append([]string{"../"}, sList...)
		}
		if err := t.Execute(w, sList); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ro.logger.Errf("t.Execute(w, sList=[%s]): [err=%s]", sList, err)
			return
		}
		ro.logger.Infof("Serving dir list: [url=%s%s, from=%s]", path, r.URL.Path, r.RemoteAddr)
	}
}

func (ro *routes) HandleProxy(name, surl string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Debugf("HandleProxy: dumping request: %+v", r)
		if ok, err := regexp.MatchString("go-get=1", r.URL.RawQuery); err == nil && ok {
			module := filepath.Base(r.URL.Path)
			w.Write([]byte(fmt.Sprintf("<meta name=\"go-import\" content=\"git.alazarte.com/%s git https://git.alazarte.com/cgit.cgi/%s/\">", module)))
			return
		}
		url, err := url.Parse(fmt.Sprintf("%s%s?%s", surl, r.URL.Path, r.URL.RawQuery))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ro.logger.Errf("HandleProxy: Failed to parse target as url: [url=%s, err=%s]", url)
			return
		}
		r.URL = url
		r.RequestURI = ""
		r.Host = "alazarte.com"
		client := http.Client{}
		res, err := client.Do(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ro.logger.Errf("client.Do(%#v): [err=%s]", r, err)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "https://alazarte.com")
		w.Header().Set("content-type", r.Header.Get("content-type"))
		if _, err := io.Copy(w, res.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ro.logger.Errf("Error proxying request: [err=%s]", err)
			return
		}
		if res.StatusCode != http.StatusOK {
			w.WriteHeader(res.StatusCode)
			return
		}
	}
}

func (ro *routes) HandleRoot(name, root string, extraHeaders map[string]string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Debugf("HandleRoot: dumping request: %+v", r)
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			ro.logger.Errf("HandleRoot: Invalid method: [method=%s]", r.Method)
			return
		}
		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}
		switch path.Ext(r.URL.Path) {
		case ".css":
			w.Header().Set("content-type", "text/css; charset=utf-8")
			fallthrough
		case ".ico":
			fallthrough
		case ".html":
			f, err := os.Open(path.Join(root, r.URL.Path))
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				ro.logger.Errf("HandleRoot: Failed to read html file: [err=%s]", err)
				return
			}
			for k, v := range extraHeaders {
				w.Header().Set(k, v)
			}
			if _, err := io.Copy(w, f); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				ro.logger.Errf("HandleRoot: Failed to write file to ResponseWriter: [err=%s]", err)
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
