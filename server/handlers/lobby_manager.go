package handlers

import (
	"fmt"
	"net"
	"sync"
	"time"

	"chat-server/server/ai"
	"chat-server/server/models"
	"chat-server/server/utils"
)

// LobbyManager manages chat lobbies
type LobbyManager struct {
	lobbies            map[string]*models.Lobby
	lobbyContexts      map[string]*models.LobbyContext
	lobbyConversations map[string]*ai.ConversationHistory
	mu                 sync.RWMutex
	contextMu          sync.RWMutex
	conversationsMu    sync.RWMutex
}

// NewLobbyManager creates a new lobby manager
func NewLobbyManager() *LobbyManager {
	return &LobbyManager{
		lobbies:            make(map[string]*models.Lobby),
		lobbyContexts:      make(map[string]*models.LobbyContext),
		lobbyConversations: make(map[string]*ai.ConversationHistory),
	}
}

// CreateDefaultLobby creates the general lobby
func (lm *LobbyManager) CreateDefaultLobby() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.lobbies["general"] = &models.Lobby{
		Name:      "general",
		IsPrivate: false,
		Password:  "",
		Creator:   "server",
		Desc:      "Welcome to the General Lobby — this is where everyone spawns when they enter the server.",
		AIPrompt:  "",
	}
}

// CreateLobby creates a new lobby
func (lm *LobbyManager) CreateLobby(name, password, desc, creator string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, exists := lm.lobbies[name]; exists {
		return fmt.Errorf("lobby already exists")
	}

	lm.lobbies[name] = &models.Lobby{
		Name:      name,
		IsPrivate: password != "",
		Password:  password,
		Creator:   creator,
		Desc:      desc,
		AIPrompt:  "",
	}
	return nil
}

// JoinLobby validates lobby join request
func (lm *LobbyManager) JoinLobby(name, password string) error {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lobby, exists := lm.lobbies[name]
	if !exists {
		return fmt.Errorf("lobby does not exist")
	}

	if lobby.IsPrivate && lobby.Password != password {
		return fmt.Errorf("incorrect password for private lobby")
	}

	return nil
}

// SetAIPrompt sets custom AI prompt for a lobby
func (lm *LobbyManager) SetAIPrompt(lobbyName, username, prompt string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lobby, exists := lm.lobbies[lobbyName]
	if !exists {
		return fmt.Errorf("lobby not found")
	}

	if lobbyName == "general" {
		return fmt.Errorf("cannot set prompt for general lobby")
	}

	if lobby.Creator != username {
		return fmt.Errorf("only the lobby creator can set the AI prompt")
	}

	lobby.AIPrompt = prompt
	ai.SetAIPromptForLobby(lobbyName, prompt)
	return nil
}

// ShowAllLobbies displays all lobbies to a connection
func (lm *LobbyManager) ShowAllLobbies(conn net.Conn) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	msg := ColorCyan + fmt.Sprintf("\n=== Available Lobbies (%d) ===\n\n", len(lm.lobbies)) + ColorReset

	for name, lobby := range lm.lobbies {
		privacyText := "public"
		if lobby.IsPrivate {
			privacyText = "private"
		}

		aiStatus := "default AI"
		if lobby.AIPrompt != "" {
			aiStatus = "custom AI"
		}

		desc := lobby.Desc
		if desc == "" {
			desc = "No description"
		}

		msg += fmt.Sprintf("%sLobby: %s%s\n", ColorWhite, name, ColorReset)
		msg += fmt.Sprintf("  Privacy: %s | AI: %s | Created by: %s\n", privacyText, aiStatus, lobby.Creator)
		msg += fmt.Sprintf("  Description: %s\n\n", desc)
	}

	conn.Write([]byte(msg))
}

// StoreMessage stores a message in lobby context
func (lm *LobbyManager) StoreMessage(lobbyName, userProfile, username, text string) {
	lm.contextMu.Lock()
	defer lm.contextMu.Unlock()

	ctx, exists := lm.lobbyContexts[lobbyName]
	if !exists {
		ctx = &models.LobbyContext{
			RecentMessages: []models.LobbyMessage{},
			Mu:             &sync.RWMutex{},
		}
		lm.lobbyContexts[lobbyName] = ctx
	}

	mu := ctx.Mu.(*sync.RWMutex)
	mu.Lock()
	defer mu.Unlock()

	ctx.RecentMessages = append(ctx.RecentMessages, models.LobbyMessage{
		Username:    username,
		Text:        text,
		UserProfile: userProfile,
		Timestamp:   time.Now(),
	})

	if len(ctx.RecentMessages) > 5 {
		ctx.RecentMessages = ctx.RecentMessages[len(ctx.RecentMessages)-5:]
	}
}

// GetLobbyContext returns formatted lobby context
func (lm *LobbyManager) GetLobbyContext(lobbyName string) string {
	lm.contextMu.RLock()
	ctx, exists := lm.lobbyContexts[lobbyName]
	lm.contextMu.RUnlock()

	if !exists || ctx == nil {
		return ""
	}

	mu := ctx.Mu.(*sync.RWMutex)
	mu.RLock()
	defer mu.RUnlock()

	if len(ctx.RecentMessages) == 0 {
		return ""
	}

	var result string
	for _, msg := range ctx.RecentMessages {
		result += fmt.Sprintf("%s: %s\n", msg.Username, msg.Text)
	}
	return result
}

// GetRecentMessages returns recent messages from lobby
func (lm *LobbyManager) GetRecentMessages(lobbyName string, duration time.Duration) string {
	lm.contextMu.RLock()
	ctx, exists := lm.lobbyContexts[lobbyName]
	lm.contextMu.RUnlock()

	if !exists || ctx == nil {
		return ""
	}

	mu := ctx.Mu.(*sync.RWMutex)
	mu.RLock()
	defer mu.RUnlock()

	if len(ctx.RecentMessages) == 0 {
		return ""
	}

	since := time.Now().Add(-duration)
	var result string

	for _, msg := range ctx.RecentMessages {
		if msg.Timestamp.Before(since) {
			continue
		}
		result += utils.FormatMessage(
    msg.UserProfile,
    msg.Username,
    msg.Text,
    utils.ColorYellow,
    utils.ColorWhite,
    utils.ColorCyan,
    utils.ColorReset,
    msg.Timestamp,
)
	}

	return result
}

// GetConversations returns the conversations map (for AI)
func (lm *LobbyManager) GetConversations() map[string]*ai.ConversationHistory {
	return lm.lobbyConversations
}

// GetConversationsMutex returns the mutex for conversations
func (lm *LobbyManager) GetConversationsMutex() interface{} {
	return &lm.conversationsMu
}

// CleanupInactiveContexts removes old lobby contexts
func (lm *LobbyManager) CleanupInactiveContexts() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		lm.contextMu.Lock()
		for name, ctx := range lm.lobbyContexts {
			if name == "general" {
				continue
			}

			mu := ctx.Mu.(*sync.RWMutex)
			mu.RLock()
			shouldDelete := false
			if len(ctx.RecentMessages) > 0 {
				lastMsg := ctx.RecentMessages[len(ctx.RecentMessages)-1]
				if time.Since(lastMsg.Timestamp) > 2*time.Hour {
					shouldDelete = true
				}
			}
			mu.RUnlock()

			if shouldDelete {
				delete(lm.lobbyContexts, name)
			}
		}
		lm.contextMu.Unlock()
	}
}
