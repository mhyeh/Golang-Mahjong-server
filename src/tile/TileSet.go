
package tile

import (
	"strconv"
	"math/rand"
)

// Set represents a set of tile
type Set [3]Color

// NewSet creates a new tile set
func NewSet(full bool) Set {
	var set Set
	for i := 0; i < 3; i++ {
		if full {
			t, _ := strconv.ParseUint("100100100100100100100100100", 2, 32)
			set[i] = Color(t)
		} else {
			set[i] = Color(0)
		}
	}
	return set
}

// ArrayToSet converts tile array to tile set
func ArrayToSet(tile []Tile) Set {
	set := NewSet(false)
	set.Add(tile)
	return set
}

// IsEmpty returns if tile set is empty
func (set Set) IsEmpty() bool {
	return (set[0] + set[1] + set[2]) == 0
}

// IsContainColor return if tile set contain color
func (set Set) IsContainColor(color int) bool {
	return set[color].Count() > 0
}

// Have returns if tile set has tlie
func (set Set) Have(tile Tile) bool {
	return set[tile.Color].GetIndex(tile.Value) > 0
}

// At returns idx th tile in tile set
func (set Set) At(idx int) Tile {
	amount := 0
	for c := 0; c < 3; c++ {
		for v := uint(0); v < 9; v++ {
			amount += int(set[c].GetIndex(v))
			if amount > idx {
				return NewTile(c, v)
			}
		}
	}
	return NewTile(-1, 0)
}

// Count returns amount of tile set
func (set Set) Count() uint {
	amount := uint(0)
	for i := 0; i < 3; i++ {
		amount += set[i].Count()
	}
	return amount
}

// Draw draws a tile from tile set
func (set *Set) Draw() Tile {
	amount := int32(set.Count())
	tile   := set.At(int(rand.Int31n(amount)))
	set.Sub(tile)
	return tile
}

// Translate translates tile set to uint64
func (set Set) Translate(lack int) uint64 {
	first  := true
	result := uint64(0)
	for i := 0; i < 3; i++ {
		if i != lack {
			result |= uint64(set[i])
			if (first) {
				result <<= 27
				first = false
			}
		}
	}
	return result
}

// ToStringArray converts tile set to string array
func (set Set) ToStringArray() []string {
	var result []string
	colorChar := [3]string {"c", "d", "b"}
	for c := 0; c < 3; c++ {
		for v := uint(0); v < 9; v++ {
			for n := uint(0); n < set[c].GetIndex(v); n++ {
				result = append(result, colorChar[c] + strconv.Itoa(int(v + 1)))
			}
		}
	}
	return result
}

// Add adds a tile or a tile set to a tile set
func (set *Set) Add(input interface{}) {
	switch input.(type) {
	case []Tile:
		for _, tile := range input.([]Tile) {
			if set[tile.Color].GetIndex(tile.Value) < 4 {
				set.Add(tile)
			}
		}
	case Tile:
		tile := input.(Tile)
		if set[tile.Color].GetIndex(tile.Value) < 4 {
			set[tile.Color].Add(tile.Value)
		}
	}
}

// Sub subs a tile or a tile set from a tile set
func (set *Set) Sub(input interface{}) {
	switch input.(type) {
	case []Tile:
		for _, tile := range input.([]Tile) {
			if set[tile.Color].GetIndex(tile.Value) > 0 {
				set.Sub(tile)
			}
		}
	case Tile:
		tile := input.(Tile)
		if set[tile.Color].GetIndex(tile.Value) > 0 {
			set[tile.Color].Sub(tile.Value)
		}
	}
}