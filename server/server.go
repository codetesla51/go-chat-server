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
		conn.Write([]byte(ColorRed + "Too many connections from your IP. Try again later.\n" + ColorReset))
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
				fmt.Sprintf("%s%s%s has left the lobby", ColorRed, client.Username, ColorReset))
		}
	}()

	scanner := bufio.NewScanner(conn)
	var username string

	// Username selection
	for {
		conn.Write([]byte(ColorYellow + "Enter your username: " + ColorReset))
		if scanner.Scan() {
			username = strings.TrimSpace(scanner.Text())
		}
		if username == "" {
			username = conn.RemoteAddr().String()
		}

		valid, errMsg := utils.IsValidUsername(username)
		if !valid {
			conn.Write([]byte(ColorRed + errMsg + "\n" + ColorReset))
			continue
		}

		if !s.clientManager.IsUsernameTaken(username) {
			break
		}
		conn.Write([]byte(ColorRed + "Username already taken, try another.\n" + ColorReset))
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
		fmt.Sprintf("%s%s%s has joined the lobby", ColorGreen, username, ColorReset))

	recent := s.lobbyManager.GetRecentMessages("general", 5*time.Minute)
	conn.Write([]byte(recent))

	// Read messages from client
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		if len(text) > utils.MaxMessageLength {
			conn.Write([]byte(ColorRed + fmt.Sprintf("Message too long (max %d chars)\n", utils.MaxMessageLength) + ColorReset))
			continue
		}

		if strings.HasPrefix(text, "/") {
			s.commandHandler.HandleCommand(conn, text, newClient)
			continue
		}

		canSend, errMsg := middleware.CanSendMessage(newClient)
		if !canSend {
			conn.Write([]byte(ColorRed + "⚠ " + errMsg + ColorReset + "\n"))
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
	banner := `
╔══════════════════════════════════════╗
║   WELCOME TO GO CHAT SERVER          ║
║                                      ║
║   Type /help to get started          ║
╚══════════════════════════════════════╝

`
	conn.Write([]byte(ColorCyan + banner + ColorReset))
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
