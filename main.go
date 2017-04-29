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
	"strconv"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/go-redis/redis"
)

var (
	address     string
	pubPath     string
	queueCap    int
	cookieAge   time.Duration
	lovelornAge time.Duration
	redisAddr   string
	redisPass   string
	redisDB     int

	// global
	redisClient *redis.Client
	locker      sync.Mutex
)

func init() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("INIT ERROR:", err)
			os.Exit(1)
		}
	}()

	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	ok := false
	var err error = nil

	if address, ok = os.LookupEnv("BP_ADDR"); !ok {
		address = "localhost:56833"
	}

	if pubPath, ok = os.LookupEnv("BP_ROOT_PATH"); !ok {
		pubPath = "./public"
	}
	if pubPath, err = filepath.Abs(pubPath); err != nil {
		panic("BP_ROOT_PATH: " + err.Error())
	}

	if m, ok := os.LookupEnv("BP_QUEUE_CAP"); ok {
		if queueCap, err = strconv.Atoi(m); err != nil {
			panic("BP_QUEUE_CAP: " + err.Error())
		}
	} else {
		queueCap = 300
	}
	singleQueue = make(chan *Client, queueCap)   // declared in match.go
	lovelornQueue = make(chan *Client, queueCap) // declared in match.go

	// used in token.go
	if e, ok := os.LookupEnv("BP_COOKIE_AGE"); ok {
		if cookieAge, err = time.ParseDuration(e); err != nil {
			panic("BP_COOKIE_AGE: " + err.Error())
		}
	} else {
		cookieAge = time.Hour * 168 // 168 == 24 * 7
	}

	if e, ok := os.LookupEnv("BP_LOVELORN_AGE"); ok {
		if lovelornAge, err = time.ParseDuration(e); err != nil {
			panic("BP_LOVELORN_AGE: " + err.Error())
		}
	} else {
		lovelornAge = time.Minute * 90
	}

	if redisAddr, ok = os.LookupEnv("BP_REDIS_ADDR"); !ok {
		redisAddr = "localhost:6379"
	}

	if redisPass, ok = os.LookupEnv("BP_REDIS_PASS"); !ok {
		redisPass = ""
	}

	if d, ok := os.LookupEnv("BP_REDIS_DB"); ok {
		if redisDB, err = strconv.Atoi(d); err != nil {
			panic("BP_REDIS_DB: " + err.Error())
		}
	} else {
		redisDB = 0
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
	})
	if err = redisClient.Ping().Err(); err != nil {
		panic("REDIS INIT ERROR: " + err.Error())
	}

	go handleBroadcast() // declared in conn.go
	go matchBus()        // declared in match.go
	go reunionBus()      // declared in match.go
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tokenDist(w, r)
		http.ServeFile(w, r, filepath.Join(pubPath, "/index.html"))
	})
	http.HandleFunc("/loveStream", handleConnections)
	http.Handle("/asset/", http.FileServer(http.Dir(pubPath)))

	log.Println("Serving at " + address + ", GOOD LUCK!")
	log.Println("http://" + address)
	log.Fatal(http.ListenAndServe(address, nil))
}
