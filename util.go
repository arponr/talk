package main

import (
	"html/template"
	"net/http"
	"regexp"

	"github.com/russross/blackfriday"
)

var (
	censor   = regexp.MustCompile(`\$\$[^\$]+\$\$|\$[^\$]+\$`)
	uncensor = regexp.MustCompile(`\${3,}`)
)

func replace(vals [][]byte) func([]byte) []byte {
	i := -1
	return func(b []byte) []byte {
		i++
		if i < len(vals) {
			return vals[i]
		}
		return b
	}
}

func markdown(input []byte) string {
	matches := censor.FindAll(input, -1)
	tex := make([][]byte, len(matches))
	for i, m := range matches {
		tex[i] = make([]byte, len(m))
		for j := range m {
			tex[i][j], m[j] = m[j], '$'
		}
	}
	output := blackfriday.MarkdownCommon(input)
	return string(uncensor.ReplaceAllFunc(output, replace(tex)))
}

func safe(s string) interface{} { return template.HTML(s) }

func buildTemplate(names ...string) *template.Template {
	files := []string{"html/base.html"}
	for _, f := range names {
		files = append(files, "html/"+f+".html")
	}
	return template.Must(template.New("").Funcs(template.FuncMap{
		"safe": safe,
	}).ParseFiles(files...))
}

var templates = map[string]*template.Template{
	"login":  buildTemplate("login"),
	"root":   buildTemplate("main", "root"),
	"thread": buildTemplate("main", "thread"),
}

func render(w http.ResponseWriter, tmpl string, data interface{}) error {
	return templates[tmpl].ExecuteTemplate(w, "base.html", data)
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func serveDNE(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
