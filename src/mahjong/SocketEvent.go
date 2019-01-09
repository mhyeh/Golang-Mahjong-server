package mahjong

import (
	"github.com/googollee/go-socket.io"
	"log"
)

// SocketError is callback of socket error event
func SocketError(so socketio.Socket, err error) {
	log.Println("error:", err)
}

// SocketConnect is callback of socket connect event
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

func socketReady(uuid string, room string) int {
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
	if game.Rooms[room] == nil {
		return []string{}
	}
	return game.Rooms[room].GetReadyPlayers()
}

func getHand(uuid string, room string) []string {
	if !Auth(room, uuid) || game.Rooms[room].State < DealTile {
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
	if PlayerList[index].State != READY {
		return -1
	}
	return PlayerList[index].Index
}

func getPlayerList(room string) []string {
	if game.Rooms[room] == nil {
		return []string{}
	}
	return game.Rooms[room].GetPlayerList()
}

func getLack(room string) []int {
	if game.Rooms[room] == nil {
		return []int{}
	}
	return game.Rooms[room].GetLack()
}

func getHandCount(room string) []int {
	if game.Rooms[room] == nil {
		return []int{}
	}
	return game.Rooms[room].GetHandCount()
}

func getRemainCount(room string) int {
	if game.Rooms[room] == nil {
		return 56
	}
	return game.Rooms[room].GetRemainCount()
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
	if game.Rooms[room] == nil {
		return -1
	}
	return game.Rooms[room].GetCurrentIdx()
}

func getScore(room string) []int {
	if game.Rooms[room] == nil {
		return []int{}
	}
	return game.Rooms[room].GetScore()
}