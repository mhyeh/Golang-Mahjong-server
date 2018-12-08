package util

import (
	"time"
	"math"
	"encoding/json"

	"tile"
	"action"
)

// ChangeTile emits to client to get the change cards
func (player *Player) ChangeTile() []tile.Tile {
	defaultChange := tile.ArrayToSet(player.defaultChangeCard()).ToStringArray()
	waitingTime   := 30 * time.Second
	player.Socket().Emit("change", defaultChange, waitingTime / 1000000)
	val    := player.waitForSocket("changeCard", defaultChange, waitingTime)
	valArr := val.([]interface{})
	var changeCards []tile.Tile
	for i:= 0; i < 3; i++ {
		changeCards = append(changeCards, tile.StringToTile(valArr[i].(string)))
	}
	player.Hand.Sub(changeCards)
	player.room.BroadcastChange(player.ID)
	return changeCards
}

// ChooseLack emits to client to get the choose lack
func (player *Player) ChooseLack() int {
	defaultLack := float64(0)
	waitingTime := 10 * time.Second
	go player.Socket().Emit("lack", defaultLack, waitingTime / 1000000)
	val        := player.waitForSocket("chooseLack", defaultLack, waitingTime)
	player.Lack = int(val.(float64))
	return player.Lack
}

// Throw emits to client to get the throw Tile
func (player *Player) Throw() tile.Tile {
	defaultCard := player.Hand.At(0).ToString()
	waitingTime := 10 * time.Second
	go player.Socket().Emit("throw", defaultCard, waitingTime / 1000000)
	val       := player.waitForSocket("throwCard", defaultCard, waitingTime)
	throwCard := tile.StringToTile(val.(string))
	player.Hand.Sub(throwCard)
	player.room.BroadcastThrow(player.ID, throwCard)
	return throwCard
}

// Draw draws a Tile
func (player *Player) Draw(drawCard tile.Tile) action.Action {
	player.Hand.Add(drawCard)
	player.Socket().Emit("draw", drawCard.ToString())

	var tai int
	player.CheckHu(tile.Tile {Color: -1, Value: 0}, &tai)
	actionSet, command := player.getAvaliableAction(true, drawCard, tai)
	act                := action.NewAction(action.NONE, drawCard, 0)

	if command == action.NONE {
		act.Command = action.NONE
		act.Tile    = drawCard
	} else if player.IsHu {
		if (command & action.ZIMO) != 0 {
			act.Command = action.ZIMO
		} else if (command & action.ONGON) != 0 {
			act.Command = action.ONGON
		} else if (command & action.PONGON) != 0 {
			act.Command = action.PONGON
		}
		act.Tile = actionSet[act.Command][0]
	} else {
		act = player.Command(actionSet, command)
	}

	player.procCommand(drawCard, &act, tai)
	return act
}

// Command emits to client to get command
func (player *Player) Command(actionSet action.Set, command int) action.Action {
	type Tmp struct {
		Key   int
		Value []string
	}
	var tmpSet []Tmp
	for key, value := range actionSet {
		t := tile.ArrayToSet(value)
		tmp := Tmp {key, t.ToStringArray()}
		tmpSet = append(tmpSet, tmp)
	}

	type TmpAct struct {
		Command int
		Tile    string
		Score   int
	}
	defaultCommand := TmpAct {action.NONE, "c0", 0}
	waitingTime    := 10 * time.Second
	actionJSON, _  := json.Marshal(actionSet)
	commandJSON, _ := json.Marshal(defaultCommand)
	go player.Socket().Emit("command", string(actionJSON), command, waitingTime / 1000000)
	val := player.waitForSocket("sendCommand", string(commandJSON), waitingTime)
	var t TmpAct
	json.Unmarshal([]byte(val.(string)), &t)
	return action.NewAction(t.Command, tile.StringToTile(t.Tile), 0)
}

// Fail emits to client to notice the command is failed
func (player *Player) Fail(command int) {
	player.Socket().Emit("fail", command)
}

// Success emits to client to notice the command is successed
func (player *Player) Success(from int, command int, Tile tile.Tile, score int) {
	player.Socket().Emit("success", from, command, Tile.ToString(), score)
	player.room.BroadcastCommand(from, player.ID, command, Tile, score)
}

func (player *Player) defaultChangeCard() []tile.Tile {
	var result []tile.Tile;
	count := 0;
	for c := 0; c < 3 && count < 3; c++ {
		if (player.Hand[c].Count()) >= 3 {
			for v := uint(0); count < 3 && v < 9; v++ {
				for n := uint(0); count < 3 && n < player.Hand[c].GetIndex(v); n++ {
					result = append(result, tile.Tile {Color: c, Value: v})
					count++;
				}
			}
		}
	}
	return result;
}

func (player *Player) waitForSocket(eventName string, defaultValue interface{}, waitingTime time.Duration) interface{} {
	c := make(chan interface{}, 1)
	var val interface{}
	go func() {
		player.Socket().On(eventName, func(v interface{}) {
			c<-v
		})
	}()
	select {
	case val = <-c:
	case <-time.After(waitingTime):
		val = defaultValue
	}
	return val
}

func (player *Player) getAvaliableAction(isDraw bool, Tile tile.Tile, tai int) (action.Set, int) {
	actionSet := action.NewSet()
	command   := 0

	if isDraw {
		actionSet, command = player.checkDrawAction(Tile, tai)
	} else {
		actionSet, command = player.checkNonDrawAction(Tile, tai)
	}
	return actionSet, command
}

func (player *Player) checkDrawAction(Tile tile.Tile, tai int) (action.Set, int) {
	actionSet := action.NewSet()
	command   := 0
	if tai > 0 {
		command |= action.ZIMO
		actionSet[action.ZIMO] = append(actionSet[action.ZIMO], Tile)
	}
	for c := 0; c < 3; c++ {
		for v :=uint(0); v < 9; v++ {
			tmpCard := tile.Tile {Color: c, Value: v}
			if player.Hand[c].GetIndex(v) == 4 {
				if player.CheckGon(tmpCard) {
					command |= action.ONGON
					actionSet[action.ONGON] = append(actionSet[action.ONGON], tmpCard)
				}
			} else if player.Hand[c].GetIndex(v) == 1 && player.Door[c].GetIndex(v) == 3 {
				if player.CheckGon(tmpCard) {
					command |= action.PONGON
					actionSet[action.PONGON] = append(actionSet[action.PONGON], tmpCard)
				}
			}
		}
	}
	return actionSet, command
}

func (player *Player) checkNonDrawAction(Tile tile.Tile, tai int) (action.Set, int) {
	actionSet := action.NewSet()
	command := 0
	if tai > 0 {
		command |= action.HU
		actionSet[action.HU] = append(actionSet[action.HU], Tile)
	}
	if player.Hand[Tile.Color].GetIndex(Tile.Value) == 3 {
		if player.CheckGon(Tile) {
			command |= action.GON
			actionSet[action.GON] = append(actionSet[action.GON], Tile)
		}
	}
	if player.CheckPon(Tile) {
		command |= action.PON
		actionSet[action.PON] = append(actionSet[action.PON], Tile)
	}
	return actionSet, command
}

func (player *Player) procCommand(drawCard tile.Tile, act *action.Action, tai int) {
	if (act.Command & action.ZIMO) != 0 {
		player.ziMo(act, tai)
	} else if (act.Command & action.ONGON) != 0 {
		player.onGon(act)
	} else if (act.Command & action.PONGON) != 0 {
		player.onGon(act)
	} else {
		if player.IsHu {
			act.Tile = drawCard
			player.room.BroadcastThrow(player.ID, drawCard)
		} else {
			act.Tile = player.Throw()
		}
		player.Hand.Sub(act.Tile)
	}
}

func (player *Player) ziMo(action *action.Action, tai int) {
	player.IsHu = true
	player.HuTiles.Add(action.Tile)
	player.room.HuTiles.Add(action.Tile)
	player.Hand.Sub(action.Tile)
	Tai := tai + 1
	if player.JustGon {
		Tai++ 
	}
	score := int(math.Pow(2, float64(Tai)))
	action.Score = score
	for i := 0; i < 4; i++ {
		if player.ID != i {
			player.Credit += score
			if player.MaxTai < tai {
				player.MaxTai = tai
			}
			player.room.Players[i].Credit -= score
		}
	}
} 

func (player *Player) onGon(action *action.Action) {
	player.Gon(action.Tile, false)
	action.Score = 2
	for i := 0; i < 4; i++ {
		if i != player.ID {
			player.Credit += 2
			player.GonRecord[i] += 2
			player.room.Players[i].Credit -= 2
		}
	}
} 

func (player *Player) ponGon(action *action.Action) {
	player.Gon(action.Tile, true)
	action.Score = 1
	for i := 0; i < 4; i++ {
		if i != player.ID {
			player.Credit++
			player.GonRecord[i]++
			player.room.Players[i].Credit--
		}
	}
} 