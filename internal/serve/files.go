package serve

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
)

const (
	DIR_TEMPLATE = `<head>
<link rel="stylesheet" type="text/css" href="/style.css" />
</head>
<body>
<h2>File explorer</h2>
<table>
{{range .}}
<tr>
<td><a href='{{.Link}}'>{{.Text}}</a></td>
</tr>
{{end}}
</table>
</body>
`
)

type linkEntry struct {
	Link string
	Text string
}

func handlePublicFiles(publicFilesRoot string, renderRoot bool, filepath string, w http.ResponseWriter, r *http.Request) {
	if renderRoot {
		renderIndexPage(publicFilesRoot, filepath, w, r)
		return
	}
	handleFile(publicFilesRoot, filepath, w, r)
}

func renderIndexPage(pathPrefix string, filepath string, w http.ResponseWriter, r *http.Request) {
	page, err := renderDirContents(pathPrefix, filepath)
	if err != nil {
		log.Printf("Couldn't read contents filepath=%s err=%s", filepath, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	w.Write(page)
}

func renderDirContents(pathPrefix string, filepath string) ([]byte, error) {
	t := template.Must(template.New("dir").Parse(DIR_TEMPLATE))

	readFrom := fmt.Sprintf("%s%s", pathPrefix, filepath)

	sList, err := listDirEntries(readFrom, filepath)
	if err != nil {
		return nil, err
	}

	if filepath != "" {
		prevFolder := getUpFolderPath(filepath)
		log.Printf("prevFolder=%s filepath=%s", prevFolder, filepath)
		sList = append([]linkEntry{{prevFolder, "../"}}, sList...)
	}

	buffer := bytes.NewBuffer(nil)
	if err := t.Execute(buffer, sList); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func listDirEntries(readFrom string, filepath string) ([]linkEntry, error) {
	list, err := os.ReadDir(readFrom)
	if err != nil {
		return []linkEntry{}, err
	}
	log.Printf("os.ReadFir(%s) = %v", readFrom, list)

	files := []linkEntry{}
	for _, d := range list {
		name := d.Name()
		if d.IsDir() {
			name = name + "/"
		}

		path := fmt.Sprintf("%s%s", filepath, name)

		fmt.Println("path:", path)
		files = append(files, linkEntry{
			path,
			name,
		})
	}
	return files, nil
}

func getUpFolderPath(filepath string) string {
	i := len(filepath) - 2
	for ; i >= 0 && filepath[i] != '/'; i-- {
	}
	if i == -1 {
		return "/"
	}
	return filepath[:i+1]
}
