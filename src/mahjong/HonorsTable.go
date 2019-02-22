package mahjong

// NewHonorsTable creates a new honors huTable.table
func NewHonorsTable() *HonorsTable {
	huTable     := NewHuTable()
	honorsTable := &HonorsTable{}
	honorsTable.huTable = huTable
	honorsTable.init()
	return honorsTable
}

// HonorsTable represenpts honors huTable.table
type HonorsTable struct {
	huTable *HuTable
}

func (honorsTable *HonorsTable) init() {
	honorsTable.setCanHu(5, 0, SIZE)
	honorsTable.setWindsAndDragons();
	honorsTable.setTinOnlyOne();
	honorsTable.filter();
}

func (honorsTable *HonorsTable) setCanHu(sets uint, curr uint, all uint) {
	if sets > 0 {
		for i := 0; i < 8; i++ {
			if honorsTable.huTable.Have(all, honorsTable.huTable.Melds[i]) {
				honorsTable.setCanHu(sets - 1, curr + honorsTable.huTable.Melds[i], all - honorsTable.huTable.Melds[i])
			}
		}
	} else {
		(*honorsTable.huTable.PreTable)[curr] |= Bonus | CanHu
		//有眼
		for i := 0; i < 7; i++ {
			if honorsTable.huTable.Have(all, honorsTable.huTable.AllEyes[i]) {
				(*honorsTable.huTable.PreTable)[curr + honorsTable.huTable.AllEyes[i]] |= Bonus | HasEye | CanHu
			}
		}
	}
}

func (honorsTable *HonorsTable) setWindsAndDragons() {
	pretable := *honorsTable.huTable.PreTable
	for i := uint(0); i < SIZE; i++ {
		w := []uint{ East, South, West, North, Red, Green, White }
		if pretable[i] & 1 != 0 {
			for j := 0; j < 7; j++ {
				if honorsTable.huTable.Have(i, honorsTable.huTable.Melds[j + 1]) {
					pretable[i] |= w[j]
				}
			}

			winds        := i & 4095
			dragons      := i & 2093056
			clearWinds   := ^(East | South | West | North)
			clearDragons := ^(Red | Green | White)

			if winds == 1754 || winds == 1747 || winds == 1691 || winds == 1243 {
				pretable[i] = (pretable[i] | LittleFourWinds) & clearWinds

			} else if winds == 1755 {
				pretable[i] = (pretable[i] | BigFourWinds) & clearWinds
			} else if dragons == 892928 || dragons == 864256 || dragons == 634880 {
				pretable[i] = (pretable[i] | LittleThreeDragons) & clearDragons
			} else if dragons == 897024 {
				pretable[i] = (pretable[i] | BigThreeDragons) & clearDragons
			}
		}	
	}
}

func (honorsTable *HonorsTable) setTinOnlyOne() {
	for tiles := uint(1); tiles < SIZE; tiles++ {
		if honorsTable.huTable.CanHu(tiles) {
			for j := uint(0); j < 7; j++ {
				if honorsTable.huTable.Have(tiles, honorsTable.huTable.AllSingle[j]) {
					huFeature   := (*honorsTable.huTable.PreTable)[tiles]
					zimoFeature := (*honorsTable.huTable.PreTable)[tiles]
					tilesOfTin  := tiles - honorsTable.huTable.AllSingle[j]
					tinCnt      := 0

					for k := 0; k < 9; k++ {
						temp := tilesOfTin + honorsTable.huTable.AllSingle[k];
						if temp > SIZE {
							continue
						}
						if honorsTable.huTable.CanHu(temp) {
							tinCnt++
						}
					}
					// 胡刻子，非自摸暗刻數-1，無獨聽
					if honorsTable.huTable.Have(tiles, honorsTable.huTable.Melds[j + 1]) {
						if honorsTable.huTable.CanHu(tiles - honorsTable.huTable.Melds[j + 1]) {
							
							tinCnt++
						}
					}
					// 獨聽
					if tinCnt == 1 {
						huFeature   |= TingOnlyOne
						zimoFeature |= TingOnlyOne
					}
					
					//if(temp != 0)
					honorsTable.huTable.Table[((j + 1) << 27) | (tiles - honorsTable.huTable.AllSingle[j])]             = huFeature
					honorsTable.huTable.Table[(1 << 31) | ((j + 1) << 27) | (tiles - honorsTable.huTable.AllSingle[j])] = zimoFeature
				}
			}
		}
	}
}

func (honorsTable *HonorsTable) filter() {
	for i := uint(0); i < SIZE; i++ {
		if honorsTable.huTable.CanHu(i) {
			honorsTable.huTable.Table[i] = (*honorsTable.huTable.PreTable)[i]
		}
	}
}