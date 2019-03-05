package mahjong

import (
	"math/rand"
	"sync"
	"time"
)

// Game State
const (
	BeforeStart = iota
	DealTile
	BuHua
	IdxTurn
)

// GameResult represents the result of mahjong
type GameResult struct {
	Hand     []string
	Door     [][]string
	Score    int
	ScoreLog ScoreRecord
}

// Run runs mahjong logic
func (room *Room) Run() {
	room.NumKeepWin = 0
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			room.setWindAndRound(i, j)

			if i == 0 && j == 0 && room.NumKeepWin == 0 {
				room.openDoor(true)
				room.Banker = room.OpenIdx
			} else {
				room.Banker = (room.Banker + 1) % 4
			}
			room.BroadcastBanker(room.Banker);

			for room.KeepWin = true; room.KeepWin; {
				if !(i == 0 && j == 0 && room.NumKeepWin == 0) {
					room.openDoor(false)
				}
				currentIdx := room.Banker
				room.preproc(currentIdx)

				flag := false
				for k := 0; k < 4; k++ {
					if room.Players[k].Flowers.Count() == 8 {
						flag = true
						room.Players[k].Hu(room.Players[k].Flowers.At(0), TaiData{ 8, "八仙過海 " }, COMMAND["ZIMO"], false, false, -1)
					}
				}
				if flag {
					if currentIdx != room.Banker {
						room.KeepWin    = false
						room.NumKeepWin = 0
					} else {
						room.KeepWin = true
						room.NumKeepWin++
					}
				} else {
					onlyThrow := false
					gameOver  := false
					for {
						curPlayer   := room.Players[currentIdx]
						throwTile   := NewTile(-1, 0)
						act         := NewAction(COMMAND["NONE"], throwTile, 0)
						sevenFlower := false
						room.State   = IdxTurn + currentIdx

						if onlyThrow {
							throwTile = curPlayer.Throw(throwTile)
							onlyThrow = false
						} else {
							act, gameOver, sevenFlower = curPlayer.Draw()
							throwTile = act.Tile
						}

						if gameOver {
							if sevenFlower && currentIdx != room.Banker {
								room.KeepWin    = false
								room.NumKeepWin = 0
							} else {
								room.KeepWin = true
								room.NumKeepWin++
							}
							break
						}

						robGon, huIdxArray, gonIdx, ponIdx, eatAction := room.checkAction(currentIdx, act, throwTile)
						if robGon {
							curPlayer.Fail(act.Command)
							room.BroadcastRobGon(curPlayer.ID, act.Tile)
						} else if act.Command != COMMAND["NONE"] && (act.Command & COMMAND["ZIMO"]) == 0 {
							curPlayer.Success(currentIdx, act.Command, act.Tile, act.Score)
						} else if (act.Command & COMMAND["ZIMO"]) != 0 {
							huIdxArray = append(huIdxArray, curPlayer.ID)
						}
						curPlayer.JustGon = false
						
						currentIdx, onlyThrow = room.doAction(currentIdx, throwTile, huIdxArray, gonIdx, ponIdx, eatAction)
						if currentIdx == curPlayer.ID && len(huIdxArray) == 0 && (act.Command & COMMAND["ONGON"]) == 0 && (act.Command & COMMAND["PONGON"]) == 0 {
							curPlayer.DiscardTiles.Add(throwTile)
							currentIdx = (currentIdx + 1) % 4
						}
						
						if len(huIdxArray) > 0 {
							room.end()
							flag := false
							for _, huIdx := range huIdxArray {
								if huIdx == room.Banker {
									room.KeepWin = true
									room.NumKeepWin++
									flag = true
									break
								}
							}
							if !flag {
								room.KeepWin    = false
								room.NumKeepWin = 0
							}
							break
						}
						if room.Deck.Count() <= 16 {
							room.end()
							room.KeepWin = true
							room.NumKeepWin++
							break
						}
					}
				}
				time.Sleep(20 * time.Second)
			}
		}
	}

	players := FindPlayerListInRoom(room.Name, 1)
	for _, player := range players {
		player.State = WAITING
	}
}

func (room *Room) setWindAndRound(wind int, round int) {
	room.Info.Winds = 0
	if wind == round {
		room.Info.Winds = room.Info.Winds | 16
	}
	temp := (1 << uint(wind)) | (1 << uint(round))
	room.Info.Winds = room.Info.Winds | Wind(temp)
	room.Wind  = wind
	room.Round = round
	room.BroadcastWindAndRound(wind, round)
}

func (room *Room) preproc(startIdx int) {
	time.Sleep(1 * time.Second)
	room.init(startIdx)
	time.Sleep(1 * time.Second)
	room.buHua(startIdx)
	time.Sleep(1 * time.Second)
}

func (room *Room) openDoor(isFirst bool) {
	room.OpenIdx = int(rand.Int31n(4))
	room.BroadcastOpenDoor(room.OpenIdx)
	if isFirst {
		room.EastIdx = room.OpenIdx
		room.BroadcastSetSeat(room.EastIdx)
	}
}

func (room *Room) init(startIdx int) {
	room.Deck = NewSuitSet(true)

	for i := 0; i < 4; i++ {
		player := room.Players[(i + startIdx) % 4]
		player.Init()
		for j := 0; j < 16; j++ {
			result := room.Deck.Draw()
			player.Hand.Add(result)
		}
	}
	room.State = DealTile
}

func (room *Room) buHua(startIdx int) {
	flag := true
	for flag {
		flag = false
		for i := 0; i < 4; i++ {
			player := room.Players[(i + startIdx) % 4]
			need   := player.Hand[4].Count()
			if need > 0 {
				flag = true
				player.Flowers[4] += player.Hand[4]
				player.Hand[4] = Suit(0)
				for j := 0; j < int(need); j++ {
					result := room.Deck.Draw()
					player.Hand.Add(result)
				}
			}
		}
	}

	var data [][]string
	for i := 0; i < 4; i++ {
		player := room.Players[i]
		player.Socket().Emit("dealTile", player.Hand.ToStringArray())
		data = append(data, player.Flowers.ToStringArray())
	}
	room.BroadcastBuHua(data)
	room.State = BuHua
}

func (room *Room) checkSevenFlower(player *Player, tile Tile) bool {
	if room.SevenFlower {
		if room.SevenID != player.ID {
			room.Players[room.SevenID].Hu(tile, TaiData{ 8, "七搶一 " }, COMMAND["HU"], false, false, player.ID)
		} else {
			player.Hu(tile, TaiData{ 8, "八仙過海 " }, COMMAND["ZIMO"], false, false, -1)
		}
		return true
	} else if player.Flowers.Count() == 7 {
		room.SevenFlower = true
		room.SevenID     = player.ID
	}
	return false
}

func (room *Room) checkAction(currentIdx int, playerAct Action, throwTile Tile) (bool, []int, int, int, EatAction) {
	ponIdx, gonIdx, huIdxArray, eatAction := -1, -1, []int{}, EatAction{-1, throwTile, throwTile}
	fail := false

	if (playerAct.Command & COMMAND["PONGON"]) != 0 {
		fail = room.checkRobGon(currentIdx, playerAct.Tile, &huIdxArray)
	} else if (playerAct.Command & COMMAND["ZIMO"]) == 0 && (playerAct.Command & COMMAND["ONGON"]) == 0 {
		room.checkOthers(currentIdx, throwTile, &huIdxArray, &gonIdx, &ponIdx, &eatAction)
	}

	return fail, huIdxArray, gonIdx, ponIdx, eatAction
}

func (room *Room) checkRobGon(currentIdx int, gonTile Tile, huIdxArray *[]int) bool {
	var waitGroup  sync.WaitGroup
	var playersAct [3]Action
	waitGroup.Add(3)
	for i := 1; i < 4; i++ {
		id  := (i + currentIdx) % 4
		tai := TaiData{ -1, "" }
		if room.Players[id].CheckHu(gonTile, 0, &tai) {
			actionSet := NewActionSet()
			actionSet[COMMAND["HU"]] = append(actionSet[COMMAND["HU"]], gonTile.ToString())
			go func(i int) {
				playersAct[i - 1] = room.Players[i].Command(actionSet, COMMAND["HU"], currentIdx)
				waitGroup.Done()
			}(id)
		} else {
			waitGroup.Done()
		}
	}
	waitGroup.Wait()
	return room.robGon(currentIdx, playersAct, gonTile, huIdxArray)
}

func (room *Room) robGon(currentIdx int, playersAct [3]Action, huTile Tile, huIdxArray *[]int) bool {
	fail      := false
	curPlayer := room.Players[currentIdx]
	for i := 1; i < 4; i++ {
		id        := (i + currentIdx) % 4
		playerAct := playersAct[i - 1]
		if (playerAct.Command & COMMAND["HU"]) != 0 {
			tai := TaiData{ -1, "" }
			room.Players[id].CheckHu(huTile, 0, &tai)
			room.Players[id].Hu(huTile, tai, COMMAND["HU"], true, !fail, currentIdx)
			// score := room.Players[id].Hu(huTile, tai, COMMAND["HU"], true, !fail, currentIdx)
			// room.Players[id].Success(currentIdx, COMMAND["HU"], huTile, score)
			if !fail {
				curPlayer.GonTiles.Sub(huTile)
				curPlayer.PonTiles.Add(huTile)
			}
			*huIdxArray = append(*huIdxArray, id)
			fail        = true
		}
	}
	return fail
}

func (room *Room) checkOthers(currentIdx int, throwTile Tile, huIdxArray *[]int, gonIdx *int, ponIdx *int, eatAction *EatAction) {
	playerAct := NewAction(COMMAND["NONE"], throwTile, 0)
	var playersAct [3]Action
	var waitGroup  sync.WaitGroup
	waitGroup.Add(3)
	for i := 1; i < 4; i++ {
		otherPlayer := room.Players[(i + currentIdx) % 4]
		tai         := TaiData{ -1, "" }

		otherPlayer.CheckHu(throwTile, 0, &tai)
		actionSet, command := otherPlayer.getAvaliableAction(currentIdx, false, throwTile, tai)
		if command == COMMAND["NONE"] {
			playerAct.Command = COMMAND["NONE"]
			playersAct[i - 1] = playerAct
			waitGroup.Done()
		} else {
			go func(i int) {
				playersAct[i - 1] = otherPlayer.Command(actionSet, command, currentIdx)
				waitGroup.Done()
			}(i)
		}
	}
	waitGroup.Wait()
	for i := 1; i < 4; i++ {
		playerID    := (i + currentIdx) % 4
		otherPlayer := room.Players[playerID]
		playerAct    = playersAct[i - 1]

		if (playerAct.Command & COMMAND["HU"]) != 0 {
			tai := TaiData{ -1, "" }
			otherPlayer.CheckHu(throwTile, 0, &tai)
			otherPlayer.Hu(playerAct.Tile, tai, COMMAND["HU"], false, len(*huIdxArray) == 0, currentIdx)
			// score      := otherPlayer.Hu(playerAct.Tile, tai, COMMAND["HU"], false, len(*huIdxArray) == 0, currentIdx)
			*huIdxArray = append(*huIdxArray, playerID)
			// otherPlayer.Success(currentIdx, COMMAND["HU"], playerAct.Tile, score)
		} else if (playerAct.Command & COMMAND["GON"]) != 0 {
			if len(*huIdxArray) == 0 && *gonIdx == -1 {
				*gonIdx = playerID
			} else {
				otherPlayer.Fail(playerAct.Command)
			}
		} else if (playerAct.Command & COMMAND["PON"]) != 0 {
			if len(*huIdxArray) == 0 && *gonIdx == -1 && *ponIdx == -1 {
				*ponIdx = playerID
			} else {
				otherPlayer.Fail(playerAct.Command)
			}
		} else if (playerAct.Command & COMMAND["EAT"]) != 0 {
			if len(*huIdxArray) == 0 && *gonIdx == -1 && *ponIdx == -1 && (*eatAction).Idx == -1 {
				(*eatAction).Idx    = playerID
				(*eatAction).First  = playerAct.Tile
				(*eatAction).Center = throwTile
			} else {
				otherPlayer.Fail(playerAct.Command)
			}
		}
	}
}

func (room *Room) doAction(currentIdx int, throwTile Tile, huIdxArray []int, gonIdx int, ponIdx int, eatAction EatAction) (int, bool) {
	onlyThrow := false

	if len(huIdxArray) != 0 {
		currentIdx = (huIdxArray[len(huIdxArray) - 1] + 1) % 4
		if gonIdx != -1 {
			room.Players[gonIdx].Fail(COMMAND["GON"])
		}
		if ponIdx != -1 {
			room.Players[ponIdx].Fail(COMMAND["PON"])
		}
		if eatAction.Idx != -1 {
			room.Players[eatAction.Idx].Fail(COMMAND["EAT"])
		}
	} else if gonIdx != -1 {
		room.Players[gonIdx].Success(currentIdx, COMMAND["GON"], throwTile, 0)
		room.Players[gonIdx].Gon(throwTile, COMMAND["GON"], currentIdx)
		currentIdx = gonIdx
	} else if ponIdx != -1 {
		room.Players[ponIdx].Success(currentIdx, COMMAND["PON"], throwTile, 0)
		room.Players[ponIdx].Pon(throwTile)
		currentIdx = ponIdx
		onlyThrow  = true
	} else if eatAction.Idx != -1 {
		room.Players[eatAction.Idx].Success(currentIdx, COMMAND["EAT"], eatAction, 0)
		room.Players[eatAction.Idx].Eat(eatAction)
		currentIdx = eatAction.Idx
		onlyThrow  = true
	}
	return currentIdx, onlyThrow
}

func (room *Room) end() {
	var data []GameResult
	for _, player := range room.Players {
		data = append(data, GameResult{
			player.Hand.ToStringArray(),
			[][]string{
				player.EatTiles.ToStringArray(),
				player.PonTiles.ToStringArray(),
				player.GonTiles.ToStringArray(),
				player.OngonTiles.ToStringArray(),
			},
			player.Credit,
			player.ScoreLog,
		})
	}
	room.BroadcastEnd(data)
	// players := FindPlayerListInRoom(room.Name, 1)
	// for _, player := range players {
	// 	player.State = WAITING
	// }
}
