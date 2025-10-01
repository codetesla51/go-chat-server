package main

import (
        "fmt"
        "log"
        "net"
        "time"

        "chat-server/server"
        "chat-server/server/ai"
        "chat-server/server/utils"
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

        // Display startup banner
        displayStartupBanner(port)

        // Initialize AI (optional)
        if err := ai.InitAI(); err != nil {
                fmt.Println(utils.ColorYellow + "    [!] AI features disabled (no API key)" + utils.ColorReset)
        } else {
                fmt.Println(utils.ColorGreen + "    [✓] AI features enabled" + utils.ColorReset)
        }

        fmt.Println(utils.ColorCyan + "\n    >> Server ready - Waiting for connections...\n" + utils.ColorReset)

        for {
                conn, err := listener.Accept()
                if err != nil {
                        log.Println("Failed to accept connection:", err)
                        continue
                }
                go srv.HandleConnection(conn)
        }
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

        // Clear screen
        fmt.Print("\033[2J\033[H")

        // Animate banner
        for _, line := range bannerLines {
                fmt.Println(utils.ColorPurple + line + utils.ColorReset)
                time.Sleep(35 * time.Millisecond)
        }

        // Server info - simple and clean
        fmt.Println(utils.ColorCyan + "    Port:        " + utils.ColorGold + port + utils.ColorReset)
        fmt.Println(utils.ColorCyan + "    Lobby:       " + utils.ColorGreen + "general" + utils.ColorReset)
        fmt.Println(utils.ColorCyan + "    Protocol:    " + utils.ColorWhite + "TCP" + utils.ColorReset)
        fmt.Println()
}