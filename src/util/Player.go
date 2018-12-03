package util

import (
	"time"
	"math"
	"encoding/json"
	"strconv"

	"github.com/googollee/go-socket.io"

	. "MJCard"
)

func NewPlayer(game *GameManager, id int, uuid string) *Player {
	return &Player {game: game, id: id, uuid: uuid} 
}

type Player struct {
	Lack   int
	Credit int

	Hand         Cards
	Door         Cards
	VisiableDoor Cards
	HuCards      Cards

	GonRecord [4]int
	MaxTai    int

	IsHu    bool
	IsTing  bool
	JustGon bool

	game *GameManager
	id   int
	uuid string
}

type Action struct {
	Command int
	Card    Card
	Score   int
}

func (this Player) ID() int {
	return this.id 
}

func (this Player) Name() string {
	index := this.game.PlayerManager.FindPlayerByUUID(this.uuid)
	return this.game.PlayerManager[index].Name
}

func (this Player) Room() string {
	index := this.game.PlayerManager.FindPlayerByUUID(this.uuid)
	return this.game.PlayerManager[index].Room
}

func (this Player) Socket() socketio.Socket {
	index := this.game.PlayerManager.FindPlayerByUUID(this.uuid)
	return *this.game.PlayerManager[index].Socket
}

func (this Player) UUID() string {
	return this.uuid
}

func (this *Player) Init() {
	index := this.game.PlayerManager.FindPlayerByUUID(this.uuid)
	this.game.PlayerManager[index].State = PLAYING
	for i := 0; i < 3; i++ {
		this.Door[i]         = 0
		this.VisiableDoor[i] = 0
		this.Hand[i]         = 0
		this.HuCards[i]      = 0
	}

	this.GonRecord[0] = 0;
	this.GonRecord[1] = 0;
	this.GonRecord[2] = 0;
	this.GonRecord[3] = 0;

	this.Credit  = 0;
	this.MaxTai  = 0;
	this.IsHu    = false;
	this.IsTing  = false;
	this.JustGon = false;
	this.Lack    = -1;
}

func (this *Player) CheckGon(card Card) bool {
	if card.Color == this.Lack {
		return false
	}

	if !this.IsHu {
		return true
	}

	handCount := int(this.Hand[card.Color].GetIndex(card.Value))
	oldTai := SSJ(this.Hand.Translate(this.Lack), this.Door.Translate(this.Lack))
	
	for i := 0; i < handCount; i++ {
		this.Hand.Sub(card)
		this.Door.Add(card)
	}
	newTai := SSJ(this.Hand.Translate(this.Lack), this.Door.Translate(this.Lack))
	if newTai > 0 {
		newTai -= 1
	}
	for i := 0; i < handCount; i++ {
		this.Hand.Add(card);
		this.Door.Sub(card);
	}
	return (oldTai == newTai);
}

func (this *Player) CheckPon(card Card) bool {
	if card.Color == this.Lack || this.IsHu {
		return false;
	}
	count := this.Hand[card.Color].GetIndex(card.Value)
	return count >= 2;
}

func (this *Player) CheckHu(card Card, tai *int) bool {
	*tai = 0
	if this.Hand[this.Lack].Count() > 0 {
		return false;
	}
	if card.Color == -1 {
		*tai = SSJ(this.Hand.Translate(this.Lack), this.Door.Translate(this.Lack))
	} else {
		if card.Color == this.Lack {
			return false;
		}
		this.Hand.Add(card)
		*tai = SSJ(this.Hand.Translate(this.Lack), this.Door.Translate(this.Lack))
		this.Hand.Sub(card)
	}
	return *tai > 0
}

func (this *Player) CheckTing(max *int) bool {
	*max = 0;
	t_Hand := this.Hand.Translate(this.Lack)
	t_Door := this.Door.Translate(this.Lack)
	total := t_Hand + t_Door
	for i := uint(0); i < 18; i++ {
		if ((total >> (i * 3)) & 7) < 4 {
			newHand := t_Hand + (1 << (i * 3));
			tai := SSJ(newHand, t_Door);
			if tai > *max {
				*max = tai;
			}
		}
	}
	return *max > 0;
}

func (this *Player) Gon(card Card, visible bool) {
	this.JustGon = true
	this.Door.Add(card)
	this.Door.Add(card)
	this.Door.Add(card)
	this.Door.Add(card)

	if visible {
		this.VisiableDoor.Add(card)
		this.VisiableDoor.Add(card)
		this.VisiableDoor.Add(card)
		this.VisiableDoor.Add(card)
	}

	this.Hand.Sub(card)
	this.Hand.Sub(card)
	this.Hand.Sub(card)
	this.Hand.Sub(card)
}

func (this *Player) Pon(card Card) {
	this.Door.Add(card)
	this.Door.Add(card)
	this.Door.Add(card)

	this.VisiableDoor.Add(card)
	this.VisiableDoor.Add(card)
	this.VisiableDoor.Add(card)

	this.Hand.Sub(card)
	this.Hand.Sub(card)
}

func (this *Player) Tai(card Card) int {
	this.Hand.Add(card)
	result := SSJ(this.Hand.Translate(this.Lack), this.Door.Translate(this.Lack))
	this.Hand.Sub(card)
	return result;
}

func (this *Player) ChangeCard() []Card {
	defaultChange := this.defaultChangeCard()
	t := CardArrayToCards(defaultChange)
	waitingTime := 30 * time.Second
	this.Socket().Emit("change", t.ToStringArray(), waitingTime)

	c := make(chan []Card)
	var changeCards []Card
	go func() {
		this.Socket().On("changeCard", func (cards []string) {
			c<-StringArrayToCardArray(cards)
		})
	}()
	select {
	case changeCards = <-c:
	case <-time.After(waitingTime):
		changeCards = defaultChange
	}
	
	this.Hand.Sub(changeCards)
	this.game.Rooms[this.Room()].BroadcastChange(this.id)
	return defaultChange
}

func (this *Player) ChooseLack() int {
	defaultLack := 0
	waitingTime := 10 * time.Second
	this.Socket().Emit("lack", defaultLack, waitingTime)

	c := make(chan int)
	go func() {
		this.Socket().On("chooseLack", func (lack int) {
			c<-lack
		})
	}()
	select {
	case this.Lack = <-c:
	case <-time.After(waitingTime):
		this.Lack = defaultLack
	}
	return this.Lack
}

func (this *Player) ThrowCard() Card {
	defaultCard := this.Hand.At(0)
	waitingTime := 10 * time.Second
	this.Socket().Emit("throw", defaultCard.ToString(), waitingTime)

	c := make(chan Card)
	var throwCard Card
	go func() {
		this.Socket().On("throwCard", func (card string) {
			c<-StringToCard(card)
		})
	}()
	select {
	case throwCard = <-c:
	case <-time.After(waitingTime):
		throwCard = defaultCard
	}
	this.Hand.Sub(throwCard)
	this.game.Rooms[this.Room()].BroadcastThrow(this.id, throwCard)
	return throwCard
}

func (this *Player) Draw(drawCard Card) Action {
	var actions map[int][]Card
	tai     := 0
	command := 0
	this.Hand.Add(drawCard)
	this.Socket().Emit("draw", drawCard.ToString())

	if this.CheckHu(Card {-1, 0}, &tai) {
		command |= COMMAND_ZIMO
		actions[COMMAND_ZIMO] = append(actions[COMMAND_ZIMO], drawCard)
	}
	for c := 0; c < 3; c++ {
		for v :=uint(0); v < 9; v++ {
			tmpCard := Card {c, v}
			if this.Hand[c].GetIndex(v) == 4 {
				if this.CheckGon(tmpCard) {
					command |= COMMAND_ONGON
					actions[COMMAND_ONGON] = append(actions[COMMAND_ONGON], tmpCard)
				}
			} else if this.Hand[c].GetIndex(v) == 1 && this.Door[c].GetIndex(v) == 3 {
				if this.CheckGon(tmpCard) {
					command |= COMMAND_PONGON
					actions[COMMAND_PONGON] = append(actions[COMMAND_PONGON], tmpCard)
				}
			}
		}
	}

	action := Action {NONE, drawCard, 0}
	if command == NONE {
		action.Command = NONE
		action.Card    = drawCard
	} else if this.IsHu {
		if (command & COMMAND_ZIMO) != 0 {
			action.Command = COMMAND_ZIMO
			action.Card    = actions[COMMAND_ZIMO][0]
		} else if (command & COMMAND_ONGON) != 0 {
			action.Command = COMMAND_ONGON
			action.Card    = actions[COMMAND_ONGON][0]
		} else if (command & COMMAND_PONGON) != 0 {
			action.Command = COMMAND_PONGON
			action.Card    = actions[COMMAND_PONGON][0]
		}
	} else {
		action = this.OnCommand(actions, command, 0)
	}

	if (action.Command & COMMAND_ZIMO) != 0 {
		this.IsHu = true
		this.HuCards.Add(action.Card)
		this.game.Rooms[this.Room()].HuTiles.Add(action.Card)
		this.Hand.Sub(action.Card)
		action.Card.Color = -1
		Tai := tai
		if this.JustGon {
			Tai++ 
		}
		score := int(math.Pow(2, float64(Tai)))
		for i := 0; i < 4; i++ {
			if this.id != i {
				this.Credit += score
				action.Score += score
				if this.MaxTai < tai {
					this.MaxTai = tai
				}
				this.game.Rooms[this.Room()].Players[i].Credit -= score
			}
		}
	} else if (action.Command & COMMAND_ONGON) != 0 {
		this.Gon(action.Card, false)
		for i := 0; i < 4; i++ {
			if i != this.id {
				this.Credit += 2
				action.Score += 2
				this.GonRecord[i] += 2
				this.game.Rooms[this.Room()].Players[i].Credit -= 2
			}
		}
	} else if (action.Command & COMMAND_PONGON) != 0 {
		this.Gon(action.Card, true)
		for i := 0; i < 4; i++ {
			if i != this.id {
				this.Credit += 1
				action.Score += 1
				this.GonRecord[i] += 1
				this.game.Rooms[this.Room()].Players[i].Credit -= 1
			}
		}
	} else {
		if this.IsHu {
			action.Card = drawCard
		} else {
			action.Card = this.ThrowCard()
		}
		if this.Hand.Have(action.Card) {
			this.Hand.Sub(action.Card)
		}
	}
	return action
}

func (this *Player) OnCommand(cards map[int][]Card, command int, from int) Action {
	type ActionSet struct {
		key string
		value []string
	}
	var actions []ActionSet
	for key, value := range cards {
		t := CardArrayToCards(value)
		actionSet := ActionSet {strconv.Itoa(key), t.ToStringArray()}
		actions = append(actions, actionSet)
	}

	defaultCommand := Action {NONE, Card {-1, 0}, 0}
	waitingTime := 10 * time.Second
	b, _ := json.Marshal(actions)
	this.Socket().Emit("command", string(b), command, waitingTime)
	
	c := make(chan Action)
	var action Action
	go func() {
		this.Socket().On("throwCard", func (command int, card string) {
			c<-Action {command, StringToCard(card), 0}
		})
	}()
	select {
	case action = <-c:
	case <-time.After(waitingTime):
		action = defaultCommand
	}
	return action
}

func (this *Player) OnFail(command int) {
	this.Socket().Emit("fail", command)
}

func (this *Player) OnSuccess(from int, command int, card Card, score int) {
	this.Socket().Emit("success", from, command, card.ToString(), score)
	this.game.Rooms[this.Room()].BroadcastCommand(from, this.id, command, card, score)
}

func (this *Player) defaultChangeCard() []Card {
	var result []Card;
	count := 0;
	for c := 0; c < 3 && count < 3; c++ {
		if (this.Hand[c].Count()) >= 3 {
			for v := uint(0); count < 3 && v < 9; v++ {
				for n := uint(0); count < 3 && n < this.Hand[c].GetIndex(v); n++ {
					result = append(result, Card {c, v})
					count++;
				}
			}
		}
	}
	return result;
}