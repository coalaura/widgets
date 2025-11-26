package main

import (
	"encoding/json"
	"html/template"
	"path/filepath"
	"strings"
)

var templates = make(map[string]*template.Template)

func init() {
	funcMap := template.FuncMap{
		"json": func(v any) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
	}

	baseLayout := template.Must(template.New("layout").Funcs(funcMap).ParseFiles("templates/layout.html"))

	files, err := filepath.Glob("templates/widgets/*.html")
	log.MustFail(err)

	for _, file := range files {
		name := filepath.Base(file)
		name = strings.TrimSuffix(name, ".html")

		tmpl := template.Must(baseLayout.Clone())

		template.Must(tmpl.ParseFiles(file))

		templates[name] = tmpl
	}
}
