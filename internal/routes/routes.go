package routes

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

type Routes struct {
	Logger interface {
		Infof(string, ...interface{})
		Errf(string, ...interface{})
		Debugf(string, ...interface{})
	}
}

func (ro Routes) HandleApi(surl string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			ro.Logger.Errf("Invalid method: [method=%s]", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// TODO to accept multiple clients, should compare to
		// previously known hosts and return the origin
		w.Header().Set("Access-Control-Allow-Origin", "https://alazarte.com")
		url, err := url.Parse(surl)
		if err != nil {
			ro.Logger.Errf("Failed to parse target as url: [url=%s, err=%s]", url)
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
			ro.Logger.Errf("%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			ro.Logger.Errf("Failed reading body from API response: [err=%s]", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(res.StatusCode)
		if _, err := w.Write(b); err != nil {
			ro.Logger.Errf("%s", err)
		}
		return
	}
}

func (ro Routes) HandlePublicFiles(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ro.Logger.Infof("Serving file: [url=%s%s]", path, r.URL.Path)
		http.ServeFile(w, r, fmt.Sprintf("%s%s", path, r.URL.Path))
	}
}

func (ro Routes) HandleRoot(root string, extraHeaders map[string]string, customPaths map[string]func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ro.Logger.Infof("request: %s, %s, %s", r.Method, r.Host, r.URL.Path)
		if h, ok := customPaths[r.URL.Path]; ok {
			ro.Logger.Infof("HandleRoot: Handling custom path: [path=%s]", r.URL.Path)
			h(w, r)
			return
		}
		if r.Method != http.MethodGet {
			ro.Logger.Errf("HandleRoot: Invalid method: [method=%s]", r.Method)
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
				ro.Logger.Errf("HandleRoot: Failed to read html file: [err=%s]", err)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			for k, v := range extraHeaders {
				w.Header().Set(k, v)
			}
			if _, err := io.Copy(w, f); err != nil {
				ro.Logger.Errf("HandleRoot: Failed to write file to ResponseWriter: [err=%s]", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}
