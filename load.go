package main

import (
	"bufio"
	"fmt"

	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	connectedUsers   int64
	messagesSent     int64
	messagesReceived int64
	errors           int64
	startTime        time.Time
)

// Simulated user behavior
type User struct {
	id       int
	conn     net.Conn
	username string
}

var messages = []string{
	"Hey everyone!",
	"What's up?",
	"Anyone here?",
	"lol",
	"This is cool",
	"How's it going?",
	"Nice server!",
	"Testing testing",
	"Hello world",
	"Anyone wanna chat?",
	"This is awesome",
	"brb",
	"back",
	"gg",
	"👋",
}

var commands = []string{
	"/users",
	"/lobbies",
	"/sp cat",
	"/sp cool",
	"/theme matrix",
	"/theme cyberpunk",
}

func main() {
	serverAddr := "localhost:8080"
	numUsers := 10

	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Println("║   CHAT SERVER LOAD TESTER                 ║")
	fmt.Println("╚═══════════════════════════════════════════╝")
	fmt.Printf("\nTarget: %s\n", serverAddr)
	fmt.Printf("Spawning %d simulated users...\n\n", numUsers)

	startTime = time.Now()

	// Start stats reporter
	go reportStats()

	// Spawn users gradually
	var wg sync.WaitGroup

	// Spawn in waves to avoid overwhelming the server instantly
	batchSize := 100
	for i := 0; i < numUsers; i += batchSize {
		end := i + batchSize
		if end > numUsers {
			end = numUsers
		}

		for j := i; j < end; j++ {
			wg.Add(1)
			go spawnUser(j, serverAddr, &wg)
			time.Sleep(time.Millisecond * 2) // Small delay between spawns
		}

		fmt.Printf("Spawned batch %d-%d\n", i, end)
		time.Sleep(time.Millisecond * 100) // Delay between batches
	}

	// Keep test running for a while
	fmt.Println("\n✓ All users spawned. Running for 60 seconds...\n")
	time.Sleep(60 * time.Second)

	fmt.Println("\n⏹ Test complete. Shutting down...")
	printFinalStats()
}

func spawnUser(id int, serverAddr string, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		atomic.AddInt64(&errors, 1)
		return
	}
	defer conn.Close()

	user := &User{
		id:       id,
		conn:     conn,
		username: fmt.Sprintf("user%d", id),
	}

	atomic.AddInt64(&connectedUsers, 1)

	// Set username
	_, err = conn.Write([]byte(user.username + "\n"))
	if err != nil {
		atomic.AddInt64(&errors, 1)
		return
	}

	// Start reading messages
	go readMessages(user)

	// Simulate user behavior
	simulateUserBehavior(user)
}

func readMessages(user *User) {
	scanner := bufio.NewScanner(user.conn)
	for scanner.Scan() {
		atomic.AddInt64(&messagesReceived, 1)
	}
}

func simulateUserBehavior(user *User) {
	// Random behavior: some users are active, some lurk
	activityLevel := rand.Float64()

	if activityLevel < 0.3 {
		// 30% are lurkers (read only, rarely send)
		time.Sleep(time.Duration(rand.Intn(120)) * time.Second)
		return
	}

	// Active users send messages periodically
	ticker := time.NewTicker(time.Duration(2+rand.Intn(10)) * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(30+rand.Intn(60)) * time.Second)

	for {
		select {
		case <-ticker.C:
			// 80% chance to send message, 20% chance to send command
			if rand.Float64() < 0.8 {
				sendMessage(user)
			} else {
				sendCommand(user)
			}
		case <-timeout:
			return
		}
	}
}

func sendMessage(user *User) {
	msg := messages[rand.Intn(len(messages))]
	_, err := user.conn.Write([]byte(msg + "\n"))
	if err != nil {
		atomic.AddInt64(&errors, 1)
		return
	}
	atomic.AddInt64(&messagesSent, 1)
}

func sendCommand(user *User) {
	cmd := commands[rand.Intn(len(commands))]
	_, err := user.conn.Write([]byte(cmd + "\n"))
	if err != nil {
		atomic.AddInt64(&errors, 1)
		return
	}
	atomic.AddInt64(&messagesSent, 1)
}

func reportStats() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		printCurrentStats()
	}
}

func printCurrentStats() {
	elapsed := time.Since(startTime).Seconds()
	connected := atomic.LoadInt64(&connectedUsers)
	sent := atomic.LoadInt64(&messagesSent)
	received := atomic.LoadInt64(&messagesReceived)
	errs := atomic.LoadInt64(&errors)

	msgRate := float64(sent) / elapsed

	fmt.Printf("⏱ %.0fs | 👥 %d users | 📤 %d sent (%.1f msg/s) | 📥 %d received | ❌ %d errors\n",
		elapsed, connected, sent, msgRate, received, errs)
}

func printFinalStats() {
	elapsed := time.Since(startTime).Seconds()
	connected := atomic.LoadInt64(&connectedUsers)
	sent := atomic.LoadInt64(&messagesSent)
	received := atomic.LoadInt64(&messagesReceived)
	errs := atomic.LoadInt64(&errors)

	fmt.Println("\n╔═══════════════════════════════════════════╗")
	fmt.Println("║   FINAL RESULTS                           ║")
	fmt.Println("╚═══════════════════════════════════════════╝")
	fmt.Printf("\n⏱  Duration: %.1f seconds\n", elapsed)
	fmt.Printf("👥 Connected Users: %d\n", connected)
	fmt.Printf("📤 Messages Sent: %d (%.1f msg/s)\n", sent, float64(sent)/elapsed)
	fmt.Printf("📥 Messages Received: %d\n", received)
	fmt.Printf("❌ Errors: %d\n", errs)

	if errs > 0 {
		fmt.Printf("\n⚠️  Error rate: %.2f%%\n", float64(errs)/float64(sent)*100)
	}

	// Performance assessment
	fmt.Println("\n📊 Performance Assessment:")
	if errs < 10 && connected > 4000 {
		fmt.Println("   🔥 EXCELLENT - Server handled load like a boss!")
	} else if errs < 100 && connected > 3000 {
		fmt.Println("   ✅ GOOD - Server performed well under pressure")
	} else if errs < 500 && connected > 2000 {
		fmt.Println("   ⚠️  MODERATE - Server struggled but survived")
	} else {
		fmt.Println("   ❌ POOR - Server needs optimization")
	}

	fmt.Println()
}
