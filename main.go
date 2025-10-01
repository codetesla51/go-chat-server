package main

import (
	"chat-server/server"
	"fmt"
	"log"
	"net"
)

func main() {
	port := ":8080"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()
	
	server.CreateGenralLobby()

	// Initialize AI (optional - won't fail if no API key)
	if err := server.InitAI(); err != nil {
		fmt.Println(server.ColorYellow + "⚠ AI features disabled (no API key)" + server.ColorReset)
	} else {
		fmt.Println(server.ColorGreen + "✓ AI features enabled" + server.ColorReset)
	}

	fmt.Println(server.ColorGreen + "==================================" + server.ColorReset)
	fmt.Println(server.ColorCyan + "  GO CHAT SERVER RUNNING" + server.ColorReset)
	fmt.Println(server.ColorGreen + "==================================" + server.ColorReset)
	fmt.Printf("Listening on port %s%s%s\n", server.ColorYellow, port, server.ColorReset)
	fmt.Println("Default lobby 'general' created")
	fmt.Println("Rate limiting enabled:")
	fmt.Printf("  - Max %d messages per %v\n", server.MaxMessagesPerWindow, server.RateLimitWindow)
	fmt.Printf("  - Max %d connections per IP\n", server.MaxConnectionsPerIP)
	fmt.Println("Waiting for connections...\n")

	go server.BroadcastMessages()
	go server.CleanupInactiveLobbyContexts()  // ADD THIS LINE

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go server.HandleConnection(conn)
	}
}