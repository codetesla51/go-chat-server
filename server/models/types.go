package server 

import (
	"net"

	"sync"
	"time"
)

type Lobby struct {
	name      string
	isPrivate bool
	password  string
	creator   string
	desc string 
	aiPrompt  string
}

type Client struct {
	username     string
	userProfile  string
	currentLobby string
	conn         net.Conn
	lastMessage  time.Time
	messageCount int
	windowStart  time.Time
}

type ConversationHistory struct {
	messages   []map[string]interface{}
	lastActive time.Time
	mu         sync.RWMutex
}

type LobbyMessage struct {
	username  string
	text      string
	userProfile string
	timestamp time.Time
}

type LobbyContext struct {
	recentMessages []LobbyMessage
	mu             sync.RWMutex
}

// Message struct for broadcasting
type Message struct {
	from Client
	text string
	timestamp time.Time
}

// Response struct for AI API
type Response struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}
