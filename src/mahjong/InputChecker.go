package mahjong

import (
	"encoding/json"
)

func (player *Player) checkChangeTiles(val interface{}) bool {
	switch val.(type) {
	case []interface{}:
	default:
		return false
	}
	valArr := val.([]interface{})
	for i:= 0; i < 3; i++ {
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
	type Tmp struct {
		Command int
		Tile    string
		Score   int
	}
	var t Tmp
	err := json.Unmarshal([]byte(val.(string)), &t)
	if err != nil {
		return false
	}
	if t.Command != NONE && t.Command & (PON | GON | ONGON | PONGON | HU | ZIMO) == 0 {
		return false
	}
	return IsValidTile(t.Tile)
}