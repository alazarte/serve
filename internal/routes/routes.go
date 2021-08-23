package routes

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

type logger interface {
	Println(...interface{})
	Printf(string, ...interface{})
}

type Routes struct {
	InfoLogger  logger
	ErrLogger   logger
	DebugLogger logger
}

func (ro Routes) HandleApi(url *url.URL) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			ro.ErrLogger.Println("invalid method:", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if url == nil {
			ro.ErrLogger.Println("missing url to post to...")
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
			ro.ErrLogger.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			ro.ErrLogger.Println("error reading response from API", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(res.StatusCode)
		if _, err := w.Write(b); err != nil {
			ro.ErrLogger.Println(err)
		}
		return
	}
}

func (ro Routes) HandlePublicFiles(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ro.InfoLogger.Printf("serving file: %s%s", path, r.URL.Path)
		http.ServeFile(w, r, fmt.Sprintf("%s%s", path, r.URL.Path))
	}
}

func (ro Routes) HandleRoot(root string, customPaths map[string]func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ro.InfoLogger.Printf("request: %s, %s, %s", r.Method, r.Host, r.URL.Path)
		if h, ok := customPaths[r.URL.Path]; ok {
			ro.InfoLogger.Println("handling custom path")
			h(w, r)
			return
		}
		if r.Method != http.MethodGet {
			ro.ErrLogger.Println("invalid method:", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}
		var out []byte
		switch path.Ext(r.URL.Path) {
		case ".css":
			w.Header().Set("content-type", "text/css; charset=utf-8")
			fallthrough
		case ".html":
			f, err := os.ReadFile(path.Join(root, r.URL.Path))
			if err != nil {
				ro.ErrLogger.Println(err)
				w.WriteHeader(http.StatusNotFound)
				out = []byte(http.StatusText(http.StatusNotFound))
			} else {
				out = f
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			out = []byte(http.StatusText(http.StatusBadRequest))
		}

		if _, err := w.Write(out); err != nil {
			ro.ErrLogger.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
