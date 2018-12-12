package main

import (
	"log"
	"net/http"
	"math/rand"
	"time"

	"github.com/rs/cors"

	"mahjong"
)

func main() {
	rand.Seed(time.Now().Unix())

	err := mahjong.NewGameManager()
	if err {
		return
	}
	go mahjong.Exec()

	mahjong.GetServer().On("connection", mahjong.SocketConnect)
	mahjong.GetServer().On("error",      mahjong.SocketError)

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", mahjong.GetServer())
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://140.118.127.157:9000"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)
	baseURL := "140.118.127.157:3000"
	log.Println("Serving at", baseURL, "...")
	log.Fatal(http.ListenAndServe(baseURL, handler))
}