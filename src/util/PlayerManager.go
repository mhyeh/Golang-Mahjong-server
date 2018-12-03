package util

import (
	"github.com/googollee/go-socket.io"
	"github.com/satori/go.uuid"
)

const (
	WAITING = 0
	MATCHED = 1
	READY   = 2
	PLAYING = 4
	LEAVE   = 8
)

type IPlayer struct {
	Name   string
	Uuid   string
	Room   string
	Socket *socketio.Socket
	State  int
}

func GetNameList(list []*IPlayer) []string {
	var nameList []string
	for _, player := range list {
		nameList = append(nameList, player.Name)
	}
	return nameList
}

func GetUUIDList(list []IPlayer) []string {
	var uuidList []string
	for _, player := range list {
		uuidList = append(uuidList, player.Uuid)
	}
	return uuidList
}

type PlayerManager []*IPlayer

func (this *PlayerManager) AddPlayer(name string) (string, bool) {
	if this.FindPlayerByName(name) == -1 {
		return "", true
	}
	uuid := uuid.Must(uuid.NewV4()).String()
	*this = append(*this, &IPlayer {name, uuid, "", nil, WAITING})
	return uuid, false
}

func (this *PlayerManager) RemovePlayer(id int) {
	if id >= 0 && id < len(*this) {
		*this = append((*this)[: id], (*this)[id + 1: ]...)
	}
}

func (this PlayerManager) FindPlayerByName(name string) int {
	for index, player := range this {
		if player.Name == name {
			return index
		}
	}
	return -1
}

func (this PlayerManager) FindPlayerByUUID(uuid string) int {
	for index, player := range this {
		if player.Uuid == uuid {
			return index
		}
	}
	return -1
}

func (this PlayerManager) FindPlayerBySocket(socket socketio.Socket) int {
	for index, player := range this {
		if (*player.Socket).Id() == socket.Id() {
			return index
		}
	}
	return -1
}

func (this PlayerManager) FindPlayersInRoom(room string) []*IPlayer {
	var list []*IPlayer
	for _, player := range this {
		if player.Room == room {
			list = append(list, player)
			if len(list) == 4 {
				break
			}
		}
	}
	return list
}

func (this PlayerManager) FindPlayersIsSameState(state int) []*IPlayer {
	var list []*IPlayer
	for _, player := range this {
		if player.State == state {
			list = append(list, player)
			if len(list) == 4 {
				break
			}
		}
	}
	return list
}

func (this PlayerManager) Auth(room string, uuid string) bool {
	for _, player := range this {
		if player.Room == room && player.Uuid == uuid {
			return true
		}
	}
	return false
}