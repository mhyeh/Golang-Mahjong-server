package util

import (
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io";
	"github.com/satori/go.uuid";
)

// command type
const  (
	NONE   = 0
	PON    = 1
	GON    = 2
	ONGON  = 4
	PONGON = 8
	HU     = 16
	ZIMO   = 32
)

// NewGameManager creates a new gameManager
func NewGameManager() GameManager {
	rooms := make(map[string]*Room)
	var playerManager PlayerManager
	game := GameManager {rooms, playerManager}
	return game
}

// GameManager represents a gameManager
type GameManager struct {
	Rooms map[string]*Room
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
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(10 * time.Second)
		}
	}
}

// WaitingNum returns the number of player which state are waiting
func (gManager *GameManager) WaitingNum() int {
	return len(gManager.PlayerManager.FindPlayersIsSameState(WAITING))
}

// CreateRoom creates a new room and add player to that room
func (gManager *GameManager) CreateRoom() {
	var roomName string
	for ;; {
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

// SSJ cals tai
func SSJ(hand uint64, door uint64) int {
	idx := (((gData[hand & 134217727] | 4) &
		(gData[hand >> 27] | 4) &
		(gData[door & 134217727] | 64) &
		(gData[door >> 27] | 64) &
		(484 | ((gData[door & 134217727] & gData[door >> 27] & 16) >> 1))) |
		(((gData[hand & 134217727] & (gData[door & 134217727] | 3)) | (gData[hand >> 27] & (gData[door >> 27] | 3))) & 19) |
		((gData[(hand & 134217727) + (door & 134217727)] & 3584) + (gData[(hand >> 27) + (door >> 27)] & 3584)))
	return gData[size + idx];
}

// InitHuTable intis the hu table
func InitHuTable() bool {
	var i uint
	gGroup = append(gGroup, 0)
	for i = 0; i < 9; i++ {
		gGroup = append(gGroup, 3 << (i * 3))
	}
	for i = 0; i < 7; i++ {
		gGroup = append(gGroup, 73 << (i * 3))
	}
	for i = 0; i < 9; i++ {
		gEye = append(gEye, 2 << (i * 3))
	}
	b01(4, 0, size)
	b2(4, 0, 1)
	b3(7, 0, 1)
	b4()
	b5(4, 0, 1)
	b6()
	b7()
	b8(4, 0, size)
	b9UP()
	t()
	println("Initialization Completed!")
	return true
}

const size = 76695844
const max  = 76699939

var gData  [max]int
var gGroup []int
var gEye   []int

func have(m int, s int) bool {
	for i := uint(0); i < 9; i++ {
		if ((m >> (i * 3)) & 7) < ((s >> (i * 3)) & 7) {
			return false
		}
	}
	return true
}

func b01(n int, d int, p int) {
	var i int
	if n != 0 {
		for i = 0; i < 17; i++ {
			if have(p, gGroup[i]) {
				b01(n - 1, d + gGroup[i], p - gGroup[i])
			}
		}
	} else {
		gData[d] |= 1
		for i = 0; i < 9; i++ {
			if have(p, gEye[i]) {
				gData[d + gEye[i]] |= 2
			}
		}
	}
}

func b2(n int, d int, c int) {
	gData[d] |= 4;
	gData[d] |= 32;
	if (d & 16777208) == 0 {
		gData[d] |= 256;
	}
	if n != 0 {
		for i := c; i <= 9; i++ {
			b2(n - 1, d + gGroup[i], i + 1);
			b2(n - 1, d + gGroup[i] / 3 * 4, i + 1);
		}
	}
}

func b3(n int, d int, c int) {
	gData[d] |= 8;
	if n != 0 {
		for i := c; i <= 9; i++ {
			b3(n - 1, d + gGroup[i] / 3 * 2, i + 1);
			b3(n - 2, d + gGroup[i] / 3 * 4, i + 1);
		}
	}
}

func b4() {
	gData[0] |= 16
}

func b5(n int, d int, c int) {
	var i int;
	gData[d] |= 32;
	for i = 0; i < 9; i++ {
		if have(size - d, gEye[i]) {
			gData[d + gEye[i]] |= 32;
		}
	}
	if n != 0 {
		for i = c; i <= 9; i++ {
			b5(n - 1, d + gGroup[i], i + 1);
		}
	}
}

func b6() {
	gData[0] |= 64;
	for i := 0; i < 9; i++ {
		gData[gEye[i]] |= 64;
	}
}

func b7() {
	for i := 0; i < size; i++ {
		if (i & 119508935) == 0 {
			gData[i] |= 128;
		}
	}
}

func b8(n int, d int, p int) {
	var i int;
	if n != 0 {
		for i = 0; i < 17; i++ {
			if have(p, gGroup[i]) && (i == 0 || i == 1 || i == 9 || i == 10 || i == 16) {
				b8(n - 1, d + gGroup[i], p - gGroup[i]);
			}
		}
	} else {
		gData[d] |= 256;
		for i = 0; i < 9; i++ {
			if have(p, gEye[i]) && (i == 0 || i == 8) {
				gData[d + gEye[i]] |= 256;
			}
		}
	}
}

func b9UP() {
	for i := 0; i < size; i++ {
		k := 0;
		for j := uint(0); j < 9; j++ {
			if (i & (4 << (j * 3))) != 0 {
				k++;
			}
		}
		if k > 7 {
			k = 7;
		}
		gData[i] |= (k << 9);
	}
}

func t() {
	for i := uint(0); i < 4095; i++ {
		k := uint(0);
		if (i & 7) == 7 {
			k = 1;
			if (i & 32) == 32 {
				k = 2;
			}
			if (i & 16) == 16 {
				k = 3;
			}
			if (i & 64) == 64 {
				k = 3;
			}
			if (i & 256) == 256 {
				k = 3;
			}
			if (i & 48) == 48 {
				k = 4;
			}
			if (i & 160) == 160 {
				k = 4;
			}
			if (i & 272) == 272 {
				k = 5;
			}
			if (i & 80) == 80 {
				k = 5;
			}
			if (i & 192) == 192 {
				k = 5;
			}
			k += (i >> 9);
		} else if (i & 8) != 0 {
			k = 3;
			if (i & 16) == 16 {
				k = 5;
			}
			if (i >> 9) != 0 {
				k = 4;
			}
			if (i & 16) == 16 && (i >> 9) != 0 {
				k = 5;
			}
			k += (i >> 9);
		}
		gData[size + i] = int(k);
	}
}