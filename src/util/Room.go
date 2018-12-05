package util

import (
	"math"
	"time"
	"math/rand"
	"sync"

	"github.com/googollee/go-socket.io"
	// "github.com/fanliao/go-promise"

	"MJCard"
)

// NewRoom creates a new room
func NewRoom(game *GameManager, name string) *Room {
	return &Room {game: game, name: name, Waiting: false}
	
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

	name         string
}

// GameResult represents the result of mahjong
type GameResult struct {
	hand  []string
	score int
}

// Name returns the room name
func (room Room) Name() string {
	return room.name
}

// NumPlayer returns the number of player in the room
func (room Room) NumPlayer() int {
	list := room.game.PlayerManager.FindPlayersInRoom(room.name)
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
		room.game.PlayerManager[index].Room = room.name
	}
	list := room.game.PlayerManager.FindPlayersInRoom(room.name)
	nameList := GetNameList(list)
	for _, player := range list {
		(*player.Socket).Emit("readyToStart", room.name, nameList)
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
	go room.Run()
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
	callback(room.NumPlayer())
	room.Players = append(room.Players, NewPlayer(room.game, room.NumPlayer(), player.UUID))
	room.game.PlayerManager[index].State = READY
	return
}

// GetPlayerList returns the list of player's name
func (room Room) GetPlayerList() []string {
	var nameList []string
	for _, player := range room.Players {
		nameList = append(nameList, player.Name())
	}
	return nameList
}

// BroadcastStopWaiting broadcasts stop waiting signal
func (room Room) BroadcastStopWaiting() {
	room.IO.BroadcastTo(room.name, "stopWaiting")
}

// BroadcastReady broadcasts the player's name who is ready
func (room Room) BroadcastReady(name string) {
	room.IO.BroadcastTo(room.name, "broadcastReady", name)
}

// BroadcastGameStart broadcasts player list
func (room Room) BroadcastGameStart() {
	room.IO.BroadcastTo(room.name, "broadcastGameStart", room.GetPlayerList())
}

// BroadcastChange broadcasts the player's id who already change cards
func (room Room) BroadcastChange(id int) {
	room.IO.BroadcastTo(room.name, "broadcastChange", id)
}

// BroadcastLack broadcasts the player's id who already choose lack
func (room Room) BroadcastLack() {
	room.IO.BroadcastTo(room.name, "afterLack", room.ChoosedLack)
}

// BroadcastDraw broadcasts the player's id who draw a card
func (room Room) BroadcastDraw(id int) {
	room.IO.BroadcastTo(room.name, "broadcastDraw", id)
}

// BroadcastThrow broadcasts the player's id and the card he threw
func (room Room) BroadcastThrow(id int, card MJCard.Card) {
	room.IO.BroadcastTo(room.name, "broadcastThrow", id, card.ToString())
}

// BroadcastCommand broadcasts the player's id and the command he made
func (room Room) BroadcastCommand(from int, to int, command int, card MJCard.Card, score int) {
	if command == ONGON {
		room.IO.BroadcastTo(room.name, "broadcastCommand", from, to, command, "", score)
	} else {
		room.IO.BroadcastTo(room.name, "broadcastCommand", from, to, command, card.ToString(), score)
	}
}

// BroadcastEnd broadcasts the game result
func (room Room) BroadcastEnd(data []GameResult) {
	room.IO.BroadcastTo(room.name, "end", data)
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
			action := curPlayer.Draw(drawCard)
			throwCard = action.Card
		}

		ponIdx, gonIdx, huIdx := -1, -1, -1
		fail := false
		if (action.Command & PONGON) != 0 {
			var waitGroup sync.WaitGroup
			waitGroup.Add(4)
			act := Action {NONE, MJCard.Card {Color: -1, Value: 0}, 0 }
			var actionSet [4]Action
			actionSet[0] = act

			for i := 1; i < 4; i++ {
				id := (i + currentID) % 4;
				tai := 0
				if room.Players[id].CheckHu(action.Card, &tai) {
					cards := make(map[int][]MJCard.Card)
					cards[HU] = append(cards[HU], action.Card)
					go func () {
						actionSet[i] = room.Players[i].OnCommand(cards, HU, (4 + currentID - id) % 4)
						waitGroup.Done()
					}()
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

					if !fail {
						curPlayer.Door.Sub(action.Card)
						curPlayer.VisiableDoor.Sub(action.Card)
						room.HuTiles.Add(action.Card)
					}

					huIdx = id
					fail = true
				}
			}
		} else if (action.Command & ZIMO) != 0 && (action.Command & ONGON) != 0 {
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
			room.Players[gonIdx].Gon(throwCard, true)
			curPlayer.Credit -= 2
			room.Players[gonIdx].Credit += 2
			room.Players[gonIdx].GonRecord[currentID] += 2
			currentID = gonIdx
			room.Players[gonIdx].OnSuccess(currentID, GON, throwCard, 2)
			if ponIdx != -1 {
				room.Players[ponIdx].OnFail(PON)
			}
		} else if ponIdx != -1 {
			room.Players[ponIdx].Pon(throwCard)
			currentID = ponIdx
			onlyThrow = true
			room.Players[ponIdx].OnSuccess(currentID, PON, throwCard, 0)
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

func (room *Room) init() {
	room.Deck         = MJCard.NewCards(true)
	room.DiscardTiles = MJCard.NewCards(false)
	room.HuTiles      = MJCard.NewCards(false)

	len := room.Deck.Count()
	for _, player := range room.Players {
		player.Init()
		for j := 0; j < 13; j++ {
			idx := rand.Int31n(int32(len))
			result := room.Deck.At(int(idx))
			room.Deck.Sub(result)
			player.Hand.Add(result)
			len--
		}
		player.Socket().Emit("dealCard", player.Hand.ToStringArray())
	}
}

func (room *Room) changeCard() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	for i := 0; i < 4; i++ {
		go func (id int) {
			room.ChangedTiles[id] = room.Players[id].ChangeCard()
			waitGroup.Done()
		}(i)
	}
	waitGroup.Wait()

	rand := rand.Int31n(3)
	var tmp []MJCard.Card
	if rand == 0 {
		tmp = room.ChangedTiles[0]
		for i := 0; i < 3; i++ {
			room.ChangedTiles[i] = room.ChangedTiles[i + 1]
		}
		room.ChangedTiles[3] = tmp
	} else if rand == 1 {
		tmp = room.ChangedTiles[3]
		for i := 3; i >= 3; i-- {
			room.ChangedTiles[i] = room.ChangedTiles[i - 1]
		}
		room.ChangedTiles[0] = tmp
	} else {
		for i := 0; i < 2; i++ {
			tmp = room.ChangedTiles[i];
			room.ChangedTiles[i] = room.ChangedTiles[i + 2];
			room.ChangedTiles[i + 2] = tmp;
		}
	}

	for i := 0; i < 4; i++ {
		room.Players[i].Hand.Add(room.ChangedTiles[i])
		t := MJCard.CardArrayToCards(room.ChangedTiles[i])
		room.Players[i].Socket().Emit("afterChange", t.ToStringArray(), rand)
	}
}

func (room *Room) chooseLack() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	for i := 0; i < 4; i++ {
		go func (id int) {
			room.ChoosedLack[id] = room.Players[id].ChooseLack()
			waitGroup.Done()
		}(i)
	}
	waitGroup.Wait()
	room.BroadcastLack()
}

func (room *Room) checkOthers(currentID int, throwCard MJCard.Card, huIdx *int, gonIdx *int, ponIdx *int) {
	action := Action {NONE, throwCard, 0}
	var playerCommand [4]Action
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)
	playerCommand[0] = action
	for i := 1; i < 4; i++ {
		id := (i + currentID) % 4
		actions := make(map[int][]MJCard.Card)
		otherPlayer := room.Players[id]

		command := 0
		tai     := 0
		if otherPlayer.CheckHu(throwCard, &tai) {
			if !otherPlayer.IsHu {
				command |= HU
				actions[HU] = append(actions[HU], throwCard)
			}
		}

		if otherPlayer.Hand[throwCard.Color].GetIndex(throwCard.Value) == 3 {
			if otherPlayer.CheckGon(throwCard) {
				command |= GON
				actions[GON] = append(actions[GON], throwCard)
			}
		}

		if otherPlayer.CheckPon(throwCard) {
			command |= PON
			actions[PON] = append(actions[PON], throwCard)
		}

		if command == NONE {
			action.Command = NONE
			playerCommand[i] = action
			waitGroup.Done()
		} else if otherPlayer.IsHu {
			if (command & HU) != 0 {
				action.Command = HU
			} else if (command & GON) != 0 {
				action.Command = GON
			}
			action.Card = throwCard
			playerCommand[i] = action
			waitGroup.Done()
		} else {
			go func() {
				playerCommand[i] = otherPlayer.OnCommand(actions, command, ((4 + currentID - otherPlayer.ID()) % 4))
				waitGroup.Done()
			}()
		}
	}
	waitGroup.Wait()
	for i := 1; i < 4; i++ {
		playerID := (i + currentID) % 4
		otherPlayer := room.Players[playerID]
		action = playerCommand[i]
		tai := 0
		otherPlayer.CheckHu(throwCard, &tai)

		if (action.Command & HU) != 0 {
			otherPlayer.IsHu = true
			otherPlayer.HuCards.Add(action.Card)
			if *huIdx == -1 {
				room.HuTiles.Add(action.Card)
			}
			*huIdx = playerID
			Tai := tai
			if room.Players[currentID].JustGon {
				Tai++ 
			}
			score := int(math.Pow(2, float64(Tai - 1)))
			otherPlayer.Credit += score
			if otherPlayer.MaxTai < tai {
				otherPlayer.MaxTai = tai
			}
			room.Players[currentID].Credit -= score
			otherPlayer.OnSuccess(currentID, HU, action.Card, score)
		} else if (action.Command & GON) != 0 {
			if *huIdx == -1 && *gonIdx == -1 {
				*gonIdx = playerID
			} else {
				otherPlayer.OnFail(action.Command)
			}
		} else if (action.Command & PON) != 0 {
			if *huIdx == -1 && *gonIdx == -1 && *ponIdx == -1 {
				*ponIdx = playerID
			} else {
				otherPlayer.OnFail(action.Command)
			}
		}
	}
}

func (room *Room) huUnder2() bool {
	count := 0
	for i := 0; i < 4; i++ {
		if room.Players[i].IsHu {
			count++
		}
	}
	if count <= 2 {
		for i := 0; i < 4; i++ {
			if !room.Players[i].IsHu {
				room.Players[i].IsTing = room.Players[i].CheckTing(&room.Players[i].MaxTai)
			}
		}
		return true;
	}
	return false;
}

func (room *Room) lackPenalty() {
	for i := 0; i < 4; i++ {
		if (room.Players[i].Hand.ContainColor(room.Players[i].Lack)) {
			for j := 0; j < 4; j++ {
				if room.Players[j].Hand[room.Players[j].Lack].Count() == 0 && i != j {
					room.Players[i].Credit -= 16;
					room.Players[j].Credit += 16;
				}
			}
		}
	}
}

func (room *Room) noTingPenalty() {
	for i := 0; i < 4; i++ {
		if !room.Players[i].IsTing && !room.Players[i].IsHu {
			for j := 0; j < 4; j++ {
				if room.Players[j].IsTing && i != j {
					room.Players[i].Credit -= int(math.Pow(2, float64(room.Players[j].MaxTai - 1)));
					room.Players[j].Credit += int(math.Pow(2, float64(room.Players[j].MaxTai - 1)));
				}
			}
		}
	}
}

func (room *Room) returnMoney() {
	for i := 0; i < 4; i++ {
		if !room.Players[i].IsTing && !room.Players[i].IsHu {
			for j := 0; j < 4; j++ {
				if i != j {
					room.Players[i].Credit -= room.Players[i].GonRecord[j];
					room.Players[j].Credit += room.Players[i].GonRecord[j];
				}
			}
		}
	}
}

func (room *Room) end() {
	var data []GameResult
	for _, player := range room.Players {
		data = append(data, GameResult {player.Hand.ToStringArray(), player.Credit})
	}
	room.BroadcastEnd(data)
}