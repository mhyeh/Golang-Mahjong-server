package main

import (
	"log"
	"net/http"
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io"
	"github.com/rs/cors"

	"mahjong"
)

func main() {
	rand.Seed(time.Now().Unix())

	game := mahjong.NewGameManager()
	mahjong.InitHuTable()

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	game.Server = server
	
	go game.Exec()

	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")

		so.Emit("auth")
		
		so.On("join", func(name string) (string, bool) {
			if name == "" {
				return "", true
			}
			_uuid, _err := game.Login(name, so)
			return _uuid, _err
		})

		so.On("auth", func(uuid string, room string) int {
			if uuid == "" {
				return -1
			}

			index := mahjong.FindPlayerByUUID(uuid)
			if index == -1 {
				return -1
			}

			player := mahjong.PlayerList[index]
			if (player.State & (mahjong.MATCHED | mahjong.READY | mahjong.PLAYING)) != 0 && !(room == "") && player.Room == room {
				so.Join(room)
			}
			player.Socket = &so
			return player.State
		})

		so.On("ready", func(room string, uuid string) int {
			if !mahjong.Auth(room, uuid) {
				return -1
			}

			c := make(chan int, 1)
			fn := func(id int) {
				c<-id
			}
			go game.Rooms[room].Accept(uuid, fn)
			return <-c
		})

		so.On("getRoomInfo", func(uuid string) (string, []string, bool) {
			if uuid == "" {
				return "", []string{}, true
			}
			index := mahjong.FindPlayerByUUID(uuid)
			if index == -1 {
				return "", []string{}, true
			}
			player := mahjong.PlayerList[index]
			room   := player.Room
			return room, mahjong.GetNameList(mahjong.FindPlayerListInRoom(room)), false
		})

		so.On("getID", func(uuid string, room string) int {
			if !mahjong.Auth(room, uuid) {
				return -1
			}
			index := mahjong.FindPlayerByUUID(uuid)
			if mahjong.PlayerList[index].State != mahjong.READY {
				return -1
			}
			return mahjong.PlayerList[index].Index
		})

		so.On("getReadyPlayer", func(room string) []string {
			if (game.Rooms[room] == nil) {
				return []string{}
			}
			return game.Rooms[room].GetReadyPlayers()
		})

		so.On("getHand", func(uuid string, room string) []string {
			if !mahjong.Auth(room, uuid) || game.Rooms[room].State < mahjong.DealCard {
				return []string{}
			}
			index := mahjong.FindPlayerByUUID(uuid)
			id    := mahjong.PlayerList[index].Index
			return game.Rooms[room].Players[id].Hand.ToStringArray()
		})

		so.On("getPlayerList", func(room string) []string {
			if game.Rooms[room] == nil {
				return []string{}
			}
			return game.Rooms[room].GetPlayerList()
		})

		so.On("getLack", func(room string) []int {
			if game.Rooms[room] == nil {
				return []int{}
			}
			return game.Rooms[room].GetLack()
		})

		so.On("getHandCount", func(room string) []int {
			if game.Rooms[room] == nil {
				return []int{}
			}
			return game.Rooms[room].GetHandCount()
		})

		so.On("getRemainCount", func(room string) int {
			if game.Rooms[room] == nil {
				return 56
			}
			return game.Rooms[room].GetRemainCount()
		})

		so.On("getDoor", func(uuid string, room string) ([][]string, []int, bool) {
			if !mahjong.Auth(room, uuid) {
				return [][]string{}, []int{}, true
			}
			index := mahjong.FindPlayerByUUID(uuid)
			id    := mahjong.PlayerList[index].Index
			return game.Rooms[room].GetDoor(id)
		})

		so.On("getSea", func(room string) ([][]string, bool) {
			if game.Rooms[room] == nil {
				return [][]string{}, true
			}
			return game.Rooms[room].GetSea()
		})

		so.On("getHu", func(room string) ([][]string, bool) {
			if game.Rooms[room] == nil {
				return [][]string{}, true
			}
			return game.Rooms[room].GetHu()
		})

		so.On("getCurrentIdx", func(room string) int {
			if game.Rooms[room] == nil {
				return -1
			}
			return game.Rooms[room].GetCurrentIdx()
		})

		so.On("getScore", func(room string) []int {
			if game.Rooms[room] == nil {
				return []int{}
			}
			return game.Rooms[room].GetScore()
		})

		so.On("disconnection", func() {
			log.Println("on disconnect")
			game.Logout(so)
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", server)

	
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://140.118.127.157:9000"},
		AllowCredentials: true,
	})
	
	
	handler := c.Handler(mux)
	baseURL := "140.118.127.157:3000"

	log.Println("Serving at", baseURL, "...")
	log.Fatal(http.ListenAndServe(baseURL, handler))
}