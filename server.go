// Copyright 2016 Vladimir Osintsev <osintsev@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
// See the COPYING file in the main directory for details.

package tornote

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	DB   *sql.DB
	Host string
	Key  string
}

// Open database connection.
func (s *Server) OpenDB(path string) (err error) {
	// XXX: Check err from sql.Open
	s.DB, err = sql.Open("sqlite3", path)
	err = s.DB.Ping()
	return
}

// Running daemon process.
func (s *Server) Run() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", frontPageHandler).Methods("GET")

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Handle("/note", saveNoteHandler(s.DB)).Methods("POST")
	api.Handle("/note/{id}", readNoteHandler(s.DB)).Methods("GET")

	log.Printf("Starting tornote server on %s", s.Host)
	log.Fatal(http.ListenAndServe(s.Host, router))
}
