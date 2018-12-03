package util

import (
	"math"
	"time"
	"math/rand"
	"sync"

	"github.com/googollee/go-socket.io"
	// "github.com/fanliao/go-promise"

	. "MJCard"
)

func NewRoom(game *GameManager, name string) *Room {
	return &Room {game: game, name: name}
	
}

type Room struct {
	Players      []*Player
	ChangedTiles [4][]Card
	ChoosedLack  [4]int
	Deck         Cards
	DiscardTiles Cards
	HuTiles      Cards

	IO           *socketio.Server

	game         *GameManager

	name         string
}

type GameResult struct {
	hand  []string
	score int
}

func (this Room) Name() string {
	return this.name
}

func (this Room) NumPlayer() int {
	list := this.game.PlayerManager.FindPlayersInRoom(this.name)
	num := 0
	for _, player := range list {
		if (player.State & (READY | PLAYING)) != 0 {
			num++
		}
	}
	return num
}

func (this *Room) AddPlayer(playerList []string) {
	for _, uuid := range playerList {
		index := this.game.PlayerManager.FindPlayerByUUID(uuid)
		this.game.PlayerManager[index].Room = this.name
	}
	list := this.game.PlayerManager.FindPlayersInRoom(this.name)
	nameList := GetNameList(list)
	for _, player := range list {
		(*player.Socket).Emit("readyToStart", this.name, nameList)
	}
}

func (this *Room) RemovePlayer(id int) {
	this.Players = append(this.Players[: id], this.Players[id + 1: ]...)
}

func (this *Room) WaitToStart() {
	for ; this.NumPlayer() < 4; {
		time.Sleep(0)
	}
	this.BroadcastGameStart()

	this.Run()
}

func (this *Room) Accept(uuid string) int {
	index := this.game.PlayerManager.FindPlayerByUUID(uuid)
	if index == -1 {
		return -1
	}
	player := this.game.PlayerManager[index]
	this.Players = append(this.Players, NewPlayer(this.game, this.NumPlayer(), player.Uuid))
	this.game.PlayerManager[index].State = READY
	this.BroadcastReady(player.Name)
	return this.NumPlayer() - 1
}

func (this Room) GetPlayerList() []string {
	var nameList []string
	for _, player := range this.Players {
		nameList = append(nameList, player.Name())
	}
	return nameList
}

func (this Room) BroadcastReady(name string) {
	this.IO.BroadcastTo(this.name, "broadcastReady", name)
}

func (this Room) BroadcastGameStart() {
	this.IO.BroadcastTo(this.name, "broadcastGameStart", this.GetPlayerList())
}

func (this Room) BroadcastChange(id int) {
	this.IO.BroadcastTo(this.name, "broadcastChange", id)
}

func (this Room) BroadcastLack() {
	this.IO.BroadcastTo(this.name, "afterLack", this.ChoosedLack)
}

func (this Room) BroadcastDraw(id int) {
	this.IO.BroadcastTo(this.name, "broadcastDraw", id)
}

func (this Room) BroadcastThrow(id int, card Card) {
	this.IO.BroadcastTo(this.name, "broadcastThrow", id, card.ToString())
}

func (this Room) BroadcastCommand(from int, to int, command int, card Card, score int) {
	if command == COMMAND_ONGON {
		this.IO.BroadcastTo(this.name, "broadcastCommand", from, to, command, "", score)
	} else {
		this.IO.BroadcastTo(this.name, "broadcastCommand", from, to, command, card.ToString(), score)
	}
}

func (this Room) BroadcastEnd(data []GameResult) {
	this.IO.BroadcastTo(this.name, "end", data)
}
func (this *Room) Run() {
	time.Sleep(2 * time.Second)
	this.init()
	time.Sleep(3 * time.Second)
	this.changeCard()
	time.Sleep(5 * time.Second)
	this.chooseLack()

	currentId := 0
	onlyThrow := false
	gameOver  := false
	for ; gameOver; {
		curPlayer := this.Players[currentId]
		throwCard := Card {-1, 0}
		action := Action {NONE, throwCard, 0}

		if onlyThrow {
			throwCard = curPlayer.ThrowCard()
			curPlayer.Hand.Sub(throwCard)
			onlyThrow = false
		} else {
			drawCard := this.Deck.Draw()
			this.BroadcastDraw(currentId)
			action := curPlayer.Draw(drawCard)
			throwCard = action.Card
		}

		ponIdx, gonIdx, huIdx := -1, -1, -1
		fail := false
		if (action.Command & COMMAND_PONGON) != 0 {
			var waitGroup sync.WaitGroup
			waitGroup.Add(4)
			act := Action {NONE, Card {-1, 0}, 0 }
			var actionSet [4]Action
			actionSet[0] = act

			for i := 1; i < 4; i++ {
				id := (i + currentId) % 4;
				tai := 0
				if this.Players[id].CheckHu(action.Card, &tai) {
					var cards map[int][]Card;
					cards[COMMAND_HU] = append(cards[COMMAND_HU], action.Card)
					go func () {
						actionSet[i] = this.Players[i].OnCommand(cards, COMMAND_HU, (4 + currentId - id) % 4)
						waitGroup.Done()
					}()
				} else {
					waitGroup.Done()
				}
			}
			waitGroup.Wait()
			for i := 1; i < 4; i++ {
				id := (i + currentId) % 4
				act = actionSet[i]
				if (act.Command & COMMAND_HU) != 0 {
					tai := 0
					this.Players[id].CheckHu(action.Card, &tai)

					score := int(math.Pow(2, float64(tai)))
					curPlayer.Credit        -= score
					this.Players[id].Credit += score
					this.Players[id].HuCards.Add(action.Card)

					if !fail {
						curPlayer.Door.Sub(action.Card)
						curPlayer.VisiableDoor.Sub(action.Card)
						this.HuTiles.Add(action.Card)
					}

					huIdx = id
					fail = true
				}
			}
		} else if (action.Command & COMMAND_ZIMO) != 0 && (action.Command & COMMAND_ONGON) != 0 {
			this.checkOthers(currentId, throwCard, &huIdx, &gonIdx, &ponIdx)
		}

		if fail {
			curPlayer.OnFail(action.Command)
		} else {
			curPlayer.OnSuccess(currentId, action.Command, action.Card, action.Score)
		}

		curPlayer.JustGon = false

		if huIdx != -1 {
			currentId = (huIdx + 1) % 4
			if gonIdx != -1 {
				this.Players[gonIdx].OnFail(COMMAND_GON)
			}
			if ponIdx != -1 {
				this.Players[ponIdx].OnFail(COMMAND_PON)
			}
		} else if gonIdx != -1 {
			this.Players[gonIdx].Gon(throwCard, true)
			curPlayer.Credit -= 2
			this.Players[gonIdx].Credit += 2
			this.Players[gonIdx].GonRecord[currentId] += 2
			currentId = gonIdx
			this.Players[gonIdx].OnSuccess(currentId, COMMAND_GON, throwCard, 2)
			if ponIdx != -1 {
				this.Players[ponIdx].OnFail(COMMAND_PON)
			}
		} else if ponIdx != -1 {
			this.Players[ponIdx].Pon(throwCard)
			currentId = ponIdx
			onlyThrow = true
			this.Players[ponIdx].OnSuccess(currentId, COMMAND_PON, throwCard, 0)
		} else if !fail && (action.Command & COMMAND_ONGON) != 0 || (action.Command & COMMAND_ONGON) != 0 {
			currentId = currentId
		} else {
			if throwCard.Color > 0 {
				this.DiscardTiles.Add(throwCard)
			}
			currentId = (currentId + 1) % 4
		}
		if this.Deck.IsEmpty() {
			gameOver = true
		}
	}
	if this.huUnder2() {
		this.lackPenalty()
		this.noTingPenalty()
		this.returnMoney()
	}
	this.end()
}

func (this *Room) Stop() {
	// TODO
}

func (this *Room) init() {
	this.Deck         = NewCards(true)
	this.DiscardTiles = NewCards(false)
	this.HuTiles      = NewCards(false)

	len := this.Deck.Count()
	for _, player := range this.Players {
		player.Init()
		for j := 0; j < 13; j++ {
			idx := rand.Int31n(int32(len))
			result := this.Deck.At(int(idx))
			this.Deck.Sub(result)
			player.Hand.Add(result)
			len -= 1
		}
		player.Socket().Emit("dealCard", player.Hand.ToStringArray())
	}
}

func (this *Room) changeCard() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	for i := 0; i < 4; i++ {
		go func () {
			this.ChangedTiles[i] = this.Players[i].ChangeCard()
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()

	rand := rand.Int31n(3)
	var tmp []Card
	if rand == 0 {
		tmp = this.ChangedTiles[0]
		for i := 0; i < 3; i++ {
			this.ChangedTiles[i] = this.ChangedTiles[i + 1]
		}
		this.ChangedTiles[3] = tmp
	} else if rand == 1 {
		tmp = this.ChangedTiles[3]
		for i := 3; i >= 3; i-- {
			this.ChangedTiles[i] = this.ChangedTiles[i - 1]
		}
		this.ChangedTiles[0] = tmp
	} else {
		for i := 0; i < 2; i++ {
			tmp = this.ChangedTiles[i];
			this.ChangedTiles[i] = this.ChangedTiles[i + 2];
			this.ChangedTiles[i + 2] = tmp;
		}
	}

	for i := 0; i < 4; i++ {
		this.Players[i].Hand.Add(this.ChangedTiles[i])
		t := CardArrayToCards(this.ChangedTiles[i])
		this.Players[i].Socket().Emit("afterChange", t.ToStringArray())
	}
}

func (this *Room) chooseLack() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	for i := 0; i < 4; i++ {
		go func () {
			this.ChoosedLack[i] = this.Players[i].ChooseLack()
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	this.BroadcastLack()
}

func (this *Room) checkOthers(currentId int, throwCard Card, huIdx *int, gonIdx *int, ponIdx *int) {
	action := Action {NONE, throwCard, 0}
	var playerCommand [4]Action
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)
	playerCommand[0] = action
	for i := 1; i < 4; i++ {
		id := (i + currentId) % 4
		var actions map[int][]Card
		otherPlayer := this.Players[id]

		command := 0
		tai     := 0
		if otherPlayer.CheckHu(throwCard, &tai) {
			if !otherPlayer.IsHu {
				command |= COMMAND_HU
				actions[COMMAND_HU] = append(actions[COMMAND_HU], throwCard)
			}
		}

		if otherPlayer.Hand[throwCard.Color].GetIndex(throwCard.Value) == 3 {
			if otherPlayer.CheckGon(throwCard) {
				command |= COMMAND_GON
				actions[COMMAND_GON] = append(actions[COMMAND_GON], throwCard)
			}
		}

		if otherPlayer.CheckPon(throwCard) {
			command |= COMMAND_PON
			actions[COMMAND_PON] = append(actions[COMMAND_PON], throwCard)
		}

		if command == NONE {
			action.Command = NONE
			playerCommand[i] = action
			waitGroup.Done()
		} else if otherPlayer.IsHu {
			if (command & COMMAND_HU) != 0 {
				action.Command = COMMAND_HU
			} else if (command & COMMAND_GON) != 0 {
				action.Command = COMMAND_GON
			}
			action.Card = throwCard
			playerCommand[i] = action
			waitGroup.Done()
		} else {
			go func() {
				playerCommand[i] = otherPlayer.OnCommand(actions, command, ((4 + currentId - otherPlayer.ID()) % 4))
				waitGroup.Done()
			}()
		}
	}
	waitGroup.Wait()
	for i := 1; i < 4; i++ {
		playerId := (i + currentId) % 4
		otherPlayer := this.Players[playerId]
		action = playerCommand[i]
		tai := 0
		otherPlayer.CheckHu(throwCard, &tai)

		if (action.Command & COMMAND_HU) != 0 {
			otherPlayer.IsHu = true
			otherPlayer.HuCards.Add(action.Card)
			if *huIdx == -1 {
				this.HuTiles.Add(action.Card)
			}
			*huIdx = playerId
			Tai := tai
			if this.Players[currentId].JustGon {
				Tai++ 
			}
			score := int(math.Pow(2, float64(Tai - 1)))
			otherPlayer.Credit += score
			if otherPlayer.MaxTai < tai {
				otherPlayer.MaxTai = tai
			}
			this.Players[currentId].Credit -= score
			otherPlayer.OnSuccess(currentId, COMMAND_HU, action.Card, score)
		} else if (action.Command & COMMAND_GON) != 0 {
			if *huIdx == -1 && *gonIdx == -1 {
				*gonIdx = playerId
			} else {
				otherPlayer.OnFail(action.Command)
			}
		} else if (action.Command & COMMAND_PON) != 0 {
			if *huIdx == -1 && *gonIdx == -1 && *ponIdx == -1 {
				*ponIdx = playerId
			} else {
				otherPlayer.OnFail(action.Command)
			}
		}
	}
}

func (this *Room) huUnder2() bool {
	count := 0
	for i := 0; i < 4; i++ {
		if this.Players[i].IsHu {
			count++
		}
	}
	if count <= 2 {
		for i := 0; i < 4; i++ {
			if !this.Players[i].IsHu {
				this.Players[i].IsTing = this.Players[i].CheckTing(&this.Players[i].MaxTai)
			}
		}
		return true;
	}
	return false;
}

func (this *Room) lackPenalty() {
	for i := 0; i < 4; i++ {
		if (this.Players[i].Hand.ContainColor(this.Players[i].Lack)) {
			for j := 0; j < 4; j++ {
				if this.Players[j].Hand[this.Players[j].Lack].Count() == 0 && i != j {
					this.Players[i].Credit -= 16;
					this.Players[j].Credit += 16;
				}
			}
		}
	}
}

func (this *Room) noTingPenalty() {
	for i := 0; i < 4; i++ {
		if !this.Players[i].IsTing && !this.Players[i].IsHu {
			for j := 0; j < 4; j++ {
				if this.Players[j].IsTing && i != j {
					this.Players[i].Credit -= int(math.Pow(2, float64(this.Players[j].MaxTai - 1)));
					this.Players[j].Credit += int(math.Pow(2, float64(this.Players[j].MaxTai - 1)));
				}
			}
		}
	}
}

func (this *Room) returnMoney() {
	for i := 0; i < 4; i++ {
		if !this.Players[i].IsTing && !this.Players[i].IsHu {
			for j := 0; j < 4; j++ {
				if i != j {
					this.Players[i].Credit -= this.Players[i].GonRecord[j];
					this.Players[j].Credit += this.Players[i].GonRecord[j];
				}
			}
		}
	}
}

func (this *Room) end() {
	var data []GameResult
	for _, player := range this.Players {
		data = append(data, GameResult {player.Hand.ToStringArray(), player.Credit})
	}
	this.BroadcastEnd(data)
}