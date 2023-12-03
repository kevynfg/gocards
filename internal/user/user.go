package user

import (
	"math/rand"
	"time"
)

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUser(username, password, email string) *User {
	return &User{
		ID:       rand.Intn(1000),
		Username: username,
		Password: password,
		Email:    email,
		CreatedAt: time.Now().UTC(),
	}
}

type Service interface {
	CreateUser(username, password, email string) *User
}

type Repository interface {
	CreateUser(username, password, email string) *User
}