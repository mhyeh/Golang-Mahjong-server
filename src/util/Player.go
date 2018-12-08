package util

import (
	"github.com/googollee/go-socket.io"

	"tile"
	"manager"
	"ssj"
)

// NewPlayer creates a new player
func NewPlayer(room *Room, id int, uuid string) *Player {
	return &Player{room: room, ID: id, UUID: uuid}
}

// Player represents a player in mahjong
type Player struct {
	Lack         int
	Credit       int
	MaxTai       int
	GonRecord    [4]int
	Hand         tile.Set
	Door         tile.Set
	VisiableDoor tile.Set
	DiscardTiles tile.Set
	HuTiles      tile.Set
	IsHu         bool
	IsTing       bool
	JustGon      bool
	ID           int
	UUID         string
	room         *Room
}

// Name returns the player's name
func (player Player) Name() string {
	index := manager.FindPlayerByUUID(player.UUID)
	return manager.PlayerList[index].Name
}

// Room returns the player's room
func (player Player) Room() string {
	index := manager.FindPlayerByUUID(player.UUID)
	return manager.PlayerList[index].Room
}

// Socket returns the player's socket
func (player Player) Socket() socketio.Socket {
	index := manager.FindPlayerByUUID(player.UUID)
	return *manager.PlayerList[index].Socket
}

// Init inits the player's state
func (player *Player) Init() {
	index := manager.FindPlayerByUUID(player.UUID)
	manager.PlayerList[index].State = manager.PLAYING
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
func (player *Player) CheckGon(tile tile.Tile) bool {
	if tile.Color == player.Lack {
		return false
	}
	if !player.IsHu {
		return true
	}

	handCount := int(player.Hand[tile.Color].GetIndex(tile.Value))
	oldTai    := ssj.CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))

	for i := 0; i < handCount; i++ {
		player.Hand.Sub(tile)
		player.Door.Add(tile)
	}
	newTai := ssj.CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	if newTai > 0 {
		newTai--
	}
	for i := 0; i < handCount; i++ {
		player.Hand.Add(tile)
		player.Door.Sub(tile)
	}
	return (oldTai == newTai)
}

// CheckPon checks if the player can pon
func (player *Player) CheckPon(tile tile.Tile) bool {
	if tile.Color == player.Lack || player.IsHu {
		return false
	}
	return player.Hand[tile.Color].GetIndex(tile.Value) >= 2
}

// CheckHu checks if the player can hu
func (player *Player) CheckHu(tile tile.Tile, tai *int) bool {
	*tai = 0
	if player.Hand[player.Lack].Count() > 0 {
		return false
	}
	if tile.Color == -1 {
		*tai = ssj.CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	} else {
		if tile.Color == player.Lack {
			return false
		}
		player.Hand.Add(tile)
		*tai = ssj.CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
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
			tai := ssj.CalTai(newHand, tDoor)
			if tai > *max {
				*max = tai
			}
		}
	}
	return *max > 0
}

// Gon gons the tile
func (player *Player) Gon(tile tile.Tile, visible bool) {
	player.JustGon = true
	for i := 0; i < 4; i++ {
		player.Door.Add(tile)
		if visible {
			player.VisiableDoor.Add(tile)
		}
		player.Hand.Sub(tile)
	}
}

// Pon pons the tile
func (player *Player) Pon(tile tile.Tile) {
	for i := 0; i < 3; i++ {
		player.Door.Add(tile)
		player.VisiableDoor.Add(tile)
	}
	player.Hand.Sub(tile)
	player.Hand.Sub(tile)
}

// Tai cals the tai
func (player *Player) Tai(tile tile.Tile) int {
	player.Hand.Add(tile)
	result := ssj.CalTai(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	player.Hand.Sub(tile)
	return result
}
