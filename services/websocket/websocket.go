package websocket

import (
	"encoding/json"
	"inv_app/database"
	"inv_app/services/materials"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type UserRole string

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients      = make(map[*websocket.Conn]UserRole)
	clientsMutex sync.Mutex
)

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func addClient(conn *websocket.Conn, userRole UserRole) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	clients[conn] = userRole
}

func removeClient(conn *websocket.Conn) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	delete(clients, conn)
}

func reader(conn *websocket.Conn) {
	defer removeClient(conn)
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage error:", err)
			return
		}

		msgType := string(p)

		switch msgType {
		case "materialsUpdated":
			handleSendMaterial()
		case "vaultUpdated":
			handleSendVault()
		case "requestedMaterialsUpdated":
			isUpdated := true
			handleRequestMaterial(isUpdated)
		case "requestedMaterialsRemoved":
			isUpdated := false
			handleRequestMaterial(isUpdated)
		default:
			log.Println("WS unknown message type:", msgType)
		}
	}
}

func handleSendMaterial() {
	db, _ := database.ConnectToDB()
	count, err := materials.GetIncomingWarehouseMaterialsCount(db)
	if err != nil {
		log.Println("WS error getting warehouse materials count:", err)
		return
	}

	msg := Message{Type: "incomingMaterialsQty", Data: count}
	broadcastMessage(msg)
}

func handleSendVault() {
	db, _ := database.ConnectToDB()
	count, err := materials.GetIncomingVaultMaterialsCount(db)
	if err != nil {
		log.Println("WS error getting vault materials count:", err)
		return
	}

	msg := Message{Type: "incomingVaultQty", Data: count}
	broadcastMessage(msg)
}

// Handle the current requested materials quantity based on the state provided.
func handleRequestMaterial(isUpdated bool) {
	db, _ := database.ConnectToDB()
	count, err := materials.GetRequestedMaterialsCount(db)

	if err != nil {
		log.Println("WS error getting requested materials count:", err)
		return
	}

	var msg Message
	msg.Data = count
	if isUpdated {
		msg.Type = "requestedMaterialsQty" // user will be notified.
	} else {
		msg.Type = "requestedMaterialsQtyRemoved" // no notification.
	}

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

	for client, userRole := range clients {
		// If message type is "requestedMaterialsQty", only send to warehouse group
		// since this socket triggers notification on the client side.
		if message.Type == "requestedMaterialsQty" && userRole != "warehouse" {
			continue
		}

		if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("WS Broadcast WriteMessage error:", err)
			client.Close()
			removeClient(client)
		}
	}
}
