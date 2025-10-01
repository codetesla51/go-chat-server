package main

import (
	"fmt"
	"log"
	"net"

	"chat-server/server"
	"chat-server/server/ai"
)

func main() {
	port := ":8080"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	// Initialize server
	srv := server.NewServer()
	srv.Start()

	// Initialize AI (optional)
	if err := ai.InitAI(); err != nil {
		fmt.Println("\033[33m⚠ AI features disabled (no API key)\033[0m")
	} else {
		fmt.Println("\033[32m✓ AI features enabled\033[0m")
	}

	fmt.Println("\033[32m==================================\033[0m")
	fmt.Println("\033[36m  GO CHAT SERVER RUNNING\033[0m")
	fmt.Println("\033[32m==================================\033[0m")
	fmt.Printf("Listening on port \033[33m%s\033[0m\n", port)
	fmt.Println("Default lobby 'general' created")
	fmt.Println("Waiting for connections...\n")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go srv.HandleConnection(conn)
	}
}
