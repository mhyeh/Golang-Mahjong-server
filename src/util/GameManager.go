package util

import (
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io";
	"github.com/satori/go.uuid";

	"manager"
)

// NewGameManager creates a new gameManager
func NewGameManager() GameManager {
	rooms := make(map[string]*Room)
	game  := GameManager {rooms, nil}
	return game
}

// GameManager represents a gameManager
type GameManager struct {
	Rooms  map[string]*Room
	Server *socketio.Server
}

// Login handles player's login
func (gManager *GameManager) Login(name string, socket socketio.Socket) (string, bool) {
	uuid, err := manager.AddPlayer(name)
	if err {
		return "", true
	}
	index := manager.FindPlayerByUUID(uuid)
	manager.PlayerList[index].Socket = &socket
	manager.PlayerList[index].State  = manager.WAITING

	return uuid, false
}

// Logout handles player's logout
func (gManager *GameManager) Logout(socket socketio.Socket) {
	index := manager.FindPlayerBySocket(socket)
	if index >= 0 && index < len(manager.PlayerList) {
		if manager.PlayerList[index].State == manager.WAITING {
			manager.RemovePlayer(index)
		} 
		// else if manager.PlayerList[index].State == MATCHED {
		// 	gManager.RemoveRoom(manager.PlayerList[index].Room)
		// 	RemovePlayer(index)
		// }
	}
}

// Exec executes the whole game
func (gManager *GameManager) Exec() {
	for {
		if gManager.WaitingNum() >= 4 {
			go gManager.CreateRoom()
			time.Sleep(2 * time.Second)
		}
		time.Sleep(5 * time.Second)
	}
}

// WaitingNum returns the number of player which state are waiting
func (gManager *GameManager) WaitingNum() int {
	return len(manager.FindPlayerListIsSameState(manager.WAITING))
}

// CreateRoom creates a new room and add player to that room
func (gManager *GameManager) CreateRoom() {
	var roomName string
	for {
		roomName = uuid.Must(uuid.NewV4()).String()
		if gManager.Rooms[roomName] == nil {
			break
		}
	}
	gManager.Rooms[roomName]    = NewRoom(roomName)
	gManager.Rooms[roomName].IO = gManager.Server
	matchPlayer := gManager.Match()
	gManager.Rooms[roomName].AddPlayer(matchPlayer)
	gManager.Rooms[roomName].WaitToStart()
	gManager.RemoveRoom(roomName)
}

// RemoveRoom removes a room by room name
func (gManager *GameManager) RemoveRoom(name string) {
	if gManager.Rooms[name].Waiting {
		gManager.Rooms[name].StopWaiting()
	}
	playerList := manager.FindPlayerListInRoom(name)
	for _, player := range playerList {
		var index int
		index = manager.FindPlayerByUUID(player.UUID)
		if gManager.Rooms[name].Waiting {
			manager.PlayerList[index].State = manager.WAITING
		} else {
			manager.RemovePlayer(index)
		}
	}
	delete(gManager.Rooms, name)
}

// Match matchs 4 player into a room
func (gManager *GameManager) Match() []string {
	waitingList := manager.FindPlayerListIsSameState(manager.WAITING)
	var sample []string
	for i := 0; i < 4; i++ {
		index := rand.Int31n(int32(len(waitingList)))
		sample = append(sample, waitingList[index].UUID)
		waitingList = append(waitingList[: index], waitingList[index + 1: ]...)
	}
	for _, uuid := range sample {
		index := manager.FindPlayerByUUID(uuid)
		manager.PlayerList[index].State = manager.MATCHED
	}
	return sample
}