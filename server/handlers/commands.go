package handlers

import (
	"chat-server/server/ai"
	"chat-server/server/middleware"
	"chat-server/server/models"
	"context "
	"fmt"
	"log"
	"net"
	"strings"
)

// CommandHandler holds dependencies for command handling
type CommandHandler struct {
	ClientManager *ClientManager
	LobbyManager  *LobbyManager
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(cm *ClientManager, lm *LobbyManager) *CommandHandler {
	return &CommandHandler{
		ClientManager: cm,
		LobbyManager:  lm,
	}
}

// HandleCommand processes user commands
func (h *CommandHandler) HandleCommand(conn net.Conn, cmd string, client *models.Client) {
	if cmd == "/quit" {
		conn.Write([]byte(ColorYellow + "Disconnecting from server. Goodbye!\n" + ColorReset))
		conn.Close()
		return
	}

	canSend, errMsg := middleware.CanSendMessage(client)
	if !canSend {
		conn.Write([]byte(ColorRed + "⚠ " + errMsg + ColorReset + "\n"))
		return
	}
	middleware.RecordMessage(client)

	switch {
	case cmd == "/users":
		h.showLobbyUsers(conn, client)
	case cmd == "/help":
		showHelpMessage(conn)
	case strings.HasPrefix(cmd, "/ai "):
		h.handleAICommand(conn, client, cmd)
	case cmd == "/lobbies":
		h.LobbyManager.ShowAllLobbies(conn)
	case strings.HasPrefix(cmd, "/tag "):
		h.handleTagCommand(conn, client, cmd)
	case strings.HasPrefix(cmd, "/setai "):
		h.handleSetAI(conn, client, cmd)
	case strings.HasPrefix(cmd, "/msg "):
		h.handlePrivateMessage(conn, client, cmd)
	case strings.HasPrefix(cmd, "/create "):
		h.handleCreateLobby(conn, client, cmd)
	case strings.HasPrefix(cmd, "/join "):
		h.handleJoinLobby(conn, client, cmd)
	case strings.HasPrefix(cmd, "/sp"):
		h.handleSetProfile(conn, client, cmd)
	default:
		conn.Write([]byte(ColorRed + "Unknown command. Type /help for available commands.\n" + ColorReset))
	}
}

func (h *CommandHandler) handleAICommand(conn net.Conn, client *models.Client, cmd string) {
	content := strings.TrimPrefix(cmd, "/ai ")
	userText := strings.TrimSpace(content)

	if userText == "" {
		conn.Write([]byte(ColorRed + "Usage: /ai <your question>\n" + ColorReset))
		return
	}

	if len(userText) > 1000 {
		conn.Write([]byte(ColorRed + "AI question too long. Max: 1000 characters\n" + ColorReset))
		return
	}

	if ai.GetAPIKey() == "" {
		conn.Write([]byte(ColorRed + "AI is not available on this server\n" + ColorReset))
		return
	}

	conn.Write([]byte(ColorMagenta + "[AI] Thinking...\n" + ColorReset))
	h.ClientManager.BroadcastToLobby(client.CurrentLobby,
		fmt.Sprintf("%s%s%s asked AI: %s", ColorCyan, client.Username, ColorReset, userText))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reply, err := ai.HandleAIChat(ctx, userText, client.CurrentLobby, client.Username,
		h.LobbyManager.GetConversations(), h.LobbyManager.GetConversationsMutex(),
		h.LobbyManager.GetLobbyContext)

	if err != nil {
		log.Printf("AI error for user %s: %v", client.Username, err)
		errMsg := ai.FormatAIError(err)
		conn.Write([]byte(ColorRed + errMsg + "\n" + ColorReset))
		return
	}

	h.ClientManager.BroadcastToLobby(client.CurrentLobby,
		fmt.Sprintf("%s[AI Response to %s]%s\n%s", ColorMagenta, client.Username, ColorReset, reply))
}

func (h *CommandHandler) showLobbyUsers(conn net.Conn, client *models.Client) {
	users := h.ClientManager.GetLobbyUsers(client.CurrentLobby)
	msg := ColorCyan + fmt.Sprintf("\n=== Users in '%s' (%d) ===\n", client.CurrentLobby, len(users)) + ColorReset
	for _, user := range users {
		msg += fmt.Sprintf("  %s %s%s%s\n", user.UserProfile, ColorWhite, user.Username, ColorReset)
	}
	msg += "\n"
	conn.Write([]byte(msg))
}

func showHelpMessage(conn net.Conn) {
	helpMsg := ColorCyan + "\n=== Available Commands ===\n" + ColorReset
	helpMsg += "  /users  - Show users in current lobby\n"
	helpMsg += "  /lobbies - List all lobbies\n"
	helpMsg += "  /create <name> [password] <desc> - Create new lobby\n"
	helpMsg += "  /join <name> [password] - Join a lobby\n"
	helpMsg += "  /sp <name> - Set profile picture\n"
	helpMsg += "  /sp list - List available profile pictures\n"
	helpMsg += "  /msg <user> <message> - Send private message\n"
	helpMsg += "  /tag <user> <message> - Tag someone in lobby\n"
	helpMsg += "  /ai <question> - Ask AI a question\n"
	helpMsg += "  /setai <prompt> - Set custom AI (creator only)\n"
	helpMsg += "  /quit   - Disconnect from server\n\n"
	conn.Write([]byte(helpMsg))
}

// Color constants
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
)
