package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

/*
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
*/

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var (
	locker      sync.Mutex
	onlineUsers = 0
	clientsPool = make(map[*websocket.Conn]bool)
	broadcast   = make(chan Message)
	upgrader    = websocket.Upgrader{}
)

func main() {
	http.HandleFunc("/stream", handleConnections)
	http.HandleFunc("/access", access)
	//http.Handle("/", http.FileServer(http.Dir("./public")))
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

	go handleMessages()

	log.Println("Serving at *:5000...")
	log.Println("http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

type applicantJSON struct {
	Likes []string `json:"likes"`
}

func access(w http.ResponseWriter, r *http.Request) {
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
	// 然后将标签化为 int flag，保存到 redis
	// 这里暂且用 map
	log.Println(t.Likes)
	// TODO: 返回成功 JSON
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clientsPool[ws] = true

	for {
		var msg Message
		if err := ws.ReadJSON(&msg); err != nil {
			log.Printf("recv error: %v", err)
			delete(clientsPool, ws)
			break
		}

		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clientsPool {
			if err := client.WriteJSON(msg); err != nil {
				log.Printf("send error: %v", err)
				client.Close()
				delete(clientsPool, client)
			}
		}
	}
}

/*
func match() {
	// Vanilla Yuri Yaoi Hentai Loli Shota Slice-of-Life Schoolgirl
	// 1       2    4    8      16   32    64			 128
	// 0x1     0x2  0x4  0x8    0x10 0x20  0x40			 0x80
	randomGuy0 := 0x1 | 0x2 | 0x10 | 0x80
	randomGuy1 := 0x2 | 0x4 | 0x8 | 0x10 | 0x20 | 0x80
	and := randomGuy0 & randomGuy1
	var similarity int
	for similarity = 0; and > 0; similarity++ {
		and &= (and - 1)
	}
	log.Printf("%d\n", similarity)
}
*/
