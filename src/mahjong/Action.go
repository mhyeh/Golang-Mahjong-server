package mahjong

import (
	"encoding/json"
	"strings"
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
	"EAT":    64,
	"TING":   128,
}

// NewAction creates a new action
func NewAction(command int, tile Tile, score int) Action {
	return Action{command, tile, score}
}

// NewActionSet creates a new action set
func NewActionSet() ActionSet {
	return make(ActionSet)
}

// Action represents a command made by player
type Action struct {
	Command int
	Tile    Tile
	Score   int
}

// EatAction represents
type EatAction struct {
	Idx    int
	First  Tile
	Center Tile
}

// ToJSON converts action to json string
func (act Action) ToJSON() string {
	type Tmp struct {
		Command int
		Tile    string
		Score   int
	}
	tmp     := Tmp{ act.Command, act.Tile.ToString(), act.Score }
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
type ActionSet map[int][]string

// Throw emits to client to get the throw Tile
func (player *Player) Throw(drawTile Tile) Tile {
	if drawTile.Suit == -1 {
		drawTile = player.Hand.At(0)
	}
	
	var throwTile Tile
	waitingTime := 10 * time.Second

	if player.IsTing {
		throwTile = drawTile
	} else {
		defaultTile := drawTile.ToString()
		go player.Socket().Emit("throw", defaultTile, waitingTime / microSec)
		val := player.waitForSocket("throwTile", defaultTile, waitingTime)
		if player.checkThrow(val) {
			throwTile = StringToTile(val.(string))
		} else {
			throwTile = StringToTile(defaultTile)
		}
	}

	player.Hand.Sub(throwTile)
	player.room.BroadcastThrow(player.ID, throwTile)
	if player.CheckTing() {
		go func() {
			waitingTime = 5 * time.Second
			go player.Socket().Emit("ting", waitingTime / microSec)
			val := player.waitForSocket("sendTing", false, waitingTime)
			if player.checkTing(val) && val.(bool) {
				player.IsTing = true
				player.room.BroadcastTing(player.ID)
			}
		}()
	}
	return throwTile
}

// Draw draws a Tile
func (player *Player) Draw() (Action, bool, bool) {
	for player.room.Deck.Count() > 16 {
		drawTile := player.room.Deck.Draw()
		player.room.BroadcastDraw(player.ID, player.room.Deck.Count())
		player.Socket().Emit("draw", drawTile.ToString())
		if drawTile.Suit == 4 {
			player.Flowers.Add(drawTile)
			time.Sleep(1 * time.Second)
			player.room.BroadcastHua(player.ID, drawTile)
			if player.room.checkSevenFlower(player, drawTile) {
				return NewAction(COMMAND["NONE"], NewTile(-1, 0), 0), true, true
			}
			continue
		}
		var tai TaiData
		player.CheckHu(drawTile, 1, &tai)
		player.Hand.Add(drawTile)
		actionSet, command := player.getAvaliableAction(player.ID, true, drawTile, tai)
		playerAct          := NewAction(COMMAND["NONE"], drawTile, 0)

		if command == COMMAND["NONE"] {
			playerAct.Command = COMMAND["NONE"]
			playerAct.Tile    = drawTile
		} else {
			playerAct = player.Command(actionSet, command, -1)
		}

		player.procDrawCommand(drawTile, &playerAct, tai)
		player.FirstDraw = false
		return playerAct, false, false
	}

	return NewAction(COMMAND["NONE"], NewTile(-1, 0), 0), true, false
}

// Command emits to client to get command
func (player *Player) Command(actionSet ActionSet, command int, idx int) Action {
	defaultCommand := NewAction(COMMAND["NONE"], NewTile(-1, 0), 0).ToJSON()
	waitingTime    := 10 * time.Second
	go player.Socket().Emit("command", actionSet, command, idx, waitingTime / microSec)
	val := player.waitForSocket("sendCommand", defaultCommand, waitingTime)
	if player.checkCommand(val) {
		a := JSONToAction(val.(string))
		return a
	}
	return JSONToAction(defaultCommand)
}

// Fail emits to client to notice the command is failed
func (player *Player) Fail(command int) {
	player.Socket().Emit("fail", command)
}

// Success emits to client to notice the command is successed
func (player *Player) Success(from int, command int, tile interface{}, score int) {
	var str string
	switch tile.(type) {
	case Tile:
		str = tile.(Tile).ToString()
	case EatAction:
		str = strings.Join([]string{ tile.(EatAction).First.ToString(), tile.(EatAction).Center.ToString() }, ",")
	}
	player.Socket().Emit("success", from, command, str, score)
	player.room.BroadcastCommand(from, player.ID, command, str, score)
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

func (player *Player) getAvaliableAction(id int, isDraw bool, tile Tile, tai TaiData) (ActionSet, int) {
	actionSet := NewActionSet()
	command   := 0

	if isDraw {
		actionSet, command = player.checkDrawAction(tile, tai)
	} else {
		actionSet, command = player.checkNonDrawAction(id, tile, tai)
	}
	return actionSet, command
}

func (player *Player) checkDrawAction(tile Tile, tai TaiData) (ActionSet, int) {
	actionSet := NewActionSet()
	command   := 0
	if tai.Tai >= 0 {
		command |= COMMAND["ZIMO"]
		actionSet[COMMAND["ZIMO"]] = append(actionSet[COMMAND["ZIMO"]], tile.ToString())
	}
	for s := 0; s < 4; s++ {
		for v := uint(0); v < SuitTileCount[s]; v++ {
			tmpTile    := NewTile(s, v)
			tmpTileStr := tmpTile.ToString()
			if player.Hand[s].GetIndex(v) == 4 {
				command |= COMMAND["ONGON"]
				actionSet[COMMAND["ONGON"]] = append(actionSet[COMMAND["ONGON"]], tmpTileStr)
			} else if player.Hand[s].GetIndex(v) == 1 && player.PonTiles.Have(tmpTile) && !player.IsTing {
				command |= COMMAND["PONGON"]
				actionSet[COMMAND["PONGON"]] = append(actionSet[COMMAND["PONGON"]], tmpTileStr)
			}
		}
	}
	return actionSet, command
}

func (player *Player) checkNonDrawAction(id int, tile Tile, tai TaiData) (ActionSet, int) {
	actionSet := NewActionSet()
	command   := 0
	tileStr   := tile.ToString()
	if tai.Tai >= 0 {
		command |= COMMAND["HU"]
		actionSet[COMMAND["HU"]] = append(actionSet[COMMAND["HU"]], tileStr)
	}
	if player.Hand[tile.Suit].GetIndex(tile.Value) == 3 && (id + 1) % 4 != player.ID && !player.IsTing {
		command |= COMMAND["GON"]
		actionSet[COMMAND["GON"]] = append(actionSet[COMMAND["GON"]], tileStr)
	}
	if player.CheckPon(tile) {
		command |= COMMAND["PON"]
		actionSet[COMMAND["PON"]] = append(actionSet[COMMAND["PON"]], tileStr)
	}
	if (id + 1) % 4 == player.ID && player.CheckEat(tile) {
		command |= COMMAND["EAT"]
		player.Hand.Add(tile)
		actionSet[COMMAND["EAT"]] = append(actionSet[COMMAND["EAT"]], tileStr)
		for i := int(tile.Value) - 2; i <= int(tile.Value); i++ {
			flag := true
			for j := 0; j < 3; j++ {
				if (i + j < 0) || (i + j >= 9) || !player.Hand.Have(NewTile(tile.Suit, uint(i + j))) {
					flag = false
					i   += j
					break
				}
			}
			if flag {
				actionSet[COMMAND["EAT"]] = append(actionSet[COMMAND["EAT"]], NewTile(tile.Suit, uint(i)).ToString())
			}
		}
		player.Hand.Sub(tile)
	}
	return actionSet, command
}

func (player *Player) procDrawCommand(drawTile Tile, act *Action, tai TaiData) {
	if (act.Command & COMMAND["ZIMO"]) != 0 {
		act.Score = player.Hu(act.Tile, tai, act.Command, false, true, -1)
	} else if (act.Command & (COMMAND["ONGON"] | COMMAND["PONGON"])) != 0 {
		player.Gon(act.Tile, act.Command, -1)
	} else {
		time.Sleep(1 * time.Second)
		act.Tile = player.Throw(drawTile)
	}
}
