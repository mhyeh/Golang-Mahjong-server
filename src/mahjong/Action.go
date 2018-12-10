package mahjong

import (
	"time"
	"math"
	"encoding/json"
)

// Command type
const  (
	NONE   = 0
	PON    = 1
	GON    = 2
	ONGON  = 4
	PONGON = 8
	HU     = 16
	ZIMO   = 32
)

// NewAction creates a new action
func NewAction(command int, tile Tile, score int) Action {
	action := Action {command, tile, score}
	return action
}

// NewActionSet creates a new action set
func NewActionSet() ActionSet {
	set := make(ActionSet)
	return set
}

// Action represent a command made by player
type Action struct {
	Command int
	Tile    Tile
	Score   int
}

// ActionSet represents a set of action
type ActionSet map[int][]Tile

// ChangeTiles emits to client to get the change cards
func (player *Player) ChangeTiles() []Tile {
	defaultChange := ArrayToSuitSet(player.defaultChangeCard()).ToStringArray()
	waitingTime   := 30 * time.Second
	t := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		t[i] = defaultChange[i]
	}

	player.Socket().Emit("change", defaultChange, waitingTime / 1000000)
	val := player.waitForSocket("changeCard", t, waitingTime)
	var changeTiles []Tile
	if player.checkChangeTiles(val) {
		valArr := val.([]interface{})
		for i:= 0; i < 3; i++ {
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
	go player.Socket().Emit("lack", defaultLack, waitingTime / 1000000)
	val := player.waitForSocket("chooseLack", defaultLack, waitingTime)
	if player.checkLack(val) {
		player.Lack = int(val.(float64))
	} else {
		player.Lack = 0
	}
	return player.Lack
}

// Throw emits to client to get the throw Tile
func (player *Player) Throw() Tile {
	defaultTile := player.Hand.At(0).ToString()
	waitingTime := 10 * time.Second
	go player.Socket().Emit("throw", defaultTile, waitingTime / 1000000)
	val := player.waitForSocket("throwCard", defaultTile, waitingTime)
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
func (player *Player) Draw(drawCard Tile) Action {
	player.Hand.Add(drawCard)
	player.Socket().Emit("draw", drawCard.ToString())

	var tai int
	player.CheckHu(NewTile(-1, 0), &tai)
	actionSet, command := player.getAvaliableAction(true, drawCard, tai)
	playerAct          := NewAction(NONE, drawCard, 0)

	if command == NONE {
		playerAct.Command = NONE
		playerAct.Tile    = drawCard
	} else if player.IsHu {
		if (command & ZIMO) != 0 {
			playerAct.Command = ZIMO
		} else if (command & ONGON) != 0 {
			playerAct.Command = ONGON
		} else if (command & PONGON) != 0 {
			playerAct.Command = PONGON
		}
		playerAct.Tile = actionSet[playerAct.Command][0]
	} else {
		playerAct = player.Command(actionSet, command)
	}

	player.procDrawCommand(drawCard, &playerAct, tai)
	return playerAct
}

// Command emits to client to get command
func (player *Player) Command(actionSet ActionSet, command int) Action {
	type Tmp struct {
		Key   int
		Value []string
	}
	type TmpAct struct {
		Command int
		Tile    string
		Score   int
	}

	var tmpSet []Tmp
	for key, value := range actionSet {
		t := ArrayToSuitSet(value)
		tmp := Tmp {key, t.ToStringArray()}
		tmpSet = append(tmpSet, tmp)
	}
	
	defaultCommand := TmpAct {NONE, "c0", 0}
	waitingTime    := 10 * time.Second
	actionJSON, _  := json.Marshal(actionSet)
	commandJSON, _ := json.Marshal(defaultCommand)
	go player.Socket().Emit("command", string(actionJSON), command, waitingTime / 1000000)
	val := player.waitForSocket("sendCommand", string(commandJSON), waitingTime)
	var commandStr string
	if player.checkCommand(val) {
		commandStr = val.(string)
	} else {
		commandStr = string(commandJSON)
	}
	var t TmpAct
	json.Unmarshal([]byte(commandStr), &t)
	return NewAction(t.Command, StringToTile(t.Tile), 0)
}

// Fail emits to client to notice the command is failed
func (player *Player) Fail(command int) {
	player.Socket().Emit("fail", command)
}

// Success emits to client to notice the command is successed
func (player *Player) Success(from int, command int, Tile Tile, score int) {
	player.Socket().Emit("success", from, command, Tile.ToString(), score)
	player.room.BroadcastCommand(from, player.ID, command, Tile, score)
}

func (player *Player) defaultChangeCard() []Tile {
	var result []Tile;
	count := 0;
	for c := 0; c < 3 && count < 3; c++ {
		if (player.Hand[c].Count()) >= 3 {
			for v := uint(0); count < 3 && v < 9; v++ {
				for n := uint(0); count < 3 && n < player.Hand[c].GetIndex(v); n++ {
					result = append(result, Tile {Suit: c, Value: v})
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

func (player *Player) getAvaliableAction(isDraw bool, Tile Tile, tai int) (ActionSet, int) {
	actionSet := NewActionSet()
	command   := 0

	if isDraw {
		actionSet, command = player.checkDrawAction(Tile, tai)
	} else {
		actionSet, command = player.checkNonDrawAction(Tile, tai)
	}
	return actionSet, command
}

func (player *Player) checkDrawAction(Tile Tile, tai int) (ActionSet, int) {
	actionSet := NewActionSet()
	command   := 0
	if tai > 0 {
		command |= ZIMO
		actionSet[ZIMO] = append(actionSet[ZIMO], Tile)
	}
	for s := 0; s < 3; s++ {
		for v :=uint(0); v < 9; v++ {
			tmpCard := NewTile(s, v)
			if player.Hand[s].GetIndex(v) == 4 {
				if player.CheckGon(tmpCard) {
					command |= ONGON
					actionSet[ONGON] = append(actionSet[ONGON], tmpCard)
				}
			} else if player.Hand[s].GetIndex(v) == 1 && player.Door[s].GetIndex(v) == 3 {
				if player.CheckGon(tmpCard) {
					command |= PONGON
					actionSet[PONGON] = append(actionSet[PONGON], tmpCard)
				}
			}
		}
	}
	return actionSet, command
}

func (player *Player) checkNonDrawAction(Tile Tile, tai int) (ActionSet, int) {
	actionSet := NewActionSet()
	command := 0
	if tai > 0 {
		command |= HU
		actionSet[HU] = append(actionSet[HU], Tile)
	}
	if player.Hand[Tile.Suit].GetIndex(Tile.Value) == 3 {
		if player.CheckGon(Tile) {
			command |= GON
			actionSet[GON] = append(actionSet[GON], Tile)
		}
	}
	if player.CheckPon(Tile) {
		command |= PON
		actionSet[PON] = append(actionSet[PON], Tile)
	}
	return actionSet, command
}

func (player *Player) procDrawCommand(drawCard Tile, act *Action, tai int) {
	if (act.Command & ZIMO) != 0 {
		player.ziMo(act, tai)
	} else if (act.Command & ONGON) != 0 {
		player.onGon(act)
	} else if (act.Command & PONGON) != 0 {
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

func (player *Player) ziMo(action *Action, tai int) {
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

func (player *Player) onGon(action *Action) {
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

func (player *Player) ponGon(action *Action) {
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