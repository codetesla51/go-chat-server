package server

import (
	"fmt"
	"net"
	"strings"
	"os"
	"log"
)

func handleCommand(conn net.Conn, cmd string, client *Client) {
	// Allow quit without rate limit
	if cmd == "/quit" {
		conn.Write([]byte(ColorYellow + "Disconnecting from server. Goodbye!\n" + ColorReset))
		conn.Close()
		return
	}

	// Rate limit all other commands
	canSend, errMsg := client.canSendMessage()
	if !canSend {
		conn.Write([]byte(ColorRed + "⚠ " + errMsg + ColorReset + "\n"))
		return
	}
	client.recordMessage()

	switch {
	case cmd == "/users":
		showLobbyUsers(conn, client)
	case cmd == "/help":
		showHelpMessage(conn)
	case strings.HasPrefix(cmd, "/ai "):
		content := strings.TrimPrefix(cmd, "/ai ")
		userText := strings.TrimSpace(content)
		if userText == "" {
			conn.Write([]byte(ColorRed + "Usage: /ai <your question>\n" + ColorReset))
			return
		}

		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			conn.Write([]byte(ColorRed + "AI API Key not configured\n" + ColorReset))
			return
		}

		conn.Write([]byte(ColorMagenta + "[AI] Thinking...\n" + ColorReset))

		// Broadcast to lobby that user asked AI
		broadcastLobbyMessage(client.currentLobby,
			fmt.Sprintf("%s%s%s asked AI: %s", ColorCyan, client.username, ColorReset, userText))

		reply, err := handleAichatWithContext(apiKey, userText, client.currentLobby, client.username)

errMsg := "AI Error: Please try again later." 

if err != nil {
    log.Printf("AI error for user %s: %v", client.username, err) 
    e := err.Error()
    switch {
    case strings.Contains(e, "rate limit"):
        errMsg = "AI Error: Rate limit reached. Please wait and try again."
    case strings.Contains(e, "quota"):
        errMsg = "AI Error: Quota reached. Try later."
    case strings.Contains(e, "invalid prompt"):
        errMsg = "AI Error: Your prompt is invalid."
    }

    conn.Write([]byte(ColorRed + errMsg + "\n" + ColorReset))
    return
}
		// Broadcast AI response to lobby
		broadcastLobbyMessage(client.currentLobby,
			fmt.Sprintf("%s[AI Response to %s]%s\n%s", ColorMagenta, client.username, ColorReset, reply))

	case cmd == "/lobbies":
		showAllLobbies(conn)
			case strings.HasPrefix(cmd, "/tag"):
		content := strings.TrimPrefix(cmd, "/tag ")
		parts := strings.SplitN(content, " ", 2)
		if len(parts) < 2 {
			conn.Write([]byte(ColorRed + "Usage: /tag <username> <message>\n" + ColorReset))
			return
		}
		target := parts[0]
		message := parts[1]
		handleTagMessage(client,target, message)
	case strings.HasPrefix(cmd, "/setai "):
		content := strings.TrimSpace(strings.TrimPrefix(cmd, "/setai "))
		handleSetAIPrompt(conn, client, content)
	case strings.HasPrefix(cmd, "/msg "):
		content := strings.TrimPrefix(cmd, "/msg ")
		parts := strings.SplitN(content, " ", 2)
		if len(parts) < 2 {
			conn.Write([]byte(ColorRed + "Usage: /msg <username> <message>\n" + ColorReset))
			return
		}
		target := parts[0]
		message := parts[1]
		handlePrivateMessage(client, target, message)
	case strings.HasPrefix(cmd, "/create "):
		content := strings.TrimSpace(strings.TrimPrefix(cmd, "/create "))
		parts := strings.SplitN(content, " ", 2)
		lobbyName := parts[0]
		password := ""
		if len(parts) == 2 {
			password = parts[1]
		}
		handleCreateLobby(conn, client, lobbyName, password)
	case strings.HasPrefix(cmd, "/join "):
		content := strings.TrimSpace(strings.TrimPrefix(cmd, "/join "))
		parts := strings.SplitN(content, " ", 2)
		lobbyName := parts[0]
		password := ""
		if len(parts) == 2 {
			password = parts[1]
		}
		handleJoinLobby(conn, client, lobbyName, password)
	case strings.HasPrefix(cmd, "/sp"):
		content := strings.TrimSpace(strings.TrimPrefix(cmd, "/sp"))
		if content == "" || content == "default" {
			client.userProfile = profilePics["default"]
			conn.Write([]byte(ColorGreen + "Profile picture reset to default.\n" + ColorReset))
			return
		}
		if content == "list" {
			showProfilePics(conn)
			return
		}
		pic, exists := profilePics[content]
		if !exists {
			conn.Write([]byte(ColorRed + "Profile picture not found. Use /sp list to see available options.\n" + ColorReset))
			return
		}
		client.userProfile = pic
		conn.Write([]byte(ColorGreen + fmt.Sprintf("Profile picture changed to: %s\n", pic) + ColorReset))
	default:
		conn.Write([]byte(ColorRed + "Unknown command. Type /help to see available commands.\n" + ColorReset))
	}
}
