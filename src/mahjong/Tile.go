package mahjong

import (
	"strconv"
)

// NewTile creates a new tile
func NewTile(suit int, value uint) Tile {
	tile := Tile {Suit: suit, Value: value}
	return tile
}
// Tile represents a mahjong tile
type Tile struct {
	Suit  int
	Value uint
}

// ToString converts tile to string
func (tile Tile) ToString() string {
	suit := [3]string {"c", "d", "b"}
	return suit[tile.Suit] + strconv.Itoa(int(tile.Value + 1))
}

// StringArrayToTileArray converts string array to tile array
func StringArrayToTileArray(tiles []string) []Tile {
	suit := make(map[string]int)
	suit["c"] = 0
	suit["d"] = 1
	suit["b"] = 2

	var res []Tile
	for _, tile := range tiles {
		r    := []rune(tile)
		s    := string(r[0])
		v, _ := strconv.Atoi(string(r[1]))

		res = append(res, NewTile(suit[s], uint(v - 1)))
	}
	return res
}

// StringToTile converts string to tile
func StringToTile(tile string) Tile {
	suit := make(map[string]int)
	suit["c"] = 0
	suit["d"] = 1
	suit["b"] = 2

	r    := []rune(tile)
	s    := string(r[0])
	v, _ := strconv.Atoi(string(r[1]))

	return Tile {suit[s], uint(v - 1)}
}

// IsValidTile checks if tile string is vaild
func IsValidTile(tile string) bool {
	r    := []rune(tile)
	s    := string(r[0])
	v, _ := strconv.Atoi(string(r[1]))
	if s != "C" && s != "b" && s != "d" {
		return false
	}
	return v > 0 && v <= 9
}