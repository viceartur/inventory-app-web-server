package websocket

import (
	"log"
	"net/http"
)

// Web Socket Endpoint without any Parameters
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	addClient(ws)
	go reader(ws)
}
