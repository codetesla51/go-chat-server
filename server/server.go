package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"chat-server/server/handlers"
	"chat-server/server/middleware"
	"chat-server/server/models"
	"chat-server/server/utils"
)

// Server represents the chat server
type Server struct {
	clientManager  *handlers.ClientManager
	lobbyManager   *handlers.LobbyManager
	commandHandler *handlers.CommandHandler
	messages       chan *models.Message
}

// NewServer creates a new chat server instance
func NewServer() *Server {
	cm := handlers.NewClientManager()
	lm := handlers.NewLobbyManager()
	ch := handlers.NewCommandHandler(cm, lm)

	return &Server{
		clientManager:  cm,
		lobbyManager:   lm,
		commandHandler: ch,
		messages:       make(chan *models.Message, 100),
	}
}

// Start starts the server components
func (s *Server) Start() {
	s.lobbyManager.CreateDefaultLobby()
	go s.broadcastMessages()
	go s.lobbyManager.CleanupInactiveContexts()
}

// HandleConnection handles a new client connection
func (s *Server) HandleConnection(conn net.Conn) {
	ip := middleware.GetIP(conn)

	if !middleware.CanAcceptConnection(ip) {
		conn.Write([]byte(utils.ColorRed + "Too many connections from your IP. Try again later.\n" + utils.ColorReset))
		conn.Close()
		return
	}

	middleware.IncrementIPConnection(ip)
	defer middleware.DecrementIPConnection(ip)

	sendWelcomeBanner(conn)
	defer func() {
		conn.Close()
		if client := s.clientManager.RemoveClient(conn); client != nil {
			fmt.Printf("%s disconnected from the server\n", client.Username)
			s.clientManager.BroadcastToLobby(client.CurrentLobby,
				fmt.Sprintf("%s%s%s has left the lobby", utils.ColorRed, client.Username, utils.ColorReset))
		}
	}()

	scanner := bufio.NewScanner(conn)
	var username string

	// Username selection
	for {
		conn.Write([]byte(utils.ColorYellow + "Enter your username: " + utils.ColorReset))
		if scanner.Scan() {
			username = strings.TrimSpace(scanner.Text())
		}
		if username == "" {
			username = conn.RemoteAddr().String()
		}

		valid, errMsg := utils.IsValidUsername(username)
		if !valid {
			conn.Write([]byte(utils.ColorRed + errMsg + "\n" + utils.ColorReset))
			continue
		}

		if !s.clientManager.IsUsernameTaken(username) {
			break
		}
		conn.Write([]byte(utils.ColorRed + "Username already taken, try another.\n" + utils.ColorReset))
	}

	newClient := &models.Client{
		Conn:         conn,
		Username:     username,
		UserProfile:  "[@_@]",
		CurrentLobby: "general",
		WindowStart:  time.Now(),
	}

	s.clientManager.AddClient(conn, newClient)
	fmt.Printf("%s connected to the server (lobby: general)\n", username)
	s.clientManager.BroadcastToLobby("general",
		fmt.Sprintf("%s%s%s has joined the lobby", utils.ColorGreen, username, utils.ColorReset))

	recent := s.lobbyManager.GetRecentMessages(newClient.CurrentLobby, 5*time.Minute)
	conn.Write([]byte(recent))

	// Read messages from client
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		if len(text) > utils.MaxMessageLength {
			conn.Write([]byte(utils.ColorRed + fmt.Sprintf("Message too long (max %d chars)\n", utils.MaxMessageLength) + utils.ColorReset))
			continue
		}

		if strings.HasPrefix(text, "/") {
			s.commandHandler.HandleCommand(conn, text, newClient)
			continue
		}

		canSend, errMsg := middleware.CanSendMessage(newClient)
		if !canSend {
			conn.Write([]byte(utils.ColorRed + "⚠ " + errMsg + utils.ColorReset + "\n"))
			continue
		}

		middleware.RecordMessage(newClient)
		s.lobbyManager.StoreMessage(newClient.CurrentLobby, newClient.UserProfile, newClient.Username, text)

		s.messages <- &models.Message{
			From:      newClient,
			Text:      text,
			Timestamp: time.Now(),
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Connection error:", err)
	}
}

func (s *Server) broadcastMessages() {
	for msg := range s.messages {
		s.clientManager.BroadcastMessage(msg, func(profile, username, text, colorYellow, colorWhite, colorCyan, colorReset string) string {
			return utils.FormatMessage(profile, username, text, colorYellow, colorWhite, colorCyan, colorReset, msg.Timestamp)
		})
	}
}

func sendWelcomeBanner(conn net.Conn) {
	bannerLines := []string{
		"╔══════════════════════════════════════════════════════════════╗",
		"║                                                              ║",
		"║     ██████╗  ██████╗        ██████╗██╗  ██╗ █████╗ ████████╗║",
		"║    ██╔════╝ ██╔═══██╗      ██╔════╝██║  ██║██╔══██╗╚══██╔══╝║",
		"║    ██║  ███╗██║   ██║█████╗██║     ███████║███████║   ██║   ║",
		"║    ██║   ██║██║   ██║╚════╝██║     ██╔══██║██╔══██║   ██║   ║",
		"║    ╚██████╔╝╚██████╔╝      ╚██████╗██║  ██║██║  ██║   ██║   ║",
		"║     ╚═════╝  ╚═════╝        ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ║",
		"║                                                              ║",
		"║              " + utils.Bold + ">> Welcome to the Ultimate Chat Server <<" + utils.ColorReset + utils.ColorPurple + "      ║",
		"║                                                              ║",
		"║    ┌────────────────────────────────────────────────────┐   ║",
		"║    │  " + utils.Underline + "/help" + utils.ColorReset + utils.ColorPurple + "     - Display available commands            │   ║",
		"║    │  " + utils.Underline + "/users" + utils.ColorReset + utils.ColorPurple + "    - Show who's online                     │   ║",
		"║    │  " + utils.Underline + "/quit" + utils.ColorReset + utils.ColorPurple + "     - Disconnect from server                │   ║",
		"║    └────────────────────────────────────────────────────┘   ║",
		"║                                                              ║",
		"╚══════════════════════════════════════════════════════════════╝",
	}

	conn.Write([]byte("\033[2J\033[H"))

	// Send each line with slight delay for animation effect
	for _, line := range bannerLines {
		conn.Write([]byte(utils.ColorPurple + line + utils.ColorReset + "\n"))
		time.Sleep(30 * time.Millisecond)
	}

	// Blinking status indicator
	conn.Write([]byte("\n"))
	statusMsg := "    " + utils.ColorGold + utils.Bold + "[" + utils.Blink + "●" + utils.ColorReset + utils.ColorGold + utils.Bold +
		" CONNECTED]" + utils.ColorReset + " " + utils.ColorCyan +
		"You're live! Start chatting now...\n\n" + utils.ColorReset
	conn.Write([]byte(statusMsg))
}
func (s *Server) Shutdown() {
	msg := utils.BuildColor("server is shutting Down GoodBye", "yellow")

	allClients := s.clientManager.ClientsSnapshot()
	for _, client := range allClients {
		client.Conn.Write([]byte(msg))
		time.Sleep(50 * time.Millisecond)
		client.Conn.Close()
	}

	close(s.messages)
}
