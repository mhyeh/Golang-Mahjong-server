package mahjong

func (player *Player) checkThrow(val interface{}) bool {
	switch val.(type) {
	case string:
		return IsValidTile(val.(string))
	default:
		return false
	}
}

func (player *Player) checkCommand(val interface{}) bool {
	switch val.(type) {
	case string:
	default:
		return false
	}
	act  := JSONToAction(val.(string))
	flag := false
	for _, command := range COMMAND {
		if act.Command == command {
			flag = true
		}
	}
	return flag && act.Tile.Suit != -1
}

func (player *Player) checkTing(val interface{}) bool {
	switch val.(type) {
	case bool:
		return true
	default:
		return false
	}
}