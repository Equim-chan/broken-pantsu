package main

import (
	"github.com/gorilla/websocket"
)

const (
	Yuri = 1 << iota
	Cosplay
	Crossdressing
	Cuddling
	Eyebrows
	Fangs
	Fantasy
	Futanari
	Genderbend
	Glasses
	Hentai
	HoldingHands
	Horror
	Housewife
	Humiliation
	Idol
	Incest
	Loli
	Maid
	Miko
	MonsterGirl
	Muscles
	Netorare
	Nurse
	OfficeLady
	Oppai
	Schoolgirl
	SciFi
	Shota
	SliceOfLife
	Socks
	Spread
	Stockings
	Swimsuit
	Teacher
	Tentacles
	Tomboy
	Tsundere
	Vanilla
	WarmSmiles
	Western
	Yandere
	Yaoi
	Yukata // = 2 ^ 43
)

type Client struct {
	Conn     *websocket.Conn
	Username string
	Likes    uint64
	// Token string
}

func (c *Client) SimilarityWith(p *Client) uint8 {
	var similarity uint8
	bitAnd := c.Likes & p.Likes
	for similarity = 0; bitAnd > 0; similarity++ {
		bitAnd &= (bitAnd - 1)
	}

	return similarity
}
