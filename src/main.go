package main

import (
	"io"
	"net/http"

	"chats/src/chat"
	"golang.org/x/net/websocket"
)

func echoHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

func main() {
	hub := chat.CreateHub()
	go hub.Listen()

	http.Handle("/", http.FileServer(http.Dir("ui")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
