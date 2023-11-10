package main

import (
	"bytes"
	"fmt"
	// TODO format log with file and lines
	"log"
	"net/http"
	"os"
	"path"
	"text/template"
)

type linkEntry struct {
	Link string
	Text string
}

const (
	DIR_TEMPLATE = `<head>
<link rel="stylesheet" type="text/css" href="/style.css" />
</head>
<body>
<h2>File explorer</h2>
<table>
{{range .}}
<tr>
<td><a href='/{{.Link}}'>{{.Text}}</a></td>
</tr>
{{end}}
</table>
</body>
`
)

func handleUploadingFiles(w http.ResponseWriter, r *http.Request) {

}

func handlePublicFiles(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
	filename := r.URL.Path[1:]

	setMimeForServedFile(w, filename)
	if err := servePathContents(w, r, filename); err != nil {
		log.Printf("error: %s %s %s %s", r.Method, r.URL.Path, r.RemoteAddr, err)
		return
	}
}

func setMimeForServedFile(w http.ResponseWriter, filename string) {
	switch path.Ext(filename) {
	case ".css":
		w.Header().Set("content-type", "text/css; charset=utf-8")
	}
}

func serveFileContents(w http.ResponseWriter, r *http.Request, filename string) error {
	return serveContents(w, r, filename, fileContents)
}

func servePathContents(w http.ResponseWriter, r *http.Request, filename string) error {
	return serveContents(w, r, filename, pathContents)
}

func serveContents(w http.ResponseWriter, r *http.Request, filename string, getContents func(string) ([]byte, error)) error {
	contents, err := getContents(filename)
	if err != nil {
		return err
	}

	_, err = w.Write(contents)
	return err
}

func pathContents(filename string) ([]byte, error) {
	return openFilepathOrNotFound(filename, true)
}

func fileContents(filename string) ([]byte, error) {
	return openFilepathOrNotFound(filename, false)
}

func openFilepathOrNotFound(filename string, renderIfDir bool) ([]byte, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return readFile(filename)
	}

	if renderIfDir {
		return renderDirContents(filename)
	}

	return nil, fmt.Errorf("bad request")
}

func lookBackwardsSlash(filepath string) string {
	// /asd/qwe/
	i := len(filepath) - 2
	for ; i >= 0 && filepath[i] != '/'; i-- {
	}
	if i < 0 {
		return filepath
	}
	return filepath[:i]
}

func renderDirContents(filepath string) ([]byte, error) {
	t, err := template.New("dir").Parse(DIR_TEMPLATE)
	if err != nil {
		return nil, err
	}

	sList, err := listDirEntries(filepath)
	if err != nil {
		return nil, err
	}

	prevFolder := lookBackwardsSlash(filepath)
	sList = append([]linkEntry{{prevFolder, "../"}}, sList...)

	buffer := bytes.NewBuffer(nil)
	if err := t.Execute(buffer, sList); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func listDirEntries(filepath string) ([]linkEntry, error) {
	list, err := os.ReadDir(filepath)
	if err != nil {
		return []linkEntry{}, err
	}

	files := []linkEntry{}
	for _, d := range list {
		name := d.Name()
		if d.IsDir() {
			name = name + "/"
		}

		files = append(files, linkEntry{
			fmt.Sprintf("%s%s", filepath, name),
			name,
		})
	}
	return files, nil
}

func readFile(filename string) ([]byte, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return file, nil
}
