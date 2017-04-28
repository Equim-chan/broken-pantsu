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
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type InBoundMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type OutBoundMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type MatchedNotify struct {
	Type    string      `json:"type"`
	Message *ClientJSON `json:"partnerInfo"`
}

var (
	address     string
	pubPath     string
	queueCap    int
	cookieAge   time.Duration
	lovelornAge time.Duration
	redisAddr   string
	redisPass   string
	redisDB     int

	redisClient *redis.Client

	locker        sync.Mutex
	singleQueue   chan *Client
	lovelornQueue chan *Client
	upgrader      = websocket.Upgrader{}
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
		queueCap = 1000
	}
	singleQueue = make(chan *Client, queueCap)
	lovelornQueue = make(chan *Client, queueCap)

	if e, ok := os.LookupEnv("BP_COOKIE_AGE"); ok {
		if cookieAge, err = time.ParseDuration(e); err != nil {
			panic("BP_COOKIE_AGE: " + err.Error())
		}
	} else {
		cookieAge = time.Hour * 168 // 24 * 7
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

	go handleBroadcast()
	go matchingBus()
	go reunionBus()
}

func main() {
	http.HandleFunc("/", tokenDist)
	http.HandleFunc("/loveStream", handleConnections)
	http.Handle("/asset/", http.FileServer(http.Dir(pubPath)))

	log.Println("Serving at " + address + ", GOOD LUCK!")
	log.Println("http://" + address)
	log.Fatal(http.ListenAndServe(address, nil))
}

type applicantJSON struct {
	Likes []string `json:"likes"`
}

func tokenDist(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("token"); err != nil {
		token := uuid.NewV4().String()
		exp := time.Now().Add(cookieAge)

		tokenCookie := &http.Cookie{Name: "token", Value: token, Expires: exp}
		http.SetCookie(w, tokenCookie)
		log.Println("HANDOUT THE TOKEN:", token)
	}

	http.ServeFile(w, r, filepath.Join(pubPath, "/index.html"))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer ws.Close()

	// 获取来自客户端的第一条消息，即自己的资料
	// 只有这步是由 main.go 负责的，因为这里还没有实例化 Client
	// 实例化后的收发队列处理均在 client.go 内
	var identity Identity
	if err := ws.ReadJSON(&identity); err != nil || !identity.IsValid() {
		ws.WriteJSON(&OutBoundMessage{"reject", "Malformed request."})
		return
	} else if exist, _ := redisClient.SIsMember("online", identity.Token).Result(); exist {
		ws.WriteJSON(&OutBoundMessage{"reject", "Sorry but you can't be a playboy."})
		return
	}
	redisClient.SAdd("online", identity.Token)
	defer redisClient.SRem("online", identity.Token)

	// 根据客户端身份定义 client 对象
	c := NewClient(ws, &identity)
	log.Println("CONNECTED:", c.Token)
	c.SendQueue <- &OutBoundMessage{"approved", "Valid request."}

	isInitiativeDisconnect := false

	// 匹配
	if t, _ := redisClient.Get(c.Token).Result(); t != "" {
		// 如果此人是之前断线的
		log.Println("FOUND A HEARTBROKEN WISHING TO FIND:", t)
		lovelornQueue <- c
		select {
		case c.Partner = <-c.PartnerReceiver:
			break
		case <-c.DisconnectionSignal:
			return
		}
		c.SendQueue <- &MatchedNotify{"reunion", c.Partner.ToJsonStruct()}
	} else {
		// 如果此人是新来的
		singleQueue <- c
		select {
		case c.Partner = <-c.PartnerReceiver:
			break
		case <-c.DisconnectionSignal:
			return
		}
		c.SendQueue <- &MatchedNotify{"matched", c.Partner.ToJsonStruct()}
	}
	defer func() {
		log.Println("DEFER IS TRIGGERED FOR:", c.Token)

		if isInitiativeDisconnect || c.Partner == nil {
			return
		}

		locker.Lock()
		_, ok := clientsPool[c.Partner]
		locker.Unlock()
		if !ok {
			return
		}

		c.Partner.HeartbrokenSignal <- 1
	}()

	// This is like an event loop
	for {
		select {
		case inMsg := <-c.RecvQueue:
			switch inMsg.Type {
			case "chat":
				// TODO: 确认是否会阻塞，而影响对断线的判断
				c.Partner.SendQueue <- &OutBoundMessage{"chat", inMsg.Message}
			case "typing":
				c.Partner.SendQueue <- &OutBoundMessage{"typing", inMsg.Message}
			case "offline":
				log.Println("INITIATIVE DISCONNECT FROM:", c.Token)
				c.Partner.SendQueue <- &OutBoundMessage{"switch", ""}
				c.Partner.GotSwitchedSignal <- 1

				isInitiativeDisconnect = true

				// TODO: 是否要检查 c.Partner 在不在连接池中
				c.Partner.Partner = nil
				singleQueue <- c.Partner
				return
			case "switch":
				log.Println("SWITCH IS TRIGGERED FOR:", c.Token)
				c.Partner.SendQueue <- &OutBoundMessage{"switch", ""}
				c.Partner.GotSwitchedSignal <- 1

				c.Partner = nil
				singleQueue <- c

				select {
				case c.Partner = <-c.PartnerReceiver:
					break
				case <-c.DisconnectionSignal:
					return
				}

				c.SendQueue <- &MatchedNotify{"matched", c.Partner.ToJsonStruct()}
			}

		case <-c.HeartbrokenSignal:
			p := c.Partner
			c.Partner = nil

			c.SendQueue <- &OutBoundMessage{"panic", ""}

			multi := redisClient.Pipeline()
			multi.Set(c.Token, p.Token, lovelornAge)
			multi.Set(p.Token, c.Token, lovelornAge)
			if _, err := multi.Exec(); err != nil {
				log.Println("REDIS ERROR:", err)
				// ...
			}
			log.Println("SET NEW LOVELORN PAIR IN REDIS:", c.Token, "<❤>", p.Token)

			p = nil
			lovelornQueue <- c

			select {
			case c.Partner = <-c.PartnerReceiver:
				break
			case <-c.DisconnectionSignal:
				log.Println("DISCONNECTED BEFORE RE-MATCHED")
				return
			}

			c.SendQueue <- &MatchedNotify{"reunion", c.Partner.ToJsonStruct()}

		case <-c.GotSwitchedSignal:
			c.Partner = nil
			singleQueue <- c

			// TODO: 这个 select 可不可以移植到外层的那个 select 里
			select {
			case c.Partner = <-c.PartnerReceiver:
				break
			case <-c.DisconnectionSignal:
				return
			}

			c.SendQueue <- &MatchedNotify{"matched", c.Partner.ToJsonStruct()}

		case <-c.DisconnectionSignal:
			// 直接触发 defer
			return
		}
	}
}

func matchingBus() {
	bufferQueue := []*Client{}
	for {
		// TODO: 考虑更苛刻的条件，比如 maxSim < 3
		c := <-singleQueue // c主动，p被动
		var p *Client = nil
		var maxSim uint8 = 0
		for {
			someSingle := <-singleQueue

			// 这里会检查两者中有没有其中哪个在等待的过程中下线了
			locker.Lock()
			_, ok0 := clientsPool[c]
			_, ok1 := clientsPool[someSingle]
			locker.Unlock()
			if !ok0 {
				c = someSingle
				continue
			}
			if !ok1 {
				continue
			}

			sim := c.SimilarityWith(someSingle)
			// 匹配相似度最高的。如果遇到相似度相同的，则匹配对方喜好数最小的
			if sim > maxSim || sim == maxSim && maxSim > 0 && someSingle.LikesCount() < p.LikesCount() {
				p = someSingle
				maxSim = sim
			}

			bufferQueue = append(bufferQueue, someSingle)

			if len(singleQueue) <= 0 {
				// 把 buffer 给 dump 出来
				for _, v := range bufferQueue {
					// p可能为nil，也可能为匹配到的人
					if p != v {
						singleQueue <- v
					}
				}
				bufferQueue = nil
				if p != nil {
					log.Println("MATCHED:", c.Token, "<❤>", p.Token)
					break
				}
				log.Println("P IS NIL")
				c = <-singleQueue
				maxSim = 0
			}
		}

		c.PartnerReceiver <- p
		p.PartnerReceiver <- c
	}
}

// 力挽狂澜
func reunionBus() {
	bufferQueue := []*Client{}
	for {
		c := <-lovelornQueue
		var p *Client = nil
		for {
			heartBroken := <-lovelornQueue

			locker.Lock()
			_, ok0 := clientsPool[c]
			_, ok1 := clientsPool[heartBroken]
			locker.Unlock()
			if !ok0 {
				c = heartBroken
				continue
			}
			if !ok1 {
				continue
			}

			if t, _ := redisClient.Get(c.Token).Result(); t == heartBroken.Token {
				p = heartBroken
			} else {
				bufferQueue = append(bufferQueue, heartBroken)
			}

			if len(lovelornQueue) <= 0 {
				// 把 buffer 给 dump 出来
				for _, v := range bufferQueue {
					// p可能为nil，也可能为匹配到的人
					if p != v {
						lovelornQueue <- v
					}
				}
				bufferQueue = nil
				if p != nil {
					log.Println("RE-MATCHED! CONGRATZ!", c.Token, "<❤>", p.Token)
					break
				}
				log.Println("P IS NIL")
				c = <-lovelornQueue
			}
		}

		c.PartnerReceiver <- p
		p.PartnerReceiver <- c

		multi := redisClient.Pipeline()
		multi.Del(c.Token)
		multi.Del(p.Token)
		if _, err := multi.Exec(); err != nil {
			log.Println("REDIS ERROR:", err)
			// ...
		}
		log.Println("REMOVED LOVELORN PAIR FROM REDIS:", c.Token, "<❤>", p.Token)
	}
}
