package main

import (
	"io"
	"log"
	"net/http"
)

var (
	filepathPrefix string = "./root"
)

func redirect(url string, w http.ResponseWriter, r *http.Request) error {
	res, err := http.Get(url)
	if err != nil {
		log.Println("redirect Get()", err)
		return err
	}

	contents, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("redirect ReadAll()", err)
		return err
	}

	w.Write(contents)
	return nil
}

func main() {
	http.HandleFunc("/", handleServeFiles)
	log.Fatal(http.ListenAndServe(defaultPort, nil))
}
