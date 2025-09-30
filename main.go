package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

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
	
	// Rate limiting constants
	MaxMessagesPerWindow = 5
	RateLimitWindow      = 10 * time.Second
	MaxConnectionsPerIP  = 10
)

// Theme struct defines color schemes
type Theme struct {
	name       string
	primary    string
	secondary  string
	accent     string
	text       string
	border     string
}

// Available themes
var themes = map[string]Theme{
	"default": {
		name:      "default",
		primary:   ColorCyan,
		secondary: ColorMagenta,
		accent:    ColorGreen,
		text:      ColorWhite,
		border:    ColorBlue,
	},
	"matrix": {
		name:      "matrix",
		primary:   "\033[92m",
		secondary: "\033[32m",
		accent:    "\033[92m",
		text:      "\033[32m",
		border:    "\033[92m",
	},
	"cyberpunk": {
		name:      "cyberpunk",
		primary:   "\033[95m",
		secondary: "\033[96m",
		accent:    "\033[93m",
		text:      "\033[97m",
		border:    "\033[95m",
	},
	"ocean": {
		name:      "ocean",
		primary:   "\033[94m",
		secondary: "\033[96m",
		accent:    "\033[36m",
		text:      "\033[97m",
		border:    "\033[34m",
	},
	"sunset": {
		name:      "sunset",
		primary:   "\033[91m",
		secondary: "\033[93m",
		accent:    "\033[95m",
		text:      "\033[97m",
		border:    "\033[33m",
	},
	"hacker": {
		name:      "hacker",
		primary:   "\033[32m",
		secondary: "\033[90m",
		accent:    "\033[37m",
		text:      "\033[37m",
		border:    "\033[32m",
	},
}

// Predefined profile pictures
var profilePics = map[string]string{
	"default": "[@_@]",
	"cat":     "(=^･^=)",
	"dog":     "(ᵔᴥᵔ)",
	"cool":    "(⌐■_■)",
	"bear":    "ʕ•ᴥ•ʔ",
	"happy":   "(◕‿◕)",
	"star":    "☆彡",
	"fire":    "(🔥)",
	"alien":   "[👽]",
	"robot":   "[▪‿▪]",
	"love":    "(♥‿♥)",
	"wink":    "(^_~)",
	"dead":    "(x_x)",
	"shrug":   "¯\\_(ツ)_/¯",
	"music":   "♪(┌・。・)┌",
	"ninja":   "[忍]",
	"king":    "(♔‿♔)",
	"queen":   "(♕‿♕)",
	"devil":   "(ψ｀∇´)ψ",
	"angel":   "(◕ᴗ◕✿)",
	"sleep":   "(-.-)zzZ",
	"cry":     "(╥﹏╥)",
	"laugh":   "(≧▽≦)",
	"angry":   "(╬ಠ益ಠ)",
	"confused": "(・_・ヾ",
	"shocked":  "(°ロ°)",
	"peace":    "(✌ﾟ∀ﾟ)☞",
	"skull":    "[☠]",
	"heart":    "[❤]",
	"coffee":   "c[_]",
	"pizza":    "[🍕]",
	"ghost":    "(ー'`ー)",
	"fox":      "ᓚᘏᗢ",
	"owl":      "(◉Θ◉)",
	"penguin":  "(°<°)",
	"frog":     "( ･ั﹏･ั)",
	"bunny":    "(\\(•ᴗ•)/)",
	"snake":    "~>°)~~~",
	"dino":     "<コ:彡",
	"wizard":   "⊂(◉‿◉)つ",
	"pirate":   "(✪‿✪)ノ",
	"nerd":     "(⌐□_□)",
	"party":    "ヽ(^o^)ノ",
	"think":    "(¬‿¬)",
	"flex":     "ᕦ(ò_óˇ)ᕤ",
	"dance":    "┏(･o･)┛",
	"flip":     "(ノಠ益ಠ)ノ彡┻━┻",
}

// Lobby struct represents a chat lobby/room
type Lobby struct {
	name      string
	isPrivate bool
	password  string
	creator   string
}

// Client struct to store client information
type Client struct {
	username     string
	userProfile  string
	currentLobby string
	theme        Theme
	conn         net.Conn
	lastMessage  time.Time
	messageCount int
	windowStart  time.Time
}

// Global variables
var (
	clients        = make(map[net.Conn]*Client)
	lobbies        = make(map[string]*Lobby)
	ipConnections  = make(map[string]int)
	clientsMutex   sync.RWMutex
	lobbiesMutex   sync.RWMutex
	ipMutex        sync.RWMutex
	messages       = make(chan Message, 100)
)

// Message struct for broadcasting
type Message struct {
	from Client
	text string
}

// Rate limiting methods
func (c *Client) canSendMessage() (bool, string) {
	now := time.Now()
	
	// Reset window if it's expired
	if now.Sub(c.windowStart) > RateLimitWindow {
		c.messageCount = 0
		c.windowStart = now
	}
	
	// Check if too many messages in window
	if c.messageCount >= MaxMessagesPerWindow {
		timeLeft := RateLimitWindow - now.Sub(c.windowStart)
		return false, fmt.Sprintf("Rate limited! Wait %.0f seconds.", timeLeft.Seconds())
	}
		return true, ""
		
	}
	


func (c *Client) recordMessage() {
	c.lastMessage = time.Now()
	c.messageCount++
}

func getIP(conn net.Conn) string {
	addr := conn.RemoteAddr().String()
	// Extract IP without port
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

func canAcceptConnection(ip string) bool {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	
	count := ipConnections[ip]
	return count < MaxConnectionsPerIP
}

func incrementIPConnection(ip string) {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	ipConnections[ip]++
}

func decrementIPConnection(ip string) {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	if ipConnections[ip] > 0 {
		ipConnections[ip]--
	}
	if ipConnections[ip] == 0 {
		delete(ipConnections, ip)
	}
}

func handleConnection(conn net.Conn) {
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
		theme:        themes["default"],
		windowStart:  time.Now(),
	}

	clientsMutex.Lock()
	clients[conn] = newClient
	clientsMutex.Unlock()

	fmt.Printf("%s connected to the server (lobby: general)\n", username)
	broadcastLobbyMessage("general", fmt.Sprintf("%s%s%s has joined the lobby", ColorGreen, username, ColorReset))

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
		messages <- Message{from: *newClient, text: text}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Connection error:", err)
	}
}

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
	case cmd == "/lobbies":
		showAllLobbies(conn)
	case cmd == "/themes":
		showThemes(conn, client)
	case strings.HasPrefix(cmd, "/theme "):
		themeName := strings.TrimSpace(strings.TrimPrefix(cmd, "/theme "))
		handleThemeChange(conn, client, themeName)
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
		conn.Write([]byte(ColorRed + "Unknown command. Type a command to see available options.\n" + ColorReset))
	}
}

func showHelpMessage(conn net.Conn){
  	// Send help message
	helpMsg := ColorCyan + "\n=== Available Commands ===\n" + ColorReset
	helpMsg += "  /users  - Show users in current lobby\n"
	helpMsg += "  /lobbies - List all lobbies\n"
	helpMsg += "  /create <name> [password] - Create new lobby\n"
	helpMsg += "  /join <name> [password] - Join a lobby\n"
	helpMsg += "  /sp <name> - Set profile picture\n"
	helpMsg += "  /sp list - List available profile pictures\n"
	helpMsg += "  /msg <user> <message> - Send private message\n"
	helpMsg += "  /themes - List available themes"
	helpMsg += "  /theme <name> - set theme \n"
	helpMsg += "  /quit   - Disconnect from server\n\n"
	conn.Write([]byte(helpMsg))
}
func handleThemeChange(conn net.Conn, client *Client, themeName string) {
	theme, exists := themes[themeName]
	if !exists {
		conn.Write([]byte(ColorRed + "Theme not found. Use /themes to see available themes.\n" + ColorReset))
		return
	}
	
	client.theme = theme
	msg := theme.accent + fmt.Sprintf("Theme changed to '%s'!\n", themeName) + ColorReset
	conn.Write([]byte(msg))
}

func showThemes(conn net.Conn, client *Client) {
	msg := client.theme.primary + "\n╔═══════════════════════════════════╗\n"
	msg += "║     Available Themes              ║\n"
	msg += "╚═══════════════════════════════════╝\n" + ColorReset
	msg += ColorYellow + "Usage: /theme <name>\n\n" + ColorReset
	
	for name, theme := range themes {
		indicator := " "
		if name == client.theme.name {
			indicator = "→"
		}
		msg += fmt.Sprintf("%s %s%-12s%s - %sPreview text%s\n", 
			indicator, ColorWhite, name, ColorReset, theme.primary, ColorReset)
	}
	msg += "\n"
	conn.Write([]byte(msg))
}

func formatMessage(client *Client, username, text string) string {
	t := client.theme
	
	border := t.border + "┌─────────────────────────────────────────┐\n" + ColorReset
	
	content := fmt.Sprintf("%s│%s %s %s%s%s: %s%s%s │%s\n",
		t.border,
		ColorReset,
		client.userProfile,
		t.secondary,
		username,
		ColorReset,
		t.text,
		text,
		t.border,
		ColorReset)
	
	border += content
	border += t.border + "└─────────────────────────────────────────┘\n" + ColorReset
	
	return border
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
	}
	lobbies[lobbyName] = newLobby

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

		msg += fmt.Sprintf("  %s %s%-15s%s [%s] - %d users (created by %s)\n", 
			ColorWhite, name, ColorReset, privacyText, userCount, lobby.creator)
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

func broadcastMessages() {
	for msg := range messages {
		clientsMutex.Lock()
		var deadConns []net.Conn
		
		for conn, client := range clients {
			if client.currentLobby == msg.from.currentLobby {
				formattedMsg := formatMessage(client, msg.from.username, msg.text)
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
	var target *Client
	for _, client := range clients {
		if client.username == targetName {
			target = client
			break
		}
	}
	clientsMutex.RUnlock()

	if target == nil {
		sender.conn.Write([]byte(ColorRed + "User not found on server.\n" + ColorReset))
		return
	}

	targetMsg := fmt.Sprintf("%s[PM from %s%s%s]:%s %s\n",
		ColorMagenta,
		ColorWhite,
		sender.username,
		ColorMagenta,
		ColorReset,
		message)
	target.conn.Write([]byte(targetMsg))

	senderMsg := fmt.Sprintf("%s[PM to %s%s%s]:%s %s\n",
		ColorMagenta,
		ColorWhite,
		targetName,
		ColorMagenta,
		ColorReset,
		message)
	sender.conn.Write([]byte(senderMsg))
}

func sendWelcomeBanner(conn net.Conn) {
	banner := `
╔══════════════════════════════════════╗
║   WELCOME TO GO CHAT SERVER          ║
║                                      ║
║   Type commands to get started       ║
╚══════════════════════════════════════╝

`
	conn.Write([]byte(ColorCyan + banner + ColorReset))
}

func main() {
	port := ":8080"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	lobbies["general"] = &Lobby{
		name:      "general",
		isPrivate: false,
		password:  "",
		creator:   "server",
	}

	fmt.Println(ColorGreen + "==================================" + ColorReset)
	fmt.Println(ColorCyan + "  GO CHAT SERVER RUNNING" + ColorReset)
	fmt.Println(ColorGreen + "==================================" + ColorReset)
	fmt.Printf("Listening on port %s%s%s\n", ColorYellow, port, ColorReset)
	fmt.Println("Default lobby 'general' created")
	fmt.Println("Rate limiting enabled:")
	fmt.Printf("  - Max %d messages per %v\n", MaxMessagesPerWindow, RateLimitWindow)
	fmt.Printf("  - Max %d connections per IP\n", MaxConnectionsPerIP)
	fmt.Println("Waiting for connections...\n")

	go broadcastMessages()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}