/**
 * Copyright (c) 2017 Equim and other Broken Pantsu contributors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

var (
	cookieExpires time.Duration

	staticPath  string
	indexPath   string
	serveStatic func(http.ResponseWriter, *http.Request)

	upgrader = websocket.Upgrader{}
)

func init() {
	ok := false
	var err error = nil

	if staticPath, ok = os.LookupEnv("BP_ROOT_PATH"); !ok {
		staticPath = "./public"
	}
	if staticPath, err = filepath.Abs(staticPath); err != nil {
		log.Fatalln("BP_ROOT_PATH:", err)
	}
	serveStatic = http.FileServer(http.Dir(staticPath)).ServeHTTP
	indexPath = filepath.Join(staticPath, "/index.html")

	if e, ok := os.LookupEnv("BP_COOKIE_EXP"); !ok {
		cookieExpires = time.Hour * 48
	} else if cookieExpires, err = time.ParseDuration(e); err != nil {
		log.Fatalln("BP_COOKIE_EXP:", err)
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/loveStream", handleLove)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// in production, static files are better to be handled by other server application (Caddy in this case)
		serveStatic(w, r)
		return
	}

	var tokenCookie *http.Cookie
	exp := time.Now().Add(cookieExpires)

	if tokenCookie, _ = r.Cookie("token"); tokenCookie == nil {
		token := uuid.NewV4().String()
		tokenCookie = &http.Cookie{Name: "token", Value: token, Expires: exp}
	} else {
		// renew expires
		tokenCookie.Expires = exp
	}

	http.SetCookie(w, tokenCookie)
	http.ServeFile(w, r, indexPath)
}

func handleLove(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	handleConnections(ws)

	ws.Close()
}
