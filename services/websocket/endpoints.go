package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Web Socket Endpoint without any Parameters
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	vars := mux.Vars(r)
	userRole := UserRole(vars["userRole"])

	addClient(ws, userRole)
	go reader(ws)
}
