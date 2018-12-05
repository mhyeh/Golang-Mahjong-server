package util

import (
	"time"
	"math"
	"encoding/json"
	"strconv"

	"github.com/googollee/go-socket.io"

	"MJCard"
)

// NewPlayer creates a new player
func NewPlayer(game *GameManager, id int, uuid string) *Player {
	return &Player {game: game, id: id, uuid: uuid} 
}

// Player represents a player in mahjong
type Player struct {
	Lack   int
	Credit int

	Hand         MJCard.Cards
	Door         MJCard.Cards
	VisiableDoor MJCard.Cards
	HuCards      MJCard.Cards

	GonRecord [4]int
	MaxTai    int

	IsHu    bool
	IsTing  bool
	JustGon bool

	game *GameManager
	id   int
	uuid string
}

// Action represent a command made by player
type Action struct {
	Command int
	Card    MJCard.Card
	Score   int
}

// ID returns the player's id
func (player Player) ID() int {
	return player.id 
}

// Name returns the player's name
func (player Player) Name() string {
	index := player.game.PlayerManager.FindPlayerByUUID(player.uuid)
	return player.game.PlayerManager[index].Name
}

// Room returns the player's room
func (player Player) Room() string {
	index := player.game.PlayerManager.FindPlayerByUUID(player.uuid)
	return player.game.PlayerManager[index].Room
}

// Socket returns the player's socket
func (player Player) Socket() socketio.Socket {
	index := player.game.PlayerManager.FindPlayerByUUID(player.uuid)
	return *player.game.PlayerManager[index].Socket
}

// UUID returns the player's uuid
func (player Player) UUID() string {
	return player.uuid
}

// Init inits the player's state
func (player *Player) Init() {
	index := player.game.PlayerManager.FindPlayerByUUID(player.uuid)
	player.game.PlayerManager[index].State = PLAYING
	for i := 0; i < 3; i++ {
		player.Door[i]         = 0
		player.VisiableDoor[i] = 0
		player.Hand[i]         = 0
		player.HuCards[i]      = 0
	}

	player.GonRecord[0] = 0;
	player.GonRecord[1] = 0;
	player.GonRecord[2] = 0;
	player.GonRecord[3] = 0;

	player.Credit  = 0;
	player.MaxTai  = 0;
	player.IsHu    = false;
	player.IsTing  = false;
	player.JustGon = false;
	player.Lack    = -1;
}

// CheckGon checks if the player can gon
func (player *Player) CheckGon(card MJCard.Card) bool {
	if card.Color == player.Lack {
		return false
	}

	if !player.IsHu {
		return true
	}

	handCount := int(player.Hand[card.Color].GetIndex(card.Value))
	oldTai := SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	
	for i := 0; i < handCount; i++ {
		player.Hand.Sub(card)
		player.Door.Add(card)
	}
	newTai := SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	if newTai > 0 {
		newTai--
	}
	for i := 0; i < handCount; i++ {
		player.Hand.Add(card);
		player.Door.Sub(card);
	}
	return (oldTai == newTai);
}

// CheckPon checks if the player can pon
func (player *Player) CheckPon(card MJCard.Card) bool {
	if card.Color == player.Lack || player.IsHu {
		return false;
	}
	count := player.Hand[card.Color].GetIndex(card.Value)
	return count >= 2;
}

// CheckHu checks if the player can hu
func (player *Player) CheckHu(card MJCard.Card, tai *int) bool {
	*tai = 0
	if player.Hand[player.Lack].Count() > 0 {
		return false;
	}
	if card.Color == -1 {
		*tai = SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	} else {
		if card.Color == player.Lack {
			return false;
		}
		player.Hand.Add(card)
		*tai = SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
		player.Hand.Sub(card)
	}
	return *tai > 0
}

// CheckTing checks if the player is ting
func (player *Player) CheckTing(max *int) bool {
	*max = 0;
	tHand := player.Hand.Translate(player.Lack)
	tDoor := player.Door.Translate(player.Lack)
	total := tHand + tDoor
	for i := uint(0); i < 18; i++ {
		if ((total >> (i * 3)) & 7) < 4 {
			newHand := tHand + (1 << (i * 3));
			tai := SSJ(newHand, tDoor);
			if tai > *max {
				*max = tai;
			}
		}
	}
	return *max > 0;
}

// Gon gons the card
func (player *Player) Gon(card MJCard.Card, visible bool) {
	player.JustGon = true
	player.Door.Add(card)
	player.Door.Add(card)
	player.Door.Add(card)
	player.Door.Add(card)

	if visible {
		player.VisiableDoor.Add(card)
		player.VisiableDoor.Add(card)
		player.VisiableDoor.Add(card)
		player.VisiableDoor.Add(card)
	}

	player.Hand.Sub(card)
	player.Hand.Sub(card)
	player.Hand.Sub(card)
	player.Hand.Sub(card)
}

// Pon pons the card
func (player *Player) Pon(card MJCard.Card) {
	player.Door.Add(card)
	player.Door.Add(card)
	player.Door.Add(card)

	player.VisiableDoor.Add(card)
	player.VisiableDoor.Add(card)
	player.VisiableDoor.Add(card)

	player.Hand.Sub(card)
	player.Hand.Sub(card)
}

// Tai cals the tai
func (player *Player) Tai(card MJCard.Card) int {
	player.Hand.Add(card)
	result := SSJ(player.Hand.Translate(player.Lack), player.Door.Translate(player.Lack))
	player.Hand.Sub(card)
	return result;
}

// ChangeCard emits to client to get the change cards
func (player *Player) ChangeCard() []MJCard.Card {
	defaultChange := player.defaultChangeCard()
	t := MJCard.CardArrayToCards(defaultChange)
	waitingTime := 30 * time.Second
	player.Socket().Emit("change", t.ToStringArray(), waitingTime / 1000000)

	c := make(chan []MJCard.Card)
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
	waitingTime := 10 * time.Second
	player.Socket().Emit("lack", defaultLack, waitingTime / 1000000)

	c := make(chan int)
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
	waitingTime := 10 * time.Second
	player.Socket().Emit("throw", defaultCard.ToString(), waitingTime / 1000000)

	c := make(chan MJCard.Card)
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

	if (action.Command & ZIMO) != 0 {
		player.IsHu = true
		player.HuCards.Add(action.Card)
		player.game.Rooms[player.Room()].HuTiles.Add(action.Card)
		player.Hand.Sub(action.Card)
		action.Card.Color = -1
		Tai := tai
		if player.JustGon {
			Tai++ 
		}
		score := int(math.Pow(2, float64(Tai)))
		for i := 0; i < 4; i++ {
			if player.id != i {
				player.Credit += score
				action.Score += score
				if player.MaxTai < tai {
					player.MaxTai = tai
				}
				player.game.Rooms[player.Room()].Players[i].Credit -= score
			}
		}
	} else if (action.Command & ONGON) != 0 {
		player.Gon(action.Card, false)
		for i := 0; i < 4; i++ {
			if i != player.id {
				player.Credit += 2
				action.Score += 2
				player.GonRecord[i] += 2
				player.game.Rooms[player.Room()].Players[i].Credit -= 2
			}
		}
	} else if (action.Command & PONGON) != 0 {
		player.Gon(action.Card, true)
		for i := 0; i < 4; i++ {
			if i != player.id {
				player.Credit++
				action.Score++
				player.GonRecord[i]++
				player.game.Rooms[player.Room()].Players[i].Credit--
			}
		}
	} else {
		if player.IsHu {
			action.Card = drawCard
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
		key string
		value []string
	}
	var actions []ActionSet
	for key, value := range cards {
		t := MJCard.CardArrayToCards(value)
		actionSet := ActionSet {strconv.Itoa(key), t.ToStringArray()}
		actions = append(actions, actionSet)
	}

	defaultCommand := Action {NONE, MJCard.Card {Color: -1, Value: 0}, 0}
	waitingTime := 10 * time.Second
	b, _ := json.Marshal(actions)
	player.Socket().Emit("command", string(b), command, waitingTime / 1000000)
	
	c := make(chan Action)
	var action Action
	go func() {
		player.Socket().On("throwCard", func (command int, card string) {
			c<-Action {command, MJCard.StringToCard(card), 0}
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