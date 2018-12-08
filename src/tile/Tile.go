package tile

import (
	"strconv"
)

// NewTile creates a new tile
func NewTile(color int, value uint) Tile {
	tile := Tile {Color: color, Value: value}
	return tile
}
// Tile represents a mahjong tile
type Tile struct {
	Color int
	Value uint
}

// ToString converts tile to string
func (tile Tile) ToString() string {
	color := [3]string {"c", "d", "b"}
	return color[tile.Color] + strconv.Itoa(int(tile.Value + 1))
}

// StringArrayToTileArray converts string array to tile array
func StringArrayToTileArray(tiles []string) []Tile {
	color := make(map[string]int)
	color["c"] = 0
	color["d"] = 1
	color["b"] = 2

	var res []Tile
	for _, tile := range tiles {
		r := []rune(tile)
		c := string(r[0])
		v, _ := strconv.Atoi(string(r[1]))

		res = append(res, NewTile(color[c], uint(v - 1)))
	}
	return res
}

// StringToTile converts string to tile
func StringToTile(tile string) Tile {
	color := make(map[string]int)
	color["c"] = 0
	color["d"] = 1
	color["b"] = 2

	r    := []rune(tile)
	c    := string(r[0])
	v, _ := strconv.Atoi(string(r[1]))

	return Tile {color[c], uint(v - 1)}
}

// IsValidTile checks if tile string is vaild
func IsValidTile(tile string) bool {
	r    := []rune(tile)
	c    := string(r[0])
	v, _ := strconv.Atoi(string(r[1]))
	if c != "C" && c != "b" && c != "d" {
		return false
	}
	return v > 0 && v <= 9
}