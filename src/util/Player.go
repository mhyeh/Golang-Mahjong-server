package util

import (
	"github.com/googollee/go-socket.io"

	"MJCard"
)

// NewPlayer creates a new player
func NewPlayer(game *GameManager, id int, uuid string) *Player {
	return &Player{game: game, ID: id, UUID: uuid}
}

// Player represents a player in mahjong
type Player struct {
	Lack         int
	Credit       int
	MaxTai       int
	GonRecord    [4]int
	Hand         MJCard.Cards
	Door         MJCard.Cards
	VisiableDoor MJCard.Cards
	HuCards      MJCard.Cards
	IsHu         bool
	IsTing       bool
	JustGon      bool
	ID           int
	UUID         string
	game         *GameManager
}

// Name returns the player's name
func (player Player) Name() string {
	index := player.game.PlayerManager.FindPlayerByUUID(player.UUID)
	return player.game.PlayerManager[index].Name
}

// Room returns the player's room
func (player Player) Room() string {
	index := player.game.PlayerManager.FindPlayerByUUID(player.UUID)
	return player.game.PlayerManager[index].Room
}

// Socket returns the player's socket
func (player Player) Socket() socketio.Socket {
	index := player.game.PlayerManager.FindPlayerByUUID(player.UUID)
	return *player.game.PlayerManager[index].Socket
}

// Init inits the player's state
func (player *Player) Init() {
	index := player.game.PlayerManager.FindPlayerByUUID(player.UUID)
	player.game.PlayerManager[index].State = PLAYING
	for i := 0; i < 3; i++ {
		player.Door[i]         = 0
		player.VisiableDoor[i] = 0
		player.Hand[i]         = 0
		player.HuCards[i]      = 0
		player.GonRecord[i]    = 0
	}

	player.Credit  = 0
	player.MaxTai  = 0
	player.IsHu    = false
	player.IsTing  = false
	player.JustGon = false
	player.Lack    = -1
}

// CheckGon checks if the player can gon
func (player *Player) CheckGon(card MJCard.Card) bool {
	if card.Color == player.Lack {
		return false
	}
	if !player.IsHu {
		return true
	}

	handCount := int(player.Hand[card.Color].GetIndex(card.Value))
	oldTai    := SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))

	for i := 0; i < handCount; i++ {
		player.Hand.Sub(card)
		player.Door.Add(card)
	}
	newTai := SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	if newTai > 0 {
		newTai--
	}
	for i := 0; i < handCount; i++ {
		player.Hand.Add(card)
		player.Door.Sub(card)
	}
	return (oldTai == newTai)
}

// CheckPon checks if the player can pon
func (player *Player) CheckPon(card MJCard.Card) bool {
	if card.Color == player.Lack || player.IsHu {
		return false
	}
	return player.Hand[card.Color].GetIndex(card.Value) >= 2
}

// CheckHu checks if the player can hu
func (player *Player) CheckHu(card MJCard.Card, tai *int) bool {
	*tai = 0
	if player.Hand[player.Lack].Count() > 0 {
		return false
	}
	if card.Color == -1 {
		*tai = SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	} else {
		if card.Color == player.Lack {
			return false
		}
		player.Hand.Add(card)
		*tai = SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
		player.Hand.Sub(card)
	}
	return *tai > 0
}

// CheckTing checks if the player is ting
func (player *Player) CheckTing(max *int) bool {
	*max = 0
	tHand := player.Hand.Translate(player.Lack)
	tDoor := player.Door.Translate(player.Lack)
	total := tHand + tDoor
	for i := uint(0); i < 18; i++ {
		if ((total >> (i * 3)) & 7) < 4 {
			newHand := tHand + (1 << (i * 3))
			tai := SSJ(newHand, tDoor)
			if tai > *max {
				*max = tai
			}
		}
	}
	return *max > 0
}

// Gon gons the card
func (player *Player) Gon(card MJCard.Card, visible bool) {
	player.JustGon = true
	for i := 0; i < 4; i++ {
		player.Door.Add(card)
		if visible {
			player.VisiableDoor.Add(card)
		}
		player.Hand.Sub(card)
	}
}

// Pon pons the card
func (player *Player) Pon(card MJCard.Card) {
	for i := 0; i < 3; i++ {
		player.Door.Add(card)
		player.VisiableDoor.Add(card)
	}
	player.Hand.Sub(card)
	player.Hand.Sub(card)
}

// Tai cals the tai
func (player *Player) Tai(card MJCard.Card) int {
	player.Hand.Add(card)
	result := SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	player.Hand.Sub(card)
	return result
}
