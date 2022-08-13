package routes

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
)

func (ro *routes) HandleProxy(name, from string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Infof("%s: %s %s %s", name, r.Method, r.URL.Path, r.RemoteAddr)
		ro.logger.Debugf("%s: request dump: %+v", name, r)
		if isGolibRequest(r.URL.RawQuery) {
			handleGoPackageRequest(w, r)
			return
		}

		toURL, err := getParsedURL(from, r.URL)
		if err != nil {
			writeError(w, ErrInternalServerError)
			ro.logger.Errf("%s: Failed to parse url=%s err=%s", name, r.URL, err)
			return
		}

		res, err := performRequest(w, r, toURL)
		if err != nil {
			writeError(w, ErrInternalServerError)
			ro.logger.Errf("%s: client.Do(%#v) err=%s", name, r, err)
			return
		}

		if _, err := io.Copy(w, res.Body); err != nil {
			writeError(w, ErrInternalServerError)
			ro.logger.Errf("%s: Error proxying request: %s", name, err)
			return
		}
	}
}

func isGolibRequest(query string) bool {
	ok, err := regexp.MatchString("go-get=1", query)
	return err == nil && ok
}

func handleGoPackageRequest(w http.ResponseWriter, r *http.Request) {
	module := filepath.Base(r.URL.Path)
	w.Write([]byte(fmt.Sprintf("<meta name=\"go-import\" content=\"git.alazarte.com/%s git https://git.alazarte.com/cgit.cgi/%s/\">", module, module)))
}

func getParsedURL(surl string, proxiedURL *url.URL) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("%s%s?%s", surl, proxiedURL.Path, proxiedURL.RawQuery))
}

func respondInternalServerError(w http.ResponseWriter) {
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(INTERNAL_SERVER_ERROR))
}

func performRequest(w http.ResponseWriter, r *http.Request, toURL *url.URL) (*http.Response, error) {
	r.URL = toURL
	r.RequestURI = ""
	r.Host = "alazarte.com"

	w.Header().Set("Access-Control-Allow-Origin", "https://alazarte.com")
	w.Header().Set("content-type", r.Header.Get("content-type"))

	client := http.Client{}
	return client.Do(r)
}
