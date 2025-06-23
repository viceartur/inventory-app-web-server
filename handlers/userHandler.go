package handlers

import (
	"encoding/json"
	"inv_app/database"
	"inv_app/services/users"
	"net/http"
)

// Auth
func AuthUsersHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var user users.User
	json.NewDecoder(r.Body).Decode(&user)
	authUser, err := users.AuthUser(db, user)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusUnauthorized)
		return
	}
	res := SuccessResponse{Message: "User authenticated", Data: authUser}
	json.NewEncoder(w).Encode(res)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	db, _ := database.ConnectToDB()
	defer db.Close()
	var user users.User
	json.NewDecoder(r.Body).Decode(&user)
	user, err := users.CreateUser(db, user)

	if err != nil {
		errRes := ErrorResponse{Message: err.Error()}
		res, _ := json.Marshal(errRes)
		http.Error(w, string(res), http.StatusBadRequest)
		return
	}
	res := SuccessResponse{Message: "User created", Data: user}
	json.NewEncoder(w).Encode(res)
}
