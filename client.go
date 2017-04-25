package main

import (
	"log"

	"github.com/gorilla/websocket"
)

var (
	likesList = [...]string{"Yuri", "Cosplay", "Crossdressing", "Cuddling", "Eyebrows", "Fangs", "Fantasy", "Futanari", "Genderbend", "Glasses", "Hentai", "Holding Hands", "Horror", "Housewife", "Humiliation", "Idol", "Incest", "Loli", "Maid", "Miko", "Monster Girl", "Muscles", "Netorare", "Nurse", "Office Lady", "Oppai", "Schoolgirl", "Sci-Fi", "Shota", "Slice-of-Life", "Socks", "Spread", "Stockings", "Swimsuit", "Teacher", "Tentacles", "Tomboy", "Tsundere", "Vanilla", "Warm Smiles", "Western", "Yandere", "Yaoi", "Yukata"} // len = 43
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
	Conn            *websocket.Conn
	SendQueue       chan interface{}
	Partner         *Client
	PartnerReceiver chan *Client
	likesMask       uint64
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

	sendQueue := make(chan interface{}, 20)
	ret := &Client{identity, conn, sendQueue, nil, make(chan *Client), likesMask}
	go ret.runSendQueue()

	return ret
}

func (i *Identity) IsValid() bool {
	return i.Username != "" &&
		i.Token != "" &&
		i.Timezone >= -12 &&
		i.Timezone <= 12
}

// 保证不会出现并发写
func (c *Client) runSendQueue() {
	for {
		outMsg := <-c.SendQueue

		if err := c.Conn.WriteJSON(outMsg); err != nil {
			log.Printf("send error: %v", err)
			c.Conn.Close()
			locker.Lock()
			delete(clientsPool, c)
			locker.Unlock()
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
