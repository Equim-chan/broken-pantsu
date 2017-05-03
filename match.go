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
)

var (
	singleQueue   chan *Client
	lovelornQueue chan *Client
	queueCap      int
)

func init() {
	var err error = nil

	if m, ok := os.LookupEnv("BP_QUEUE_CAP"); !ok {
		queueCap = 1000
	} else if queueCap, err = strconv.Atoi(m); err != nil {
		log.Fatalln("BP_QUEUE_CAP:", err)
	}

	singleQueue = make(chan *Client, queueCap)
	lovelornQueue = make(chan *Client, queueCap)

	go matchBus()
	go reunionBus()
}

func matchBus() {
	bufferQueue := []*Client{}
	for {
		// c is initialtive and p is passive
		c := <-singleQueue
		var p *Client = nil
		var maxSim uint8 = 0
		for {
			someSingle := <-singleQueue

			// check if one of them is disconnected while waiting for matching
			locker.RLock()
			_, ok0 := clientsPool[c]
			_, ok1 := clientsPool[someSingle]
			locker.RUnlock()
			if !ok0 {
				c = someSingle
				continue
			}
			if !ok1 {
				continue
			}

			sim := c.SimilarityWith(someSingle)
			// the candidate must have higher similarity than MinimumSim (initially, 5)
			// the candidate should be the one who has the highest similarity with c among the whole queue
			// if there are many candidates with the same similarity with c, then match the one who has the least likes
			if sim >= c.MinimumSim && sim >= someSingle.MinimumSim &&
				(sim > maxSim ||
					sim == maxSim && maxSim > 0 && someSingle.LikesCount() < p.LikesCount()) {
				p = someSingle
				maxSim = sim
			}

			bufferQueue = append(bufferQueue, someSingle)

			// end of cycle
			if len(singleQueue) <= 0 {
				// dump the buffer
				for _, v := range bufferQueue {
					// p may be either nil or the matched one
					if p != v {
						singleQueue <- v
					}
				}
				bufferQueue = nil

				// matched in this cycle
				if p != nil {
					c.ResetMinimumSim()
					p.ResetMinimumSim()
					log.Println("MATCHED:", c.Token, "<❤>", p.Token)
					break
				}

				// not matched in this cycle
				// reduce the minimum
				if c.MinimumSim > 0 {
					c.MinimumSim--
				}
				singleQueue <- c

				// change the initialtive side for the next cycle
				c = <-singleQueue
				maxSim = 0
			}
		}

		select {
		case c.PartnerReceiver <- p:
		default:
			c = <-singleQueue
			continue
		}
		select {
		case p.PartnerReceiver <- c:
		default:
			continue
		}
	}
}

func reunionBus() {
	bufferQueue := []*Client{}
	for {
		c := <-lovelornQueue
		var p *Client = nil
		for {
			heartBroken := <-lovelornQueue

			locker.RLock()
			_, ok0 := clientsPool[c]
			_, ok1 := clientsPool[heartBroken]
			locker.RUnlock()
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

			if p != nil || len(lovelornQueue) <= 0 {
				for _, v := range bufferQueue {
					if p != v {
						lovelornQueue <- v
					}
				}
				bufferQueue = nil
				if p != nil {
					log.Println("REUNION:", c.Token, "<❤>", p.Token)
					break
				}
				lovelornQueue <- c
				c = <-lovelornQueue
			}
		}

		select {
		case c.PartnerReceiver <- p:
		default:
			c = <-lovelornQueue
			continue
		}
		select {
		case p.PartnerReceiver <- c:
		default:
			continue
		}

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
