package util

import (
	"MJCard"
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

// Action represent a command made by player
type Action struct {
	Command int
	Card    MJCard.Card
	Score   int
}


// ChangeCard emits to client to get the change cards
func (player *Player) ChangeCard() []MJCard.Card {
	defaultChange := MJCard.CardArrayToCards(player.defaultChangeCard()).ToStringArray()
	waitingTime   := 30 * time.Second
	input         := make([]interface{}, 3)
	for i:= 0; i < 3; i++ {
		input[i] = defaultChange[i]
	}

	player.Socket().Emit("change", defaultChange, waitingTime / 1000000)
	val := player.waitForSocket("changeCard", input, waitingTime)
	var changeCards []MJCard.Card
	for i:= 0; i < 3; i++ {
		changeCards = append(changeCards, MJCard.StringToCard(val[i].(string)))
	}
	player.Hand.Sub(changeCards)
	player.room.BroadcastChange(player.ID)
	return changeCards
}

// ChooseLack emits to client to get the choose lack
func (player *Player) ChooseLack() int {
	defaultLack := 0
	waitingTime := 10 * time.Second
	input       := make([]interface{}, 1)
	input[0]     = defaultLack
	go player.Socket().Emit("lack", defaultLack, waitingTime / 1000000)
	val        := player.waitForSocket("chooseLack", input, waitingTime)
	player.Lack = val[0].(int)
	return player.Lack
}

// ThrowCard emits to client to get the throw card
func (player *Player) ThrowCard() MJCard.Card {
	defaultCard := player.Hand.At(0).ToString()
	waitingTime := 10 * time.Second
	input       := make([]interface{}, 1)
	input[0]     = defaultCard
	go player.Socket().Emit("throw", defaultCard, waitingTime / 1000000)
	val       := player.waitForSocket("chooseLack", input, waitingTime)
	throwCard := MJCard.StringToCard(val[0].(string))
	player.Hand.Sub(throwCard)
	player.room.BroadcastThrow(player.ID, throwCard)
	return throwCard
}

// Draw draws a card
func (player *Player) Draw(drawCard MJCard.Card) Action {
	player.Hand.Add(drawCard)
	player.Socket().Emit("draw", drawCard.ToString())

	var tai int
	player.CheckHu(MJCard.Card {Color: -1, Value: 0}, &tai)
	actions, command := player.getAvaliableAction(true, drawCard, tai)
	action           := Action {NONE, drawCard, 0}

	if command == NONE {
		action.Command = NONE
		action.Card    = drawCard
	} else if player.IsHu {
		if (command & ZIMO) != 0 {
			action.Command = ZIMO
		} else if (command & ONGON) != 0 {
			action.Command = ONGON
		} else if (command & PONGON) != 0 {
			action.Command = PONGON
		}
		action.Card = actions[action.Command][0]
	} else {
		action = player.Command(actions, command)
	}

	player.procCommand(drawCard, &action, tai)
	return action
}

// Command emits to client to get command
func (player *Player) Command(cards map[int][]MJCard.Card, command int) Action {
	type ActionSet struct {
		Key   int
		Value []string
	}
	var actions []ActionSet
	for key, value := range cards {
		t := MJCard.CardArrayToCards(value)
		actionSet := ActionSet {key, t.ToStringArray()}
		actions = append(actions, actionSet)
	}

	defaultCommand := Action {NONE, MJCard.Card {Color: -1, Value: 0}, 0}
	waitingTime    := 10 * time.Second
	b, _           := json.Marshal(actions)
	input          := make([]interface{}, 1)
	input[0]        = defaultCommand
	go player.Socket().Emit("command", string(b), command, waitingTime / 1000000)
	val := player.waitForSocket("chooseLack", input, waitingTime)
	return val[0].(Action)
}

// OnFail emits to client to notice the command is failed
func (player *Player) OnFail(command int) {
	player.Socket().Emit("fail", command)
}

// OnSuccess emits to client to notice the command is successed
func (player *Player) OnSuccess(from int, command int, card MJCard.Card, score int) {
	player.Socket().Emit("success", from, command, card.ToString(), score)
	player.room.BroadcastCommand(from, player.ID, command, card, score)
}

func (player *Player) defaultChangeCard() []MJCard.Card {
	var result []MJCard.Card;
	count := 0;
	for c := 0; c < 3 && count < 3; c++ {
		if (player.Hand[c].Count()) >= 3 {
			for v := uint(0); count < 3 && v < 9; v++ {
				for n := uint(0); count < 3 && n < player.Hand[c].GetIndex(v); n++ {
					result = append(result, MJCard.Card {Color: c, Value: v})
					count++;
				}
			}
		}
	}
	return result;
}

func (player *Player) waitForSocket(eventName string, defaultValue []interface{}, waitingTime time.Duration) []interface{} {
	c := make(chan []interface{}, 1)
	var val []interface{}
	go func() {
		player.Socket().On(eventName, func(v []interface{}) {
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

func (player *Player) getAvaliableAction(isDraw bool, card MJCard.Card, tai int) (map[int][]MJCard.Card, int) {
	actions := make(map[int][]MJCard.Card)
	command := 0

	if isDraw {
		actions, command = player.checkDrawAction(card, tai)
	} else {
		actions, command = player.checkNonDrawAction(card, tai)
	}
	return actions, command
}

func (player *Player) checkDrawAction(card MJCard.Card, tai int) (map[int][]MJCard.Card, int) {
	actions := make(map[int][]MJCard.Card)
	command := 0
	if tai > 0 {
		command |= ZIMO
		actions[ZIMO] = append(actions[ZIMO], card)
	}
	for c := 0; c < 3; c++ {
		for v :=uint(0); v < 9; v++ {
			tmpCard := MJCard.Card {Color: c, Value: v}
			if player.Hand[c].GetIndex(v) == 4 {
				if player.CheckGon(tmpCard) {
					command |= ONGON
					actions[ONGON] = append(actions[ONGON], tmpCard)
				}
			} else if player.Hand[c].GetIndex(v) == 1 && player.Door[c].GetIndex(v) == 3 {
				if player.CheckGon(tmpCard) {
					command |= PONGON
					actions[PONGON] = append(actions[PONGON], tmpCard)
				}
			}
		}
	}
	return actions, command
}

func (player *Player) checkNonDrawAction(card MJCard.Card, tai int) (map[int][]MJCard.Card, int) {
	actions := make(map[int][]MJCard.Card)
	command := 0
	if tai > 0 {
		command |= HU
		actions[HU] = append(actions[HU], card)
	}
	if player.Hand[card.Color].GetIndex(card.Value) == 3 {
		if player.CheckGon(card) {
			command |= GON
			actions[GON] = append(actions[GON], card)
		}
	}
	if player.CheckPon(card) {
		command |= PON
		actions[PON] = append(actions[PON], card)
	}
	return actions, command
}

func (player *Player) procCommand(drawCard MJCard.Card, action *Action, tai int) {
	if (action.Command & ZIMO) != 0 {
		player.ziMo(action, tai)
	} else if (action.Command & ONGON) != 0 {
		player.onGon(action)
	} else if (action.Command & PONGON) != 0 {
		player.onGon(action)
	} else {
		if player.IsHu {
			action.Card = drawCard
			player.room.BroadcastThrow(player.ID, drawCard)
		} else {
			action.Card = player.ThrowCard()
		}
		player.Hand.Sub(action.Card)
	}
}

func (player *Player) ziMo(action *Action, tai int) {
	player.IsHu = true
	player.HuCards.Add(action.Card)
	player.room.HuTiles.Add(action.Card)
	player.Hand.Sub(action.Card)
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
	player.Gon(action.Card, false)
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
	player.Gon(action.Card, true)
	action.Score = 1
	for i := 0; i < 4; i++ {
		if i != player.ID {
			player.Credit++
			player.GonRecord[i]++
			player.room.Players[i].Credit--
		}
	}
} 