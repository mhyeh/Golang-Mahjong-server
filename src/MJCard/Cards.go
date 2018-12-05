
package MJCard

import (
	"strconv"
	"math/rand"
)

type Cards [3]Color

func NewCards(full bool) Cards {
	var cards Cards
	for i := 0; i < 3; i++ {
		if full {
			t, _ := strconv.ParseUint("100100100100100100100100100", 2, 32)
			cards[i] = Color(t)
		} else {
			cards[i] = Color(0)
		}
	}
	return cards
}

func CardArrayToCards(cards []Card) Cards {
	res := NewCards(false)
	res.Add(cards)
	return res
}

func (this Cards) IsEmpty() bool {
	return (this[0] + this[1] + this[2]) == 0
}

func (this Cards) ContainColor(color int) bool {
	return this[color].Count() > 0
}

func (this Cards) Have(card Card) bool {
	return this[card.Color].GetIndex(card.Value) > 0
}

func (this Cards) At(idx int) Card {
	count := 0
	for c := 0; c < 3; c++ {
		for v := uint(0); v < 9; v++ {
			count += int(this[c].GetIndex(v))
			if count > idx {
				return Card {c, v}
			}
		}
	}
	return Card {-1, 0}
}

func (this Cards) Count() uint {
	result := uint(0)
	for i := 0; i < 3; i++ {
		result += this[i].Count()
	}
	return result
}

func (this *Cards) Draw() Card {
	len := this.Count()
	result := this.At(int(rand.Int31n(int32(len))))
	this.Sub(result)
	return result
}

func (this Cards) Translate(lack int) uint64 {
	first := true
	result := uint64(0)
	for i := 0; i < 3; i++ {
		if i != lack {
			result |= uint64(this[i])
			if (first) {
				result <<= 27
				first = false
			}
		}
	}
	return result
}

func (this Cards) ToStringArray() []string {
	var result []string
	colorArr := [3]string {"c", "d", "b"}
	for c := 0; c < 3; c++ {
		for v := uint(0); v < 9; v++ {
			for n := uint(0); n < this[c].GetIndex(v); n++ {
				result = append(result, colorArr[c] + strconv.Itoa(int(v + 1)))
			}
		}
	}
	return result
}

func (this *Cards) Add(input interface{}) {
	switch input.(type) {
	case []Card:
		for _, card := range input.([]Card) {
			if this[card.Color].GetIndex(card.Value) < 4 {
				this.Add(card)
			}
		}
	case Card:
		card := input.(Card)
		if this[card.Color].GetIndex(card.Value) < 4 {
			this[card.Color].Add(card.Value)
		}
	}
}

func (this *Cards) Sub(input interface{}) {
	switch input.(type) {
	case []Card:
		for _, card := range input.([]Card) {
			if this[card.Color].GetIndex(card.Value) > 0 {
				this.Sub(card)
			}
		}
	case Card:
		card := input.(Card)
		if this[card.Color].GetIndex(card.Value) > 0 {
			this[card.Color].Sub(card.Value)
		}
	}
}