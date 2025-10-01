package ai

import (
	"sync"
	"time"
)

const (
	MaxAIContextMessages = 20
	AIContextTimeout     = 30 * time.Minute
	GeminiModel          = "gemini-2.5-flash"
)

// ConversationHistory stores AI conversation state
type ConversationHistory struct {
	Messages   []map[string]interface{}
	LastActive time.Time
	Mu         sync.RWMutex
}

// Response struct for AI API
type Response struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}
