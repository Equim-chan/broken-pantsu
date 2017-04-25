package main

import (
	"encoding/json"
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
	maxQueueLen int

	redisClient *redis.Client

	locker      sync.Mutex
	onlineUsers = 0
	clientsPool = make(map[*Client]bool) // true => not matched, false => matched
	singleQueue chan *Client
	broadcast   = make(chan *OutBoundMessage, 10)
	upgrader    = websocket.Upgrader{}
)

func init() {
	ok := false

	if address, ok = os.LookupEnv("BP_ADDR"); !ok {
		address = "localhost:56833"
	}

	if pubPath, ok = os.LookupEnv("BP_PUB_PATH"); !ok {
		pubPath = "./public"
	}
	pubPath, _ = filepath.Abs(pubPath)

	if m, ok := os.LookupEnv("BP_MAX_QUEUE_LEN"); ok {
		maxQueueLen, _ = strconv.Atoi(m)
	} else {
		maxQueueLen = 20
	}
	singleQueue = make(chan *Client, maxQueueLen)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err := redisClient.Ping().Err(); err != nil {
		log.Fatalln("REDIS SETUP ERROR:", err)
	}
}

func main() {
	http.HandleFunc("/", tokenDist)
	http.HandleFunc("/register", register)
	http.HandleFunc("/loveStream", handleConnections)
	http.HandleFunc("/chat", serveFile(filepath.Join(pubPath, "/chat.html")))
	http.Handle("/asset/", http.FileServer(http.Dir(pubPath)))

	go handleBroadcast()
	go matchingBus()

	log.Println("Serving at", address)
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

// 准备 deprecate 这个方法了
func register(w http.ResponseWriter, r *http.Request) {
	// TODO: 检查 cookie
	decoder := json.NewDecoder(r.Body)
	var t applicantJSON
	err := decoder.Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"malformed JSON request"}`)) // TODO: 转为常量
		return
	}
	// TODO: 检查标签的有效性(是否在预设的列表中)
	// 然后将标签化为 uint64 flag，保存到 redis
	// 这里暂且用 map
	//log.Println(t.Likes)
	// TODO: 返回成功 JSON
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// 获取来自客户端的第一条消息，即自己的资料
	var identity Identity
	if err := ws.ReadJSON(&identity); err != nil || !identity.IsValid() {
		ws.WriteJSON(&OutBoundMessage{"reject", "malformed request."})
		return
	}

	// 根据客户端身份定义 client 对象
	client := NewClient(ws, &identity)
	locker.Lock()
	onlineUsers++
	clientsPool[client] = true
	locker.Unlock()
	defer func() {
		locker.Lock()
		onlineUsers--
		delete(clientsPool, client)
		locker.Unlock()
		broadcast <- &OutBoundMessage{"online users", strconv.Itoa(onlineUsers)}
	}()

	broadcast <- &OutBoundMessage{"online users", strconv.Itoa(onlineUsers)}

	// 匹配
	if prevMatcherToken, err := redisClient.Get(client.Token).Result(); prevMatcherToken != "" && err != nil {
		// 如果此人是原配
		log.Println("FOUND PREV MATCHER", prevMatcherToken)
	} else {
		// 如果此人是新来的
		// 死锁可能性？
		singleQueue <- client
		client.AwaitPartner() // 这是个阻塞的方法
		client.SendQueue <- &MatchedNotify{"matched", client.Partner.ToJsonStruct()}
	}
	defer func() {
		// 要在这里，大做文章
		// 目前只是我方掉线之后对方重新回到单身队列
		// 但是我们的目标可不是这个！
		//
		// 我们把对方被动掉线后己方的状态称为失恋状态
		// 掉线的一方称为原配
		//
		// TODO:
		// * 区分主动下线与被动掉线
		//
		// 工作流:
		// * 向掉线的人的 Partner 发送对方被动掉线的消息
		// * partner 进入失恋队列，陷入等待状态，有几种情况
		//   * partner 主动下线 -> 直接销毁匹配
		//   * partner 被动掉线 -> 保留匹配，当一方上线时直接进入失恋状态
		//   * 有新 client 的 Token 为刚刚掉线的人的 Token -> 直接匹配
		partner := client.Partner
		locker.Lock()
		_, ok := clientsPool[partner]
		locker.Unlock()
		if !ok {
			return
		}
		partner.Partner = nil
		locker.Lock()
		clientsPool[partner] = true
		locker.Unlock()
		singleQueue <- partner
		// 因为下面的 Await 会阻塞，所以这里要异步进行
		go func() {
			partner.AwaitPartner()
			partner.SendQueue <- &MatchedNotify{"matched", partner.Partner.ToJsonStruct()}
		}()
	}()

	for {
		var inMsg InBoundMessage
		if err := ws.ReadJSON(&inMsg); err != nil {
			log.Printf("recv error: %v", err)
			locker.Lock()
			delete(clientsPool, client)
			locker.Unlock()
			break
		}

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

		client.Partner.SendQueue <- outMsg
	}
}

func handleBroadcast() {
	for {
		outMsg := <-broadcast

		locker.Lock()
		for client := range clientsPool {
			client.SendQueue <- outMsg
		}
		locker.Unlock()
	}
}

func matchingBus() {
	bufferQueue := []*Client{}
	for {
		var p *Client = nil
		var maxSim uint8 = 0
		// TODO: 考虑更苛刻的条件，比如 maxSim < 3
		c := <-singleQueue // c主动，p被动
		for {
			someSingle := <-singleQueue
			locker.Lock()
			_, ok0 := clientsPool[c]
			_, ok1 := clientsPool[someSingle]
			locker.Unlock()
			if !ok0 {
				singleQueue <- someSingle
				c = <-singleQueue
				continue
			}
			if !ok1 {
				continue
			}
			// TODO: 避免在极端情况下出现自己和自己匹配上的情况(利用token确认)
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
					log.Println("MATCHED")
					break
				}
				log.Println("P IS NIL")
				c = <-singleQueue
				maxSim = 0
			}
		}

		c.PartnerReceiver <- p
		p.PartnerReceiver <- c

		locker.Lock()
		clientsPool[c], clientsPool[p] = false, false
		locker.Unlock()

		multi := redisClient.Pipeline()
		multi.Set(c.Token, p.Token, time.Minute)
		multi.Set(p.Token, c.Token, time.Minute)
		if _, err := multi.Exec(); err != nil {
			log.Println("REDIS ERROR:", err)
			// ...
		}
		log.Println("SET IN REDIS:", c.Token, "<->", p.Token)
	}
}
