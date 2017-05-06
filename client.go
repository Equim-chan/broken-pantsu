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
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var likesList = [...]string{"Yuri", "Cosplay", "Crossdressing", "Cuddling", "Eyebrows", "Fangs", "Fantasy", "Futanari", "Genderbend", "Glasses", "Hentai", "Holding Hands", "Horror", "Housewife", "Humiliation", "Idol", "Incest", "Loli", "Maid", "Miko", "Monster Girl", "Muscles", "Netorare", "Nurse", "Office Lady", "Oppai", "Schoolgirl", "Sci-Fi", "Shota", "Slice-of-Life", "Socks", "Spread", "Stockings", "Swimsuit", "Teacher", "Tentacles", "Tomboy", "Tsundere", "Vanilla", "Warm Smiles", "Western", "Yandere", "Yaoi", "Yukata"} // len = 43

var (
	minimum uint8

	onlineUsers uint64 = 0
	clientsPool        = make(map[*Client]bool)
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
	MinimumSim      uint8
}

type ClientJSON struct {
	Username string   `json:"username"`
	Gender   bool     `json:"gender"`
	Likes    []string `json:"likes"`
	Timezone int8     `json:"Timezone"`
	// Token is private
}

func init() {
	if m, ok := os.LookupEnv("BP_MIN_SIM"); !ok {
		minimum = 5
	} else if u64_minimum, err := strconv.ParseUint(m, 10, 8); err != nil {
		log.Fatalln("BP_MIN_SIM:", err)
	} else {
		minimum = uint8(u64_minimum)
	}
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
		MinimumSim:                  minimum,
	}

	go c.runRecvQueue()
	go c.runSendQueue()

	c.addToPool()

	return c
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
	clientsPool[c] = true
	onlineUsers++
	locker.Unlock()
	broadcast <- &OutBoundMessage{"online users", strconv.FormatUint(onlineUsers, 10)}
}

func (c *Client) removeFromPool() {
	locker.Lock()
	delete(clientsPool, c)
	onlineUsers--
	locker.Unlock()
	broadcast <- &OutBoundMessage{"online users", strconv.FormatUint(onlineUsers, 10)}
}

// this function can guarantee no concurrent read from the same Conn
// note that this function will not handle "Indentity" messages, and works only after matched
func (c *Client) runRecvQueue() {
	for {
		var inMsg InBoundMessage

		if err := c.Conn.ReadJSON(&inMsg); err != nil {
			log.Println("DISCONNECTED:", c.Token)

			c.removeFromPool()

			c.internalDisconnectionSignal <- 1
			c.DisconnectionSignal <- 1
			// the deconstruction of Conn is handled by the outter defer func
			break
		}

		// messages arrived before matched are dropped by default, except for "ping"
		if inMsg.Type == "ping" {
			select {
			case c.SendQueue <- &OutBoundMessage{"pong", strconv.FormatInt(time.Now().UTC().UnixNano()/1e6, 10)}:
			default:
				break
			}
			continue
		}

		if c.Partner == nil {
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
			c.Conn.WriteJSON(outMsg)
			// if we cannot write (WriteJSON returns an error), then we must cannot read as well
			// therefore we assert that c.removeFromPool() has already been called
			// and what we need to do here is just ignoring this error

		case <-c.internalDisconnectionSignal:
			// now that we received this signal, then it must be from runRecvQueue
			// therefore, we make the same assertion as above
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

func (c *Client) ResetMinimumSim() {
	c.MinimumSim = minimum
}

func (c *Client) ToJsonStruct() *ClientJSON {
	return &ClientJSON{
		c.Username,
		c.Gender,
		c.Likes,
		c.Timezone,
	}
}
