package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	wsActiveClients = make(map[*websocket.Conn]bool)
	clientsMutex    sync.Mutex
)

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func addClient(conn *websocket.Conn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	wsActiveClients[conn] = true
}

func reader(conn *websocket.Conn) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage error:", err)
			return
		}

		log.Println("WS reader recieved:", string(p))

		if string(p) == "materialsUpdated" {
			handleSendMaterial()
		} else {
			handleChatMessage(string(p))
		}
	}
}

func handleChatMessage(message string) {
	// Broadcast the message to all clients
	msg := Message{Type: "chatMessage", Data: message}
	broadcastMessage(msg)
}

func handleSendMaterial() {
	db, _ := connectToDB()
	materials, err := getIncomingMaterials(db, 0)
	if err != nil {
		log.Println("WS error getting materials:", err)
		return
	}

	// Broadcast the message to all clients
	msg := Message{Type: "incomingMaterialsQty", Data: len(materials)}
	broadcastMessage(msg)
}

func broadcastMessage(message Message) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	msg, err := json.Marshal(message)
	if err != nil {
		log.Println("WS Broadcast error encoding message:", err)
		return
	}

	for client := range wsActiveClients {
		if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("WS Broadcast WriteMessage error:", err)
			client.Close()
			delete(wsActiveClients, client)
		}

	}
}
