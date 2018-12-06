package util

import (
	"math"
	"time"
	"sync"

	"github.com/googollee/go-socket.io"

	"MJCard"
)

// NewRoom creates a new room
func NewRoom(game *GameManager, name string) *Room {
	return &Room {game: game, Name: name, Waiting: false}
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
	game         *GameManager
	Name         string
}

// NumPlayer returns the number of player in the room
func (room Room) NumPlayer() int {
	list := room.game.PlayerManager.FindPlayersInRoom(room.Name)
	num := 0
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
		index := room.game.PlayerManager.FindPlayerByUUID(uuid)
		room.game.PlayerManager[index].Room = room.Name
	}
	list := room.game.PlayerManager.FindPlayersInRoom(room.Name)
	nameList := GetNameList(list)
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
	for ; room.NumPlayer() < 4 && room.Waiting; {
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
	index := room.game.PlayerManager.FindPlayerByUUID(uuid)
	if index == -1 {
		callback(-1)
		return
	}
	player := room.game.PlayerManager[index]
	room.BroadcastReady(player.Name)
	id := room.NumPlayer()
	callback(id)
	player.Index = id
	room.Players = append(room.Players, NewPlayer(room.game, id, player.UUID))
	room.game.PlayerManager[index].State = READY
}

// GetPlayerList returns the list of player's name
func (room Room) GetPlayerList() []string {
	var nameList []string
	for _, player := range room.Players {
		nameList = append(nameList, player.Name())
	}
	return nameList
}

// Run runs mahjong logic
func (room *Room) Run() {
	time.Sleep(2 * time.Second)
	room.init()
	time.Sleep(3 * time.Second)
	room.changeCard()
	time.Sleep(5 * time.Second)
	room.chooseLack()

	currentID := 0
	onlyThrow := false
	gameOver  := false
	for ; !gameOver; {
		room.BroadcastRemainCard(room.Deck.Count());
		curPlayer := room.Players[currentID]
		throwCard := MJCard.Card {Color: -1, Value: 0}
		action := Action {NONE, throwCard, 0}

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


		ponIdx, gonIdx, huIdx := -1, -1, -1
		fail := false
		if (action.Command & PONGON) != 0 {
			var waitGroup sync.WaitGroup
			waitGroup.Add(3)
			act := Action {NONE, MJCard.Card {Color: -1, Value: 0}, 0 }
			var actionSet [4]Action
			actionSet[0] = act

			for i := 1; i < 4; i++ {
				id := (i + currentID) % 4;
				tai := 0
				if room.Players[id].CheckHu(action.Card, &tai) {
					cards := make(map[int][]MJCard.Card)
					cards[HU] = append(cards[HU], action.Card)
					go func (i int) {
						actionSet[i] = room.Players[i].Command(cards, HU, (4 + currentID - id) % 4)
						waitGroup.Done()
					}(i)
				} else {
					waitGroup.Done()
				}
			}
			waitGroup.Wait()
			for i := 1; i < 4; i++ {
				id := (i + currentID) % 4
				act = actionSet[i]
				if (act.Command & HU) != 0 {
					tai := 0
					room.Players[id].CheckHu(action.Card, &tai)

					score := int(math.Pow(2, float64(tai)))
					curPlayer.Credit        -= score
					room.Players[id].Credit += score
					room.Players[id].HuCards.Add(action.Card)
					room.Players[id].OnSuccess(currentID, HU, action.Card, score)

					if !fail {
						curPlayer.Door.Sub(action.Card)
						curPlayer.VisiableDoor.Sub(action.Card)
						room.HuTiles.Add(action.Card)
					}

					huIdx = id
					fail = true
				}
			}
		} else if (action.Command & ZIMO) == 0 && (action.Command & ONGON) == 0 {
			room.checkOthers(currentID, throwCard, &huIdx, &gonIdx, &ponIdx)
		}

		if (action.Command != NONE) {
			if fail {
				curPlayer.OnFail(action.Command)
			} else {
				curPlayer.OnSuccess(currentID, action.Command, action.Card, action.Score)
			}
		}

		curPlayer.JustGon = false

		if huIdx != -1 {
			currentID = (huIdx + 1) % 4
			if gonIdx != -1 {
				room.Players[gonIdx].OnFail(GON)
			}
			if ponIdx != -1 {
				room.Players[ponIdx].OnFail(PON)
			}
		} else if gonIdx != -1 {
			room.Players[gonIdx].OnSuccess(currentID, GON, throwCard, 2)
			room.Players[gonIdx].Gon(throwCard, true)
			curPlayer.Credit -= 2
			room.Players[gonIdx].Credit += 2
			room.Players[gonIdx].GonRecord[currentID] += 2
			currentID = gonIdx
			if ponIdx != -1 {
				room.Players[ponIdx].OnFail(PON)
			}
		} else if ponIdx != -1 {
			room.Players[ponIdx].OnSuccess(currentID, PON, throwCard, 0)
			room.Players[ponIdx].Pon(throwCard)
			currentID = ponIdx
			onlyThrow = true
		} else if !fail && (action.Command & ONGON) != 0 || (action.Command & ONGON) != 0 {
		} else {
			if throwCard.Color > 0 {
				room.DiscardTiles.Add(throwCard)
			}
			currentID = (currentID + 1) % 4
		}
		if room.Deck.IsEmpty() {
			gameOver = true
		}
	}
	if room.huUnder2() {
		room.lackPenalty()
		room.noTingPenalty()
		room.returnMoney()
	}
	room.end()
}

// Stop stops this round
func (room *Room) Stop() {
	// TODO
}