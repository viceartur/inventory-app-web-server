package handlers

import (
	"encoding/json"
	"inv_app/v2/database"
	"inv_app/v2/services/users"
	"net/http"
)

// Auth
func AuthUsersHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var user users.UserJSON
	json.NewDecoder(r.Body).Decode(&user)
	authUser, err := users.AuthUser(db, user)

	if err != nil {
		errRes := ErrorResponseJSON{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusUnauthorized)
		return
	}
	res := SuccessResponseJSON{Message: "User authenticated", Data: authUser}
	json.NewEncoder(w).Encode(res)
}
