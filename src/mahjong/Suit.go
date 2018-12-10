package mahjong

// Suit represents a mahjong suit
type Suit uint32

// GetIndex returns tile amount at idx
func (suit Suit) GetIndex(idx uint) uint {
	return uint((suit >> (idx * 3)) & 7)
}

// Add add a tile to suit
func (suit *Suit) Add(idx uint) {
	*suit += (1 << (idx * 3))
}

// Sub subs a tile from suit
func (suit *Suit) Sub(idx uint) {
	*suit -= (1 << (idx * 3))
}

// Have returns if suit have tile
func (suit Suit) Have(idx int) bool {
	return ((suit >> (uint(idx) * 3)) & 7) > 0
}

// Count returns amount of suit
func (suit Suit) Count() uint {
	amount := uint(0)
	for i := uint(0); i < 9; i++ {
		amount += suit.GetIndex(i)
	}
	return amount
}
