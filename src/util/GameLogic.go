package util

import (
	"math/rand"
	"sync"
	"math"
	"encoding/json"
	"time"

	"tile"
	"manager"
	"action"
)

// Game State
const (
	BeforeStart = iota
	DealCard
	ChangeCard
	ChooseLack
	IdxTurn
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
	time.Sleep(5 * time.Second)
}

func (room *Room) init() {
	room.Deck    = tile.NewSet(true)
	room.HuTiles = tile.NewSet(false)

	for _, player := range room.Players {
		player.Init()
		for j := 0; j < 13; j++ {
			result := room.Deck.Draw()
			player.Hand.Add(result)
		}
		player.Socket().Emit("dealCard", player.Hand.ToStringArray())
	}
	room.State = DealCard
}

func (room *Room) changeCard() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	for i := 0; i < 4; i++ {
		go func (id int) {
			room.ChangedTiles[id] = room.Players[id].ChangeTiles()
			waitGroup.Done()
		}(i)
	}
	waitGroup.Wait()

	rand   := rand.Int31n(3)
	offset := [3]int {1, 2, 3}
	var tmp [4][]tile.Tile
	for i := 0; i < 4; i++ {
		tmp[(i + offset[rand]) % 4] = room.ChangedTiles[i]
	}

	for i := 0; i < 4; i++ {
		room.Players[i].Hand.Add(tmp[i])
		t := tile.ArrayToSet(tmp[i])
		room.Players[i].Socket().Emit("afterChange", t.ToStringArray(), rand)
	}
	room.State = ChangeCard
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
	room.State = ChooseLack
}

func (room *Room) checkAction(currentIdx int, playerAct action.Action, throwCard tile.Tile) (bool, int, int, int) {
	ponIdx, gonIdx, huIdx := -1, -1, -1
	fail                  := false

	if (playerAct.Command & action.PONGON) != 0 {
		fail = room.checkRobGon(currentIdx, playerAct.Tile, &huIdx)
	} else if (playerAct.Command & action.ZIMO) == 0 && (playerAct.Command & action.ONGON) == 0 {
		room.checkOthers(currentIdx, throwCard, &huIdx, &gonIdx, &ponIdx)
	}

	return fail, huIdx, gonIdx, ponIdx
}

func (room *Room) checkRobGon(currentIdx int, gonCard tile.Tile, huIdx *int) bool {
	var waitGroup sync.WaitGroup
	var playersAct [3]action.Action
	waitGroup.Add(3)
	for i := 1; i < 4; i++ {
		id := (i + currentIdx) % 4;
		tai := 0
		if room.Players[id].CheckHu(gonCard, &tai) {
			actionSet           := action.NewSet()
			actionSet[action.HU] = append(actionSet[action.HU], gonCard)
			go func (i int) {
				playersAct[i - 1] = room.Players[i].Command(actionSet, action.HU)
				waitGroup.Done()
			}(i)
		} else {
			waitGroup.Done()
		}
	}
	waitGroup.Wait()
	return room.robGon(currentIdx, playersAct, gonCard, huIdx)
}

func (room *Room) robGon(currentIdx int, playersAct [3]action.Action, huTile tile.Tile, huIdx *int) bool {
	fail      := false
	curPlayer := room.Players[currentIdx]
	for i := 1; i < 4; i++ {
		id        := (i + currentIdx) % 4
		playerAct := playersAct[i - 1]
		if (playerAct.Command & action.HU) != 0 {
			tai := 0
			room.Players[id].CheckHu(huTile, &tai)
			score := int(math.Pow(2, float64(tai)))
			curPlayer.Credit        -= score
			room.Players[id].Credit += score
			room.Players[id].HuTiles.Add(huTile)
			room.Players[id].Success(currentIdx, action.HU, huTile, score)
			if !fail {
				curPlayer.Door.Sub(huTile)
				curPlayer.VisiableDoor.Sub(huTile)
				room.HuTiles.Add(huTile)
			}
			*huIdx = id
			fail = true
		}
	}
	return fail
}

func (room *Room) checkOthers(currentIdx int, throwTile tile.Tile, huIdx *int, gonIdx *int, ponIdx *int) {
	playerAct := action.NewAction(action.NONE, throwTile, 0)
	var playersAct [3]action.Action
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)
	for i := 1; i < 4; i++ {
		otherPlayer := room.Players[(i + currentIdx) % 4]
		tai         := 0

		otherPlayer.CheckHu(throwTile, &tai)
		actionSet, command := otherPlayer.getAvaliableAction(false, throwTile, tai)
		if command == action.NONE {
			playerAct.Command = action.NONE
			playersAct[i - 1] = playerAct
			waitGroup.Done()
		} else if otherPlayer.IsHu {
			if (command & action.HU) != 0 {
				playerAct.Command = action.HU
			} else if (command & action.GON) != 0 {
				playerAct.Command = action.GON
			}
			playerAct.Tile    = throwTile
			playersAct[i - 1] = playerAct
			waitGroup.Done()
		} else {
			go func(i int) {
				playersAct[i - 1] = otherPlayer.Command(actionSet, command)
				waitGroup.Done()
			}(i)
		}
	}
	waitGroup.Wait()
	for i := 1; i < 4; i++ {
		playerID    := (i + currentIdx) % 4
		otherPlayer := room.Players[playerID]
		tai         := 0
		playerAct    = playersAct[i - 1]
		otherPlayer.CheckHu(throwTile, &tai)

		if (playerAct.Command & action.HU) != 0 {
			otherPlayer.IsHu = true
			otherPlayer.HuTiles.Add(playerAct.Tile)
			if *huIdx == -1 {
				room.HuTiles.Add(playerAct.Tile)
			}
			*huIdx = playerID
			Tai := tai
			if room.Players[currentIdx].JustGon {
				Tai++ 
			}
			score := int(math.Pow(2, float64(Tai - 1)))
			otherPlayer.Credit += score
			if otherPlayer.MaxTai < tai {
				otherPlayer.MaxTai = tai
			}
			room.Players[currentIdx].Credit -= score
			otherPlayer.Success(currentIdx, action.HU, playerAct.Tile, score)
		} else if (playerAct.Command & action.GON) != 0 {
			if *huIdx == -1 && *gonIdx == -1 {
				*gonIdx = playerID
			} else {
				otherPlayer.Fail(playerAct.Command)
			}
		} else if (playerAct.Command & action.PON) != 0 {
			if *huIdx == -1 && *gonIdx == -1 && *ponIdx == -1 {
				*ponIdx = playerID
			} else {
				otherPlayer.Fail(playerAct.Command)
			}
		}
	}
}

func (room *Room) doAction(currentIdx int, throwTile tile.Tile, huIdx int, gonIdx int, ponIdx int) (int, bool) {
	curPlayer := room.Players[currentIdx]
	onlyThrow := false

	if huIdx != -1 {
		currentIdx = (huIdx + 1) % 4
		if gonIdx != -1 {
			room.Players[gonIdx].Fail(action.GON)
		}
		if ponIdx != -1 {
			room.Players[ponIdx].Fail(action.PON)
		}
	} else if gonIdx != -1 {
		room.Players[gonIdx].Success(currentIdx, action.GON, throwTile, 2)
		room.Players[gonIdx].Gon(throwTile, true)
		room.Players[gonIdx].Credit += 2
		room.Players[gonIdx].GonRecord[currentIdx] += 2
		curPlayer.Credit -= 2
		currentIdx = gonIdx
	} else if ponIdx != -1 {
		room.Players[ponIdx].Success(currentIdx, action.PON, throwTile, 0)
		room.Players[ponIdx].Pon(throwTile)
		currentIdx = ponIdx
		onlyThrow = true
	}
	return currentIdx, onlyThrow
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
	players := manager.FindPlayerListInRoom(room.Name)
	for _, player := range players {
		player.State = manager.WAITING
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
		if (room.Players[i].Hand.IsContainColor(room.Players[i].Lack)) {
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