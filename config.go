package main

import (
	"errors"
	"net/http"
	"sync"
)

const (
	MaxLenUsername     = 20
	MaxNumRooms        = 2
	MaxNumUsersTotal   = 10
	MaxNumUsersPerRoom = 5
)

var (
	ErrTooManyRooms       = errors.New("too many rooms")
	ErrTooManyUsers       = errors.New("too many users")
	ErrUsernameUsed       = errors.New("the username is used")
	ErrTooManyUsersInRoom = errors.New("too many users in the room")
	ErrUserNotExist       = errors.New("the user does not exist")
	ErrRoomNotExist       = errors.New("the room does not exist")
)

type Config struct {
	mu     *sync.RWMutex
	server *http.Server
	rooms  map[*Room]bool
	users  map[*User]bool
}

func initializeConfig() *Config {
	return &Config{
		mu:     &sync.RWMutex{},
		server: nil,
		rooms:  map[*Room]bool{},
		users:  map[*User]bool{},
	}
}
