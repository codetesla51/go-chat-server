package server

import (
	"fmt"
	"net"
)

func isUsernameTaken(username string) bool {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	for _, client := range clients {
		if client.username == username {
			return true
		}
	}
	return false
}
func showHelpMessage(conn net.Conn) {
	helpMsg := ColorCyan + "\n=== Available Commands ===\n" + ColorReset
	helpMsg += "  /users  - Show users in current lobby\n"
	helpMsg += "  /lobbies - List all lobbies\n"
	helpMsg += "  /create <name> [password] - Create new lobby\n"
	helpMsg += "  /join <name> [password] - Join a lobby\n"
	helpMsg += "  /sp <name> - Set profile picture\n"
	helpMsg += "  /sp list - List available profile pictures\n"
	helpMsg += "  /msg <user> <message> - Send private message\n"
	helpMsg += "  /ai <question> - Ask AI a question\n"
	helpMsg += "  /ai clear - Clear AI conversation history\n"
	helpMsg += "  /setai <prompt> - Set custom AI prompt for lobby (creator only)\n"
	helpMsg += "  /showai - Show current lobby's AI prompt\n"
	helpMsg += "  /quit   - Disconnect from server\n\n"
	conn.Write([]byte(helpMsg))
}

func formatMessage(senderProfile string, username, text string) string {
	content := fmt.Sprintf("%s%s %s%s\n  %s╰─>%s %s\n",
		ColorYellow,
		senderProfile,
		username,
		ColorReset,
		ColorCyan,
		ColorReset,
		text,
	)
	return content
}
func showAllLobbies(conn net.Conn) {
	lobbiesMutex.RLock()
	defer lobbiesMutex.RUnlock()

	msg := ColorCyan + fmt.Sprintf("\n=== Available Lobbies (%d) ===\n", len(lobbies)) + ColorReset

	for name, lobby := range lobbies {
		userCount := 0
		clientsMutex.RLock()
		for _, client := range clients {
			if client.currentLobby == name {
				userCount++
			}
		}
		clientsMutex.RUnlock()

		privacyText := "public"
		if lobby.isPrivate {
			privacyText = "private"
		}

		aiStatus := "default AI"
		if lobby.aiPrompt != "" {
			aiStatus = "custom AI"
		}

		msg += fmt.Sprintf("  %s%-15s%s [%s] [%s] - %d users (created by %s)\n",
			ColorWhite, name, ColorReset, privacyText, aiStatus, userCount, lobby.creator)
	}
	msg += "\n"
	conn.Write([]byte(msg))
}

func showLobbyUsers(conn net.Conn, client *Client) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	var lobbyUsers []*Client
	for _, c := range clients {
		if c.currentLobby == client.currentLobby {
			lobbyUsers = append(lobbyUsers, c)
		}
	}

	msg := ColorCyan + fmt.Sprintf("\n=== Users in '%s' (%d) ===\n", client.currentLobby, len(lobbyUsers)) + ColorReset
	for _, user := range lobbyUsers {
		msg += fmt.Sprintf("  %s %s%s%s\n", user.userProfile, ColorWhite, user.username, ColorReset)
	}
	msg += "\n"
	conn.Write([]byte(msg))
}

func showProfilePics(conn net.Conn) {
	msg := ColorCyan + "\n=== Available Profile Pictures ===\n" + ColorReset
	msg += ColorYellow + "Usage: /sp <name>\n\n" + ColorReset
	for name, pic := range profilePics {
		msg += fmt.Sprintf("  %s%-10s%s → %s\n", ColorWhite, name, ColorReset, pic)
	}
	msg += "\n"
	conn.Write([]byte(msg))
}
func removeClient(conn net.Conn) {
	clientsMutex.Lock()
	client, exists := clients[conn]
	if exists {
		delete(clients, conn)
	}
	clientsMutex.Unlock()

	if exists {
		fmt.Printf("%s disconnected from the server\n", client.username)
		broadcastLobbyMessage(client.currentLobby, fmt.Sprintf("%s%s%s has left the lobby", ColorRed, client.username, ColorReset))
	}
}
func getAIPromptForLobby(lobbyName string) string {
	lobbiesMutex.RLock()
	defer lobbiesMutex.RUnlock()

	lobby, exists := lobbies[lobbyName]
	if !exists || lobby.aiPrompt == "" {
		return DefaultAiGuideline
	}
	return lobby.aiPrompt
}
func sendWelcomeBanner(conn net.Conn) {
	banner := `
╔══════════════════════════════════════╗
║   WELCOME TO GO CHAT SERVER          ║
║                                      ║
║   Type /help to get started          ║
╚══════════════════════════════════════╝

`
	conn.Write([]byte(ColorCyan + banner + ColorReset))
}
