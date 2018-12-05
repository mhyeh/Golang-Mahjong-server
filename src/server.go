package main

import (
	"log"
	"net/http"
	"math/rand"
	"time"

	"github.com/googollee/go-socket.io"
	"github.com/rs/cors"

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
	
	go game.Exec()

	server.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		
		so.On("join", func(name string) (string, bool) {
			_uuid, _err := game.Login(name, so)
			return _uuid, _err
		})

		so.On("auth", func(room string, uuid string) string {
			if !game.PlayerManager.Auth(room, uuid) {
				return "auth failed"
			}

			so.Join(room)
			game.Rooms[room].IO = server
			index := game.PlayerManager.FindPlayerByUUID(uuid)
			game.PlayerManager[index].Socket = &so
			return ""
		})

		so.On("ready", func(room string, uuid string) int {
			if !game.PlayerManager.Auth(room, uuid) {
				return -1
			}

			c := make(chan int)
			fn := func(id int) {
				c<-id
			}
			go game.Rooms[room].Accept(uuid, fn)
			return <-c
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