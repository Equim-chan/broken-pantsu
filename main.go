package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type InBoundMessage struct {
	Username string `json:"username"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

type OutBoundMessage struct {
	Username string `json:"username"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

var (
	locker      sync.Mutex
	onlineUsers = 0
	clientsPool = make(map[*Client]bool) // true => not matched, false => matched
	clientsMap  = make(map[*Client]*Client)
	broadcast   = make(chan *OutBoundMessage, 10)
	upgrader    = websocket.Upgrader{}
)

func main() {
	http.Handle("/asset/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("token")
		if err != nil {
			exp := time.Now().Add(time.Minute * 2)
			// TODO: 随机化 Token，加入查重检查等
			tokenCookie := &http.Cookie{Name: "token", Value: "PENDING TO HAVE A HASH HERE", Expires: exp}
			http.SetCookie(w, tokenCookie)
		} else {
			// TODO: 检查 token
			log.Println(token)
		}
		http.ServeFile(w, r, "./public/index.html")
	})
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/chat.html")
	})
	http.HandleFunc("/register", register)
	http.HandleFunc("/love", handleConnections)

	go handleBroadcast()

	log.Println("Serving at localhost:56833...")
	log.Println("http://localhost:56833")
	log.Fatal(http.ListenAndServe("localhost:56833", nil))
}

type applicantJSON struct {
	Likes []string `json:"likes"`
}

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
	log.Println(t.Likes)
	// TODO: 返回成功 JSON
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: 保证 map 等的操作
	defer ws.Close()

	// TODO: 保证 client 的合法性
	client := &Client{}
	client.Conn = ws
	// TEST ONLY
	client.Username = "Equim"
	client.Likes = Yuri | Loli | Schoolgirl | Vanilla
	// |||

	locker.Lock()
	onlineUsers++
	clientsPool[client] = true
	locker.Unlock()

	broadcast <- &OutBoundMessage{"", "online users", strconv.Itoa(onlineUsers)}

	// 确认池里有至少两个人
	for {
		if len(clientsPool) > 1 {
			break
		}
	}

	// 死锁可能性？
	locker.Lock()
	partner, ok := clientsMap[client]
	if !ok {
		var maxSim uint8 = 0
		// TODO: 考虑更苛刻的条件，比如 maxSim < 3
		for partner == nil {
			for p, available := range clientsPool {
				if !available || p == client {
					continue
				}
				log.Println(client.SimilarityWith(p))
				if sim := client.SimilarityWith(p); sim > maxSim {
					partner = p
					maxSim = sim
				}
			}
		}
		clientsPool[client], clientsPool[partner] = false, false
		clientsMap[client], clientsMap[partner] = partner, client
	}
	locker.Unlock()

	log.Println(clientsMap[client].Username)

	for {
		var inMsg InBoundMessage
		if err := ws.ReadJSON(&inMsg); err != nil {
			log.Printf("recv error: %v", err)
			delete(clientsPool, client)
			break
		}

		outMsg := &OutBoundMessage{}
		outMsg.Username = inMsg.Username

		// TODO: 有没有 switch 的必要？
		switch inMsg.Type {
		case "chat":
			outMsg.Type = "chat"
			outMsg.Message = inMsg.Message
		case "typing":
			outMsg.Type = "typing"
			outMsg.Message = inMsg.Message
			log.Println("typing:", inMsg.Message)
		}

		if err := partner.Conn.WriteJSON(outMsg); err != nil {
			log.Printf("send error: %v", err)
			partner.Conn.Close()
			delete(clientsPool, partner)
		}

		// TODO: 这里超过缓存量的话会阻塞，考虑一下延迟的问题。如果对方大量 spam，或者甚至对方断线的情况
		// broadcast <- outMsg
	}
}

func handleBroadcast() {
	for {
		outMsg := <-broadcast

		// 线程安全？
		for client := range clientsPool {
			if err := client.Conn.WriteJSON(outMsg); err != nil {
				log.Printf("send error: %v", err)
				client.Conn.Close()
				delete(clientsPool, client)
			}
		}
	}
}
