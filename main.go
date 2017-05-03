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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	_ "net/http/pprof"

	"github.com/go-redis/redis"
)

var (
	VERSION string

	address   string
	redisAddr string
	redisPass string
	redisDB   int

	// global
	redisClient *redis.Client
	locker      sync.RWMutex
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	ok := false
	var err error = nil

	if _, ok = os.LookupEnv("BP_QUIET"); ok {
		log.SetOutput(ioutil.Discard)
	}

	if address, ok = os.LookupEnv("BP_ADDR"); !ok {
		address = "localhost:56833"
	}

	if redisAddr, ok = os.LookupEnv("BP_REDIS_ADDR"); !ok {
		redisAddr = "localhost:6379"
	}

	if redisPass, ok = os.LookupEnv("BP_REDIS_PASS"); !ok {
		redisPass = ""
	}

	if d, ok := os.LookupEnv("BP_REDIS_DB"); !ok {
		redisDB = 0
	} else if redisDB, err = strconv.Atoi(d); err != nil {
		log.Fatalln("BP_REDIS_DB:", err)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
	})
	if err = redisClient.Ping().Err(); err != nil {
		log.Fatalln("REDIS INIT ERROR:", err)
	}

	go hookInterruptAndCleanUp()
}

func hookInterruptAndCleanUp() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	<-c
	log.Println("SIGINT detected, now cleaning up...")
	// delete online set only, as it doesn't have expire. No change to lovelorn records
	if err := redisClient.Del("online").Err(); err != nil {
		log.Println("REDIS CLEAN UP ERROR:", err)
		log.Fatalln("exited with code 1")
	} else {
		log.Println("gracefully exited with code 0")
		os.Exit(0)
	}
}

func main() {
	log.Printf(`
　　 ＿＿＿＿＿    Broken Pantsu
　　(＼　 ∞　ﾉ      %s
　　 ＼ヽ　　/
　　　 ヽ)⌒ﾉ      Serving at %s, PID: %d
　　　　　￣         GOOD LUCK!
`, VERSION, address, os.Getpid())
	log.Fatalln(http.ListenAndServe(address, nil))
}
