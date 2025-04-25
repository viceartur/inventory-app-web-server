package websocket

import (
	"encoding/json"
	"inv_app/database"
	"inv_app/services/materials"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients      = make(map[*websocket.Conn]bool)
	clientsMutex sync.Mutex
)

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func addClient(conn *websocket.Conn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	clients[conn] = true
}

func reader(conn *websocket.Conn) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage error:", err)
			return
		}

		msgType := string(p)

		if msgType == "materialsUpdated" {
			handleSendMaterial()
		} else if msgType == "vaultUpdated" {
			handleSendVault()
		}
	}
}

func handleSendMaterial() {
	db, _ := database.ConnectToDB()
	materials, err := materials.GetIncomingMaterials(db, 0)
	count := 0
	for _, material := range materials {
		if material.MaterialType != "CARDS" && material.MaterialType != "CHIPS" {
			count++
		}
	}

	if err != nil {
		log.Println("WS error getting materials:", err)
		return
	}

	// Broadcast the message to all clients
	msg := Message{Type: "incomingMaterialsQty", Data: count}
	broadcastMessage(msg)
}

func handleSendVault() {
	db, _ := database.ConnectToDB()
	materials, err := materials.GetIncomingMaterials(db, 0)
	count := 0
	for _, material := range materials {
		if material.MaterialType == "CARDS" || material.MaterialType == "CHIPS" {
			count++
		}
	}

	if err != nil {
		log.Println("WS error getting materials:", err)
		return
	}

	// Broadcast the message to all clients
	msg := Message{Type: "incomingVaultQty", Data: count}
	broadcastMessage(msg)
}

// Send a message to all connected clients
func broadcastMessage(message Message) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	msg, err := json.Marshal(message)
	if err != nil {
		log.Println("WS Broadcast error encoding message:", err)
		return
	}

	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("WS Broadcast WriteMessage error:", err)
			client.Close()
			delete(clients, client)
		}

	}
}
