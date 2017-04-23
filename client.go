package main

import (
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
	// Token string
}

type Client struct {
	*Identity
	Conn      *websocket.Conn
	likesMask uint64
}

func NewClient(conn *websocket.Conn, identity *Identity) *Client {
	var likesMask uint64 = 0
	sanitizedLikes := []string{}

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
	return &Client{identity, conn, likesMask}
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
