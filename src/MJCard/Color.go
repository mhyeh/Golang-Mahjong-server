package MJCard

type Color uint32

func (this Color) GetIndex(idx uint) uint {
	return uint((this >> (idx * 3)) & 7)
}

func (this *Color) Add(idx uint) {
	*this += (1 << (idx * 3))
}

func (this *Color) Sub(idx uint) {
	*this -= (1 << (idx * 3))
}

func (this Color) Have(idx int) bool {
	return ((this >> (uint(idx) * 3)) & 7) > 0
}

func (this Color) Count() uint {
	result := uint(0)
	for i := uint(0); i < 9; i++ {
		result += uint(((this >> (i * 3)) & 7))
	}
	return result
}
