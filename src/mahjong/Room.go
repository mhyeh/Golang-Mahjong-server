package mahjong

import (
	"time"

	socketio "github.com/googollee/go-socket.io"
)

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room {Name: name, Waiting: false, State: BeforeStart}
}

// Room represents a round of mahjong
type Room struct {
	Players      []*Player
	ChangedTiles [4][]Tile
	ChoosedLack  [4]int
	Deck         SuitSet
	HuTiles      SuitSet
	Waiting      bool
	IO           *socketio.Server
	Name         string
	State        int
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

// Run runs mahjong logic
func (room *Room) Run() {
	room.preproc()
	currentIdx := 0
	onlyThrow  := false
	gameOver   := false
	for !gameOver {
		curPlayer := room.Players[currentIdx]
		throwTile := NewTile(-1, 0)
		act       := NewAction(COMMAND["NONE"], throwTile, 0)
		room.State = IdxTurn + currentIdx

		if onlyThrow {
			throwTile = curPlayer.Throw(throwTile)
			onlyThrow = false
		} else {
			drawTile := room.Deck.Draw()
			room.BroadcastDraw(currentIdx)
			act       = curPlayer.Draw(drawTile)
			throwTile = act.Tile
		}
		room.BroadcastRemainTile(room.Deck.Count())

		robGon, huIdx, gonIdx, ponIdx := room.checkAction(currentIdx, act, throwTile)
		if robGon {
			curPlayer.Fail(act.Command)
			room.BroadcastRobGon(curPlayer.ID, act.Tile)
		} else if act.Command != COMMAND["NONE"] {
			curPlayer.Success(currentIdx, act.Command, act.Tile, act.Score)
		}
		curPlayer.JustGon = false

		currentIdx, onlyThrow = room.doAction(currentIdx, throwTile, huIdx, gonIdx, ponIdx)
		if currentIdx == curPlayer.ID && huIdx == -1 && (act.Command & COMMAND["ONGON"]) == 0 && (act.Command & COMMAND["PONGON"]) == 0 {
			curPlayer.DiscardTiles.Add(throwTile)
			currentIdx = (currentIdx + 1) % 4
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
