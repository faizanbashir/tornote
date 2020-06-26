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
	"crypto/sha256"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

type server struct {
	// Listen port
	Port uint64
	// Data source name
	DSN string
	// PostgreSQL connection
	db *pg.DB
	// Mux router
	router *mux.Router
	// Compiled templates
	templates map[string]*template.Template
}

type Server interface {
	Init()
	Listen() error
}

// Constructor for new server.
func NewServer(port uint64, dsn string) *server {
	_, err := pg.ParseURL(dsn)
	if err != nil {
		panic(err)
	}
	return &server{Port: port, DSN: dsn}
}

// Open and check database connection.
func (s *server) connectDB() error {
	opt, err := pg.ParseURL(s.DSN)
	if err != nil {
		return err
	}
	s.db = pg.Connect(opt)

	// XXX: Ping postgres connection
	//if err = s.db.Ping(); err != nil {
	//	return err
	//}
	return nil
}

// Creates database tables for notes if not exists.
func (s *server) createSchema() error {
	err := s.db.CreateTable(&Note{}, &orm.CreateTableOptions{
		IfNotExists: true,
	})
	if err != nil {
		return err
	}
	return nil
}

// Generate hash from server secret key.
func (s *server) getSecretHash() []byte {
	h := sha256.New()
	h.Write([]byte("hello world\n"))
	return h.Sum(nil)
}

// Compiles templates from templates/ dir into global map.
func (s *server) compileTemplates() (err error) {
	if s.templates == nil {
		s.templates = make(map[string]*template.Template)
	}
	// XXX:
	layout := "templates/base.html"
	pages := []string{
		"templates/index.html",
		"templates/note.html",
	}
	for _, file := range pages {
		baseName := strings.TrimLeft(file, "templates/")
		s.templates[baseName], err = template.New("").ParseFiles(file, layout)
		if err != nil {
			return err
		}
	}
	return nil
}

// Wrapper around template.ExecuteTemplate method.
func (s *server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// XXX: data is context may be...
	tmpl, ok := s.templates[name]
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

// Initialize server.
func (s *server) Init() {
	s.router = mux.NewRouter().StrictSlash(true)

	// Setup middlewares
	csrfMiddleware := csrf.Protect(
		s.getSecretHash(),
		csrf.FieldName("csrf_token"),
		csrf.SameSite(csrf.SameSiteStrictMode),
	)
	s.router.Use(csrfMiddleware)

	// HTTP handlers
	s.router.Handle("/", MainFormHandler(s)).Methods("GET")
	s.router.HandleFunc("/healthz", HealthStatusHandler).Methods("GET")
	s.router.PathPrefix("/public/").HandlerFunc(PublicFileHandler).Methods("GET")
	s.router.Handle("/note", CreateNoteHandler(s)).Methods("POST")
	s.router.Handle("/{id}", ReadNoteHandler(s)).Methods("GET")

	// Pre-compile templates
	if err := s.compileTemplates(); err != nil {
		panic(err)
	}
}

// Running server.
func (s *server) Listen() error {
	// Connecting to database
	if err := s.connectDB(); err != nil {
		return err
	}
	defer s.db.Close()

	// Bootstrap database if not exists
	if err := s.createSchema(); err != nil {
		return err
	}

	// Listen server on specified port
	log.Printf("Starting server on :%d", s.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.router))
	return nil
}
