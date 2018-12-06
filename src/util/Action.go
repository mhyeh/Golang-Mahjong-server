package util

import (
	"MJCard"
	"time"
	"math"
	"encoding/json"
)

// command type
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
	defaultChange := player.defaultChangeCard()
	t := MJCard.CardArrayToCards(defaultChange)
	waitingTime := 30 * time.Second
	player.Socket().Emit("change", t.ToStringArray(), waitingTime / 1000000)

	c := make(chan []MJCard.Card, 1)
	var changeCards []MJCard.Card
	go func() {
		player.Socket().On("changeCard", func (cards []string) {
			c<-MJCard.StringArrayToCardArray(cards)
		})
	}()
	select {
	case changeCards = <-c:
	case <-time.After(waitingTime):
		changeCards = defaultChange
	}
	
	player.Hand.Sub(changeCards)
	player.game.Rooms[player.Room()].BroadcastChange(player.id)
	return changeCards
}

// ChooseLack emits to client to get the choose lack
func (player *Player) ChooseLack() int {
	defaultLack := 0
	waitingTime := 100 * time.Second
	player.Socket().Emit("lack", defaultLack, waitingTime / 1000000)

	c := make(chan int, 1)
	go func() {
		player.Socket().On("chooseLack", func (lack int) {
			c<-lack
		})
	}()
	select {
	case player.Lack = <-c:
	case <-time.After(waitingTime):
		player.Lack = defaultLack
	}
	return player.Lack
}

// ThrowCard emits to client to get the throw card
func (player *Player) ThrowCard() MJCard.Card {
	defaultCard := player.Hand.At(0)
	waitingTime := 1000 * time.Second
	player.Socket().Emit("throw", defaultCard.ToString(), waitingTime / 1000000)

	c := make(chan MJCard.Card, 1)
	var throwCard MJCard.Card
	go func() {
		player.Socket().On("throwCard", func (card string) {
			c<-MJCard.StringToCard(card)
		})
	}()
	select {
	case throwCard = <-c:
	case <-time.After(waitingTime):
		throwCard = defaultCard
	}
	player.Hand.Sub(throwCard)
	player.game.Rooms[player.Room()].BroadcastThrow(player.id, throwCard)
	return throwCard
}

// Draw draws a card
func (player *Player) Draw(drawCard MJCard.Card) Action {
	actions := make(map[int][]MJCard.Card)
	tai     := 0
	command := 0
	player.Hand.Add(drawCard)
	player.Socket().Emit("draw", drawCard.ToString())

	if player.CheckHu(MJCard.Card {Color: -1, Value: 0}, &tai) {
		command |= ZIMO
		actions[ZIMO] = append(actions[ZIMO], drawCard)
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

	action := Action {NONE, drawCard, 0}
	if command == NONE {
		action.Command = NONE
		action.Card    = drawCard
	} else if player.IsHu {
		if (command & ZIMO) != 0 {
			action.Command = ZIMO
			action.Card    = actions[ZIMO][0]
		} else if (command & ONGON) != 0 {
			action.Command = ONGON
			action.Card    = actions[ONGON][0]
		} else if (command & PONGON) != 0 {
			action.Command = PONGON
			action.Card    = actions[PONGON][0]
		}
	} else {
		action = player.OnCommand(actions, command, 0)
	}

	if (action.Command != NONE) {
	}

	if (action.Command & ZIMO) != 0 {
		player.IsHu = true
		player.HuCards.Add(action.Card)
		player.game.Rooms[player.Room()].HuTiles.Add(action.Card)
		player.Hand.Sub(action.Card)
		Tai := tai + 1
		if player.JustGon {
			Tai++ 
		}
		score := int(math.Pow(2, float64(Tai)))
		action.Score = score
		for i := 0; i < 4; i++ {
			if player.id != i {
				player.Credit += score
				if player.MaxTai < tai {
					player.MaxTai = tai
				}
				player.game.Rooms[player.Room()].Players[i].Credit -= score
			}
		}
	} else if (action.Command & ONGON) != 0 {
		player.Gon(action.Card, false)
		action.Score = 2
		for i := 0; i < 4; i++ {
			if i != player.id {
				player.Credit += 2
				player.GonRecord[i] += 2
				player.game.Rooms[player.Room()].Players[i].Credit -= 2
			}
		}
	} else if (action.Command & PONGON) != 0 {
		player.Gon(action.Card, true)
		action.Score = 1
		for i := 0; i < 4; i++ {
			if i != player.id {
				player.Credit++
				player.GonRecord[i]++
				player.game.Rooms[player.Room()].Players[i].Credit--
			}
		}
	} else {
		if player.IsHu {
			action.Card = drawCard
			player.game.Rooms[player.Room()].BroadcastThrow(player.id, drawCard)
		} else {
			action.Card = player.ThrowCard()
		}
		if player.Hand.Have(action.Card) {
			player.Hand.Sub(action.Card)
		}
	}
	return action
}

// OnCommand emits to client to get command
func (player *Player) OnCommand(cards map[int][]MJCard.Card, command int, from int) Action {
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
	waitingTime := 1000 * time.Second
	b, _ := json.Marshal(actions)
	player.Socket().Emit("command", string(b), command, waitingTime / 1000000)
	
	c := make(chan Action, 1)
	var action Action
	go func() {
		player.Socket().On("sendCommand", func (command int, card string) {
			if (command == NONE) {
				c<-Action {NONE, MJCard.Card {Color: -1, Value: 0}, 0}
			} else {
				c<-Action {command, MJCard.StringToCard(card), 0}
			}
		})
	}()
	select {
	case action = <-c:
	case <-time.After(waitingTime):
		action = defaultCommand
	}
	return action
}

// OnFail emits to client to notice the command is failed
func (player *Player) OnFail(command int) {
	player.Socket().Emit("fail", command)
}

// OnSuccess emits to client to notice the command is successed
func (player *Player) OnSuccess(from int, command int, card MJCard.Card, score int) {
	player.Socket().Emit("success", from, command, card.ToString(), score)
	player.game.Rooms[player.Room()].BroadcastCommand(from, player.id, command, card, score)
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