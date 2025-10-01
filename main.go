package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat-server/server"
	"chat-server/server/ai"
	"chat-server/server/utils"
)

func main() {
	port := ":8080"
	tlsPort := ":8443"

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// TCP listener
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer listener.Close()

	// Optional TLS listener
	var tlsListener net.Listener
	hasTLS := false
	if fileExists("server.crt") && fileExists("server.key") {
		cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
		if err != nil {
			log.Println("Failed to load TLS certificate:", err)
		} else {
			config := &tls.Config{Certificates: []tls.Certificate{cert}}
			tlsListener, err = tls.Listen("tcp", tlsPort, config)
			if err != nil {
				log.Println("Failed to listen on TLS port:", err)
			} else {
				hasTLS = true
				defer tlsListener.Close()
				fmt.Println(utils.ColorGreen + "TLS enabled on port 8443" + utils.ColorReset)
			}
		}
	}

	// Initialize server
	srv := server.NewServer()
	srv.Start()

	// Display startup banner
	displayStartupBanner(port)

	// Initialize AI (optional)
	if err := ai.InitAI(); err != nil {
		fmt.Println(utils.ColorYellow + "    [!] AI features disabled (no API key)" + utils.ColorReset)
	} else {
		fmt.Println(utils.ColorGreen + "    [✓] AI features enabled" + utils.ColorReset)
	}

	fmt.Println(utils.ColorCyan + "\n    >> Server ready - Waiting for connections...\n" + utils.ColorReset)

	// TCP accept loop
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					log.Println("Failed to accept TCP connection:", err)
					continue
				}
			}
			go srv.HandleConnection(conn)
		}
	}()

	// TLS accept loop (if enabled)
	if hasTLS {
		fmt.Println(utils.ColorGreen + "TLS listener successfully started on 8443" + utils.ColorReset)
		go func() {
			for {
				conn, err := tlsListener.Accept()
				if err != nil {
					select {
					case <-ctx.Done():
						return
					default:
						log.Println("Failed to accept TLS connection:", err)
						continue
					}
				}
				go srv.HandleConnection(conn)
			}
		}()
	}

	// Wait for shutdown signal
	<-ctx.Done()
	fmt.Println(utils.ColorYellow + "Server is shutting down..." + utils.ColorReset)

	listener.Close()
	if hasTLS {
		tlsListener.Close()
	}
	srv.Shutdown()

	fmt.Println(utils.ColorGreen + "GoodBye" + utils.ColorReset)
}

func displayStartupBanner(port string) {
	bannerLines := []string{
		"",
		"  ═════════════════════════════════════════════",
		"",
		"    ██████╗  ██████╗       ██████╗██╗  ██╗ █████╗ ████████╗",
		"   ██╔════╝ ██╔═══██╗     ██╔════╝██║  ██║██╔══██╗╚══██╔══╝",
		"   ██║  ███╗██║   ██║     ██║     ███████║███████║   ██║   ",
		"   ██║   ██║██║   ██║     ██║     ██╔══██║██╔══██║   ██║   ",
		"   ╚██████╔╝╚██████╔╝     ╚██████╗██║  ██║██║  ██║   ██║   ",
		"    ╚═════╝  ╚═════╝       ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ",
		"",
		"              " + utils.Bold + "Realtime Chat Server v1.0" + utils.ColorReset + utils.ColorPurple,
		"",
		"  ═════════════════════════════════════════════",
		"",
	}

	fmt.Print("\033[2J\033[H") // clear screen

	for _, line := range bannerLines {
		fmt.Println(utils.ColorPurple + line + utils.ColorReset)
		time.Sleep(35 * time.Millisecond)
	}

	fmt.Println(utils.ColorCyan + "    Port:        " + utils.ColorGold + port + utils.ColorReset)
	fmt.Println(utils.ColorCyan + "    Lobby:       " + utils.ColorGreen + "general" + utils.ColorReset)
	fmt.Println(utils.ColorCyan + "    Protocol:    " + utils.ColorWhite + "TCP" + utils.ColorReset)
	fmt.Println()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
