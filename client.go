package main

import (
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

var likesList = [...]string{"Yuri", "Cosplay", "Crossdressing", "Cuddling", "Eyebrows", "Fangs", "Fantasy", "Futanari", "Genderbend", "Glasses", "Hentai", "Holding Hands", "Horror", "Housewife", "Humiliation", "Idol", "Incest", "Loli", "Maid", "Miko", "Monster Girl", "Muscles", "Netorare", "Nurse", "Office Lady", "Oppai", "Schoolgirl", "Sci-Fi", "Shota", "Slice-of-Life", "Socks", "Spread", "Stockings", "Swimsuit", "Teacher", "Tentacles", "Tomboy", "Tsundere", "Vanilla", "Warm Smiles", "Western", "Yandere", "Yaoi", "Yukata"} // len = 43

var (
	onlineUsers = 0
	clientsPool = make(map[*Client]bool) // true => not matched, false => matched
	broadcast   = make(chan *OutBoundMessage, 10)
)

type Identity struct {
	Username string   `json:"username"`
	Gender   bool     `json:"gender"` // true => female, false => male
	Likes    []string `json:"likes"`
	Timezone int8     `json:"timezone"`
	Token    string   `json:"token"`
}

type Client struct {
	*Identity
	Conn                *websocket.Conn
	DisconnectionSignal chan uint8 // uint8 备用作为信号类型
	RecvQueue           chan *InBoundMessage
	SendQueue           chan interface{}
	Partner             *Client
	PartnerReceiver     chan *Client
	likesMask           uint64
}

type ClientJSON struct {
	Username string   `json:"username"`
	Gender   bool     `json:"gender"`
	Likes    []string `json:"likes"`
	Timezone int8     `json:"Timezone"`
	// Token is private
}

func NewClient(conn *websocket.Conn, identity *Identity) *Client {
	var likesMask uint64 = 0
	sanitizedLikes := []string{}

	// 将输入过滤，并得到 likesMask
	for _, item := range identity.Likes {
		var mask uint64 = 0
		for pos, value := range likesList {
			if item == value {
				mask = 1 << uint8(pos)
				break
			}
		}
		if mask != 0 {
			if t := likesMask | mask; t != likesMask {
				sanitizedLikes = append(sanitizedLikes, item)
				likesMask = t
			}
		}
	}
	identity.Likes = sanitizedLikes

	c := &Client{
		Identity:            identity,
		Conn:                conn,
		DisconnectionSignal: make(chan uint8),
		RecvQueue:           make(chan *InBoundMessage, 20),
		SendQueue:           make(chan interface{}, 20),
		Partner:             nil,
		PartnerReceiver:     make(chan *Client),
		likesMask:           likesMask,
	}
	go c.runRecvQueue()
	go c.runSendQueue()

	c.addToPool()

	broadcast <- &OutBoundMessage{"online users", strconv.Itoa(onlineUsers)}

	return c
}

func handleBroadcast() {
	for {
		outMsg := <-broadcast

		locker.Lock()
		for c := range clientsPool {
			c.SendQueue <- outMsg
		}
		locker.Unlock()
	}
}

func (i *Identity) IsValid() bool {
	return i.Username != "" &&
		len(i.Username) <= 20 &&
		i.Token != "" &&
		i.Timezone >= -12 &&
		i.Timezone <= 12
}

func (c *Client) addToPool() {
	locker.Lock()
	onlineUsers++
	clientsPool[c] = true
	locker.Unlock()
}

func (c *Client) removeFromPool() {
	locker.Lock()
	delete(clientsPool, c)
	onlineUsers--
	locker.Unlock()
}

// 保证不会出现并发读
// 注意，这个函数不会接管身份验证，只在匹配成功后有效
func (c *Client) runRecvQueue() {
	for {
		var inMsg InBoundMessage

		if err := c.Conn.ReadJSON(&inMsg); err != nil {
			log.Println("(FROM READ) DISCONNECTED:", c.Token)

			c.removeFromPool()

			broadcast <- &OutBoundMessage{"online users", strconv.Itoa(onlineUsers)}
			c.DisconnectionSignal <- 1
			// Conn 的回收由上层的 defer 负责
			break
		}

		if c.Partner == nil {
			// 默认会把匹配前发送的包丢弃
			continue
		}

		// 将消息暴露给外部处理
		c.RecvQueue <- &inMsg
	}
}

// 保证不会出现并发写
// outMsg 的类型是 interface{}，所以可以发送的对象类型不一定要是 OutBoundMessage
func (c *Client) runSendQueue() {
	for {
		select {
		case outMsg := <-c.SendQueue:
			if err := c.Conn.WriteJSON(outMsg); err != nil {
				log.Println("(FROM WRITE) DISCONNECTED:", c.Token)

				c.removeFromPool()

				broadcast <- &OutBoundMessage{"online users", strconv.Itoa(onlineUsers)}
				c.DisconnectionSignal <- 2
				return
			}
		case <-c.DisconnectionSignal:
			// 既然能收到这个信号，那么必定是从 runRecvQueue 或 runSendQueue 来的
			// 所以就断言，已经做了 removeFromPool 的处理，这里只要回收这个 goroutine 就好
			return
		}
	}
}

func (c *Client) parseMask(f uint64) uint8 {
	var ret uint8
	for ret = 0; f > 0; ret++ {
		f &= (f - 1)
	}

	return ret
}

func (c *Client) LikesCount() uint8 {
	return c.parseMask(c.likesMask)
}

func (c *Client) SimilarityWith(p *Client) uint8 {
	return c.parseMask(c.likesMask & p.likesMask)
}

func (c *Client) AwaitPartner() {
	c.Partner = <-c.PartnerReceiver
}

func (c *Client) ToJsonStruct() *ClientJSON {
	return &ClientJSON{
		c.Username,
		c.Gender,
		c.Likes,
		c.Timezone,
	}
}
