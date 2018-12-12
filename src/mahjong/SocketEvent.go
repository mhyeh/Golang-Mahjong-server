package mahjong

import (
	"github.com/googollee/go-socket.io"
	"log"
)

func SocketError(so socketio.Socket, err error) {
	log.Println("error:", err)
}

func SocketConnect(so socketio.Socket) {
	log.Println("on connection")

	so.Emit("auth")
	
	so.On("join", func(name string) (string, bool) {
		if name == "" {
			return "", true
		}
		_uuid, _err := Login(name, &so)
		return _uuid, _err
	})

	so.On("auth", func(uuid string, room string) int {
		if uuid == "" {
			return -1
		}

		index := FindPlayerByUUID(uuid)
		if index == -1 {
			return -1
		}

		player := PlayerList[index]
		if (player.State & (MATCHED | READY | PLAYING)) != 0 && !(room == "") && player.Room == room {
			so.Join(room)
		}
		player.Socket = &so
		return player.State
	})

	so.On("ready",          socketReady)
	so.On("getRoomInfo",    getRoomInfo)
	so.On("getID",          getID)
	so.On("getReadyPlayer", getReadyPlayer)
	so.On("getHand",        getHand)
	so.On("getPlayerList",  getPlayerList)
	so.On("getLack",        getLack)
	so.On("getHandCount",   getHandCount)
	so.On("getRemainCount", getRemainCount)
	so.On("getDoor",        getDoor)
	so.On("getSea",         getSea)
	so.On("getHu",          getHu)
	so.On("getCurrentIdx",  getCurrentIdx)
	so.On("getScore",       getScore)

	so.On("disconnection", func() {
		log.Println("on disconnect")
		Logout(so)
	})
}

func socketReady(room string, uuid string) int {
	if !Auth(room, uuid) {
		return -1
	}

	c := make(chan int, 1)
	fn := func(id int) {
		c<-id
	}
	go game.Rooms[room].Accept(uuid, fn)
	return <-c
}

func getRoomInfo(uuid string) (string, []string, bool) {
	if uuid == "" {
		return "", []string{}, true
	}
	index := FindPlayerByUUID(uuid)
	if index == -1 {
		return "", []string{}, true
	}
	player := PlayerList[index]
	room   := player.Room
	return room, GetNameList(FindPlayerListInRoom(room)), false
}

func getReadyPlayer(room string) []string {
	return IF(game.Rooms[room] == nil, []string{}, game.Rooms[room].GetReadyPlayers()).([]string)
}

func getHand(uuid string, room string) []string {
	if !Auth(room, uuid) || game.Rooms[room].State < DealCard {
		return []string{}
	}
	index := FindPlayerByUUID(uuid)
	id    := PlayerList[index].Index
	return game.Rooms[room].Players[id].Hand.ToStringArray()
}

func getID(uuid string, room string) int {
	if !Auth(room, uuid) {
		return -1
	}
	index := FindPlayerByUUID(uuid)
	return IF(PlayerList[index].State != READY, -1, PlayerList[index].Index).(int)
}

func getPlayerList(room string) []string {
	return IF(game.Rooms[room] == nil, []string{}, game.Rooms[room].GetPlayerList()).([]string)
}

func getLack(room string) []int {
	return IF(game.Rooms[room] == nil, []int{}, game.Rooms[room].GetLack()).([]int)
}

func getHandCount(room string) []int {
	return IF(game.Rooms[room] == nil, []int{}, game.Rooms[room].GetHandCount()).([]int)
}

func getRemainCount(room string) int {
	return IF(game.Rooms[room] == nil, 56, game.Rooms[room].GetRemainCount()).(int)
}

func getDoor(uuid string, room string) ([][]string, []int, bool) {
	if !Auth(room, uuid) {
		return [][]string{}, []int{}, true
	}
	index := FindPlayerByUUID(uuid)
	id    := PlayerList[index].Index
	return game.Rooms[room].GetDoor(id)
}

func getSea(room string) ([][]string, bool) {
	if game.Rooms[room] == nil {
		return [][]string{}, true
	}
	return game.Rooms[room].GetSea()
}

func getHu(room string) ([][]string, bool) {
	if game.Rooms[room] == nil {
		return [][]string{}, true
	}
	return game.Rooms[room].GetHu()
}

func getCurrentIdx(room string) int {
	return IF(game.Rooms[room] == nil, -1, game.Rooms[room].GetCurrentIdx()).(int)
}

func getScore(room string) []int {
	return IF(game.Rooms[room] == nil, []int{}, game.Rooms[room].GetScore()).([]int)
}