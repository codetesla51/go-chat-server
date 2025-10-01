package middleware

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"chat-server/server/models"
)

const (
	MaxMessagesPerWindow = 5
	RateLimitWindow      = 10 * time.Second
	MaxConnectionsPerIP  = 10
)

var (
	ipConnections = make(map[string]int)
	ipMutex       sync.RWMutex
)

// CanSendMessage checks if client can send a message
func CanSendMessage(c *models.Client) (bool, string) {
	now := time.Now()

	if now.Sub(c.WindowStart) > RateLimitWindow {
		c.MessageCount = 0
		c.WindowStart = now
	}

	if c.MessageCount >= MaxMessagesPerWindow {
		timeLeft := RateLimitWindow - now.Sub(c.WindowStart)
		return false, fmt.Sprintf("Rate limited! Wait %.0f seconds.", timeLeft.Seconds())
	}
	return true, ""
}

// RecordMessage records a message for rate limiting
func RecordMessage(c *models.Client) {
	c.LastMessage = time.Now()
	c.MessageCount++
}

// GetIP extracts IP from connection
func GetIP(conn net.Conn) string {
	addr := conn.RemoteAddr().String()
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// CanAcceptConnection checks if IP can connect
func CanAcceptConnection(ip string) bool {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	count := ipConnections[ip]
	return count < MaxConnectionsPerIP
}

// IncrementIPConnection increments IP connection count
func IncrementIPConnection(ip string) {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	ipConnections[ip]++
}

// DecrementIPConnection decrements IP connection count
func DecrementIPConnection(ip string) {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	if ipConnections[ip] > 0 {
		ipConnections[ip]--
	}
	if ipConnections[ip] == 0 {
		delete(ipConnections, ip)
	}
}
