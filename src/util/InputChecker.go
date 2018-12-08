package util

import (
	"encoding/json"
	
	"tile"
	"action"
)

func (player *Player) checkChangeTiles(val interface{}) bool {
	switch val.(type) {
	case []interface{}:
	default:
		return false
	}
	valArr := val.([]interface{})
	for i:= 0; i < 3; i++ {
		if !tile.IsValidTile(valArr[i].(string)) {
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
		return tile.IsValidTile(val.(string))
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
	if t.Command != action.NONE && t.Command & (action.PON | action.GON | action.ONGON | action.PONGON | action.HU | action.ZIMO) == 0 {
		return false
	}
	return tile.IsValidTile(t.Tile)
}