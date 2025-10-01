package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func HandleConnection(conn net.Conn) {
	ip := getIP(conn)

	// Check IP connection limit
	if !canAcceptConnection(ip) {
		conn.Write([]byte(ColorRed + "Too many connections from your IP. Please try again later.\n" + ColorReset))
		conn.Close()
		return
	}

	incrementIPConnection(ip)
	defer decrementIPConnection(ip)

	sendWelcomeBanner(conn)
	defer func() {
		conn.Close()
		removeClient(conn)
	}()

	scanner := bufio.NewScanner(conn)
	var username string

	// Username selection loop
	for {
		message := ColorYellow + "Enter your username: " + ColorReset
		conn.Write([]byte(message))
		if scanner.Scan() {
			username = strings.TrimSpace(scanner.Text())
		}
		if username == "" {
			username = conn.RemoteAddr().String()
		}

		// Check if username is taken
		if !isUsernameTaken(username) {
			break
		}
		message = ColorRed + "Username already taken, please try another.\n" + ColorReset
		conn.Write([]byte(message))
	}

	newClient := &Client{
		conn:         conn,
		username:     username,
		userProfile:  profilePics["default"],
		currentLobby: "general",
		windowStart:  time.Now(),
	}

	clientsMutex.Lock()
	clients[conn] = newClient
	clientsByUsername[username] = newClient
	clientsMutex.Unlock()

	fmt.Printf("%s connected to the server (lobby: general)\n", username)
	broadcastLobbyMessage("general", fmt.Sprintf("%s%s%s has joined the lobby", ColorGreen, username, ColorReset))
recent := getRecentMessages("general", 5*time.Minute)
conn.Write([]byte(recent ))
	// Read messages from client
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(text, "/") {
			handleCommand(conn, text, newClient)
			continue
		}

		// Rate limiting for regular messages
		canSend, errMsg := newClient.canSendMessage()
		if !canSend {
			conn.Write([]byte(ColorRed + "⚠ " + errMsg + ColorReset + "\n"))
			continue
		}

		newClient.recordMessage()

		// Store message in lobby context for AI awareness
		storeLobbyMessage(newClient.currentLobby,newClient.userProfile, newClient.username, text)

		messages <- Message{from: *newClient, text: text}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Connection error:", err)
	}
}
func handleCreateLobby(conn net.Conn, client *Client, lobbyName string, password string) {
	if lobbyName == "" {
		conn.Write([]byte(ColorRed + "Usage: /create <lobby_name> [password]\n" + ColorReset))
		return
	}

	lobbiesMutex.Lock()
	defer lobbiesMutex.Unlock()

	if _, exists := lobbies[lobbyName]; exists {
		conn.Write([]byte(ColorRed + "Lobby already exists!\n" + ColorReset))
		return
	}

	newLobby := &Lobby{
		name:      lobbyName,
		isPrivate: password != "",
		password:  password,
		creator:   client.username,
		aiPrompt:  "", 
	}
	lobbies[lobbyName] = newLobby
fmt.Print(newLobby)
	lobbyType := "public"
	if newLobby.isPrivate {
		lobbyType = "private"
	}

	conn.Write([]byte(ColorGreen + fmt.Sprintf("Created %s lobby '%s'. Use /join %s to enter.\n", lobbyType, lobbyName, lobbyName) + ColorReset))
	fmt.Printf("%s created %s lobby '%s'\n", client.username, lobbyType, lobbyName)
}

func handleJoinLobby(conn net.Conn, client *Client, lobbyName string, password string) {
	lobbiesMutex.RLock()
	lobby, exists := lobbies[lobbyName]
	lobbiesMutex.RUnlock()

	if !exists {
		conn.Write([]byte(ColorRed + "Lobby does not exist!\n" + ColorReset))
		return
	}

	if lobby.isPrivate && lobby.password != password {
		conn.Write([]byte(ColorRed + "Incorrect password for private lobby!\n" + ColorReset))
		return
	}

	oldLobby := client.currentLobby
	broadcastLobbyMessage(oldLobby, fmt.Sprintf("%s%s%s has left the lobby", ColorRed, client.username, ColorReset))

	client.currentLobby = lobbyName
	conn.Write([]byte(ColorGreen + fmt.Sprintf("Joined lobby '%s'\n", lobbyName) + ColorReset))
	broadcastLobbyMessage(lobbyName, fmt.Sprintf("%s%s%s has joined the lobby", ColorGreen, client.username, ColorReset))

	fmt.Printf("%s moved from '%s' to '%s'\n", client.username, oldLobby, lobbyName)
}

func handleSetAIPrompt(conn net.Conn, client *Client, prompt string) {
	if prompt == "" {
		conn.Write([]byte(ColorRed + "Usage: /setai <prompt>\n" + ColorReset))
		return
	}

	lobbiesMutex.Lock()
	defer lobbiesMutex.Unlock()

	lobby, exists := lobbies[client.currentLobby]
	if !exists {
		conn.Write([]byte(ColorRed + "Lobby not found!\n" + ColorReset))
		return
	}
	if client.currentLobby == "general" {
	  conn.Write([]byte(ColorRed + "Cannot set Prompt for General Lobby" + ColorReset))
	  return 
	}
	// Check if user is the lobby creator
	if lobby.creator != client.username {
		conn.Write([]byte(ColorRed + "Only the lobby creator can set the AI prompt!\n" + ColorReset))
		return
	}

	lobby.aiPrompt = prompt

	// Clear existing conversation history for this lobby
//	conversationsMutex.Lock()
//	delete(lobbyConversations, client.currentLobby)
//	conversationsMutex.Unlock()

	conn.Write([]byte(ColorGreen + "AI prompt updated!.\n" + ColorReset))
	broadcastLobbyMessage(client.currentLobby,
		fmt.Sprintf("%s%s%s updated the AI prompt for this lobby", ColorYellow, client.username, ColorReset))
}
func storeLobbyMessage(lobbyName, userProfile , username, text string) {
	contextsMutex.Lock()
	defer contextsMutex.Unlock()

	ctx, exists := lobbyContexts[lobbyName]
	if !exists {
		ctx = &LobbyContext{
			recentMessages: []LobbyMessage{},
		}
		lobbyContexts[lobbyName] = ctx
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// Add new message
	ctx.recentMessages = append(ctx.recentMessages, LobbyMessage{
		username:  username,
		text:      text,
		userProfile: userProfile,
		timestamp: time.Now(),
	})

	// Keep only last 5 messages
	if len(ctx.recentMessages) > 5 {
		ctx.recentMessages = ctx.recentMessages[len(ctx.recentMessages)-5:]
	}
}

func getLobbyContext(lobbyName string) string {
	contextsMutex.RLock()
	ctx, exists := lobbyContexts[lobbyName]
	contextsMutex.RUnlock()

	if !exists || ctx == nil {
		return ""
	}

	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if len(ctx.recentMessages) == 0 {
		return ""
	}

	var contextStr string


	for _, msg := range ctx.recentMessages {
		username := msg.username
		text := msg.text
		userProfile := msg.userProfile
		contextStr = formatMessage(userProfile,username,text)
	}
return contextStr
}
func getRecentMessages(lobbyName string, duration time.Duration) string {
    contextsMutex.RLock()
    ctx, exists := lobbyContexts[lobbyName]
    contextsMutex.RUnlock()
    if !exists || ctx == nil {
        return ""
    }

    ctx.mu.RLock()
    defer ctx.mu.RUnlock()
    if len(ctx.recentMessages) == 0 {
        return ""
    }

    since := time.Now().Add(-duration)
    var contextStr strings.Builder

    for _, msg := range ctx.recentMessages {
        if msg.timestamp.Before(since) {
            continue
        }
        // Use a formatting function if you want colors or profile icons
        contextStr.WriteString(formatMessage(msg.userProfile, msg.username, msg.text))
    }

    return contextStr.String()
}

func BroadcastMessages() {
	for msg := range messages {
		clientsMutex.Lock()
				
		var deadConns []net.Conn

		for conn, client := range clients {
			if client.currentLobby == msg.from.currentLobby {
				formattedMsg := formatMessage(msg.from.userProfile, msg.from.username, msg.text)
				_, err := client.conn.Write([]byte(formattedMsg))
				if err != nil {
					deadConns = append(deadConns, conn)
				}
			}
		}

		for _, conn := range deadConns {
			delete(clients, conn)
			conn.Close()
		}
			clientsMutex.Unlock()

	}
}

func broadcastLobbyMessage(lobbyName string, text string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	var deadConns []net.Conn

	for conn, client := range clients {
		if client.currentLobby == lobbyName {
			message := ColorBlue + ColorBold + "[LOBBY] " + ColorReset + text + "\n"
			_, err := client.conn.Write([]byte(message))
			if err != nil {
				deadConns = append(deadConns, conn)
			}
		}
	}

	for _, conn := range deadConns {
		delete(clients, conn)
		conn.Close()
	}
}



func handlePrivateMessage(sender *Client, targetName, message string) {
	clientsMutex.RLock()
defer clientsMutex.RUnlock()
	target, exists := clientsByUsername[targetName]
    if !exists {
        sender.conn.Write([]byte(ColorRed + "User not found.\n" + ColorReset))
        return
    }

	// Message to receiver
	targetMsg := fmt.Sprintf("%s[DM]%s %s%s%s —» You\n  ╰─> %s\n",
		ColorMagenta,
		ColorReset,
		ColorCyan,
		sender.username,
		ColorReset,
		message)
	target.conn.Write([]byte(targetMsg))

	// Confirmation to sender
	senderMsg := fmt.Sprintf("%s[DM]%s You —» %s%s%s\n  ╰─> %s\n",
		ColorMagenta,
		ColorReset,
		ColorCyan,
		targetName,
		ColorReset,
		message)
	sender.conn.Write([]byte(senderMsg))
}
func handleTagMessage(sender *Client, targetName, message string) {
    clientsMutex.RLock()
    target, exists := clientsByUsername[targetName]
    clientsMutex.RUnlock()

    if !exists {
        sender.conn.Write([]byte(ColorRed + "User not found.\n" + ColorReset))
        return
    }

    // Format for lobby
    broadcastText := fmt.Sprintf("%s%s —» %s: %s%s", ColorBold, sender.username, targetName,ColorReset, message)
    broadcastLobbyMessage(sender.currentLobby, broadcastText)

    // Notification for tagged user
    if target.conn != nil && target != sender {
        target.conn.Write([]byte(fmt.Sprintf("%s%s replied to you: %s%s\n", ColorMagenta, sender.username, message, ColorReset)))
    }

    // Notification for sender themselves
    sender.conn.Write([]byte(fmt.Sprintf("%sYou%s —» %s: %s\n", ColorMagenta, ColorReset, targetName, message)))
}

func CreateGenralLobby() {
	lobbies["general"] = &Lobby{
		name:      "general",
		isPrivate: false,
		password:  "",
		creator:   "server",
		aiPrompt:  "", // Uses default prompt
	}
}
