package util

import (
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io";
	"github.com/satori/go.uuid";
)

// NewGameManager creates a new gameManager
func NewGameManager() GameManager {
	var playerManager PlayerManager
	rooms := make(map[string]*Room)
	game  := GameManager {rooms, playerManager}
	return game
}

// GameManager represents a gameManager
type GameManager struct {
	Rooms         map[string]*Room
	PlayerManager PlayerManager
}

// Login handles player's login
func (gManager *GameManager) Login(name string, socket socketio.Socket) (string, bool) {
	uuid, err := gManager.PlayerManager.AddPlayer(name)
	if err {
		return "", true
	}
	index := gManager.PlayerManager.FindPlayerByUUID(uuid)
	gManager.PlayerManager[index].Socket = &socket
	gManager.PlayerManager[index].State  = WAITING

	return uuid, false
}

// Logout handles player's logout
func (gManager *GameManager) Logout(socket socketio.Socket) {
	index := gManager.PlayerManager.FindPlayerBySocket(socket)
	if index >= 0 && index < len(gManager.PlayerManager) {
		if gManager.PlayerManager[index].State == WAITING {
			gManager.PlayerManager.RemovePlayer(index)
		} 
		// else if gManager.PlayerManager[index].State == MATCHED {
		// 	gManager.RemoveRoom(gManager.PlayerManager[index].Room)
		// 	gManager.PlayerManager.RemovePlayer(index)
		// }
	}
}

// Exec executes the whole game
func (gManager *GameManager) Exec() {
	for {
		if gManager.WaitingNum() >= 4 {
			go gManager.CreateRoom()
		}
		time.Sleep(0)
	}
}

// WaitingNum returns the number of player which state are waiting
func (gManager *GameManager) WaitingNum() int {
	return len(gManager.PlayerManager.FindPlayersIsSameState(WAITING))
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
	gManager.Rooms[roomName] = NewRoom(gManager, roomName)
	matchPlayer := gManager.Match()
	gManager.Rooms[roomName].AddPlayer(matchPlayer)
	gManager.Rooms[roomName].WaitToStart()
	gManager.RemoveRoom(roomName)
}

// Match matchs 4 player into a room
func (gManager *GameManager) Match() []string {
	waitingList := gManager.PlayerManager.FindPlayersIsSameState(WAITING)
	var sample []string
	for i := 0; i < 4; i++ {
		index := rand.Int31n(int32(len(waitingList)))
		sample = append(sample, waitingList[index].UUID)
		waitingList = append(waitingList[: index], waitingList[index + 1: ]...)
	}
	for _, uuid := range sample {
		index := gManager.PlayerManager.FindPlayerByUUID(uuid)
		gManager.PlayerManager[index].State = MATCHED
	}
	return sample
}

// RemoveRoom removes a room by room name
func (gManager *GameManager) RemoveRoom(name string) {
	if gManager.Rooms[name].Waiting {
		gManager.Rooms[name].StopWaiting()
	}
	playerList := gManager.PlayerManager.FindPlayersInRoom(name)
	for _, player := range playerList {
		var index int
		index = gManager.PlayerManager.FindPlayerByUUID(player.UUID)
		if gManager.Rooms[name].Waiting {
			gManager.PlayerManager[index].State = WAITING
		} else {
			gManager.PlayerManager.RemovePlayer(index)
		}
	}
	delete(gManager.Rooms, name)
}
