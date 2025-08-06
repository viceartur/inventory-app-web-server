package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Web Socket Endpoint without any Parameters
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Println("WebSocket connection attempt from:", r.RemoteAddr)

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	vars := mux.Vars(r)
	userRole := UserRole(vars["userRole"])

	log.Printf("New WebSocket client connected. Role: %s, IP: %s\n", userRole, r.RemoteAddr)

	addClient(ws, userRole)
	go reader(ws)
}
