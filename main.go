package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var clients = make(map[*websocket.Conn]string)
var payload = make(chan message)
var msg message

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	ws.ReadJSON(&msg)
	clients[ws] = msg.Username

	log.Printf("Client Connected: %v", msg.Username)
	log.Printf("Clients: %v", clients)
	go writer()
	reader(ws)
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	for {
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("readError: %v", err)
			delete(clients, ws)
			break
		}
		payload <- msg
	}
}

func writer() {
	for {
		msg := <-payload
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("msgError: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWs)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
