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
	address   string
	pubPath   string
	queueCap  int
	expires   time.Duration
	redisAddr string
	redisPass string
	redisDB   int

	redisClient *redis.Client

	locker        sync.Mutex
	singleQueue   chan *Client
	lovelornQueue chan *Client
	upgrader      = websocket.Upgrader{}
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	ok := false

	if address, ok = os.LookupEnv("BP_ADDR"); !ok {
		address = "localhost:56833"
	}

	if pubPath, ok = os.LookupEnv("BP_ROOT_PATH"); !ok {
		pubPath = "./public"
	}
	pubPath, _ = filepath.Abs(pubPath)

	if m, ok := os.LookupEnv("BP_QUEUE_CAP"); ok {
		queueCap, _ = strconv.Atoi(m)
	} else {
		queueCap = 20
	}
	singleQueue = make(chan *Client, queueCap)
	lovelornQueue = make(chan *Client, queueCap)

	if e, ok := os.LookupEnv("BP_SESSION_AGE"); ok {
		m, _ := strconv.Atoi(e)
		expires = time.Minute * time.Duration(m)
	} else {
		expires = time.Minute * 2
	}

	if redisAddr, ok = os.LookupEnv("BP_REDIS_ADDR"); !ok {
		redisAddr = "localhost:6379"
	}

	if redisPass, ok = os.LookupEnv("BP_REDIS_PASS"); !ok {
		redisPass = ""
	}

	if d, ok := os.LookupEnv("BP_REDIS_DB"); ok {
		redisDB, _ = strconv.Atoi(d)
	} else {
		redisDB = 0
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
	})
	if err := redisClient.Ping().Err(); err != nil {
		log.Fatalln("REDIS INIT ERROR:", err)
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
		exp := time.Second * 20
		expStamp := time.Now().Add(exp)

		tokenCookie := &http.Cookie{Name: "token", Value: token, Expires: expStamp}
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
	}

	// 根据客户端身份定义 client 对象
	c := NewClient(ws, &identity)
	log.Println("CONNECTED:", c.Token)

	// 匹配
	// TODO: 处理在匹配未完成时下线的情况
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
		// 要在这里，大做文章
		// 目前只是我方掉线之后对方重新回到单身队列
		// 但是我们的目标可不是这个！
		//
		// 我们把对方被动掉线后己方的状态称为失恋状态
		// 掉线的一方称为原配
		//
		// 工作流:
		// * 向掉线的人的 Partner 发送对方被动掉线的消息
		// * partner 进入失恋队列，陷入等待状态，有几种情况
		//   * partner 主动下线 -> 直接销毁匹配
		//   * partner 被动掉线 -> 保留匹配，当一方上线时直接进入失恋状态
		//   * 有新 client 的 Token 为刚刚掉线的人的 Token -> 直接匹配
		p := c.Partner

		locker.Lock()
		_, ok := clientsPool[p]
		locker.Unlock()
		if !ok {
			return
		}

		p.Partner = nil

		// TODO: 在这里检测是否是主动下线，如果是被动下线则继续执行下面的
		multi := redisClient.Pipeline()
		multi.Set(c.Token, p.Token, expires)
		multi.Set(p.Token, c.Token, expires)
		if _, err := multi.Exec(); err != nil {
			log.Println("REDIS ERROR:", err)
			// ...
		}
		log.Println("SET NEW LOVELORN PAIR IN REDIS:", c.Token, "<❤>", p.Token)

		//singleQueue <- p
		lovelornQueue <- p

		// 因为下面的 Await 会阻塞而影响后面的 defer（其实也就个 ws.Close），所以这里要异步进行
		go func() {
			// TODO: 处理在这个时候 p 断连的情况
			// 可以换个想法，比如这时候把这个信号发给 p，就可以在 p 处处理，而不是在这个 defer 里处理了
			select {
			case p.Partner = <-p.PartnerReceiver:
				break
			case <-p.DisconnectionSignal:
				log.Println("DISCONNECTED BEFORE RE-MATCHED")
				return
			}
			//p.SendQueue <- &MatchedNotify{"matched", p.Partner.ToJsonStruct()}
			p.SendQueue <- &MatchedNotify{"reunion", p.Partner.ToJsonStruct()}
		}()
	}()

	for {
		select {
		case inMsg := <-c.RecvQueue:
			outMsg := &OutBoundMessage{}

			// TODO: 有没有 switch 的必要？
			switch inMsg.Type {
			case "chat":
				outMsg.Type = "chat"
				outMsg.Message = inMsg.Message
			case "typing":
				outMsg.Type = "typing"
				outMsg.Message = inMsg.Message
			}

			// TODO: 确认是否会阻塞，而影响对断线的判断
			c.Partner.SendQueue <- outMsg
		case s := <-c.DisconnectionSignal:
			// 为了方便上面 defer 的 go func 里监听这个 channel 的 select
			c.DisconnectionSignal <- s
			// 触发 defer
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
