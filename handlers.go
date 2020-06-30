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
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

// MainFormHandler renders main form.
func MainFormHandler(s *server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.renderTemplate(w, "index.html", map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
		})
	})
}

// PublicFileHandler get file from bindata or return not found error.
func PublicFileHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path[1:]
	http.ServeFile(w, r, uri)
}

// Return status for health checks.
func HealthStatusHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Ping database connection
	w.WriteHeader(http.StatusOK)
}

// ReadNoteHandler print encrypted data for client-side decrypt and destroy note.
func ReadNoteHandler(s *server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		raw, _ := base64.RawURLEncoding.DecodeString(vars["id"])
		id, err := uuid.FromBytes(raw)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		n := &Note{UUID: id}

		// Get encrypted n or return 404
		err = s.db.Select(n)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		// Deferred n deletion
		defer func() {
			s.db.Delete(n)
		}()

		// Print encrypted n to user
		s.renderTemplate(w, "note.html", string(n.Data))
	})
}

// CreateNoteHandler save secret note to persistent datastore and return note ID.
func CreateNoteHandler(s *server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := &Note{
			UUID: uuid.New(),
			Data: []byte(r.FormValue("body")),
		}

		err := s.db.Insert(n)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, n.String())
	})
}
