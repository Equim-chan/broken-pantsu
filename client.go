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

type Identity struct {
	Username string   `json:"username"`
	Gender   bool     `json:"gender"` // true => female, false => male
	Likes    []string `json:"likes"`
	Timezone int8     `json:"timezone"`
	// Token string
}

// 出于安全考虑，不继承 Identity（因为 Likes 可以被 spam）
type Client struct {
	Conn     *websocket.Conn
	Username string
	Gender   bool
	Likes    uint64
	Timezone int8
	// Token string
}

func NewClient(conn *websocket.Conn, identity *Identity) *Client {
	var likesFlag uint64 = 0
	for _, item := range identity.Likes {
		switch item {
		case "Yuri":
			likesFlag |= Yuri
		case "Cosplay":
			likesFlag |= Cosplay
		case "Crossdressing":
			likesFlag |= Crossdressing
		case "Cuddling":
			likesFlag |= Cuddling
		case "Eyebrows":
			likesFlag |= Eyebrows
		case "Fangs":
			likesFlag |= Fangs
		case "Fantasy":
			likesFlag |= Fantasy
		case "Futanari":
			likesFlag |= Futanari
		case "Genderbend":
			likesFlag |= Genderbend
		case "Glasses":
			likesFlag |= Glasses
		case "Hentai":
			likesFlag |= Hentai
		case "Holding Hands":
			likesFlag |= HoldingHands
		case "Horror":
			likesFlag |= Horror
		case "Housewife":
			likesFlag |= Housewife
		case "Humiliation":
			likesFlag |= Humiliation
		case "Idol":
			likesFlag |= Idol
		case "Incest":
			likesFlag |= Incest
		case "Loli":
			likesFlag |= Loli
		case "Maid":
			likesFlag |= Maid
		case "Miko":
			likesFlag |= Miko
		case "Monster Girl":
			likesFlag |= MonsterGirl
		case "Muscles":
			likesFlag |= Muscles
		case "Netorare":
			likesFlag |= Netorare
		case "Nurse":
			likesFlag |= Nurse
		case "Office Lady":
			likesFlag |= OfficeLady
		case "Oppai":
			likesFlag |= Oppai
		case "School girl":
			likesFlag |= Schoolgirl
		case "Sci-Fi":
			likesFlag |= SciFi
		case "Shota":
			likesFlag |= Shota
		case "Slice-of-Life":
			likesFlag |= SliceOfLife
		case "Socks":
			likesFlag |= Socks
		case "Spread":
			likesFlag |= Spread
		case "Stockings":
			likesFlag |= Stockings
		case "Swimsuit":
			likesFlag |= Swimsuit
		case "Teacher":
			likesFlag |= Teacher
		case "Tentacles":
			likesFlag |= Tentacles
		case "Tomboy":
			likesFlag |= Tomboy
		case "Tsundere":
			likesFlag |= Tsundere
		case "Vanilla":
			likesFlag |= Vanilla
		case "Warm Smiles":
			likesFlag |= WarmSmiles
		case "Western":
			likesFlag |= Western
		case "Yandere":
			likesFlag |= Yandere
		case "Yaoi":
			likesFlag |= Yaoi
		case "Yukata":
			likesFlag |= Yukata
		}
	}
	return &Client{
		conn,
		identity.Username,
		identity.Gender,
		likesFlag,
		identity.Timezone,
	}
}

func (c *Client) parseFlag(f uint64) uint8 {
	var ret uint8
	for ret = 0; f > 0; ret++ {
		f &= (f - 1)
	}

	return ret
}

func (c *Client) LikesCount() uint8 {
	return c.parseFlag(c.Likes)
}

func (c *Client) SimilarityWith(p *Client) uint8 {
	return c.parseFlag(c.Likes & p.Likes)
}
