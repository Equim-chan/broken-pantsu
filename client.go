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
	"strconv"

	"github.com/gorilla/websocket"
)

var likesList = [...]string{"Yuri", "Cosplay", "Crossdressing", "Cuddling", "Eyebrows", "Fangs", "Fantasy", "Futanari", "Genderbend", "Glasses", "Hentai", "Holding Hands", "Horror", "Housewife", "Humiliation", "Idol", "Incest", "Loli", "Maid", "Miko", "Monster Girl", "Muscles", "Netorare", "Nurse", "Office Lady", "Oppai", "Schoolgirl", "Sci-Fi", "Shota", "Slice-of-Life", "Socks", "Spread", "Stockings", "Swimsuit", "Teacher", "Tentacles", "Tomboy", "Tsundere", "Vanilla", "Warm Smiles", "Western", "Yandere", "Yaoi", "Yukata"} // len = 43

var (
	onlineUsers = 0
	clientsPool = make(map[*Client]bool)
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
	likesFlag uint64

	Conn *websocket.Conn

	DisconnectionSignal         chan uint8 // uint8 is reserved for indicating signal types
	internalDisconnectionSignal chan uint8
	HeartbrokenSignal           chan uint8
	GotSwitchedSignal           chan uint8

	RecvQueue chan *InBoundMessage
	SendQueue chan interface{}

	Partner         *Client
	PartnerReceiver chan *Client
}

type ClientJSON struct {
	Username string   `json:"username"`
	Gender   bool     `json:"gender"`
	Likes    []string `json:"likes"`
	Timezone int8     `json:"Timezone"`
	// Token is private
}

func NewClient(conn *websocket.Conn, identity *Identity) *Client {
	var likesFlag uint64 = 0
	sanitizedLikes := []string{}

	// filter the input and get likesFlag
	for _, item := range identity.Likes {
		var mask uint64 = 0
		for pos, value := range likesList {
			if item == value {
				mask = 1 << uint8(pos)
				break
			}
		}
		if mask != 0 {
			if t := likesFlag | mask; t != likesFlag {
				sanitizedLikes = append(sanitizedLikes, item)
				likesFlag = t
			}
		}
	}
	identity.Likes = sanitizedLikes

	c := &Client{
		Identity:                    identity,
		likesFlag:                   likesFlag,
		Conn:                        conn,
		DisconnectionSignal:         make(chan uint8),
		internalDisconnectionSignal: make(chan uint8),
		HeartbrokenSignal:           make(chan uint8),
		GotSwitchedSignal:           make(chan uint8),
		RecvQueue:                   make(chan *InBoundMessage, 20),
		SendQueue:                   make(chan interface{}, 20),
		Partner:                     nil,
		PartnerReceiver:             make(chan *Client),
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

// this function can guarantee no concurrent read from the same Conn
// note that this function will not handle "Indentity" messages, and works only after matched
func (c *Client) runRecvQueue() {
	for {
		var inMsg InBoundMessage

		if err := c.Conn.ReadJSON(&inMsg); err != nil {
			log.Println("(FROM READ) DISCONNECTED:", c.Token)

			c.removeFromPool()

			broadcast <- &OutBoundMessage{"online users", strconv.Itoa(onlineUsers)}
			c.internalDisconnectionSignal <- 1
			c.DisconnectionSignal <- 1
			// the deconstruction of Conn is handled by the outter defer func
			break
		}

		if c.Partner == nil {
			// messages sent before matched are dropped by default
			continue
		}

		// expose the message to outter process
		c.RecvQueue <- &inMsg
	}
}

// this function can guarantee no concurrent write to the same Conn
// the type of outMsg is interface{}, therefore the object to be sent is not restricted to OutBoundMessage
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
		case <-c.internalDisconnectionSignal:
			// now that we received this signal, then it must be from runRecvQueue
			// therefore we assert that c.removeFromPool() has already been called
			// and what we need to do here is just destroying this goroutine
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
	return c.parseMask(c.likesFlag)
}

func (c *Client) SimilarityWith(p *Client) uint8 {
	return c.parseMask(c.likesFlag & p.likesFlag)
}

func (c *Client) ToJsonStruct() *ClientJSON {
	return &ClientJSON{
		c.Username,
		c.Gender,
		c.Likes,
		c.Timezone,
	}
}
