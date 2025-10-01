package handlers

import (
	
	"net"
	"sync"

	"chat-server/server/models"
)

// ClientManager manages connected clients
type ClientManager struct {
	clients           map[net.Conn]*models.Client
	clientsByUsername map[string]*models.Client
	mu                sync.RWMutex
}

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:           make(map[net.Conn]*models.Client),
		clientsByUsername: make(map[string]*models.Client),
	}
}

// AddClient adds a new client
func (cm *ClientManager) AddClient(conn net.Conn, client *models.Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[conn] = client
	cm.clientsByUsername[client.Username] = client
}

// RemoveClient removes a client
func (cm *ClientManager) RemoveClient(conn net.Conn) *models.Client {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	client, exists := cm.clients[conn]
	if exists {
		delete(cm.clients, conn)
		delete(cm.clientsByUsername, client.Username)
	}
	return client
}

// GetClient gets client by connection
func (cm *ClientManager) GetClient(conn net.Conn) *models.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients[conn]
}

// GetClientByUsername gets client by username
func (cm *ClientManager) GetClientByUsername(username string) *models.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clientsByUsername[username]
}

// IsUsernameTaken checks if username is already in use
func (cm *ClientManager) IsUsernameTaken(username string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, exists := cm.clientsByUsername[username]
	return exists
}

// GetLobbyUsers returns all users in a lobby
func (cm *ClientManager) GetLobbyUsers(lobbyName string) []*models.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	var users []*models.Client
	for _, client := range cm.clients {
		if client.CurrentLobby == lobbyName {
			users = append(users, client)
		}
	}
	return users
}

// BroadcastToLobby broadcasts a message to all users in a lobby
func (cm *ClientManager) BroadcastToLobby(lobbyName string, text string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var deadConns []net.Conn
	message := ColorBlue + ColorBold + "[LOBBY] " + ColorReset + text + "\n"

	for conn, client := range cm.clients {
		if client.CurrentLobby == lobbyName {
			_, err := client.Conn.Write([]byte(message))
			if err != nil {
				deadConns = append(deadConns, conn)
			}
		}
	}

	for _, conn := range deadConns {
		delete(cm.clients, conn)
		conn.Close()
	}
}

// BroadcastMessage broadcasts a user message to lobby
func (cm *ClientManager) BroadcastMessage(msg *models.Message, formatFn func(string, string, string, string, string, string, string) string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var deadConns []net.Conn

	for conn, client := range cm.clients {
		if client.CurrentLobby == msg.From.CurrentLobby {
			formattedMsg := formatFn(msg.From.UserProfile, msg.From.Username, msg.Text,
				ColorYellow, ColorWhite, ColorCyan, ColorReset)
			_, err := client.Conn.Write([]byte(formattedMsg))
			if err != nil {
				deadConns = append(deadConns, conn)
			}
		}
	}

	for _, conn := range deadConns {
		delete(cm.clients, conn)
		conn.Close()
	}
}
