package models

// Handles all things user related such as creating them, logging in, updating
// attributes, and deleting them

import (
	"github.com/derekdowling/bursa/renaissance/authentication"
	"time"
)

type Role int

// Our role definitions as specified by renaissance/firewall
const (
	Visitor Role = 1 << iota
	Authenticated
)

type User struct {
	Id        int64
	Name      string `sql:"size:255"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Password  string `sql:"size:255"`
	Salt      string `sql:"size:64"`

	Email string `sql:"size:255"`
}

func CreateUser(email string, password string) {

	// hash and salt password
	salt, hash := authentication.CreatePassword(password)

	// create user object
	user := &User{
		Email:    email,
		Salt:     salt,
		Password: hash,
	}

	// create/save user
	db, _ := Connect()
	db.Create(&user)
}

// Test's whether or not a user has authenticated successfully
func AttemptLogin(email string, password string) *User {
	// Todo: this is broken
	db, _ := Connect()
	var user *User
	db.Where("email = ?", email).First(&user)
	// match := authentication.PasswordMatch(password, user.salt, user.hash)
	return user
}
