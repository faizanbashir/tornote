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

//import (
//
//	sw "github.com/osminogin/tornote/go"
//)
//
//
//func main() {
//
//	router := sw.NewRouter()
//
//	log.Fatal(http.ListenAndServe(":8080", router))
//}

package main

import (
	"log"

	"github.com/osminogin/tornote"
	"github.com/spf13/viper"
)

var (
	GitCommit string
)

func main() {
	// Configuration settings.
	v := viper.New()
	v.SetDefault("PORT", 8000)
	v.SetDefault("DATABASE_URL", "postgres://postgres:postgres@localhost/postgres")
	v.SetDefault("VERSION", GitCommit)

	v.SetConfigName(".env")
	v.SetConfigType("dotenv")
	v.AddConfigPath(".")
	v.ReadInConfig()
	v.AutomaticEnv()

	// Server init and run.
	var s tornote.Server
	s = tornote.NewServer(v.GetUint64("PORT"), v.GetString("DATABASE_URL"))
	s.Init()
	log.Fatal(s.Listen())
}
