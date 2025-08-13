//go:build !embed

package web

import (
	"html/template"
	"io/fs"
	"os"
)

func Content() fs.FS {
	return os.DirFS("./web/html")
}

var tmpl = template.Must(template.ParseFS(os.DirFS("./web/template"), "*.tmpl"))

func Templates() *template.Template {
	return tmpl
}

