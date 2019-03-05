package mahjong

// SIZE represents table size
const SIZE = 76695844

// NewHuTable creates a new hu table
func NewHuTable() *HuTable {
	huTable := &HuTable{}
	arr := make([]uint, SIZE)
	huTable.PreTable = &arr
	huTable.Table = make(map[uint]uint)
	huTable.Melds = append(huTable.Melds, 0)
	for i := 0; i < 9; i++ {
		huTable.Melds = append(huTable.Melds, 3 << uint(i * 3))
	}
	for i := 0; i < 7; i++ {
		huTable.Melds = append(huTable.Melds, 73 << uint(i * 3))
	}
	for i := 0; i < 9; i++ {
		huTable.AllEyes = append(huTable.AllEyes, 2 << uint(i * 3))
	}
	for i := 0; i < 9; i++ {
		huTable.AllSingle = append(huTable.AllSingle, 1 << uint(i * 3))
	}

	return huTable
}

// HuTable represents hu table
type HuTable struct {
	Melds     []uint
	AllEyes   []uint
	AllSingle []uint
	PreTable  *[]uint
	Table     map[uint]uint
}

// Get gets table value
func (huTable HuTable) Get(idx uint) uint {
	it, ok := huTable.Table[idx]
	if ok {
		return it
	}
	return 0
}

// Have checks if hand has feature
func (huTable HuTable) Have(now uint, check uint) bool {
	for i := uint(0); i < 9; i++ {
		if ((now >> (i * 3)) & 7) < ((check >> (i * 3)) & 7) {
			return false;
		}
	}
	return true;
}

// CanHu check if hand can hu
func (huTable HuTable) CanHu(t uint) bool {
	return ((*huTable.PreTable)[t & 134217727] & 1) != 0
}