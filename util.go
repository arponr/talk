package main

import (
	"html/template"
	"net/http"
	"regexp"

	"github.com/russross/blackfriday"
)

var (
	censor   = regexp.MustCompile(`\$\$[^\$]+\$\$|\$[^\$]+\$`)
	uncensor = regexp.MustCompile(`\$+`)
)

func replace(vals [][]byte) func([]byte) []byte {
	i := -1
	return func(b []byte) []byte {
		i++
		return vals[i]
	}
}

func markdown(input []byte) []byte {
	matches := censor.FindAll(input, -1)
	tex := make([][]byte, len(matches))
	for i, m := range matches {
		tex[i] = make([]byte, len(m))
		for j := range m {
			tex[i][j], m[j] = m[j], '$'
		}
	}
	output := blackfriday.MarkdownCommon(input)
	return uncensor.ReplaceAllFunc(output, replace(tex))
}

func buildTemplate(files ...string) *template.Template {
	files = append(files, "html/base.html")
	return template.Must(template.New("").ParseFiles(files...))
}

var templates = map[string]*template.Template{
	"root":  buildTemplate("html/root.html"),
	"login": buildTemplate("html/login.html"),
}

func render(w http.ResponseWriter, tmpl string, data interface{}) error {
	return templates[tmpl].ExecuteTemplate(w, "base.html", data)
}
