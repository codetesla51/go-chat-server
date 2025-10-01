package models

import (
	"net"
	"time"
)

// Lobby represents a chat room
type Lobby struct {
	Name      string
	IsPrivate bool
	Password  string
	Creator   string
	Desc      string
	AIPrompt  string
}

// Client represents a connected user
type Client struct {
	Username     string
	UserProfile  string
	CurrentLobby string
	Conn         net.Conn
	LastMessage  time.Time
	MessageCount int
	WindowStart  time.Time
}

// LobbyMessage represents a message in a lobby
type LobbyMessage struct {
	Username    string
	Text        string
	UserProfile string
	Timestamp   time.Time
}

// LobbyContext stores recent messages for context
type LobbyContext struct {
	RecentMessages []LobbyMessage
	Mu             interface{} // sync.RWMutex
}

// Message struct for broadcasting
type Message struct {
	From      *Client
	Text      string
	Timestamp time.Time
}
