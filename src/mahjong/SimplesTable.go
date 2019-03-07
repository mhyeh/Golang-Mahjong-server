package mahjong

// NewSimplesTable creates a new simples table
func NewSimplesTable() *SimplesTable {
	huTable      := NewHuTable()
	simplesTable := &SimplesTable{}
	simplesTable.huTable = huTable
	simplesTable.init()
	return simplesTable
}

// SimplesTable represents simples table
type SimplesTable struct {
	huTable *HuTable
}

func (simplesTable *SimplesTable) init() {
	simplesTable.setCanHu(5, 0, SIZE)
	simplesTable.setOnlyHasEye()
	simplesTable.setPingHu(5, 0, SIZE)
	simplesTable.setSameSuit()
	simplesTable.setPonPair(5, 0, 1)
	simplesTable.setConcealedPonCnt()
	simplesTable.setEyeNumber()
	simplesTable.setTinOnlyOne()
	simplesTable.filter()
}

func (simplesTable *SimplesTable) setCanHu(sets uint, curr uint, all uint) {
	if sets > 0 {
		for i := 0; i < 17; i++ {
			if simplesTable.huTable.Have(all, simplesTable.huTable.Melds[i]) {
				simplesTable.setCanHu(sets - 1, curr + simplesTable.huTable.Melds[i], all - simplesTable.huTable.Melds[i])
			}
		}
	} else {
		(*simplesTable.huTable.PreTable)[curr] |= CanHu
		//有眼
		for i := 0; i < 9; i++ {
			if simplesTable.huTable.Have(all, simplesTable.huTable.AllEyes[i]) {
				(*simplesTable.huTable.PreTable)[curr + simplesTable.huTable.AllEyes[i]] |= HasEye | CanHu
			}
		}
	}
}

func (simplesTable *SimplesTable) setOnlyHasEye() {
	(*simplesTable.huTable.PreTable)[0] |= OnlyHasEye
	for i := 0; i < 9; i++ {
		(*simplesTable.huTable.PreTable)[simplesTable.huTable.AllEyes[i]] |= OnlyHasEye
	}
}

func (simplesTable *SimplesTable) setPingHu(sets uint, curr uint, all uint) {
	if sets > 0 {
		for i := 0; i < 17; i++ {
			if simplesTable.huTable.Have(all, simplesTable.huTable.Melds[i]) && (i == 0 || i > 9) {
				simplesTable.setPingHu(sets - 1, curr + simplesTable.huTable.Melds[i], all - simplesTable.huTable.Melds[i])
			}
		}
	} else {
		(*simplesTable.huTable.PreTable)[curr] |= PingHu
		//有眼
		for i := 0; i < 9; i++ {
			if simplesTable.huTable.Have(all, simplesTable.huTable.AllEyes[i]) {
				(*simplesTable.huTable.PreTable)[curr + simplesTable.huTable.AllEyes[i]] |= PingHu
			}
		}
	}
}

func (simplesTable *SimplesTable) setSameSuit() {
	(*simplesTable.huTable.PreTable)[0] |= SameSuit
}

func (simplesTable *SimplesTable) setPonPair(sets uint, curr uint, c uint) {
	(*simplesTable.huTable.PreTable)[curr] |= PonPair
	for i := 0; i < 9; i++ {
		if simplesTable.huTable.Have(SIZE - curr, simplesTable.huTable.AllEyes[i]) {
			(*simplesTable.huTable.PreTable)[curr + simplesTable.huTable.AllEyes[i]] |= PonPair
		}

	}

	if sets > 0 {
		for i := c; i <= 9; i++ {
			simplesTable.setPonPair(sets - 1, curr + simplesTable.huTable.Melds[i], i + 1)
		}
	}
}

func (simplesTable *SimplesTable) setConcealedPonCnt() {
	for i := uint(0); i < SIZE; i++ {
		ponCnt := uint(0)
		if simplesTable.huTable.CanHu(i) {
			for j := 0; j < 9; j++ {
				if simplesTable.huTable.Have(i, simplesTable.huTable.Melds[j + 1]) {
					if simplesTable.huTable.CanHu(i - simplesTable.huTable.Melds[j + 1]) {
						ponCnt++
					}
				}
			}
			(*simplesTable.huTable.PreTable)[i] |= IF(ponCnt < 7, ponCnt, uint(7)).(uint) * ConcealedPonCnt / 7
		}
	}
}

func (simplesTable *SimplesTable) setEyeNumber() {
	for i := uint(0); i < SIZE; i++ {
		pos := 0
		if simplesTable.huTable.CanHu(i) {
			for j := 0; j < 9; j++ {
				if simplesTable.huTable.Have(i, simplesTable.huTable.AllEyes[j]) {
					if simplesTable.huTable.CanHu(i - simplesTable.huTable.AllEyes[j]) {
						pos = j + 1
					}
				}
			}
			(*simplesTable.huTable.PreTable)[i] |= uint(pos << 12)
		}
	}
}

func (simplesTable *SimplesTable) setTinOnlyOne() {
	for tiles := uint(1); tiles < SIZE; tiles++ {
		if simplesTable.huTable.CanHu(tiles) {
			for j := uint(0); j < 9; j++ {
				if simplesTable.huTable.Have(tiles, simplesTable.huTable.AllSingle[j]) {
					huFeature   := (*simplesTable.huTable.PreTable)[tiles]
					zimoFeature := (*simplesTable.huTable.PreTable)[tiles]
					readyHand   := tiles - simplesTable.huTable.AllSingle[j]
					tinCnt      := 0

					for k := 0; k < 9; k++ {
						temp := readyHand + simplesTable.huTable.AllSingle[k]
						if temp > SIZE {
							continue
						}
						if simplesTable.huTable.CanHu(temp) {
							tinCnt++
						}
					}
					// 胡刻子，非自摸暗刻數-1，無獨聽
					if simplesTable.huTable.Have(tiles, simplesTable.huTable.Melds[j + 1]) {
						if simplesTable.huTable.CanHu(tiles - simplesTable.huTable.Melds[j + 1]) {
							huFeature -= 1 << 9
							tinCnt++
						}
					}
					// 獨聽
					if tinCnt == 1 {
						huFeature   |= TingOnlyOne
						zimoFeature |= TingOnlyOne
					}
					// 平胡
					if (((*simplesTable.huTable.PreTable)[tiles] >> 12) & 15) == j + 1 || tinCnt == 1 {
						huFeature &= ^PingHu
					}

					simplesTable.huTable.Table[((j + 1) << 27) | (tiles - simplesTable.huTable.AllSingle[j])]             = huFeature;
					simplesTable.huTable.Table[(1 << 31) | ((j + 1) << 27) | (tiles - simplesTable.huTable.AllSingle[j])] = zimoFeature;
				}
			}
		}
	}
}

func (simplesTable *SimplesTable) filter() {
	for i := uint(0); i < SIZE; i++ {
		if simplesTable.huTable.CanHu(i) {
			simplesTable.huTable.Table[i] = (*simplesTable.huTable.PreTable)[i]
		}
	}
}
