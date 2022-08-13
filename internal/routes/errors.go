package routes

import (
	"errors"
	"net/http"
)

var (
	// TODO replace with http.StatusText()
	ErrBadRequest          = errors.New("Bad request")
	ErrFileNotFound        = errors.New("File not found")
	ErrInternalServerError = errors.New("Internal server error")
)

func writeError(w http.ResponseWriter, err error) {
	switch err {
	case ErrFileNotFound:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
	case ErrInternalServerError:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
	}
}
