package main

import (
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/mitchellb613/tarantulabox.git/ui"
)

type templateData struct {
	CurrentYear     int
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

var functions = template.FuncMap{}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.html",
			// "html/partials/*.html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	return cache, nil
}
