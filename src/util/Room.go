package util

import (
	"time"

	"github.com/googollee/go-socket.io"

	"MJCard"
	"PManager"
)

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room {Name: name, Waiting: false, State: BeforeStart}
}

// Room represents a round of mahjong
type Room struct {
	Players      []*Player
	ChangedTiles [4][]MJCard.Card
	ChoosedLack  [4]int
	Deck         MJCard.Cards
	DiscardTiles MJCard.Cards
	HuTiles      MJCard.Cards
	Waiting      bool
	IO           *socketio.Server
	Name         string
	State        int
}

// NumPlayer returns the number of player in the room
func (room Room) NumPlayer() int {
	list := PManager.FindPlayersInRoom(room.Name)
	num := 0
	for _, player := range list {
		if (player.State & (PManager.READY | PManager.PLAYING)) != 0 {
			num++
		}
	}
	return num
}

// AddPlayer adds 4 player into this room
func (room *Room) AddPlayer(playerList []string) {
	for _, uuid := range playerList {
		index := PManager.FindPlayerByUUID(uuid)
		PManager.Players[index].Room = room.Name
	}
	list := PManager.FindPlayersInRoom(room.Name)
	nameList := PManager.GetNameList(list)
	for _, player := range list {
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
	index := PManager.FindPlayerByUUID(uuid)
	if index == -1 {
		callback(-1)
		return
	}
	player := PManager.Players[index]
	id     := room.NumPlayer()
	room.BroadcastReady(player.Name)
	callback(id)
	player.Index = id
	room.Players = append(room.Players, NewPlayer(room, id, player.UUID))
	PManager.Players[index].State = PManager.READY
}

// Run runs mahjong logic
func (room *Room) Run() {
	room.preproc()
	currentID := 0
	onlyThrow := false
	gameOver  := false
	for !gameOver {
		room.BroadcastRemainCard(room.Deck.Count());
		curPlayer := room.Players[currentID]
		throwCard := MJCard.Card {Color: -1, Value: 0}
		action    := Action {NONE, throwCard, 0}
		room.State = IDTurn + currentID

		if onlyThrow {
			throwCard = curPlayer.ThrowCard()
			curPlayer.Hand.Sub(throwCard)
			onlyThrow = false
		} else {
			drawCard := room.Deck.Draw()
			room.BroadcastDraw(currentID)
			action = curPlayer.Draw(drawCard)
			throwCard = action.Card
		}

		fail, huIdx, gonIdx, ponIdx := room.checkAction(currentID, action, throwCard)
		if fail {
			curPlayer.OnFail(action.Command)
		} else if action.Command != NONE {
			curPlayer.OnSuccess(currentID, action.Command, action.Card, action.Score)
		}
		curPlayer.JustGon = false

		currentID, onlyThrow = room.doAction(currentID, throwCard, huIdx, gonIdx, ponIdx)
		if currentID == curPlayer.ID {
			if fail || (action.Command & ONGON) == 0 && (action.Command & PONGON) == 0 {
				if throwCard.Color > 0 {
					room.DiscardTiles.Add(throwCard)
					curPlayer.DiscardTiles.Add(throwCard)
				}
				currentID = (currentID + 1) % 4
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