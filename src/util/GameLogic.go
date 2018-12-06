package util

import (
	"math/rand"
	"sync"
	"math"
	"encoding/json"

	"MJCard"
)

// GameResult represents the result of mahjong
type GameResult struct {
	Hand  []string
	Score int
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
			command |= HU
			actions[HU] = append(actions[HU], throwCard)
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
			go func(id int) {
				playerCommand[id] = otherPlayer.Command(actions, command, ((4 + currentID - otherPlayer.ID) % 4))
				waitGroup.Done()
			}(i)
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

func (room *Room) end() {
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