package routes

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"
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
)

func (ro *routes) HandlePublicFiles(name, path string) {
	ro.mux.handlers[name] = func(w http.ResponseWriter, r *http.Request) {
		ro.logger.Infof("%s: %s %s %s", name, r.Method, r.URL.Path, r.RemoteAddr)
		ro.logger.Debugf("%s: dumping request: %+v", name, r)
		filename := fmt.Sprintf("%s%s", path, r.URL.Path)

		if err := servePathContents(w, r, filename); err != nil {
			ro.logger.Errf("%s: %s %s %s", name, r.Method, r.URL.Path, r.RemoteAddr)
			writeError(w, err)
			return
		}
	}
}

func serveFileContents(w http.ResponseWriter, r *http.Request, filename string) error {
	return serveContents(w, r, filename, fileContents)
}

func servePathContents(w http.ResponseWriter, r *http.Request, filename string) error {
	return serveContents(w, r, filename, pathContents)
}

func serveContents(w http.ResponseWriter, r *http.Request, filename string, getContents func(string) (io.ReadSeeker, error)) error {
	contents, err := getContents(filename)
	if err != nil {
		return err
	}

	modtime, err := getFileModtime(filename)
	if err != nil {
		return err
	}

	http.ServeContent(w, r, filename, modtime, contents)
	return nil
}

func pathContents(filename string) (io.ReadSeeker, error) {
	return openFilepathOrNotFound(filename, true)
}

func fileContents(filename string) (io.ReadSeeker, error) {
	return openFilepathOrNotFound(filename, false)
}

func openFilepathOrNotFound(filename string, renderIfDir bool) (io.ReadSeeker, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		if filename == "404.html" {
			return nil, ErrFileNotFound
		}
		return openFilepathOrNotFound("404.html", false)
	}

	if !stat.IsDir() {
		return readFile(filename)
	}

	if renderIfDir {
		return readDirContents(filename)
	}

	return nil, ErrBadRequest
}

func readDirContents(filepath string) (io.ReadSeeker, error) {
	list, err := os.ReadDir(filepath)
	if err != nil {
		return nil, ErrInternalServerError
	}

	t, err := template.New("dir").Parse(DIR_TEMPLATE)
	if err != nil {
		return nil, ErrInternalServerError
	}

	sList := listDirEntries(list)
	if filepath != "/" {
		sList = append([]string{"../"}, sList...)
	}

	buffer := bytes.NewBuffer(nil)
	if err := t.Execute(buffer, sList); err != nil {
		return nil, ErrInternalServerError
	}

	return bytes.NewReader(buffer.Bytes()), nil
}

func listDirEntries(dirs []os.DirEntry) []string {
	list := []string{}
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			name = name + "/"
		}
		list = append(list, name)
	}
	return list
}

func getFileModtime(filename string) (time.Time, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return time.Now(), err
	}
	return stat.ModTime(), nil
}

func readFile(filename string) (io.ReadSeeker, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, ErrInternalServerError
	}
	return f, nil
}
