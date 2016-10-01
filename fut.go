package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var (
	defaultNamespace = "und:"
	ttt              *template.Template
)

func setupTemplates() {
	ttt = template.New("root").Funcs(template.FuncMap{"hasns": hasns})
	ttt.New("top").Parse(`
	<h1><a href="/">Fedora Utility Tool</a></h1>
	<form action="/items" method="post">
	<label for="item">PID</label>
	<input type="text" name="item" id="item">
	</form>`)
	ttt.New("flink").Parse(`
	{{ if hasns . }}
	<a href="/items/{{ . }}">{{ . }}</a>
	{{else}}{{ . }}
	{{end}}`)
	ttt.New("home").Parse(`
	<body>
	{{ template "top" }}
	Welcome!
	</body>`)
	ttt.New("items").Parse(`
	<body>
	{{ template "top" }}
	PID = {{ template "flink" . }}
	</body>`)
}

func hasns(name string) bool {
	return strings.ContainsRune(name, ':')
}

func main() {
	setupTemplates()

	r := mux.NewRouter()
	r.HandleFunc("/items/{pid}", ItemView)
	r.HandleFunc("/items", ItemPost).Methods("POST")
	r.HandleFunc("/", WelcomeHandler)
	http.ListenAndServe(":8000", r)
}

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	ttt.ExecuteTemplate(w, "home", nil)
}

func ItemPost(w http.ResponseWriter, r *http.Request) {
	item := r.FormValue("item")
	if !strings.ContainsRune(item, ':') {
		item = defaultNamespace + item
	}
	http.Redirect(w, r, "/items/"+item, http.StatusSeeOther)
}

func ItemView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ttt.ExecuteTemplate(w, "items", vars["pid"])
}
