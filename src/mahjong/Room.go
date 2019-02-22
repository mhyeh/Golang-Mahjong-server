package mahjong

import (
	"time"

	socketio "github.com/googollee/go-socket.io"
)

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room { Name: name, Waiting: false, State: BeforeStart }
}

// Room represents a round of mahjong
type Room struct {
	Players      []*Player
	Deck         SuitSet
	Waiting      bool
	IO           *socketio.Server
	Name         string
	State        int
	Info         HuInfo
	KeepWin      bool
	NumKeepWin   int
	Banker       int
	SevenFlower  bool
	SevenID      int
	Wind         int
	Round        int
	OpenIdx      int
}

// NumPlayer returns the number of player in the room
func (room Room) NumPlayer() int {
	list := FindPlayerListInRoom(room.Name)
	num  := 0
	for _, player := range list {
		if (player.State & (READY | PLAYING)) != 0 {
			num++
		}
	}
	return num
}

// AddPlayer adds 4 player into this room
func (room *Room) AddPlayer(playerList []string) {
	for _, uuid := range playerList {
		index := FindPlayerByUUID(uuid)
		PlayerList[index].Room = room.Name
	}
	playerLsit := FindPlayerListInRoom(room.Name)
	nameList   := GetNameList(playerLsit)
	for _, player := range playerLsit {
		(*player.Socket).Emit("readyToStart", room.Name, nameList)
	}
}

// RemovePlayer reomves id th player from this room
func (room *Room) RemovePlayer(id int) {
	room.Players = append(room.Players[:id], room.Players[id+1:]...)
}

// WaitToStart checks if all player in this room are ready
// and run the mahjong logic
func (room *Room) WaitToStart() {
	room.Waiting = true
	for room.NumPlayer() < 4 && room.Waiting {
		time.Sleep(2 * time.Second)
	}

	if !room.Waiting {
		return
	}
	room.Waiting = false
	room.BroadcastGameStart()
	room.Run()
}

// StopWaiting stops waiting
func (room *Room) StopWaiting() {
	room.BroadcastStopWaiting()
	room.Waiting = false
	for i := 0; i < room.NumPlayer(); i++ {
		room.Players[i] = nil
	}
}

// Accept checks player's info and constructs the player
func (room *Room) Accept(uuid string, callback func(int)) {
	if !room.Waiting {
		callback(-1)
		return
	}
	index := FindPlayerByUUID(uuid)
	if index == -1 {
		callback(-1)
		return
	}
	player := PlayerList[index]
	idx    := room.NumPlayer()
	room.BroadcastReady(player.Name)
	callback(idx)
	player.Index = idx
	room.Players = append(room.Players, NewPlayer(room, idx, player.UUID))
	PlayerList[index].State = READY
}
