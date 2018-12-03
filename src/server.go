package main

import (
	"log"
	"net/http"
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io"

	. "util"
)

func main() {
	rand.Seed(time.Now().Unix())

	game := NewGameManager()
	InitHuTable()

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	
	game.Exec()

	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		
		so.On("join", func(name string, callback func(uuid string, err bool)) {
			callback(game.Login(name, so))
		})

		so.On("auth", func(room string, uuid string, callback func(err string)) {
			if !game.PlayerManager.Auth(room, uuid) {
				callback("auth failed")
				return
			}

			so.Join(room)
			game.Rooms[room].IO = server
			index := game.PlayerManager.FindPlayerByUUID(uuid)
			game.PlayerManager[index].Socket = &so
			callback("")
		})

		so.On("ready", func(room string, uuid string, callback func(err int)) {
			if !game.PlayerManager.Auth(room, uuid) {
				callback(-1)
				return
			}

			id := game.Rooms[room].Accept(uuid)
			println("Accept", id)
			callback(id)
		})

		so.On("disconnection", func() {
			log.Println("on disconnect")
			game.Logout(so)
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	http.Handle("/", server)
	log.Println("Serving at localhost:3000...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}