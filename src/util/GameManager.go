package util

import (
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io";
	"github.com/satori/go.uuid";
)

const  (
	NONE           = 0
	COMMAND_PON    = 1
	COMMAND_GON    = 2
	COMMAND_ONGON  = 4
	COMMAND_PONGON = 8
	COMMAND_HU     = 16
	COMMAND_ZIMO   = 32
)

func NewGameManager() GameManager {
	rooms := make(map[string]*Room)
	var playerManager PlayerManager
	game := GameManager {rooms, playerManager}
	return game
}

type GameManager struct {
	Rooms map[string]*Room
	PlayerManager PlayerManager
}

func (this *GameManager) Login(name string, socket socketio.Socket) (string, bool) {
	uuid, err := this.PlayerManager.AddPlayer(name)
	if err {
		return "", true
	}
	index := this.PlayerManager.FindPlayerByUUID(uuid)
	this.PlayerManager[index].Socket = &socket
	this.PlayerManager[index].State  = WAITING

	return uuid, false
}

func (this *GameManager) Logout(socket socketio.Socket) {
	index := this.PlayerManager.FindPlayerBySocket(socket)
	if index >= 0 && index < len(this.PlayerManager) && this.PlayerManager[index].State == WAITING {
		this.PlayerManager.RemovePlayer(index)
	}
}

func (this *GameManager) Exec() {
	for {
		if this.WaitingNum() >= 4 {
			this.CreateRoom()
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(10 * time.Second)
		}
	}
}

func (this *GameManager) WaitingNum() int {
	return len(this.PlayerManager.FindPlayersIsSameState(WAITING))
}

func (this *GameManager) CreateRoom() {
	roomName := uuid.Must(uuid.NewV4()).String()
	this.Rooms[roomName] = NewRoom(this, roomName)
	matchPlayer := this.Match()
	this.Rooms[roomName].AddPlayer(matchPlayer)
	this.Rooms[roomName].WaitToStart()
}

func (this *GameManager) Match() []string {
	waitingList := this.PlayerManager.FindPlayersIsSameState(WAITING)
	var sample []string
	for i := 0; i < 4; i++ {
		index := rand.Int31n(int32(len(waitingList)))
		sample = append(sample, waitingList[index].Uuid)
		waitingList = append(waitingList[: index], waitingList[index + 1: ]...)
	}
	for _, uuid := range sample {
		index := this.PlayerManager.FindPlayerByUUID(uuid)
		this.PlayerManager[index].State = MATCHED
	}
	return sample
}

func (this *GameManager) RemoveRoom(name string) {
	delete(this.Rooms, name)
}

func SSJ(hand uint64, door uint64) int {
	idx := (((g_data[hand & 134217727] | 4) &
		(g_data[hand >> 27] | 4) &
		(g_data[door & 134217727] | 64) &
		(g_data[door >> 27] | 64) &
		(484 | ((g_data[door & 134217727] & g_data[door >> 27] & 16) >> 1))) | 
		(((g_data[hand & 134217727] & (g_data[door & 134217727] | 3)) | (g_data[hand >> 27] & (g_data[door >> 27] | 3))) & 19) |
		((g_data[(hand & 134217727) + (door & 134217727)] & 3584) + (g_data[(hand >> 27) + (door >> 27)] & 3584)))
	println(size, idx, size + idx, max)
	return g_data[size + idx];
}

func InitHuTable() bool {
	var i uint
	g_group = append(g_group, 0)
	for i = 0; i < 9; i++ {
		g_group = append(g_group, 3 << i * 3)
	}
	for i = 0; i < 7; i++ {
		g_group = append(g_group, 73 << i * 3)
	}
	for i = 0; i < 9; i++ {
		g_eye = append(g_eye, 2 << i * 3)
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

var g_data  [max]int
var g_group []int
var g_eye   []int

func have(m int, s int) bool {
	for i := uint(0); i < 9; i++ {
		if ((m >> i * 3) & 7) < ((s >> i * 3) & 7) {
			return false
		}
	}
	return true
}

func b01(n int, d int, p int) {
	var i int
	if n != 0 {
		for i = 0; i < 17; i++ {
			if have(p, g_group[i]) {
				b01(n - 1, d + g_group[i], p - g_group[i])
			}
		}
	} else {
		g_data[d] |= 1
		for i = 0; i < 9; i++ {
			if have(p, g_eye[i]) {
				g_data[d + g_eye[i]] |= 2
			}
		}
	}
}

func b2(n int, d int, c int) {
	g_data[d] |= 4;
	g_data[d] |= 32;
	if (d & 16777208) == 0 {
		g_data[d] |= 256;
	}
	if n != 0 {
		for i := c; i <= 9; i++ {
			b2(n - 1, d + g_group[i], i + 1);
			b2(n - 1, d + g_group[i] / 3 * 4, i + 1);
		}
	}
}

func b3(n int, d int, c int) {
	g_data[d] |= 8;
	if n != 0 {
		for i := c; i <= 9; i++ {
			b3(n - 1, d + g_group[i] / 3 * 2, i + 1);
			b3(n - 2, d + g_group[i] / 3 * 4, i + 1);
		}
	}
}

func b4() {
	g_data[0] |= 10
}

func b5(n int, d int, c int) {
	var i int;
	g_data[d] |= 32;
	for i = 0; i < 9; i++ {
		if have(size - d, g_eye[i]) {
			g_data[d + g_eye[i]] |= 32;
		}
	}
	if n != 0 {
		for i = c; i <= 9; i++ {
			b5(n - 1, d + g_group[i], i + 1);
		}
	}
}

func b6() {
	g_data[0] |= 64;
	for i := 0; i < 9; i++ {
		g_data[g_eye[i]] |= 64;
	}
}

func b7() {
	for i := 0; i < size; i++ {
		if (i & 119508935) == 0 {
			g_data[i] |= 128;
		}
	}
}

func b8(n int, d int, p int) {
	var i int;
	if n != 0 {
		for i = 0; i < 17; i++ {
			if have(p, g_group[i]) && (i == 0 || i == 1 || i == 9 || i == 10 || i == 16) {
				b8(n - 1, d + g_group[i], p - g_group[i]);
			}
		}
	} else {
		g_data[d] |= 256;
		for i = 0; i < 9; i++ {
			if have(p, g_eye[i]) && (i == 0 || i == 8) {
				g_data[d + g_eye[i]] |= 256;
			}
		}
	}
}

func b9UP() {
	for i := 0; i < size; i++ {
		k := 0;
		for j := uint(0); j < 9; j++ {
			if (i & (4 << j * 3)) != 0 {
				k++;
			}
		}
		if k > 7 {
			k = 7;
		}
		g_data[i] |= (k << 9);
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
		g_data[size + i] = int(k);
	}
}