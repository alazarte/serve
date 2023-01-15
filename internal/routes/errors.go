package routes

import (
	"errors"
	"net/http"
	"os"
)

var (
	// TODO replace with http.StatusText()
	ErrBadRequest          = errors.New("Bad request")
	ErrFileNotFound        = errors.New("File not found")
	ErrInternalServerError = errors.New("Internal server error")

	ErrBadRequestFilepath          = "/www/400.html"
	ErrFileNotFoundFilepath        = "/www/404.html"
	ErrInternalServerErrorFilepath = "/www/500.html"
)

func writeError(w http.ResponseWriter, err error) {
	status, contents := readResponseContents(err)

	w.WriteHeader(status)
	w.Write(contents)
}

func readResponseContents(err error) (int, []byte) {
	status := http.StatusOK
	filepath := ""

	switch err {
	case ErrFileNotFound:
		status = http.StatusNotFound
		filepath = ErrFileNotFoundFilepath
	case ErrInternalServerError:
		status = http.StatusInternalServerError
		filepath = ErrInternalServerErrorFilepath
	case ErrBadRequest:
		status = http.StatusBadRequest
		filepath = ErrBadRequestFilepath
	}

	if _, err := os.Stat(filepath); err != nil {
		return status, []byte(err.Error())
	}

	contents, err := os.ReadFile(filepath)
	if err != nil {
		return status, []byte(err.Error())
	}

	return status, []byte(contents)
}
