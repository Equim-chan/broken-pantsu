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
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

var (
	staticPath        string
	staticHandlerFunc = http.FileServer(http.Dir(staticPath)).ServeHTTP
	indexPath         = filepath.Join(staticPath, "/index.html")
	upgrader          = websocket.Upgrader{}
)

func registerHandlersToDefaultMux() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/loveStream", handleLove)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// filtered index.html here in case of not distributing the token cookie
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		// in production, static files are better to be handled by other server application (Caddy in this case)
		staticHandlerFunc(w, r)
		return
	}

	if _, err := r.Cookie("token"); err != nil {
		token := uuid.NewV4().String()
		exp := time.Now().Add(cookieAge)

		tokenCookie := &http.Cookie{Name: "token", Value: token, Expires: exp}
		http.SetCookie(w, tokenCookie)
	}

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
