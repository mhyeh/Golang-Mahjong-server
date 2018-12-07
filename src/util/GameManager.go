package util

import (
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io";
	"github.com/satori/go.uuid";

	"PManager"
)

// NewGameManager creates a new gameManager
func NewGameManager() GameManager {
	rooms := make(map[string]*Room)
	game  := GameManager {rooms}
	return game
}

// GameManager represents a gameManager
type GameManager struct {
	Rooms map[string]*Room
}

// Login handles player's login
func (gManager *GameManager) Login(name string, socket socketio.Socket) (string, bool) {
	uuid, err := PManager.AddPlayer(name)
	if err {
		return "", true
	}
	index := PManager.FindPlayerByUUID(uuid)
	PManager.Players[index].Socket = &socket
	PManager.Players[index].State  = PManager.WAITING

	return uuid, false
}

// Logout handles player's logout
func (gManager *GameManager) Logout(socket socketio.Socket) {
	index := PManager.FindPlayerBySocket(socket)
	if index >= 0 && index < len(PManager.Players) {
		if PManager.Players[index].State == PManager.WAITING {
			PManager.RemovePlayer(index)
		} 
		// else if PManager.Players[index].State == MATCHED {
		// 	gManager.RemoveRoom(PManager.Players[index].Room)
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
	return len(PManager.FindPlayersIsSameState(PManager.WAITING))
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
	gManager.Rooms[roomName] = NewRoom(roomName)
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
	playerList := PManager.FindPlayersInRoom(name)
	for _, player := range playerList {
		var index int
		index = PManager.FindPlayerByUUID(player.UUID)
		if gManager.Rooms[name].Waiting {
			PManager.Players[index].State = PManager.WAITING
		} else {
			PManager.RemovePlayer(index)
		}
	}
	delete(gManager.Rooms, name)
}

// Match matchs 4 player into a room
func (gManager *GameManager) Match() []string {
	waitingList := PManager.FindPlayersIsSameState(PManager.WAITING)
	var sample []string
	for i := 0; i < 4; i++ {
		index := rand.Int31n(int32(len(waitingList)))
		sample = append(sample, waitingList[index].UUID)
		waitingList = append(waitingList[: index], waitingList[index + 1: ]...)
	}
	for _, uuid := range sample {
		index := PManager.FindPlayerByUUID(uuid)
		PManager.Players[index].State = PManager.MATCHED
	}
	return sample
}