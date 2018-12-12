package mahjong

import (
	"math/rand"
	"time"
	"log"

	"github.com/googollee/go-socket.io";
	"github.com/satori/go.uuid";
)

var game *GameManager
// NewGameManager creates a new gameManager
func NewGameManager() (bool) {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
		return true
	}

	go InitHuTable()

	rooms := make(map[string]*Room)
	game   = &GameManager {rooms, server}
	return false
}

// GameManager represents a gameManager
type GameManager struct {
	Rooms  map[string]*Room
	Server *socketio.Server
}

func GetServer() *socketio.Server {
	return game.Server
}

// Login handles player's login
func Login(name string, socket *socketio.Socket) (string, bool) {
	uuid, err := AddPlayer(name)
	if err {
		return "", true
	}
	index := FindPlayerByUUID(uuid)
	PlayerList[index].Socket = socket
	PlayerList[index].State  = WAITING

	return uuid, false
}

// Logout handles player's logout
func Logout(socket socketio.Socket) {
	index := FindPlayerBySocket(socket)
	if index >= 0 && index < len(PlayerList) {
		if PlayerList[index].State == WAITING {
			RemovePlayer(index)
		} 
		// else if PlayerList[index].State == MATCHED {
		// 	game.RemoveRoom(PlayerList[index].Room)
		// 	RemovePlayer(index)
		// }
	}
}

// Exec executes the whole game
func Exec() {
	for {
		if WaitingNum() >= 4 {
			go CreateRoom()
			time.Sleep(2 * time.Second)
		}
		time.Sleep(5 * time.Second)
	}
}

// WaitingNum returns the number of player which state are waiting
func WaitingNum() int {
	return len(FindPlayerListIsSameState(WAITING))
}

// CreateRoom creates a new room and add player to that room
func CreateRoom() {
	var roomName string
	for {
		roomName = uuid.Must(uuid.NewV4()).String()
		if game.Rooms[roomName] == nil {
			break
		}
	}
	matchPlayer := Match()
	game.Rooms[roomName]    = NewRoom(roomName)
	game.Rooms[roomName].IO = game.Server
	game.Rooms[roomName].AddPlayer(matchPlayer)
	game.Rooms[roomName].WaitToStart()
	RemoveRoom(roomName)
}

// RemoveRoom removes a room by room name
func RemoveRoom(name string) {
	if game.Rooms[name].Waiting {
		game.Rooms[name].StopWaiting()
	}
	playerList := FindPlayerListInRoom(name)
	for _, player := range playerList {
		var index int
		index = FindPlayerByUUID(player.UUID)
		if game.Rooms[name].Waiting {
			PlayerList[index].State = WAITING
		} else {
			RemovePlayer(index)
		}
	}
	delete(game.Rooms, name)
}

// Match matchs 4 player into a room
func Match() []string {
	waitingList := FindPlayerListIsSameState(WAITING)
	var sample []string
	for i := 0; i < 4; i++ {
		index      := rand.Int31n(int32(len(waitingList)))
		sample      = append(sample, waitingList[index].UUID)
		waitingList = append(waitingList[: index], waitingList[index + 1: ]...)
	}
	for _, uuid := range sample {
		index := FindPlayerByUUID(uuid)
		PlayerList[index].State = MATCHED
	}
	return sample
}