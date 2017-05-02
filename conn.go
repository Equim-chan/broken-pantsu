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
	"time"

	"github.com/gorilla/websocket"
)

type InBoundMessage struct {
	Type    string `json:"type"`
	Message string `json:"msg"`
}

type OutBoundMessage InBoundMessage

type MatchedNotify struct {
	Type    string      `json:"type"`
	Message *ClientJSON `json:"partnerInfo"`
}

var (
	lovelornAge time.Duration

	broadcast = make(chan *OutBoundMessage, 10)
)

func init() {
	var err error = nil

	if e, ok := os.LookupEnv("BP_LOVELORN_AGE"); !ok {
		lovelornAge = time.Minute * 90
	} else if lovelornAge, err = time.ParseDuration(e); err != nil {
		log.Fatalln("BP_LOVELORN_AGE:", err)
	}

	go handleBroadcast()
}

func handleBroadcast() {
	for {
		outMsg := <-broadcast

		locker.RLock()
		for c := range clientsPool {
			c.SendQueue <- outMsg
		}
		locker.RUnlock()
	}
}

func handleConnections(ws *websocket.Conn) {
	// read the first message from client, ie. the identity(profile)
	// only this read is handled by main.go, as Client is not instantiated yet
	// after the instantiation, reading and sending will be handled by client.go
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

	// instantiate Client object according to the identity
	c := NewClient(ws, &identity)
	log.Println("CONNECTED:", c.Token)
	c.SendQueue <- &OutBoundMessage{"approved", ""}

	isInitiativeDisconnect := false

	// MATCHING
	if t, _ := redisClient.Get(c.Token).Result(); t != "" {
		// if this client is from a previously unexpectedly disconnected match
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
		// if this client is new
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

		locker.RLock()
		_, ok := clientsPool[c.Partner]
		locker.RUnlock()
		if !ok {
			return
		}

		c.Partner.HeartbrokenSignal <- 1
	}()

	// it serves as an event loop here
	for {
		select {
		case inMsg := <-c.RecvQueue:
			switch inMsg.Type {
			case "chat":
				c.Partner.SendQueue <- &OutBoundMessage{"chat", inMsg.Message}
			case "typing":
				c.Partner.SendQueue <- &OutBoundMessage{"typing", inMsg.Message}
			case "offline":
				log.Println("INITIATIVE DISCONNECT FROM:", c.Token)
				c.Partner.SendQueue <- &OutBoundMessage{"switch", ""}
				c.Partner.GotSwitchedSignal <- 1

				isInitiativeDisconnect = true

				// TODO: make sure whether to check c.Partner is in clientsPool or not
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

			// TODO: can we migrate this select to the outter select?
			select {
			case c.Partner = <-c.PartnerReceiver:
				break
			case <-c.DisconnectionSignal:
				return
			}

			c.SendQueue <- &MatchedNotify{"matched", c.Partner.ToJsonStruct()}

		case <-c.DisconnectionSignal:
			// trigger defer directly
			return
		}
	}
}
