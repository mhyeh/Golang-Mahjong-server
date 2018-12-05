package MJCard

import (
	"strconv"
)

type Card struct {
	Color int
	Value uint
}

func (this Card) ToString() string {
	color := [3]string {"c", "d", "b"}
	return color[this.Color] + strconv.Itoa(int(this.Value + 1))
}

func StringArrayToCardArray(cards []string) []Card {
	var color map[string]int = make(map[string]int)
	color["c"] = 0
	color["d"] = 1
	color["b"] = 2

	var res []Card
	for _, card := range cards {
		r := []rune(card)
		c := string(r[0])
		v, _ := strconv.Atoi(string(r[1]))

		res = append(res, Card {color[c], uint(v - 1)})
	}
	return res
}

func StringToCard(card string) Card {
	var color map[string]int = make(map[string]int)
	color["c"] = 0
	color["d"] = 1
	color["b"] = 2

	r := []rune(card)
	c := string(r[0])
	v, _ := strconv.Atoi(string(r[1]))

	return Card {color[c], uint(v - 1)}
}
