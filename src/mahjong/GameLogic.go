package mahjong

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// Game State
const (
	BeforeStart = iota
	DealTile
	ChangeTile
	ChooseLack
	IdxTurn
)

// GameResult represents the result of mahjong
type GameResult struct {
	Hand     []string
	Door     []string
	Score    int
	ScoreLog []ScoreRecord
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
			room.BroadcastDraw(currentIdx, room.Deck.Count())
			act       = curPlayer.Draw(drawTile)
			throwTile = act.Tile
		}

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

func (room *Room) preproc() {
	time.Sleep(2 * time.Second)
	room.init()
	time.Sleep(3 * time.Second)
	room.changeTile()
	time.Sleep(5 * time.Second)
	room.chooseLack()
	time.Sleep(3 * time.Second)
}

func (room *Room) init() {
	room.Deck    = NewSuitSet(true)
	room.HuTiles = NewSuitSet(false)

	for _, player := range room.Players {
		player.Init()
		for j := 0; j < 13; j++ {
			result := room.Deck.Draw()
			player.Hand.Add(result)
		}
		player.Socket().Emit("dealTile", player.Hand.ToStringArray())
	}
	room.State = DealTile
}

func (room *Room) changeTile() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	for i := 0; i < 4; i++ {
		go func(id int) {
			room.ChangedTiles[id] = room.Players[id].ChangeTiles()
			waitGroup.Done()
		}(i)
	}
	waitGroup.Wait()

	rand   := rand.Int31n(3)
	offset := [3]int{1, 2, 3}
	var tmp [4][]Tile
	for i := 0; i < 4; i++ {
		tmp[(i + offset[rand]) % 4] = room.ChangedTiles[i]
	}

	for i := 0; i < 4; i++ {
		room.Players[i].Hand.Add(tmp[i])
		t := ArrayToSuitSet(tmp[i])
		room.Players[i].Socket().Emit("afterChange", t.ToStringArray(), rand)
	}
	room.State = ChangeTile
}

func (room *Room) chooseLack() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(4)
	for i := 0; i < 4; i++ {
		go func(id int) {
			room.ChoosedLack[id] = room.Players[id].ChooseLack()
			waitGroup.Done()
		}(i)
	}
	waitGroup.Wait()
	room.BroadcastLack()
	room.State = ChooseLack
}

func (room *Room) checkAction(currentIdx int, playerAct Action, throwTile Tile) (bool, int, int, int) {
	ponIdx, gonIdx, huIdx := -1, -1, -1
	fail                  := false

	if (playerAct.Command & COMMAND["PONGON"]) != 0 {
		fail = room.checkRobGon(currentIdx, playerAct.Tile, &huIdx)
	} else if (playerAct.Command & COMMAND["ZIMO"]) == 0 && (playerAct.Command & COMMAND["ONGON"]) == 0 {
		room.checkOthers(currentIdx, throwTile, &huIdx, &gonIdx, &ponIdx)
	}

	return fail, huIdx, gonIdx, ponIdx
}

func (room *Room) checkRobGon(currentIdx int, gonTile Tile, huIdx *int) bool {
	var waitGroup  sync.WaitGroup
	var playersAct [3]Action
	waitGroup.Add(3)
	for i := 1; i < 4; i++ {
		id  := (i + currentIdx) % 4
		tai := 0
		if room.Players[id].CheckHu(gonTile, &tai) {
			actionSet := NewActionSet()
			actionSet[COMMAND["HU"]] = append(actionSet[COMMAND["HU"]], gonTile)
			go func(i int) {
				playersAct[i - 1] = room.Players[i].Command(actionSet, COMMAND["HU"])
				waitGroup.Done()
			}(i)
		} else {
			waitGroup.Done()
		}
	}
	waitGroup.Wait()
	return room.robGon(currentIdx, playersAct, gonTile, huIdx)
}

func (room *Room) robGon(currentIdx int, playersAct [3]Action, huTile Tile, huIdx *int) bool {
	fail      := false
	curPlayer := room.Players[currentIdx]
	for i := 1; i < 4; i++ {
		id        := (i + currentIdx) % 4
		playerAct := playersAct[i - 1]
		if (playerAct.Command & COMMAND["HU"]) != 0 {
			tai := 0
			room.Players[id].CheckHu(huTile, &tai)
			score := room.Players[id].Hu(huTile, tai, COMMAND["HU"], true, !fail, currentIdx)
			room.Players[id].Success(currentIdx, COMMAND["HU"], huTile, score)
			if !fail {
				curPlayer.Door.Sub(huTile)
				curPlayer.VisiableDoor.Sub(huTile)
			}
			*huIdx = id
			fail   = true
		}
	}
	return fail
}

func (room *Room) checkOthers(currentIdx int, throwTile Tile, huIdx *int, gonIdx *int, ponIdx *int) {
	playerAct := NewAction(COMMAND["NONE"], throwTile, 0)
	var playersAct [3]Action
	var waitGroup  sync.WaitGroup
	waitGroup.Add(3)
	for i := 1; i < 4; i++ {
		otherPlayer := room.Players[(i + currentIdx) % 4]
		tai         := 0

		otherPlayer.CheckHu(throwTile, &tai)
		actionSet, command := otherPlayer.getAvaliableAction(false, throwTile, tai)
		if command == COMMAND["NONE"] {
			playerAct.Command = COMMAND["NONE"]
			playersAct[i - 1] = playerAct
			waitGroup.Done()
		} else if otherPlayer.IsHu {
			if (command & COMMAND["HU"]) != 0 {
				playerAct.Command = COMMAND["HU"]
			} else if (command & COMMAND["GON"]) != 0 {
				playerAct.Command = COMMAND["GON"]
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

		if (playerAct.Command & COMMAND["HU"]) != 0 {
			score := otherPlayer.Hu(playerAct.Tile, tai, COMMAND["HU"], false, *huIdx == -1, currentIdx)
			*huIdx = playerID
			otherPlayer.Success(currentIdx, COMMAND["HU"], playerAct.Tile, score)
		} else if (playerAct.Command & COMMAND["GON"]) != 0 {
			if *huIdx == -1 && *gonIdx == -1 {
				*gonIdx = playerID
			} else {
				otherPlayer.Fail(playerAct.Command)
			}
		} else if (playerAct.Command & COMMAND["PON"]) != 0 {
			if *huIdx == -1 && *gonIdx == -1 && *ponIdx == -1 {
				*ponIdx = playerID
			} else {
				otherPlayer.Fail(playerAct.Command)
			}
		}
	}
}

func (room *Room) doAction(currentIdx int, throwTile Tile, huIdx int, gonIdx int, ponIdx int) (int, bool) {
	onlyThrow := false

	if huIdx != -1 {
		currentIdx = (huIdx + 1) % 4
		if gonIdx != -1 {
			room.Players[gonIdx].Fail(COMMAND["GON"])
		}
		if ponIdx != -1 {
			room.Players[ponIdx].Fail(COMMAND["PON"])
		}
	} else if gonIdx != -1 {
		score := room.Players[gonIdx].Gon(throwTile, COMMAND["GON"], currentIdx)
		room.Players[gonIdx].Success(currentIdx, COMMAND["GON"], throwTile, score)
		currentIdx = gonIdx
	} else if ponIdx != -1 {
		room.Players[ponIdx].Success(currentIdx, COMMAND["PON"], throwTile, 0)
		room.Players[ponIdx].Pon(throwTile)
		currentIdx = ponIdx
		onlyThrow  = true
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
		data = append(data, GameResult {player.Hand.ToStringArray(), player.Door.ToStringArray(), player.Credit, player.ScoreLog})
	}
	room.BroadcastEnd(data)
	players := FindPlayerListInRoom(room.Name)
	for _, player := range players {
		player.State = WAITING
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
	return count <= 2
}

func (room *Room) lackPenalty() {
	score := 16
	for i := 0; i < 4; i++ {
		if room.Players[i].Hand.IsContainColor(room.Players[i].Lack) {
			for j := 0; j < 4; j++ {
				if room.Players[j].Hand[room.Players[j].Lack].Count() == 0 && i != j {
					room.Players[i].Credit -= score
					room.Players[j].Credit += score
					room.Players[i].ScoreLog = append(room.Players[i].ScoreLog, NewScoreRecord("花豬", "to", room.Players[j].Name(), "", -score))
					room.Players[j].ScoreLog = append(room.Players[j].ScoreLog, NewScoreRecord("花豬", "from", room.Players[i].Name(), "", score))
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
					score := int(math.Pow(2, float64(room.Players[j].MaxTai - 1)))
					room.Players[i].Credit -= score
					room.Players[j].Credit += score
					room.Players[i].ScoreLog = append(room.Players[i].ScoreLog, NewScoreRecord("大叫", "to", room.Players[j].Name(), "", -score))
					room.Players[j].ScoreLog = append(room.Players[j].ScoreLog, NewScoreRecord("大叫", "from", room.Players[i].Name(), "", score))
				}
			}
		}
	}
}

func (room *Room) returnMoney() {
	for i := 0; i < 4; i++ {
		if !room.Players[i].IsTing && !room.Players[i].IsHu {
			for j := 0; j < 4; j++ {
				score := room.Players[i].GonRecord[j]
				if score != 0 {
					room.Players[i].Credit -= score
					room.Players[j].Credit += score
					room.Players[i].ScoreLog = append(room.Players[i].ScoreLog, NewScoreRecord("退稅", "to", room.Players[j].Name(), "", -score))
					room.Players[j].ScoreLog = append(room.Players[j].ScoreLog, NewScoreRecord("退稅", "from", room.Players[i].Name(), "", score))
				}
			}
		}
	}
}
