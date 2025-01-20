package main

import (
	"database/sql"
	"errors"
)

type UserJSON struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

type UserDB struct {
	Username string `field:"username"`
	Password string `field:"password"`
	Role     string `field:"role"`
}

func authUser(db *sql.DB, user UserJSON) (UserJSON, error) {
	username := user.Username
	password := user.Password

	var actualUser UserDB
	db.QueryRow(`
		SELECT username, password, role FROM users WHERE username = $1
		`, username).Scan(
		&actualUser.Username,
		&actualUser.Password,
		&actualUser.Role,
	)

	if actualUser.Username == "" {
		return UserJSON{}, errors.New("No user found")
	}

	if password != actualUser.Password {
		return UserJSON{}, errors.New("Wrong password")
	}

	authUser := UserJSON{Username: actualUser.Username, Role: actualUser.Role}

	return authUser, nil
}
