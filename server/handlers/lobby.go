package handlers

import (
	"chat-server/server/models"
	"fmt"
	"net"
	"strings"
	"time"
)

func (h *CommandHandler) handleCreateLobby(conn net.Conn, client *models.Client, cmd string) {
	content := strings.TrimSpace(strings.TrimPrefix(cmd, "/create "))
	parts := strings.SplitN(content, " ", 3)

	if len(parts) < 2 {
		conn.Write([]byte(ColorRed + "Usage: /create <name> [password] <description>\n" + ColorReset))
		return
	}

	lobbyName := parts[0]
	password := ""
	desc := parts[len(parts)-1]

	if len(parts) == 3 {
		password = parts[1]
		desc = parts[2]
	}

	if err := h.LobbyManager.CreateLobby(lobbyName, password, desc, client.Username); err != nil {
		conn.Write([]byte(ColorRed + err.Error() + "\n" + ColorReset))
		return
	}

	lobbyType := "public"
	if password != "" {
		lobbyType = "private"
	}

	conn.Write([]byte(ColorGreen + fmt.Sprintf("Created %s lobby '%s'. Use /join %s to enter.\n",
		lobbyType, lobbyName, lobbyName) + ColorReset))
}

func (h *CommandHandler) handleJoinLobby(conn net.Conn, client *models.Client, cmd string) {
	content := strings.TrimSpace(strings.TrimPrefix(cmd, "/join "))
	parts := strings.SplitN(content, " ", 2)

	lobbyName := parts[0]
	password := ""
	if len(parts) == 2 {
		password = parts[1]
	}

	if err := h.LobbyManager.JoinLobby(lobbyName, password); err != nil {
		conn.Write([]byte(ColorRed + err.Error() + "\n" + ColorReset))
		return
	}

	oldLobby := client.CurrentLobby
	h.ClientManager.BroadcastToLobby(oldLobby,
		fmt.Sprintf("%s%s%s has left the lobby", ColorRed, client.Username, ColorReset))

	client.CurrentLobby = lobbyName
	conn.Write([]byte(ColorGreen + fmt.Sprintf("Joined lobby '%s'\n", lobbyName) + ColorReset))

	h.ClientManager.BroadcastToLobby(lobbyName,
		fmt.Sprintf("%s%s%s has joined the lobby", ColorGreen, client.Username, ColorReset))
	recent := h.LobbyManager.GetRecentMessages(lobbyName, 10*time.Minute)
	conn.Write([]byte(recent))
}

func (h *CommandHandler) handleSetAI(conn net.Conn, client *models.Client, cmd string) {
	prompt := strings.TrimSpace(strings.TrimPrefix(cmd, "/setai "))

	if err := h.LobbyManager.SetAIPrompt(client.CurrentLobby, client.Username, prompt); err != nil {
		conn.Write([]byte(ColorRed + err.Error() + "\n" + ColorReset))
		return
	}

	conn.Write([]byte(ColorGreen + "AI prompt updated!\n" + ColorReset))
	h.ClientManager.BroadcastToLobby(client.CurrentLobby,
		fmt.Sprintf("%s%s%s updated the AI prompt", ColorYellow, client.Username, ColorReset))
}
