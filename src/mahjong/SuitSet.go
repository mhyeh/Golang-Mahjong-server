package mahjong

import (
	"strconv"
	"math/rand"
)

// SuitSet represents a set of suit
type SuitSet [5]Suit

// SuitTileCount represents each suit's tile count
var SuitTileCount = [5]uint{ 9, 9, 9, 7, 8 }

// NewSuitSet creates a new suit suitSet
func NewSuitSet(full bool) SuitSet {
	var suitSet SuitSet
	t, _ := strconv.ParseUint("100100100100100100100100100", 2, 32)
	for s := 0; s < 3; s++ {
		suitSet[s] = IF(full, Suit(t), Suit(0)).(Suit)
	}

	t, _ = strconv.ParseUint("100100100100100100100", 2, 32) // 字、風
	suitSet[3] = IF(full, Suit(t), Suit(0)).(Suit)

	t, _ = strconv.ParseUint("001001001001001001001001", 2, 32)
	suitSet[4] = IF(full, Suit(t), Suit(0)).(Suit)

	return suitSet
}

// ArrayToSuitSet converts tile array to suit set
func ArrayToSuitSet(tile []Tile) SuitSet {
	suitSet := NewSuitSet(false)
	suitSet.Add(tile)
	return suitSet
}

// IsEmpty returns if suit set is empty
func (suitSet SuitSet) IsEmpty() bool {
	count := uint32(0)
	for i := 0; i < 5; i++ {
		count += uint32(suitSet[i])
	}
	return count == 0
}

// IsContainColor return if suit set contain color
func (suitSet SuitSet) IsContainColor(color int) bool {
	return uint32(suitSet[color]) > 0
}

// Have returns if suit set has tlie
func (suitSet SuitSet) Have(tile Tile) bool {
	return suitSet[tile.Suit].GetIndex(tile.Value) > 0
}

// At returns idx th tile in suit set
func (suitSet SuitSet) At(idx int) Tile {
	amount := 0
	for s := 0; s < 5; s++ {
		for v := uint(0); v < SuitTileCount[s]; v++ {
			amount += int(suitSet[s].GetIndex(v))
			if amount > idx {
				return NewTile(s, v)
			}
		}
	}
	return NewTile(-1, 0)
}

// Count returns amount of suit set
func (suitSet SuitSet) Count() uint {
	amount := uint(0)
	for s := 0; s < 5; s++ {
		amount += suitSet[s].Count()
	}
	return amount
}

// Draw draws a tile from suit set
func (suitSet *SuitSet) Draw() Tile {
	amount := int32(suitSet.Count())
	tile   := suitSet.At(int(rand.Int31n(amount)))
	suitSet.Sub(tile)
	return tile
}

// ToStringArray converts suit set to string array
func (suitSet SuitSet) ToStringArray() []string {
	var result []string
	for s := 0; s < 5; s++ {
		for v := uint(0); v < SuitTileCount[s]; v++ {
			for n := uint(0); n < suitSet[s].GetIndex(v); n++ {
				result = append(result, suitStr[s] + strconv.Itoa(int(v + 1)))
			}
		}
	}
	return result
}

// Add adds a tile or a suit set to a suit set
func (suitSet *SuitSet) Add(input interface{}) {
	switch input.(type) {
	case []Tile:
		for _, tile := range input.([]Tile) {
			if suitSet[tile.Suit].GetIndex(tile.Value) < 4 {
				suitSet.Add(tile)
			}
		}
	case Tile:
		tile := input.(Tile)
		if suitSet[tile.Suit].GetIndex(tile.Value) < 4 {
			suitSet[tile.Suit].Add(tile.Value)
		}
	}
}

// Sub subs a tile or a suit set from a suit set
func (suitSet *SuitSet) Sub(input interface{}) {
	switch input.(type) {
	case []Tile:
		for _, tile := range input.([]Tile) {
			if suitSet[tile.Suit].GetIndex(tile.Value) > 0 {
				suitSet.Sub(tile)
			}
		}
	case Tile:
		tile := input.(Tile)
		if suitSet[tile.Suit].GetIndex(tile.Value) > 0 {
			suitSet[tile.Suit].Sub(tile.Value)
		}
	}
}