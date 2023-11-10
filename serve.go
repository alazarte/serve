package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
)

func writeStatusPage(status int, w http.ResponseWriter) {
	w.WriteHeader(status)
	// TODO implement a template for errors
	body := fmt.Sprintf("<h1>%s</h1>", http.StatusText(status))
	w.Write([]byte(body))
}

func handleServeFiles(w http.ResponseWriter, r *http.Request) {
	localFilepath := fmt.Sprintf("%s%s", filepathPrefix, r.URL.Path)

	if !fp.IsLocal(localFilepath) {
		log.Println("Filepath not local:", localFilepath)
		writeStatusPage(http.StatusBadRequest, w)
		return
	}

	pathInfo, err := os.Stat(localFilepath)
	if err != nil {
		log.Println("Failed Stat() file:", localFilepath)
		writeStatusPage(http.StatusNotFound, w)
		return
	}

	if pathInfo.IsDir() {
		// If the path is a folder, and doesn't ends in /, the browser won't load resources located
		// in that path, like a script.js, it will try to load that script from the root folder
		if localFilepath[len(localFilepath)-1] != '/' {
			// TODO get the redirect URL some other way
			http.Redirect(w, r, "http://"+r.Host+r.RequestURI+"/", http.StatusSeeOther)
			return
		}
		localFilepath = fp.Join(localFilepath, "index.html")
		log.Println("path ends in /, should be index")
	}
	log.Println(localFilepath)

	if err := writeFileAndStatus(localFilepath, 200, w); err != nil {
		log.Println("Failed writeFile()", err)
		writeStatusPage(statusPerError(err), w)
	}
}

func writeFileAndStatus(filepath string, status int, w http.ResponseWriter) error {
	f, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	mime := guessMimeFromFilename(filepath)
	if mime == "" {
		mime = http.DetectContentType(f)
		log.Println("DetectContentType():", mime)
	}

	w.Header().Add("Content-type", mime)
	w.WriteHeader(status)
	w.Write(f)
	return nil
}
