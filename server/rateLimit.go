package server

import (
	"fmt"
	"net"
	"strings"
	"time"
)

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

// helper function to split Ip for full IP port
// this is because addr itself contains the port so we need to split,and check if : exit's with the -1
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
