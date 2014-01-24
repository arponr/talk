package main

import (
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/russross/blackfriday"
)

var (
	censor   = regexp.MustCompile(`\$\$[^\$]+\$\$|\$[^\$]+\$`)
	uncensor = regexp.MustCompile(`\${3,}`)
)

func replace(v [][]byte) func(string) string {
	i := 0
	return func(in string) (out string) {
		if i < len(v) {
			out = string(v[i])
			i++
		}
		return out
	}
}

func markdown(text string, math bool) (out string) {
	in := []byte(text)
	if math {
		matches := censor.FindAll(in, -1)
		tex := make([][]byte, len(matches))
		for i, m := range matches {
			tex[i] = make([]byte, len(m))
			for j := range m {
				tex[i][j], m[j] = m[j], '$'
			}
		}
		defer func() {
			out = uncensor.ReplaceAllStringFunc(out, replace(tex))
		}()
	}
	out = string(blackfriday.MarkdownCommon(in))
	return
}

func safe(s string) interface{}       { return template.HTML(s) }
func dateFmt(t time.Time) interface{} { return t.Format("2006-01-02") }
func timeFmt(t time.Time) interface{} { return t.Format("3:04pm") }

func buildTemplate(names ...string) *template.Template {
	files := []string{"html/base.html"}
	for _, f := range names {
		files = append(files, "html/"+f+".html")
	}
	return template.Must(template.New("").Funcs(template.FuncMap{
		"safe": safe,
		"date": dateFmt,
		"time": timeFmt,
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
