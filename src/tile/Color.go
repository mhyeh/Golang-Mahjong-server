package tile

// Color represents a mahjong color
type Color uint32

// GetIndex returns tile amount at idx
func (color Color) GetIndex(idx uint) uint {
	return uint((color >> (idx * 3)) & 7)
}

// Add add a tile to color
func (color *Color) Add(idx uint) {
	*color += (1 << (idx * 3))
}

// Sub subs a tile from color
func (color *Color) Sub(idx uint) {
	*color -= (1 << (idx * 3))
}

// Have returns if color have tile
func (color Color) Have(idx int) bool {
	return ((color >> (uint(idx) * 3)) & 7) > 0
}

// Count returns amount of color
func (color Color) Count() uint {
	amount := uint(0)
	for i := uint(0); i < 9; i++ {
		amount += color.GetIndex(i)
	}
	return amount
}
