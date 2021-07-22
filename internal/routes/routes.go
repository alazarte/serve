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
	ErrLogger   logger
	DebugLogger logger

	PostUrl *url.URL

	PublicPath   string
	HtmlFilepath string
}

func (ro Routes) Api(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ro.ErrLogger.Println("invalid method:", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if ro.PostUrl == nil {
		ro.ErrLogger.Println("missing url to post to...")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	client := http.Client{}
	r2 := new(http.Request)
	*r2 = *r
	r2.URL = ro.PostUrl
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

func (ro Routes) PublicFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, fmt.Sprintf("%s%s", ro.PublicPath, r.URL.Path))
}

func (ro Routes) Index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ro.ErrLogger.Println("invalid method:", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if path.Ext(r.URL.Path) == ".css" {
		w.Header().Set("content-type", "text/css; charset=utf-8")
	}
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
	f, err := os.ReadFile(path.Join(ro.HtmlFilepath, r.URL.Path))
	if err != nil {
		ro.ErrLogger.Println(err)
		w.WriteHeader(http.StatusNotFound)
		f = []byte(http.StatusText(http.StatusNotFound))
	}
	if _, err := w.Write(f); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ro.ErrLogger.Println(err)
	}
}
