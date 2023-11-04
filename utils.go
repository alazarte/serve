package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
)

func urlFilepath(filepath string) string {
	return fmt.Sprintf("%s/%s", schemeHostnameDefault, filepath)
}

func guessMimeFromFilename(filename string) string {
	switch fp.Ext(filename) {
	case ".css":
		return "text/css"
	default:
		log.Println("Cannot guess mime:", filename)
		return ""
	}
}

func statusPerError(err error) int {
	if os.IsNotExist(err) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
