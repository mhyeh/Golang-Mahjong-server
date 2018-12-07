package PManager

import (
	"github.com/googollee/go-socket.io"
	"github.com/satori/go.uuid"
)

// Player's state
const (
	WAITING = 0
	MATCHED = 1
	READY   = 2
	PLAYING = 4
	LEAVE   = 8
)

// IPlayer represents the player's info
type IPlayer struct {
	Name   string
	UUID   string
	Room   string
	Socket *socketio.Socket
	State  int
	Index  int
}

// PlayerManager represents the array of pointer of IPlayer
type PlayerManager []*IPlayer

// Players is a PlayerManager
var Players PlayerManager

// GetNameList returns the list of player's name
func GetNameList(list []*IPlayer) []string {
	var nameList []string
	for _, player := range list {
		nameList = append(nameList, player.Name)
	}
	return nameList
}

// GetUUIDList returns the list of player's uuid
func GetUUIDList(list []IPlayer) []string {
	var uuidList []string
	for _, player := range list {
		uuidList = append(uuidList, player.UUID)
	}
	return uuidList
}

// AddPlayer adds a new player into PlayerManager
func AddPlayer(name string) (string, bool) {
	if FindPlayerByName(name) != -1 {
		return "", true
	}
	var _uuid string
	for ;; {
		_uuid = uuid.Must(uuid.NewV4()).String()
		if FindPlayerByUUID(_uuid) == -1 {
			break
		}
	}
	Players = append(Players, &IPlayer {name, _uuid, "", nil, WAITING, -1})
	return _uuid, false
}

// RemovePlayer remove a player from Players
func RemovePlayer(id int) {
	if id >= 0 && id < len(Players) {
		Players = append((Players)[: id], (Players)[id + 1: ]...)
	}
}

// FindPlayerByName gets player's index by player's name
func FindPlayerByName(name string) int {
	for index, player := range Players {
		if player.Name == name {
			return index
		}
	}
	return -1
}

// FindPlayerByUUID gets player's index by player's uuid
func FindPlayerByUUID(uuid string) int {
	for index, player := range Players {
		if player.UUID == uuid {
			return index
		}
	}
	return -1
}

// FindPlayerBySocket gets player's index by player's socket
func FindPlayerBySocket(socket socketio.Socket) int {
	for index, player := range Players {
		if (*player.Socket).Id() == socket.Id() {
			return index
		}
	}
	return -1
}

// FindPlayersInRoom gets list of player which in the same room
func FindPlayersInRoom(room string) []*IPlayer {
	var list []*IPlayer
	for _, player := range Players {
		if player.Room == room {
			list = append(list, player)
			if len(list) == 4 {
				break
			}
		}
	}
	return list
}

// FindPlayersIsSameState gets list of player which are same state
func FindPlayersIsSameState(state int) []*IPlayer {
	var list []*IPlayer
	for _, player := range Players {
		if player.State == state {
			list = append(list, player)
			if len(list) == 4 {
				break
			}
		}
	}
	return list
}

// Auth authenticates a player
func Auth(room string, uuid string) bool {
	for _, player := range Players {
		if player.Room == room && player.UUID == uuid {
			return true
		}
	}
	return false
}