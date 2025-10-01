package middleware

import (
	"net"
	"strings"
	"testing"
	"time"

	"chat-server/server/models"
)

type fakeConn struct{ remote string }

func (f fakeConn) Read([]byte) (int, error)  { return 0, nil }
func (f fakeConn) Write([]byte) (int, error) { return 0, nil }
func (f fakeConn) Close() error              { return nil }
func (f fakeConn) LocalAddr() net.Addr       { return nil }
func (f fakeConn) RemoteAddr() net.Addr      { return fakeAddr{f.remote} }
func (f fakeConn) SetDeadline(time.Time) error      { return nil }
func (f fakeConn) SetReadDeadline(time.Time) error  { return nil }
func (f fakeConn) SetWriteDeadline(time.Time) error { return nil }

type fakeAddr struct{ s string }

func (f fakeAddr) Network() string { return "tcp" }
func (f fakeAddr) String() string  { return f.s }

func TestGetIP(t *testing.T) {
	conn := fakeConn{remote: "127.0.0.1:12345"}
	ip := GetIP(conn)
	if ip != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", ip)
	}
}


func TestCanSendMessage(t *testing.T) {
	now := time.Now()
	c := &models.Client{
		MessageCount: 0,
		WindowStart:  now,
	}

	// Initial message should be allowed
	allowed, msg := CanSendMessage(c)
	if !allowed || msg != "" {
		t.Errorf("expected allowed, got allowed=%v, msg=%q", allowed, msg)
	}

	// Simulate sending max messages
	c.MessageCount = MaxMessagesPerWindow
	allowed, msg = CanSendMessage(c)
	if allowed || !strings.Contains(msg, "Rate limited") {
		t.Errorf("expected rate limited, got allowed=%v, msg=%q", allowed, msg)
	}

	// Simulate window expiration
	c.WindowStart = now.Add(-RateLimitWindow - time.Second)
	allowed, msg = CanSendMessage(c)
	if !allowed || msg != "" {
		t.Errorf("expected allowed after window reset, got allowed=%v, msg=%q", allowed, msg)
	}

	// Test RecordMessage increments count
	c.MessageCount = 0
	RecordMessage(c)
	if c.MessageCount != 1 {
		t.Errorf("expected message count 1, got %d", c.MessageCount)
	}
}


func TestIPConnections(t *testing.T) {
	ipConnections = make(map[string]int) // reset global state

	ip := "1.2.3.4"

	// Increment until max
	for i := 0; i < MaxConnectionsPerIP; i++ {
		if !CanAcceptConnection(ip) {
			t.Errorf("expected to accept connection %d", i)
		}
		IncrementIPConnection(ip)
	}

	// Should now reject
	if CanAcceptConnection(ip) {
		t.Errorf("should reject connection after max reached")
	}

	// Decrement and check
	DecrementIPConnection(ip)
	if !CanAcceptConnection(ip) {
		t.Errorf("should accept connection after decrement")
	}

	// Decrement to remove completely
	for i := 0; i < MaxConnectionsPerIP; i++ {
		DecrementIPConnection(ip)
	}
	if _, exists := ipConnections[ip]; exists {
		t.Errorf("expected IP to be removed from map, but still exists")
	}
}