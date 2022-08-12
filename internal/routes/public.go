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
		ro.logger.Infof("HandlePublicFiles: %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		ro.logger.Debugf("HandlePublicFiles: dumping request: %+v", r)
		filename := fmt.Sprintf("%s%s", path, r.URL.Path)
		contents, err := handleFile(filename)
		if err != nil {
			handleError(w, r, err)
			return
		}
		modtime, err := getFileModtime(filename)
		if err != nil {
			handleError(w, r, ErrInternalServerError)
			return
		}
		http.ServeContent(w, r, filename, modtime, contents)
	}
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
		// ro.logger.Errf("t.Execute(w, sList=[%s]): [err=%s]", sList, err)
		return nil, ErrInternalServerError
	}

	return bytes.NewReader(buffer.Bytes()), nil
}

func handleFile(filename string) (io.ReadSeeker, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		if filename == "/404.html" {
			return nil, ErrFileNotFound
		}
		return handleFile("/404.html")
	}

	if !stat.IsDir() {
		return readFile(filename)
	}
	return readDirContents(filename)
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case ErrFileNotFound:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
	case ErrInternalServerError:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
	}
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
