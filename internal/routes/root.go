package routes

import (
	"net/http"
	"path"
)

func (ro *routes) HandleRoot(name, root string, extraHeaders map[string]string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Infof("%s: %s %s %s", name, r.Method, r.URL.Path, r.RemoteAddr)
		ro.logger.Debugf("%s: request dump %+v", name, r)
		if r.Method != http.MethodGet {
			ro.logger.Errf("%s: Invalid method=%s", name, r.Method)
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
			// TODO remove path.Join, it will produce a panic
			for k, v := range extraHeaders {
				w.Header().Set(k, v)
			}

			if err := serveFileContents(w, r, path.Join(root, r.URL.Path)); err != nil {
				writeError(w, err)
				ro.logger.Errf("%s: file=%s err=%s", name, r.URL.Path, err)
			}

		default:
			ro.logger.Errf("%s: Bad request path=%s", name, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}
