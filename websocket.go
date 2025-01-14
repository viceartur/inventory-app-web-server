package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Websocket Read Message error:", err)
			return
		}

		log.Println(string(p))

		switch message := string(p); message {
		case "sendMaterial":
			handleSendMaterial(conn, messageType)
		default:
			log.Println("WS unhandled message type:", message)
		}

	}
}

func handleSendMaterial(conn *websocket.Conn, messageType int) {
	db, _ := connectToDB()
	materials, err := getIncomingMaterials(db, 0)

	if err != nil {
		log.Println("WS error getting materials:", err)
		return
	}

	msg, err := json.Marshal(
		Message{
			Type: "incomingMaterialsQty",
			Data: len(materials),
		},
	)
	if err != nil {
		log.Println("WS error encoding JSON:", err)
		return
	}
	if err := conn.WriteMessage(messageType, []byte(msg)); err != nil {
		log.Println("WS Write Message error:", err)
		return
	}
}
