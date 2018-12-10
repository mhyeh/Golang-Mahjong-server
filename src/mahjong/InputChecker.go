package mahjong

func (player *Player) checkChangeTiles(val interface{}) bool {
	switch val.(type) {
	case []interface{}:
	default:
		return false
	}
	valArr := val.([]interface{})
	for i := 0; i < 3; i++ {
		if !IsValidTile(valArr[i].(string)) {
			return false
		}
	}
	return true
}

func (player *Player) checkLack(val interface{}) bool {
	switch val.(type) {
	case float64:
	default:
		return false
	}
	lack := int(val.(float64))
	return lack >= 0 && lack < 4
}

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
	act  := StringToAction(val.(string))
	flag := false
	for _, command := range COMMAND {
		if act.Command == command {
			flag = true
		}
	}
	return flag && act.Tile.Suit != -1
}
