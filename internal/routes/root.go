package routes

import (
	"errors"
	"fmt"
	"net/http"
	"path"
)

func (ro *routes) HandleRoot(name, root string, extraHeaders map[string]string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Infof("%s: %s %s %s", name, r.Method, r.URL.Path, r.RemoteAddr)
		ro.logger.Debugf("%s: request dump %+v", name, r)

		if err := validateMethod(r.Method); err != nil {
			ro.logger.Errf("%s: %s", name, err)
			writeError(w, ErrBadRequest)
		}

		filepath := makeValidPath(r.URL.Path)
		if err := serveRootContents(path.Join(root, filepath), extraHeaders, w, r); err != nil {
			ro.logger.Errf("%s: %s", name, err)
			writeError(w, err)
		}
	}
}

func validateMethod(method string) error {
	if method != http.MethodGet {
		return errors.New(fmt.Sprintf("Invalid method=%s", method))
	}
	return nil
}

func makeValidPath(original string) string {
	if original == "/" {
		return "/index.html"
	}
	return original
}

func serveRootContents(filepath string, extraHeaders map[string]string, w http.ResponseWriter, r *http.Request) error {
	switch path.Ext(filepath) {
	case ".css":
		w.Header().Set("content-type", "text/css; charset=utf-8")
		fallthrough
	case ".ico":
		fallthrough
	case ".js":
		fallthrough
	case ".html":
		// TODO remove path.Join, it will produce a panic
		for k, v := range extraHeaders {
			w.Header().Set(k, v)
		}
		if err := serveFileContents(w, r, filepath); err != nil {
			// TODO this was being logged before
			// ro.logger.Errf("%s: file=%s err=%s", name, path, err)
			return err
		}
	default:
		// ro.logger.Errf("%s: Bad request path=%s", name, path)
		return ErrBadRequest
	}
	return nil
}
