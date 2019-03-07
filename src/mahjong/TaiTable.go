package mahjong

// TaiData represents tai data
type TaiData struct {
	Tai     int
	Message string
}

// NewSimpleTaiTable creates a new simples tai table
func NewSimpleTaiTable() *SimplesTaiTable {
	taiTable := &SimplesTaiTable{}
	taiTable.TaiTable = make(map[uint]TaiData)
	taiTable.initTable()
	return taiTable
}

// SimplesTaiTable represents simples tai table
type SimplesTaiTable struct {
	TaiTable map[uint]TaiData
}

func (simplesTaiTable *SimplesTaiTable) initTable() {
	for i := uint(0); i < 8191; i++ {
		if (i & CanHu) != 0 && (i & HasEye) != 0 {
			tai := 0
			str := ""
			// 自摸
			if (i & Zimo) != 0 {
				tai++
				str += "自摸 "
			}
			// 全求人(自摸不計) or 門清
			if (i & OnlyHasEye) != 0 && (i & Zimo) == 0 {
				tai += 2
				str += "全求人 "
			} else if (i & Clean) != 0 {
				tai++
				str += "門清 "
				//門清自摸加一台
				tai += IF((i & Zimo) != 0, 1, 0).(int)
			}
			if (i & OnlyHasEye) != 0 && (i & Clean) != 0 {
				continue
			}

			// 獨聽(全求人不計)
			if (i & TingOnlyOne) != 0 && !((i & OnlyHasEye) != 0 && (i & Zimo) == 0) {
				tai++
				str += "獨聽 "
			}

			// 湊一色 or 清一色
			if (i & MixSuit) != 0 {
				tai += 4
				str += "湊一色 "
			} else if (i & SameSuit) != 0 {
				tai += 8
				str += "清一色 "
			} else if ( i& MixSuit) != 0 && (i & SameSuit) != 0 {
				continue
			}
			// 碰碰胡
			if (i & PonPair) != 0 {
				tai += 4
				str += "碰碰胡 "
			}

			// 三暗刻 四暗刻 五暗刻
			cnt := (i >> 9) & 7
			if cnt > 6 {
				continue
			}
			if (i & OnlyHasEye) == 0 {
				//五暗刻必門清和碰碰胡
				if (i & Clean) != 0 && (i & PonPair) != 0 {
					if cnt == 5 {
						tai += 8
						str += "五暗刻 "
					} else if cnt == 4 {
						tai += 5
						str += "四暗刻 "
					}
				} else {
					if cnt == 4 {
						tai += 5
						str += "四暗刻 "
					} else if cnt == 3 {
						tai += 2
						str += "三暗刻 "
					}
				}
			}
			
			simplesTaiTable.TaiTable[i] = TaiData{ tai, str }
		}
	}
	// 平胡
	simplesTaiTable.TaiTable[PingHu | HasEye | CanHu] = TaiData{ 2, "平胡 " }
	// 門清+平胡
	simplesTaiTable.TaiTable[Clean | PingHu | HasEye | CanHu] = TaiData{ 3, "門清 平胡 " }
}

// Get gets simplesTaiTable data
func (simplesTaiTable *SimplesTaiTable) Get(idx uint) TaiData {
	it, ok := simplesTaiTable.TaiTable[idx]
	if ok {
		return it
	}
	return TaiData{ -1, "" }
}

// NewHonorsTaiTable creates a new simples tai table
func NewHonorsTaiTable() *HonorsTaiTable {
	taiTable := &HonorsTaiTable{}
	taiTable.TaiTable = make(map[uint]TaiData)
	taiTable.initTable()
	return taiTable
}

// HonorsTaiTable represents honors tai table
type HonorsTaiTable struct {
	TaiTable map[uint]TaiData
}

func (honorsTaiTable *HonorsTaiTable) initTable() {
	for i := uint(0); i < (1 << 12); i++ {
		tai := 0
		str := ""
		if (i & (LittleFourWinds >> 4)) != 0 {
			tai += 8
			str += "小四喜 "
		} else if (i & (BigFourWinds >> 4)) != 0 {
			tai += 16
			str += "大四喜 "
		} else if (i & (LittleThreeDragons >> 4)) != 0 {
			tai += 4
			str += "小三元 "
		} else if (i & (BigThreeDragons >> 4)) != 0 {
			tai += 8
			str += "大三元 "
		}
		
		honorCnt  := 0
		honorsTai := IF((i & Bonus) != 0, 2, 1).(int)

		if (i & (East >> 4)) != 0 {
			tai += honorsTai
			str += IF((i& Bonus) != 0, "東風東 ", "東風 ").(string)
			honorCnt++
		}
		if (i & (South >> 4)) != 0 {
			tai += honorsTai
			str += IF((i& Bonus) != 0, "南風南 ", "南風 ").(string)
			honorCnt++
		}
		if (i & (West >> 4)) != 0 {
			tai += honorsTai
			str += IF((i& Bonus) != 0, "西風西 ", "西風 ").(string)
			honorCnt++
		}
		if (i & (North >> 4)) != 0 {
			tai += honorsTai
			str += IF((i& Bonus) != 0, "北風北 ", "北風 ").(string)
			honorCnt++
		}
		if (i & (Red >> 4)) != 0 {
			tai++
			str += "紅中 "
			honorCnt++
		}
		if (i & (Green >> 4)) != 0 {
			tai++
			str += "青發 "
			honorCnt++
		}
		if (i & (White >> 4)) != 0 {
			tai++
			str += "白板 "
			honorCnt++
		}
		if honorCnt > 6 {
			continue
		}
		
		honorsTaiTable.TaiTable[i] = TaiData{ tai, str }
	}
}

// Get gets honorsTaiTable data
func (honorsTaiTable *HonorsTaiTable) Get(idx uint) TaiData {
	it, ok := honorsTaiTable.TaiTable[idx]
	if ok {
		return it
	}
	return TaiData{ -1, "" }
}

// NewTaiTable creates a new tai table
func NewTaiTable() *TaiTable {
	sTable := NewSimpleTaiTable()
	hTable := NewHonorsTaiTable()

	taiTable := &TaiTable{ sTable, hTable }
	return taiTable
}

// TaiTable represents tai table
type TaiTable struct {
	sTable *SimplesTaiTable
	hTable *HonorsTaiTable
}

// Get gets taiTable data
func (taiTable *TaiTable) Get(idx uint) TaiData {
	sData := taiTable.sTable.Get(idx & 65535)
	hData := taiTable.hTable.Get(idx >> 20)
	tai   := IF(sData.Tai + hData.Tai == -2, -1, sData.Tai + hData.Tai).(int)

	return TaiData{ tai, sData.Message + hData.Message }
}