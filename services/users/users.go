package users

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
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

// AuthUser authenticates a user by checking the username and password against the database
// It returns the authenticated user information if successful, or an error if authentication fails.
func AuthUser(db *sql.DB, user UserJSON) (UserJSON, error) {
	username := user.Username
	password := user.Password

	var actualUser UserDB
	err := db.QueryRow(`
		SELECT user_id, username, password, role FROM users WHERE username = $1
		`, username).Scan(
		&actualUser.UserID,
		&actualUser.Username,
		&actualUser.Password,
		&actualUser.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return UserJSON{}, errors.New("No user found")
		}
		return UserJSON{}, err
	}

	// Compare the provided password with the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(actualUser.Password), []byte(password))
	if err != nil {
		return UserJSON{}, errors.New("Wrong password")
	}

	authUser := UserJSON{
		UserID:   actualUser.UserID,
		Username: actualUser.Username,
		Role:     actualUser.Role,
	}

	return authUser, nil
}

// CreateUser creates a new user in the database with a hashed password
// It returns the created user information or an error if the creation fails.
func CreateUser(db *sql.DB, user UserJSON) (UserJSON, error) {
	// Validate the input
	if user.Username == "" || user.Role == "" {
		return UserJSON{}, errors.New("Username and role are required")
	}

	// Check if the user already exists
	var existingUserID int
	err := db.QueryRow(`SELECT user_id FROM users WHERE username = $1`, user.Username).Scan(&existingUserID)

	if err == nil {
		return UserJSON{}, errors.New("User already exists")
	}

	if err != sql.ErrNoRows {
		return UserJSON{}, err
	}

	// Save the provided password or generate a random one if not provided
	var userPassword string
	if user.Password != "" {
		userPassword = user.Password
	} else {
		userPassword = generateRandomPassword()
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	if err != nil {
		return UserJSON{}, err
	}

	// Insert the new user into the database
	var newUserID int
	err = db.QueryRow(`INSERT INTO users (username, password, role) VALUES ($1, $2, $3) RETURNING user_id`,
		user.Username, string(hashedPassword), user.Role,
	).Scan(&newUserID)

	if err != nil {
		return UserJSON{}, err
	}

	user.UserID = newUserID      // Set the UserID created field
	user.Password = userPassword // Set the random password generated

	return user, nil
}

// Generates a random password of a specified length
func generateRandomPassword() string {
	const passwordLength = 12
	lower := "abcdefghijklmnopqrstuvwxyz"
	upper := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	special := "!@#$%^&*+-"
	all := lower + upper + digits + special

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Ensure at least one character from each category
	password := []byte{
		lower[seededRand.Intn(len(lower))],
		upper[seededRand.Intn(len(upper))],
		digits[seededRand.Intn(len(digits))],
		special[seededRand.Intn(len(special))],
	}

	// Fill the rest of the password with random characters from all sets
	for i := len(password); i < passwordLength; i++ {
		password = append(password, all[seededRand.Intn(len(all))])
	}

	// Shuffle the password
	rand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})

	return string(password)
}
