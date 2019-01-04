package mahjong

import (
	"encoding/json"
	"time"
)

const microSec = 1000000

// COMMAND is a map of command type
var COMMAND = map[string]int{
	"NONE":   0,
	"PON":    1,
	"GON":    2,
	"ONGON":  4,
	"PONGON": 8,
	"HU":     16,
	"ZIMO":   32,
}

// NewAction creates a new action
func NewAction(command int, tile Tile, score int) Action {
	return Action{command, tile, score}
}

// NewActionSet creates a new action set
func NewActionSet() ActionSet {
	return make(ActionSet)
}

// Action represent a command made by player
type Action struct {
	Command int
	Tile    Tile
	Score   int
}

// ToJSON converts action to json string
func (act Action) ToJSON() string {
	type Tmp struct {
		Command int
		Tile    string
		Score   int
	}
	tmp     := Tmp {act.Command, act.Tile.ToString(), act.Score}
	JSON, _ := json.Marshal(tmp)
	return string(JSON)
}

// JSONToAction converts json string to action
func JSONToAction(actionStr string) Action {
	type Tmp struct {
		Command int
		Tile    string
		Score   int
	}
	var t Tmp
	json.Unmarshal([]byte(actionStr), &t)
	return NewAction(t.Command, StringToTile(t.Tile), t.Score)
}

// ActionSet represents a set of action
type ActionSet map[int][]Tile

// ToJSON converts action set to json string
func (set ActionSet) ToJSON() string {
	type Tmp struct {
		Key   int
		Value []string
	}
	var tmpSet []Tmp
	for key, value := range set {
		t     := ArrayToSuitSet(value)
		tmp   := Tmp {key, t.ToStringArray()}
		tmpSet = append(tmpSet, tmp)
	}
	JSON, _ := json.Marshal(tmpSet)
	return string(JSON)
}

// ChangeTiles emits to client to get the change tiles
func (player *Player) ChangeTiles() []Tile {
	defaultChange := ArrayToSuitSet(player.defaultChangeTile()).ToStringArray()
	waitingTime   := 30 * time.Second
	t := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		t[i] = defaultChange[i]
	}

	player.Socket().Emit("change", defaultChange, waitingTime / microSec)
	val := player.waitForSocket("changeTile", t, waitingTime)
	var changeTiles []Tile
	if player.checkChangeTiles(val) {
		valArr := val.([]interface{})
		for i := 0; i < 3; i++ {
			changeTiles = append(changeTiles, StringToTile(valArr[i].(string)))
		}
	} else {
		changeTiles = StringArrayToTileArray(defaultChange)
	}
	player.Hand.Sub(changeTiles)
	player.room.BroadcastChange(player.ID)
	return changeTiles
}

// ChooseLack emits to client to get the choose lack
func (player *Player) ChooseLack() int {
	defaultLack := float64(0)
	waitingTime := 10 * time.Second
	go player.Socket().Emit("lack", defaultLack, waitingTime / microSec)
	val := player.waitForSocket("chooseLack", defaultLack, waitingTime)
	if (player.checkLack(val)) {
		player.Lack = int(val.(float64))
	} else {
		player.Lack = 0
	}
	return player.Lack
}

// Throw emits to client to get the throw Tile
func (player *Player) Throw(drawTile Tile) Tile {
	if drawTile.Suit == -1 {
		drawTile = player.Hand.At(0)
	} 
	defaultTile := drawTile.ToString()
	waitingTime := 10 * time.Second
	go player.Socket().Emit("throw", defaultTile, waitingTime / microSec)
	val := player.waitForSocket("throwTile", defaultTile, waitingTime)
	var throwTile Tile
	if player.checkThrow(val) {
		throwTile = StringToTile(val.(string))
	} else {
		throwTile = StringToTile(defaultTile)
	}
	player.Hand.Sub(throwTile)
	player.room.BroadcastThrow(player.ID, throwTile)
	return throwTile
}

// Draw draws a Tile
func (player *Player) Draw(drawTile Tile) Action {
	player.Hand.Add(drawTile)
	player.Socket().Emit("draw", drawTile.ToString())

	var tai int
	player.CheckHu(NewTile(-1, 0), &tai)
	actionSet, command := player.getAvaliableAction(true, drawTile, tai)
	playerAct          := NewAction(COMMAND["NONE"], drawTile, 0)

	if command == COMMAND["NONE"] {
		playerAct.Command = COMMAND["NONE"]
		playerAct.Tile = drawTile
	} else if player.IsHu {
		if (command & COMMAND["ZIMO"]) != 0 {
			playerAct.Command = COMMAND["ZIMO"]
		} else if (command & COMMAND["ONGON"]) != 0 {
			playerAct.Command = COMMAND["ONGON"]
		} else if (command & COMMAND["PONGON"]) != 0 {
			playerAct.Command = COMMAND["PONGON"]
		}
		playerAct.Tile = actionSet[playerAct.Command][0]
	} else {
		playerAct = player.Command(actionSet, command)
	}

	player.procDrawCommand(drawTile, &playerAct, tai)
	return playerAct
}

// Command emits to client to get command
func (player *Player) Command(actionSet ActionSet, command int) Action {
	defaultCommand := NewAction(COMMAND["NONE"], NewTile(-1, 0), 0).ToJSON()
	waitingTime    := 10 * time.Second
	go player.Socket().Emit("command", actionSet.ToJSON(), command, waitingTime / microSec)
	val := player.waitForSocket("sendCommand", defaultCommand, waitingTime)
	if player.checkCommand(val) {
		return JSONToAction(val.(string))
	} 
	return JSONToAction(defaultCommand)
}

// Fail emits to client to notice the command is failed
func (player *Player) Fail(command int) {
	player.Socket().Emit("fail", command)
}

// Success emits to client to notice the command is successed
func (player *Player) Success(from int, command int, tile Tile, score int) {
	player.Socket().Emit("success", from, command, tile.ToString(), score)
	player.room.BroadcastCommand(from, player.ID, command, tile, score)
}

func (player *Player) defaultChangeTile() []Tile {
	var result []Tile
	count := 0
	for s := 0; s < 3 && count < 3; s++ {
		if (player.Hand[s].Count()) >= 3 {
			for v := uint(0); count < 3 && v < 9; v++ {
				for n := uint(0); count < 3 && n < player.Hand[s].GetIndex(v); n++ {
					result = append(result, NewTile(s, v))
					count++
				}
			}
		}
	}
	return result
}

func (player *Player) waitForSocket(eventName string, defaultValue interface{}, waitingTime time.Duration) interface{} {
	c := make(chan interface{}, 1)
	var val interface{}
	go func() {
		player.Socket().On(eventName, func(v interface{}) {
			c <- v
		})
	}()
	select {
	case val = <-c:
	case <-time.After(waitingTime):
		val = defaultValue
	}
	return val
}

func (player *Player) getAvaliableAction(isDraw bool, tile Tile, tai int) (ActionSet, int) {
	actionSet := NewActionSet()
	command   := 0

	if isDraw {
		actionSet, command = player.checkDrawAction(tile, tai)
	} else {
		actionSet, command = player.checkNonDrawAction(tile, tai)
	}
	return actionSet, command
}

func (player *Player) checkDrawAction(tile Tile, tai int) (ActionSet, int) {
	actionSet := NewActionSet()
	command   := 0
	if tai > 0 {
		command |= COMMAND["ZIMO"]
		actionSet[COMMAND["ZIMO"]] = append(actionSet[COMMAND["ZIMO"]], tile)
	}
	for s := 0; s < 3; s++ {
		for v := uint(0); v < 9; v++ {
			tmpTile := NewTile(s, v)
			
			if player.Hand[s].GetIndex(v) == 4 {
				player.Hand.Sub(tmpTile)
				if player.CheckGon(tmpTile) {
					command |= COMMAND["ONGON"]
					actionSet[COMMAND["ONGON"]] = append(actionSet[COMMAND["ONGON"]], tmpTile)
				}
				player.Hand.Add(tmpTile)
			} else if player.Hand[s].GetIndex(v) == 1 && player.Door[s].GetIndex(v) == 3 {
				player.Hand.Sub(tmpTile)
				if player.CheckGon(tmpTile) {
					command |= COMMAND["PONGON"]
					actionSet[COMMAND["PONGON"]] = append(actionSet[COMMAND["PONGON"]], tmpTile)
				}
				player.Hand.Add(tmpTile)
			}
		}
	}
	return actionSet, command
}

func (player *Player) checkNonDrawAction(tile Tile, tai int) (ActionSet, int) {
	actionSet := NewActionSet()
	command   := 0
	if tai > 0 {
		command |= COMMAND["HU"]
		actionSet[COMMAND["HU"]] = append(actionSet[COMMAND["HU"]], tile)
	}
	if player.Hand[tile.Suit].GetIndex(tile.Value) == 3 {
		if player.CheckGon(tile) {
			command |= COMMAND["GON"]
			actionSet[COMMAND["GON"]] = append(actionSet[COMMAND["GON"]], tile)
		}
	}
	if player.CheckPon(tile) {
		command |= COMMAND["PON"]
		actionSet[COMMAND["PON"]] = append(actionSet[COMMAND["PON"]], tile)
	}
	return actionSet, command
}

func (player *Player) procDrawCommand(drawTile Tile, act *Action, tai int) {
	if (act.Command & COMMAND["ZIMO"]) != 0 {
		act.Score = player.Hu(act.Tile, tai, act.Command, true, true, -1)
	} else if (act.Command & (COMMAND["ONGON"] | COMMAND["PONGON"])) != 0 {
		act.Score = player.Gon(act.Tile, act.Command, -1)
	} else {
		if player.IsHu {
			act.Tile = drawTile
			player.Hand.Sub(act.Tile)
			player.room.BroadcastThrow(player.ID, drawTile)
		} else {
			act.Tile = player.Throw(drawTile)
		}
	}
}