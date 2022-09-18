package routes

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"text/template"
)

const (
	DIR_TEMPLATE = `<head>
<link rel="stylesheet" type="text/css" href="/style.css" />
</head>
<body>
<h1>alazarte</h1>
[<a href="https://alazarte.com">home</a>] <hr/>
<table>
{{range .}}
<tr>
<td><a href='{{.}}'>{{.}}</a></td>
</tr>
{{end}}
</table>
</body>
`

	// TODO shouldn't specify /www path
	NotFoundFilepath = "/www/404.html"
)

func (ro *routes) HandlePublicFiles(name, path string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Infof("%s: %s %s %s", name, r.Method, r.URL.Path, r.RemoteAddr)
		ro.logger.Debugf("%s: dumping request: %+v", name, r)
		filename := fmt.Sprintf("%s%s", path, r.URL.Path)

		setMimeForServedFile(w, filename)
		if err := servePathContents(w, r, filename); err != nil {
			ro.logger.Errf("%s: %s %s %s", name, r.Method, r.URL.Path, r.RemoteAddr)
			writeError(w, err)
			return
		}
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
		log.Println("File not found:", filename, renderIfDir, err)
		if filename == NotFoundFilepath {
			return []byte(http.StatusText(http.StatusNotFound)), ErrFileNotFound
		}
		return openFilepathOrNotFound(NotFoundFilepath, false)
	}

	if !stat.IsDir() {
		return readFile(filename)
	}

	if renderIfDir {
		return readDirContents(filename)
	}

	return nil, ErrBadRequest
}

func readDirContents(filepath string) ([]byte, error) {
	t, err := template.New("dir").Parse(DIR_TEMPLATE)
	if err != nil {
		return nil, ErrInternalServerError
	}

	sList, err := listDirEntries(filepath)
	if err != nil {
		return nil, ErrInternalServerError
	}

	if filepath != "/" {
		sList = append([]string{"../"}, sList...)
	}

	buffer := bytes.NewBuffer(nil)
	if err := t.Execute(buffer, sList); err != nil {
		return nil, ErrInternalServerError
	}

	return buffer.Bytes(), nil
}

func listDirEntries(filepath string) ([]string, error) {
	list, err := os.ReadDir(filepath)
	if err != nil {
		return []string{}, err
	}

	files := []string{}
	for _, d := range list {
		name := d.Name()
		if d.IsDir() {
			name = name + "/"
		}

		files = append(files, fmt.Sprintf("%s", name))
	}
	return files, nil
}

func readFile(filename string) ([]byte, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, ErrInternalServerError
	}

	return file, nil
}
