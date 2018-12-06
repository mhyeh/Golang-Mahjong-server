package util

import (
	"MJCard"
)

// BroadcastRemainCard broadcasts remain card
func (room Room) BroadcastRemainCard(num uint) {
	room.IO.BroadcastTo(room.Name, "remainCard", num)
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

// BroadcastChange broadcasts the player's id who already change cards
func (room Room) BroadcastChange(id int) {
	room.IO.BroadcastTo(room.Name, "broadcastChange", id)
}

// BroadcastLack broadcasts the player's id who already choose lack
func (room Room) BroadcastLack() {
	room.IO.BroadcastTo(room.Name, "afterLack", room.ChoosedLack)
}

// BroadcastDraw broadcasts the player's id who draw a card
func (room Room) BroadcastDraw(id int) {
	room.IO.BroadcastTo(room.Name, "broadcastDraw", id)
}

// BroadcastThrow broadcasts the player's id and the card he threw
func (room Room) BroadcastThrow(id int, card MJCard.Card) {
	room.IO.BroadcastTo(room.Name, "broadcastThrow", id, card.ToString())
}

// BroadcastCommand broadcasts the player's id and the command he made
func (room Room) BroadcastCommand(from int, to int, command int, card MJCard.Card, score int) {
	if command == ONGON {
		room.IO.BroadcastTo(room.Name, "broadcastCommand", from, to, command, "", score)
	} else {
		room.IO.BroadcastTo(room.Name, "broadcastCommand", from, to, command, card.ToString(), score)
	}
}

// BroadcastEnd broadcasts the game result
func (room Room) BroadcastEnd(data string) {
	room.IO.BroadcastTo(room.Name, "end", data)
}