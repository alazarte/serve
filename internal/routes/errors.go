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
	status := http.StatusOK
	statusContent := []byte{}

	switch err {
	case ErrFileNotFound:
		status = http.StatusNotFound
		statusContent = []byte(http.StatusText(http.StatusNotFound))
	case ErrInternalServerError:
		status = http.StatusInternalServerError
		statusContent = []byte(http.StatusText(http.StatusInternalServerError))
	case ErrBadRequest:
		status = http.StatusBadRequest
		statusContent = []byte(http.StatusText(http.StatusBadRequest))
	}

	w.WriteHeader(status)
	w.Write(statusContent)
}
