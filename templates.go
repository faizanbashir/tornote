// Copyright 2020 Vladimir Osintsev <osintsev@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tornote

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

var Templates map[string]*template.Template

// Compiles templates from templates/ dir into global map.
func compileTemplates() (err error) {
	if Templates == nil {
		Templates = make(map[string]*template.Template)
	}
	// XXX:
	layout := "templates/base.html"
	pages := []string{
		"templates/index.html",
		"templates/note.html",
	}
	for _, file := range pages {
		baseName := strings.TrimLeft(file, "templates/")
		Templates[baseName], err = template.New("").ParseFiles(file, layout)
		if err != nil {
			return err
		}
	}
	return nil
}

// Wrapper around template.ExecuteTemplate method.
func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// XXX: data is context may be...
	tmpl, ok := Templates[name]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalf("%s template file doesn't exists", name)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
