package util

import (
	"math/rand"
	"sync"
	"math"
	"encoding/json"
	"time"

	"MJCard"
)

// GameResult represents the result of mahjong
type GameResult struct {
	Hand  []string
	Score int
}

func (room *Room) preproc() {
	time.Sleep(2 * time.Second)
	room.init()
	time.Sleep(3 * time.Second)
	room.changeCard()
	time.Sleep(5 * time.Second)
	room.chooseLack()
}

func (room *Room) init() {
	room.Deck         = MJCard.NewCards(true)
	room.DiscardTiles = MJCard.NewCards(false)
	room.HuTiles      = MJCard.NewCards(false)

	for _, player := range room.Players {
		player.Init()
		for j := 0; j < 13; j++ {
			result := room.Deck.Draw()
			player.Hand.Add(result)
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
	var tmp [4][]MJCard.Card
	offset := [3]int {1, 2, 3}
	for i := 0; i < 4; i++ {
		tmp[(i + offset[rand]) % 4] = room.ChangedTiles[i]
	}

	for i := 0; i < 4; i++ {
		room.Players[i].Hand.Add(tmp[i])
		t := MJCard.CardArrayToCards(tmp[i])
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

func (room *Room) checkAction(currentID int, action Action, throwCard MJCard.Card) (bool, int, int, int) {
	ponIdx, gonIdx, huIdx := -1, -1, -1
	fail                  := false

	if (action.Command & PONGON) != 0 {
		fail = room.checkRobGon(currentID, action.Card, &huIdx)
	} else if (action.Command & ZIMO) == 0 && (action.Command & ONGON) == 0 {
		room.checkOthers(currentID, throwCard, &huIdx, &gonIdx, &ponIdx)
	}

	return fail, huIdx, gonIdx, ponIdx
}

func (room *Room) checkRobGon(currentID int, gonCard MJCard.Card, huIdx *int) bool {
	var waitGroup sync.WaitGroup
	var actionSet [3]Action
	waitGroup.Add(3)
	for i := 1; i < 4; i++ {
		id := (i + currentID) % 4;
		tai := 0
		if room.Players[id].CheckHu(gonCard, &tai) {
			cards    := make(map[int][]MJCard.Card)
			cards[HU] = append(cards[HU], gonCard)
			go func (i int) {
				actionSet[i - 1] = room.Players[i].Command(cards, HU)
				waitGroup.Done()
			}(i)
		} else {
			waitGroup.Done()
		}
	}
	waitGroup.Wait()
	return room.robGon(currentID, actionSet, gonCard, huIdx)
}

func (room *Room) robGon(currentID int, actionSet [3]Action, huCard MJCard.Card, huIdx *int) bool {
	fail      := false
	curPlayer := room.Players[currentID]
	for i := 1; i < 4; i++ {
		id  := (i + currentID) % 4
		act := actionSet[i - 1]
		if (act.Command & HU) != 0 {
			tai := 0
			room.Players[id].CheckHu(huCard, &tai)
			score := int(math.Pow(2, float64(tai)))
			curPlayer.Credit        -= score
			room.Players[id].Credit += score
			room.Players[id].HuCards.Add(huCard)
			room.Players[id].OnSuccess(currentID, HU, huCard, score)
			if !fail {
				curPlayer.Door.Sub(huCard)
				curPlayer.VisiableDoor.Sub(huCard)
				room.HuTiles.Add(huCard)
			}
			*huIdx = id
			fail = true
		}
	}
	return fail
}

func (room *Room) checkOthers(currentID int, throwCard MJCard.Card, huIdx *int, gonIdx *int, ponIdx *int) {
	action := Action {NONE, throwCard, 0}
	var playerCommand [3]Action
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)
	for i := 1; i < 4; i++ {
		otherPlayer := room.Players[(i + currentID) % 4]
		tai         := 0

		otherPlayer.CheckHu(throwCard, &tai)
		actions, command := otherPlayer.getAvaliableAction(false, throwCard, tai)
		if command == NONE {
			action.Command       = NONE
			playerCommand[i - 1] = action
			waitGroup.Done()
		} else if otherPlayer.IsHu {
			if (command & HU) != 0 {
				action.Command = HU
			} else if (command & GON) != 0 {
				action.Command = GON
			}
			action.Card          = throwCard
			playerCommand[i - 1] = action
			waitGroup.Done()
		} else {
			go func(i int) {
				playerCommand[i - 1] = otherPlayer.Command(actions, command)
				waitGroup.Done()
			}(i)
		}
	}
	waitGroup.Wait()
	for i := 1; i < 4; i++ {
		playerID    := (i + currentID) % 4
		otherPlayer := room.Players[playerID]
		tai         := 0
		action       = playerCommand[i - 1]
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

func (room *Room) doAction(currentID int, throwCard MJCard.Card, huIdx int, gonIdx int, ponIdx int) (int, bool) {
	curPlayer := room.Players[currentID]
	onlyThrow := false

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
		room.Players[gonIdx].Credit += 2
		room.Players[gonIdx].GonRecord[currentID] += 2
		curPlayer.Credit -= 2
		currentID = gonIdx
	} else if ponIdx != -1 {
		room.Players[ponIdx].OnSuccess(currentID, PON, throwCard, 0)
		room.Players[ponIdx].Pon(throwCard)
		currentID = ponIdx
		onlyThrow = true
	}
	return currentID, onlyThrow
}

func (room *Room) end() {
	if room.huUnder2() {
		room.lackPenalty()
		room.noTingPenalty()
		room.returnMoney()
	}

	var data []GameResult
	for _, player := range room.Players {
		data = append(data, GameResult {player.Hand.ToStringArray(), player.Credit})
	}
	b, _ := json.Marshal(data)
	room.BroadcastEnd(string(b))
	// players := room.game.PlayerManager.FindPlayersInRoom(room.name)
	// for _, player := range players {
	// 	player.State = WAITING
	// }
}

func (room *Room) huUnder2() bool {
	count := 0
	for i := 0; i < 4; i++ {
		if room.Players[i].IsHu {
			count++
		} else {
			room.Players[i].IsTing = room.Players[i].CheckTing(&room.Players[i].MaxTai)
		}
	}
	return count <= 2;
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