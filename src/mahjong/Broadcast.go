package mahjong

import (
	"encoding/json"
)

// BroadcastRemainTile broadcasts remain tile
func (room Room) BroadcastRemainTile(num uint) {
	room.IO.BroadcastTo(room.Name, "remainTile", num)
}

// BroadcastStopWaiting broadcasts stop waiting signal
func (room Room) BroadcastStopWaiting() {
	room.IO.BroadcastTo(room.Name, "stopWaiting")
}

// BroadcastReady broadcasts the player's name who is ready
func (room Room) BroadcastReady(name string) {
	room.IO.BroadcastTo(room.Name, "broadcastReady", name)
}

// BroadcastGameStart broadcasts player list
func (room Room) BroadcastGameStart() {
	room.IO.BroadcastTo(room.Name, "broadcastGameStart", room.GetPlayerList())
}

// BroadcastChange broadcasts the player's id who already change tiles
func (room Room) BroadcastChange(id int) {
	room.IO.BroadcastTo(room.Name, "broadcastChange", id)
}

// BroadcastLack broadcasts the player's id who already choose lack
func (room Room) BroadcastLack() {
	room.IO.BroadcastTo(room.Name, "broadcastLack", room.ChoosedLack)
}

// BroadcastDraw broadcasts the player's id who draw a tile
func (room Room) BroadcastDraw(id int, num uint) {
	room.IO.BroadcastTo(room.Name, "broadcastDraw", id, num)
}

// BroadcastThrow broadcasts the player's id and the tile he threw
func (room Room) BroadcastThrow(id int, tile Tile) {
	room.IO.BroadcastTo(room.Name, "broadcastThrow", id, tile.ToString())
}

// BroadcastCommand broadcasts the player's id and the command he made
func (room Room) BroadcastCommand(from int, to int, command int, tile Tile, score int) {
	if command == COMMAND["ONGON"] {
		room.IO.BroadcastTo(room.Name, "broadcastCommand", from, to, command, "", score)
	} else {
		room.IO.BroadcastTo(room.Name, "broadcastCommand", from, to, command, tile.ToString(), score)
	}
}

// BroadcastEnd broadcasts the game result
func (room Room) BroadcastEnd(data []GameResult) {
	result, _ := json.Marshal(data)
	room.IO.BroadcastTo(room.Name, "end", string(result))
}

// BroadcastRobGon broadcasts rob gon
func (room Room) BroadcastRobGon(id int, tile Tile) {
	room.IO.BroadcastTo(room.Name, "robGon", id, tile.ToString())
}