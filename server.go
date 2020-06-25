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
	"log"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

type server struct {
	// Listen port
	Port uint64
	// Secret used for encryption/decription
	Key string
	// Data source name
	DSN string
	// PostgreSQL connection
	db *pg.DB
}

type Server interface {
	Run() error
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

// Running server.
func (s *server) Run() error {
	r := mux.NewRouter().StrictSlash(true)

	// Setup middlewares
	csrfMiddleware := csrf.Protect(
		s.getSecretHash(),
		csrf.FieldName("csrf_token"),
		csrf.SameSite(csrf.SameSiteStrictMode),
	)
	r.Use(csrfMiddleware)

	// HTTP handlers
	r.HandleFunc("/", mainFormHandler).Methods("GET")
	//r.PathPrefix("/favicon.ico").HandlerFunc(publicFileHandler).Methods("GET")
	r.PathPrefix("/public/").HandlerFunc(publicFileHandler).Methods("GET")
	r.Handle("/note", createNoteHandler(s)).Methods("POST")
	r.Handle("/{id}", readNoteHandler(s)).Methods("GET")

	// Connecting to database
	if err := s.connectDB(); err != nil {
		return err
	}
	defer s.db.Close()

	// Bootstrap database if not exists
	if err := s.createSchema(); err != nil {
		return err
	}

	// Pre-compile templates
	if err := compileTemplates(); err != nil {
		return err
	}

	// Listen server on specified port
	log.Printf("Starting server on :%d", s.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.Port), r))
	return nil
}
