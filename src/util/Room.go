package util

import (
	"time"

	"github.com/googollee/go-socket.io"

	"tile"
	"manager"
	"action"
)

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room {Name: name, Waiting: false, State: BeforeStart}
}

// Room represents a round of mahjong
type Room struct {
	Players      []*Player
	ChangedTiles [4][]tile.Tile
	ChoosedLack  [4]int
	Deck         tile.Set
	HuTiles      tile.Set
	Waiting      bool
	IO           *socketio.Server
	Name         string
	State        int
}

// NumPlayer returns the number of player in the room
func (room Room) NumPlayer() int {
	list := manager.FindPlayerListInRoom(room.Name)
	num := 0
	for _, player := range list {
		if (player.State & (manager.READY | manager.PLAYING)) != 0 {
			num++
		}
	}
	return num
}

// AddPlayer adds 4 player into this room
func (room *Room) AddPlayer(playerList []string) {
	for _, uuid := range playerList {
		index := manager.FindPlayerByUUID(uuid)
		manager.PlayerList[index].Room = room.Name
	}
	playerLsit := manager.FindPlayerListInRoom(room.Name)
	nameList   := manager.GetNameList(playerLsit)
	for _, player := range playerLsit {
		(*player.Socket).Emit("readyToStart", room.Name, nameList)
	}
}

// RemovePlayer reomves id th player from this room
func (room *Room) RemovePlayer(id int) {
	room.Players = append(room.Players[: id], room.Players[id + 1: ]...)
}

// WaitToStart checks if all player in this room are ready
// and run the mahjong logic
func (room *Room) WaitToStart() {
	room.Waiting = true
	for room.NumPlayer() < 4 && room.Waiting {
		time.Sleep(0)
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
	index := manager.FindPlayerByUUID(uuid)
	if index == -1 {
		callback(-1)
		return
	}
	player := manager.PlayerList[index]
	idx    := room.NumPlayer()
	room.BroadcastReady(player.Name)
	callback(idx)
	player.Index = idx
	room.Players = append(room.Players, NewPlayer(room, idx, player.UUID))
	manager.PlayerList[index].State = manager.READY
}

// Run runs mahjong logic
func (room *Room) Run() {
	room.preproc()
	currentIdx := 0
	onlyThrow := false
	gameOver  := false
	for !gameOver {
		room.BroadcastRemainCard(room.Deck.Count());
		curPlayer := room.Players[currentIdx]
		throwCard := tile.NewTile(-1, 0) 
		act       := action.NewAction(action.NONE, throwCard, 0)
		room.State = IdxTurn + currentIdx

		if onlyThrow {
			throwCard = curPlayer.Throw()
			curPlayer.Hand.Sub(throwCard)
			onlyThrow = false
		} else {
			drawCard := room.Deck.Draw()
			room.BroadcastDraw(currentIdx)
			act       = curPlayer.Draw(drawCard)
			throwCard = act.Tile
		}

		fail, huIdx, gonIdx, ponIdx := room.checkAction(currentIdx, act, throwCard)
		if fail {
			curPlayer.Fail(act.Command)
		} else if act.Command != action.NONE {
			curPlayer.Success(currentIdx, act.Command, act.Tile, act.Score)
		}
		curPlayer.JustGon = false

		currentIdx, onlyThrow = room.doAction(currentIdx, throwCard, huIdx, gonIdx, ponIdx)
		if currentIdx == curPlayer.ID {
			if fail || (act.Command & action.ONGON) == 0 && (act.Command & action.PONGON) == 0 {
				if throwCard.Color > 0 {
					curPlayer.DiscardTiles.Add(throwCard)
				}
				currentIdx = (currentIdx + 1) % 4
			}
		}
		if room.Deck.IsEmpty() {
			gameOver = true
		}
	}
	room.end()
}

// Stop stops this round
func (room *Room) Stop() {
	// TODO
}