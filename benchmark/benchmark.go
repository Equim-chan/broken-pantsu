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
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

const (
	total          = 200000
	maxConcurrency = 1000
)

var likesList = [...]string{"Yuri", "Cosplay", "Crossdressing", "Cuddling", "Eyebrows", "Fangs", "Fantasy", "Futanari", "Genderbend", "Glasses", "Hentai", "Holding Hands", "Horror", "Housewife", "Humiliation", "Idol", "Incest", "Loli", "Maid", "Miko", "Monster Girl", "Muscles", "Netorare", "Nurse", "Office Lady", "Oppai", "Schoolgirl", "Sci-Fi", "Shota", "Slice-of-Life", "Socks", "Spread", "Stockings", "Swimsuit", "Teacher", "Tentacles", "Tomboy", "Tsundere", "Vanilla", "Warm Smiles", "Western", "Yandere", "Yaoi", "Yukata"}
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZабвгдеёжзийклмнопрстуфхцчшщъыьэюяяАБВГДЕЁЖЗИЙЛМНОПРСТУФХЦЧШЩЪЫЬЭЮ萌破胖次")

type Identity struct {
	Username string   `json:"username"`
	Gender   bool     `json:"gender"`
	Likes    []string `json:"likes"`
	Timezone int8     `json:"timezone"`
	Token    string   `json:"token"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.Lmicroseconds)
	log.Println("Max concurrency", maxConcurrency, "of", total, "requests")
}

func main() {
	limiter := make(chan uint8, maxConcurrency)
	doneSignal := make(chan uint8)

	dialer := websocket.Dialer{}
	url := "ws://localhost:56833/loveStream"

	for i := 1; i <= total; i++ {
		if i%(total/20) == 0 {
			log.Printf("%d/%d (%d%%)\n", i, total, 100*i/total)
		}

		if i < total {
			limiter <- 1
		} else {
			limiter <- 2
		}

		go func() {
			ws, _, _ := dialer.Dial(url, nil)
			ws.WriteJSON(randIdentity())
			//ws.Close()

			if <-limiter == 2 {
				doneSignal <- 1
			}
		}()
	}

	<-doneSignal
	log.Println("RESOLVED?")
}

func randIdentity() *Identity {
	username := randStringRunes(rand.Intn(24) + 1)
	var gender bool
	if rand.Intn(2) == 0 {
		gender = true
	} else {
		gender = false
	}
	likes := []string{}
	for j := rand.Intn(44); j > 0; j-- {
		likes = append(likes, likesList[rand.Intn(44)])
	}
	timezone := int8(rand.Intn(25) - 12)
	token := uuid.NewV4().String()

	return &Identity{
		Username: username,
		Gender:   gender,
		Likes:    likes,
		Timezone: timezone,
		Token:    token,
	}
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
