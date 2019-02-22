package mahjong

import (
	"strings"

	"github.com/googollee/go-socket.io"
)

// NewPlayer creates a new player
func NewPlayer(room *Room, id int, uuid string) *Player {
	return &Player{ room: room, ID: id, UUID: uuid }
}

// NewScoreRecord creates a new scoreRecord
func NewScoreRecord(message string, direct string, player string, tile string, score int) ScoreRecord {
	if direct != "" {
		return ScoreRecord{ Message: strings.Join([]string{ message, direct, player }, " "), Tile: tile, Score: score }
	}
	return ScoreRecord{ Message: message, Tile: tile, Score: score }

}

// ScoreRecord represents the record of score
type ScoreRecord struct {
	Message string
	Tile    string
	Score   int
}

// Player represents a player in mahjong
type Player struct {
	Hand         SuitSet
	EatTiles     SuitSet
	PonTiles     SuitSet
	GonTiles     SuitSet
	OngonTiles   SuitSet
	Flowers      SuitSet
	DiscardTiles SuitSet
	ScoreLog     []ScoreRecord
	Credit       int
	JustGon      bool
	FirstDraw    bool
	ID           int
	UUID         string
	room         *Room
}

// Name returns the player's name
func (player Player) Name() string {
	index := FindPlayerByUUID(player.UUID)
	return PlayerList[index].Name
}

// Room returns the player's room
func (player Player) Room() string {
	index := FindPlayerByUUID(player.UUID)
	return PlayerList[index].Room
}

// Socket returns the player's socket
func (player Player) Socket() socketio.Socket {
	index := FindPlayerByUUID(player.UUID)
	return *PlayerList[index].Socket
}

// Init inits the player's state
func (player *Player) Init() {
	index := FindPlayerByUUID(player.UUID)
	PlayerList[index].State = PLAYING
	for i := 0; i < 5; i++ {
		player.Hand[i]         = 0
		player.DiscardTiles[i] = 0
		player.Flowers[i]      = 0
		player.EatTiles[i]     = 0
		player.PonTiles[i]     = 0
		player.GonTiles[i]     = 0
		player.OngonTiles[i]   = 0
	}

	player.Credit  = 0
	player.JustGon = false
}


// CheckPon checks if the player can pon
func (player *Player) CheckPon(tile Tile) bool {
	if tile.Suit < 0 || tile.Suit > 3 {
		return false
	}
	return player.Hand[tile.Suit].GetIndex(tile.Value) >= 2
}

// CheckEat checks if the player can eat
func (player *Player) CheckEat(tile Tile) bool {
	if tile.Suit < 0 || tile.Suit > 2 {
		return false
	}
	player.Hand.Add(tile)
	flag := false
	for i := int(tile.Value) - 2; i <= int(tile.Value); i++ {
		flag = true
		for j := 0; j < 3; j++ {
			if !(i + j > 0) || !(i + j <= 9) || !player.Hand.Have(NewTile(tile.Suit, uint(i + j))) {
				flag = false
				i   += j
				break
			}
		}
		if flag {
			break
		}
	}
	player.Hand.Sub(tile)
	return flag
}

// CheckHu checks if the player can hu
func (player *Player) CheckHu(tile Tile, isZimo uint, tai *TaiData) bool {
	*tai  = TaiData{ -1, "" }
	info := player.room.Info
	for i := 0; i < 5; i++ {
		info.Door[i] = 0
	}
	info.CertainTile = tile

	info.Hand = player.Hand

	eatCount := int(player.EatTiles.Count())
	for i := 0; i < eatCount; i++ {
		firstTile := player.EatTiles.At(i)
		for j := uint(0); j < 3; j++ {
			info.Door.Add(NewTile(firstTile.Suit, firstTile.Value + j))
		}
	}

	ponCount := int(player.PonTiles.Count())
	for i := 0; i < ponCount; i++ {
		tile := player.PonTiles.At(i)
		for j := uint(0); j < 3; j++ {
			info.Door.Add(tile)
		}
	}

	gonCount := int(player.GonTiles.Count())
	for i := 0; i < gonCount; i++ {
		tile := player.GonTiles.At(i)
		for j := uint(0); j < 4; j++ {
			info.Door.Add(tile)
		}
	}

	ongonCount := int(player.OngonTiles.Count())
	for i := 0; i < ongonCount; i++ {
		tile := player.OngonTiles.At(i)
		for j := uint(0); j < 4; j++ {
			info.Door.Add(tile)
		}
	}

	info.AllChow = uint(IF(player.PonTiles.Count() == 0 && player.GonTiles.Count() == 0 && player.OngonTiles.Count() == 0, 1, 0).(int))
	info.AllPon  = uint(IF(player.EatTiles.Count() == 0, 1, 0).(int))
	info.IsClean = uint(IF(player.EatTiles.Count() == 0 && player.PonTiles.Count() == 0 && player.GonTiles.Count() == 0, 1, 0).(int))
	info.NoBonus = uint(IF(player.Flowers.Count()  == 0, 1, 0).(int))
	info.IsZimo  = isZimo

	*tai = scoreCalc.CountTai(info)

	return (*tai).Tai > 0
}

// Hu hus tile tile
func (player *Player) Hu(tile Tile, tai TaiData, Type int, robGon bool, addToRoom bool, fromID int) int {
	if Type == COMMAND["ZIMO"] && player.FirstDraw {
		tai.Tai     += 16
		tai.Message += IF(player.ID == player.room.Banker, "天胡 ", "地胡 ").(string)
	} else if Type == COMMAND["ZIMO"] && !strings.Contains(tai.Message, "七搶一") && !strings.Contains(tai.Message, "八仙過海") {
		tai.Tai++
		tai.Message += "自摸 "
		if player.room.Deck.Count() == 16 {
			tai.Tai++
			tai.Message += "海底撈月 "
		}
	}
	
	if robGon {
		tai.Tai++
		tai.Message += "搶槓胡 "
	}
	if player.JustGon {
		tai.Tai++
		tai.Message += "槓上花 "
	}
	season := uint((4 + player.ID - player.room.OpenIdx) % 4 + 1)
	if player.Flowers[4].GetIndex(season) > 0 || player.Flowers[4].GetIndex(season + 4) > 0 {
		tai.Tai     += int(player.Flowers[4].GetIndex(season) + player.Flowers[4].GetIndex(season + 4))
		tai.Message += "花 "
	}
	
	score := 0
	for i := 0; i < 4; i++ {
		if Type == COMMAND["ZIMO"] && i != player.ID || Type == COMMAND["HU"] && i == fromID {
			tmp := IF(player.ID == player.room.Banker || i == player.room.Banker, tai.Tai + player.room.NumKeepWin, tai.Tai).(int)
			score                          += tmp
			player.Credit                  += tmp
			player.room.Players[i].Credit  -= tmp
			player.room.Players[i].ScoreLog = append(player.room.Players[i].ScoreLog, NewScoreRecord(tai.Message, "to", player.Name(), tile.ToString(), tmp))
		}
	}
	if Type == COMMAND["HU"] {
		player.ScoreLog = append(player.ScoreLog, NewScoreRecord(tai.Message, "from", player.room.Players[fromID].Name(), tile.ToString(), score))
	} else {
		player.ScoreLog = append(player.ScoreLog, NewScoreRecord(tai.Message, "", "", tile.ToString(), score))
	}
	return score
}

// Gon gons the tile
func (player *Player) Gon(tile Tile, Type int, fromID int) int {
	player.JustGon = true
	for i := 0; i < IF(Type == COMMAND["PONGON"], 1, 4).(int); i++ {
		player.Hand.Sub(tile)
	}

	score := 2
	var message string
	switch Type {
	case COMMAND["PONGON"]:
		score   = 1
		message = "碰槓"
		player.PonTiles.Sub(tile)
		player.GonTiles.Add(tile)
	case COMMAND["ONGON"]:
		message = "暗槓"
		player.OngonTiles.Add(tile)
	default:
		message = "槓"
		player.GonTiles.Add(tile)
	}
	for i := 0; i < 4; i++ {
		if Type != COMMAND["GON"] && i != player.ID || Type == COMMAND["GON"] && i == fromID {
			player.Credit                  += score
			player.room.Players[i].Credit  -= score
			player.room.Players[i].ScoreLog = append(player.room.Players[i].ScoreLog, NewScoreRecord(message, "to", player.Name(), tile.ToString(), -score))
		}
	}
	if Type == COMMAND["GON"] {
		player.ScoreLog = append(player.ScoreLog, NewScoreRecord(message, "from", player.room.Players[fromID].Name(), tile.ToString(), score))
	} else {
		player.ScoreLog = append(player.ScoreLog, NewScoreRecord(message, "", "", tile.ToString(), score * 3))
	}
	return score
}

// Pon pons the tile
func (player *Player) Pon(tile Tile) {
	player.PonTiles.Add(tile)
	player.Hand.Sub([]Tile{tile, tile})
}

// Eat eats the tile
func (player *Player) Eat(eatAction EatAction) {
	tile := eatAction.First
	player.EatTiles.Add(tile)
	for i := uint(0); i < 3; i++ {
		if tile.Value + i != eatAction.Center.Value {
			player.Hand.Sub(NewTile(tile.Suit, tile.Value + i))
		}
	}
}
