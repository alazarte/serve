package routes

import (
	"io"
	"net/http"
	"os"
	"path"
)

func (ro *routes) HandleRoot(name, root string, extraHeaders map[string]string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Infof("HandleRoot: %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
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
		default:
			ro.logger.Errf("Bad request: [path=%s]", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}
