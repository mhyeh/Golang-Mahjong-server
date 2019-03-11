package mahjong

// Wind represents wind
type Wind byte

// HuInfo represents information of hu
type HuInfo struct {
	Hand    SuitSet
	Door    SuitSet
	Winds   Wind
	IsZimo  uint
	IsClean uint
	AllPon  uint
	AllChow uint
	NoBonus uint

	CertainTile Tile
}

// Feature
const (
	CanHu              = uint(1)
	HasEye             = uint(2)
	OnlyHasEye         = uint(4)
	TingOnlyOne        = uint(8)
	Clean              = uint(16)
	PingHu             = uint(32)
	MixSuit            = uint(64)
	SameSuit           = uint(128)
	PonPair            = uint(256)
	ConcealedPonCnt    = uint(3584)
	Zimo               = uint(4096)

	East               = uint(16)
	South              = uint(32)
	West               = uint(64)
	North              = uint(128)
	Bonus              = uint(256)
	Red                = uint(512)
	Green              = uint(1024)
	White              = uint(2048)
	LittleFourWinds    = uint(4096)
	BigFourWinds       = uint(8192)
	LittleThreeDragons = uint(16384)
	BigThreeDragons    = uint(32768)

	EyeNumber = uint(61440)
)

// NewScoreCalc creates a new ccore calculator
func NewScoreCalc() ScoreCalc {
	taiTable     := NewTaiTable()
	simplesTable := NewSimplesTable()
	honorsTable  := NewHonorsTable()
	scoreCalc    := ScoreCalc{}
	scoreCalc.TaiTable     = taiTable
	scoreCalc.SimplesTable = simplesTable
	scoreCalc.HonorsTable  = honorsTable

	return scoreCalc
}

// ScoreCalc represents score calculator
type ScoreCalc struct {
	TaiTable     *TaiTable
	SimplesTable *SimplesTable
	HonorsTable  *HonorsTable
}

var scoreCalc = NewScoreCalc()

// CountTai cal tai
func (scoreCalc ScoreCalc) CountTai(info HuInfo) TaiData {
	simples := scoreCalc.SimplesTable
	honors  := scoreCalc.HonorsTable
	hands   := info.Hand
	doors   := info.Door

	hands[info.CertainTile.Suit] |= Suit(info.IsZimo << 31 | (info.CertainTile.Value + 1) << 27)
	f := uint(0)
	c := simples.huTable.Get(uint(hands[0])) // 萬
	d := simples.huTable.Get(uint(hands[1])) // 筒
	b := simples.huTable.Get(uint(hands[2])) // 條
	h := simples.huTable.Get(uint(hands[3])) // 字

	//   bit3   bit2  bit1   bit0
	// |  字  |  條  |  筒  |  萬  |
	hasZi     := int(IF((hands[3] + doors[3]) != 0, 1, 0).(int))
	hasBamboo := int(IF((hands[2] + doors[2]) != 0, 1, 0).(int))
	hasDot    := int(IF((hands[1] + doors[1]) != 0, 1, 0).(int))
	hasChar   := int(IF((hands[0] + doors[0]) != 0, 1, 0).(int))
	suits     := (hasZi << 3) | (hasBamboo << 2) | (hasDot << 1) | hasChar

	honorsFeature := honors.huTable.Get(uint(hands[3]) + scoreCalc.kongToPong(uint(doors[3]))) // 門前+手上字牌(槓轉為碰)

	// 無法胡牌
	if ((c & d & b & honorsFeature) & CanHu) == 0 || ((c & HasEye) + (d & HasEye) + (b & HasEye) + (honorsFeature & HasEye)) != HasEye {
		return TaiData{ -1, "" }
	}

	onlyHasEye      := (c & d & b & h) & OnlyHasEye
	tingOnlyOne     := (c | d | b | honorsFeature) & TingOnlyOne
	concealedPonCnt := ((c & ConcealedPonCnt) + (d & ConcealedPonCnt) + (b & ConcealedPonCnt) + (h & ConcealedPonCnt))
	f |= concealedPonCnt | tingOnlyOne | onlyHasEye | HasEye | CanHu

	//平胡(門前全為吃牌且無花)
	if info.AllChow != 0 && info.NoBonus != 0 && int(hands[3] + doors[3]) == 0 {
		f |= (c & d & b & h) & PingHu
	}

	if suits == 9 || suits == 10 || suits == 12 {
		f |= MixSuit
	} else if suits == 1 || suits == 2 || suits == 4 || suits == 8 {
		f |= SameSuit
	}

	// 碰碰胡
	if info.AllPon != 0 {
		f |= (c & d & b) & PonPair
	}

	// 門清
	if (doors[0] + doors[1] + doors[2] + doors[3]) == 0 {
		f |= Clean
	}
	// 自摸
	f |= IF(info.IsZimo != 0, Zimo, uint(0)).(uint)

	// 判斷門風和圈風
	honorsFeature &= ((uint(info.Winds) | ^uint(31)) << 4)
	return scoreCalc.TaiTable.Get((honorsFeature << 16) | f)
}

func (scoreCalc ScoreCalc) kongToPong(t uint) uint {
	for i := uint(0); i < 9; i++ {
		cnt := (t >> (i * 3)) & 7
		if cnt == 4 {
			t -= 1 << (i * 3)
		}
	}
	return t
}
