//go:build embed

package web

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed html
var content embed.FS

//go:embed template
var templates embed.FS

func Content() fs.FS {
	r, _ := fs.Sub(content, "html")
	return r
}

var tmpl = template.Must(template.ParseFS(templates, "template/*.tmpl"))

func Templates() *template.Template {
	return tmpl
}

