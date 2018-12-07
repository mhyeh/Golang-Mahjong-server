package main

import (
	"log"
	"net/http"
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io"
	"github.com/rs/cors"

	"util"
	"PManager"
)

func main() {
	rand.Seed(time.Now().Unix())

	game := util.NewGameManager()
	util.InitHuTable()

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

			index := PManager.FindPlayerByUUID(uuid)
			if index == -1 {
				return -1
			}

			player := PManager.Players[index]
			if (player.State & (PManager.MATCHED | PManager.READY | PManager.PLAYING)) != 0 && !(room == "") && player.Room == room {
				so.Join(room)
			}
			player.Socket = &so
			return player.State
		})

		so.On("ready", func(room string, uuid string) int {
			if !PManager.Auth(room, uuid) {
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
			index := PManager.FindPlayerByUUID(uuid)
			if index == -1 {
				return "", []string{}, true
			}
			player := PManager.Players[index]
			room   := player.Room
			return room, PManager.GetNameList(PManager.FindPlayersInRoom(room)), false
		})

		so.On("getID", func(uuid string, room string) int {
			if !PManager.Auth(room, uuid) {
				return -1
			}
			index := PManager.FindPlayerByUUID(uuid)
			if PManager.Players[index].State != PManager.READY {
				return -1
			}
			return PManager.Players[index].Index
		})

		so.On("getReadyPlayer", func(room string) []string {
			playerList := PManager.FindPlayersInRoom(room)
			var nameList []string
			for _, player := range playerList {
				if player.State == PManager.READY {
					nameList = append(nameList, player.Name)
				}
			}
			return nameList
		})

		so.On("getHand", func(uuid string, room string) []string {
			if !PManager.Auth(room, uuid) || game.Rooms[room].State < util.DealCard {
				return []string{}
			}
			index := PManager.FindPlayerByUUID(uuid)
			id    := PManager.Players[index].Index
			return game.Rooms[room].Players[id].Hand.ToStringArray()
		})

		so.On("getPlayerList", func(room string) []string {
			if game.Rooms[room] == nil {
				return []string{}
			}
			return game.Rooms[room].GetPlayerList()
		})

		so.On("getLack", func(room string) []int {
			if game.Rooms[room] == nil || game.Rooms[room].State < util.ChooseLack {
				return []int{}
			}
			var res []int
			for _, player := range game.Rooms[room].Players {
				res = append(res, player.Lack)
			}
			return res
		})

		so.On("getHandCount", func(room string) []int {
			if game.Rooms[room] == nil || game.Rooms[room].State < util.IDTurn {
				return []int{}
			}
			var res []int
			for _, player := range game.Rooms[room].Players {
				res = append(res, int(player.Hand.Count()))
			}
			return res
		})

		so.On("getRemainCount", func(room string) int {
			if game.Rooms[room] == nil || game.Rooms[room].State < util.IDTurn {
				return 56
			}
			return int(game.Rooms[room].Deck.Count())
		})

		so.On("getDoor", func(uuid string, room string) ([][]string, []int, bool) {
			if !PManager.Auth(room, uuid) || game.Rooms[room].State < util.IDTurn {
				return [][]string{}, []int{}, true
			}
			index := PManager.FindPlayerByUUID(uuid)
			id    := PManager.Players[index].Index
			var inVisible []int
			var res       [][]string
			for _, player := range game.Rooms[room].Players {
				if id == player.ID {
					res       = append(res, player.Door.ToStringArray())
					inVisible = append(inVisible, 0)
				} else {
					res = append(res, player.VisiableDoor.ToStringArray())
					inVisible = append(inVisible, int(player.Door.Count() - player.VisiableDoor.Count()))
				}
			}
			return res, inVisible, false
		})

		so.On("getSea", func(room string) ([][]string, bool) {
			if game.Rooms[room] == nil || game.Rooms[room].State < util.IDTurn {
				return [][]string{}, true
			}
			var res [][]string
			for _, player := range game.Rooms[room].Players {
				res = append(res, player.DiscardTiles.ToStringArray())
			}
			return res, false
		})

		so.On("getHu", func(room string) ([][]string, bool) {
			if game.Rooms[room] == nil || game.Rooms[room].State < util.IDTurn {
				return [][]string{}, true
			}
			var res [][]string
			for _, player := range game.Rooms[room].Players {
				res = append(res, player.HuTiles.ToStringArray())
			}
			return res, false
		})

		so.On("getCurrentIdx", func(room string) int {
			if game.Rooms[room] == nil {
				return -1
			}
			id := -1
			if game.Rooms[room].State >= util.IDTurn {
				id = game.Rooms[room].State - util.IDTurn
			}
			return id
		})

		so.On("getScore", func(room string) []int {
			if game.Rooms[room] == nil {
				return []int{}
			}
			var res []int
			for _, player := range game.Rooms[room].Players {
				res = append(res, player.Credit)
			}
			return res
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

	log.Println("Serving at 140.118.127.157:3000...")
	log.Fatal(http.ListenAndServe("140.118.127.157:3000", handler))
}