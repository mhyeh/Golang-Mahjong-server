package action

import (
	"tile"
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

// NewAction creates a new action
func NewAction(command int, tile tile.Tile, score int) Action {
	action := Action {command, tile, score}
	return action
}

// NewSet creates a new action set
func NewSet() Set {
	set := make(Set)
	return set
}

// Action represent a command made by player
type Action struct {
	Command int
	Tile    tile.Tile
	Score   int
}

// Set represents a set of action
type Set map[int][]tile.Tile