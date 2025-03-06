package users

import (
	"database/sql"
	"errors"
)

type UserJSON struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

type UserDB struct {
	UserID   int    `field:"user_id"`
	Username string `field:"username"`
	Password string `field:"password"`
	Role     string `field:"role"`
}

func AuthUser(db *sql.DB, user UserJSON) (UserJSON, error) {
	username := user.Username
	password := user.Password

	var actualUser UserDB
	db.QueryRow(`
		SELECT user_id, username, password, role FROM users WHERE username = $1
		`, username).Scan(
		&actualUser.UserID,
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

	authUser := UserJSON{
		UserID:   actualUser.UserID,
		Username: actualUser.Username,
		Role:     actualUser.Role,
	}

	return authUser, nil
}
