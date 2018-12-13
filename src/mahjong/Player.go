package mahjong

import (
	"math"

	"github.com/googollee/go-socket.io"
)

// NewPlayer creates a new player
func NewPlayer(room *Room, id int, uuid string) *Player {
	return &Player {room: room, ID: id, UUID: uuid}
}

// Player represents a player in mahjong
type Player struct {
	Hand         SuitSet
	Door         SuitSet
	VisiableDoor SuitSet
	DiscardTiles SuitSet
	HuTiles      SuitSet
	GonRecord    [4]int
	Lack         int
	Credit       int
	MaxTai       int
	IsHu         bool
	IsTing       bool
	JustGon      bool
	ID           int
	UUID         string
	room         *Room
}

// Name returns the player's name
func (player Player) Name() string {
	index := FindPlayerByUUID(player.UUID)
	return PlayerList[index].Name
}

// Room returns the player's room
func (player Player) Room() string {
	index := FindPlayerByUUID(player.UUID)
	return PlayerList[index].Room
}

// Socket returns the player's socket
func (player Player) Socket() socketio.Socket {
	index := FindPlayerByUUID(player.UUID)
	return *PlayerList[index].Socket
}

// Init inits the player's state
func (player *Player) Init() {
	index := FindPlayerByUUID(player.UUID)
	PlayerList[index].State = PLAYING
	for i := 0; i < 3; i++ {
		player.Door[i]         = 0
		player.VisiableDoor[i] = 0
		player.Hand[i]         = 0
		player.HuTiles[i]      = 0
		player.GonRecord[i]    = 0
		player.DiscardTiles[i] = 0
	}

	player.Credit  = 0
	player.MaxTai  = 0
	player.IsHu    = false
	player.IsTing  = false
	player.JustGon = false
	player.Lack    = -1
}

// CheckGon checks if the player can gon
func (player *Player) CheckGon(tile Tile) bool {
	if tile.Suit == player.Lack {
		return false
	}
	if !player.IsHu {
		return true
	}

	handCount := int(player.Hand[tile.Suit].GetIndex(tile.Value))
	oldTai    := CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))

	for i := 0; i < handCount; i++ {
		player.Hand.Sub(tile)
		player.Door.Add(tile)
	}
	newTai := CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	if newTai > 0 {
		newTai--
	}
	for i := 0; i < handCount; i++ {
		player.Hand.Add(tile)
		player.Door.Sub(tile)
	}
	return oldTai == newTai
}

// CheckPon checks if the player can pon
func (player *Player) CheckPon(tile Tile) bool {
	if tile.Suit == player.Lack || player.IsHu {
		return false
	}
	return player.Hand[tile.Suit].GetIndex(tile.Value) >= 2
}

// CheckHu checks if the player can hu
func (player *Player) CheckHu(tile Tile, tai *int) bool {
	*tai = 0
	if player.Hand[player.Lack].Count() > 0 {
		return false
	}
	if tile.Suit == -1 {
		*tai = CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	} else {
		if tile.Suit == player.Lack {
			return false
		}
		player.Hand.Add(tile)
		*tai = CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
		player.Hand.Sub(tile)
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
			tai     := CalTai(newHand, tDoor)
			if tai > *max {
				*max = tai
			}
		}
	}
	return *max > 0
}

// Hu hus tile tile
func (player *Player) Hu(tile Tile, tai int, Type int, addOneTai, addToRoom bool, fromID int) int {
	player.IsHu = true
	player.HuTiles.Add(tile)
	if Type == COMMAND["ZIMO"] {
		player.Hand.Sub(tile)
	}
	if addToRoom {
		player.room.HuTiles.Add(tile)
	}
	Tai   := IF(addOneTai,      tai + 1, tai).(int)
	Tai    = IF(player.JustGon, Tai + 1, Tai).(int)
	score := int(math.Pow(2, float64(Tai - 1)))
	for i := 0; i < 4; i++ {
		if Type == COMMAND["ZIMO"] && i != player.ID || Type == COMMAND["HU"] && i == fromID {
			player.Credit += score
			player.room.Players[i].Credit -= score
		}
	}
	player.MaxTai = IF(player.MaxTai < tai, tai, player.MaxTai).(int)
	return score
}

// Gon gons the tile
func (player *Player) Gon(tile Tile, Type int, fromID int) int {
	player.JustGon = true
	for i := 0; i < IF(Type == COMMAND["PONGON"], 1, 4).(int); i++ {
		player.Door.Add(tile)
		if Type != COMMAND["ONGON"] {
			player.VisiableDoor.Add(tile)
		}
		player.Hand.Sub(tile)
	}

	score := IF(Type == COMMAND["PONGON"], 1, 2).(int)
	for i := 0; i < 4; i++ {
		if Type != COMMAND["GON"] && i != player.ID || Type == COMMAND["GON"] && i == fromID {
			player.Credit                 += score
			player.GonRecord[i]           += score
			player.room.Players[i].Credit -= score
		}
	}
	return score
}

// Pon pons the tile
func (player *Player) Pon(tile Tile) {
	for i := 0; i < 3; i++ {
		player.Door.Add(tile)
		player.VisiableDoor.Add(tile)
	}
	player.Hand.Sub(tile)
	player.Hand.Sub(tile)
}

// Tai cals the tai
func (player *Player) Tai(tile Tile) int {
	player.Hand.Add(tile)
	result := CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	player.Hand.Sub(tile)
	return result
}
