package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var geminiAPIKey string

// InitAI initializes the AI client with API key from environment
func InitAI() error {
	geminiAPIKey = os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}
	return nil
}

// GetAPIKey returns the current API key (for internal use)
func GetAPIKey() string {
	return geminiAPIKey
}

// HandleAIChat processes AI requests with conversation context
func HandleAIChat(userPrompt, lobbyName, username string,
	conversations map[string]*ConversationHistory,
	convMutex interface{},
	getLobbyContextFn func(string) string) (string, error) {

	if geminiAPIKey == "" {
		return "", fmt.Errorf("AI not initialized")
	}

	guideline := GetAIPromptForLobby(lobbyName)

	// Type assertion for mutex (passed as interface{} to avoid circular import)
	mutex, ok := convMutex.(interface {
		Lock()
		Unlock()
	})
	if !ok {
		return "", fmt.Errorf("invalid mutex type")
	}

	mutex.Lock()
	conv, exists := conversations[lobbyName]
	if !exists {
		conv = &ConversationHistory{
			Messages:   []map[string]interface{}{},
			LastActive: time.Now(),
		}
		conversations[lobbyName] = conv
	}
	mutex.Unlock()

	conv.Mu.Lock()
	defer conv.Mu.Unlock()

	// Clear old conversations
	if time.Since(conv.LastActive) > AIContextTimeout {
		conv.Messages = []map[string]interface{}{}
	}
	conv.LastActive = time.Now()

	if len(conv.Messages) == 0 {
		systemPrompt := guideline
		lobbyContext := getLobbyContextFn(lobbyName)
		if lobbyContext != "" {
			systemPrompt += "\n\n" + lobbyContext
		}

		conv.Messages = append(conv.Messages, map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": systemPrompt},
			},
		})
		conv.Messages = append(conv.Messages, map[string]interface{}{
			"role": "model",
			"parts": []map[string]string{
				{"text": "Understood! I'm ready to assist. Beep-boop!"},
			},
		})
	}

	fullPrompt := userPrompt
	lobbyContext := getLobbyContextFn(lobbyName)
	if lobbyContext != "" {
		fullPrompt = lobbyContext + "\n" + username + " asked: " + userPrompt
	} else {
		fullPrompt = username + " asked: " + userPrompt
	}

	conv.Messages = append(conv.Messages, map[string]interface{}{
		"role": "user",
		"parts": []map[string]string{
			{"text": fullPrompt},
		},
	})

	if len(conv.Messages) > MaxAIContextMessages {
		conv.Messages = conv.Messages[len(conv.Messages)-20:]
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		GeminiModel, geminiAPIKey)

	payload := map[string]interface{}{
		"contents": conv.Messages,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		aiResponse := result.Candidates[0].Content.Parts[0].Text

		conv.Messages = append(conv.Messages, map[string]interface{}{
			"role": "model",
			"parts": []map[string]string{
				{"text": aiResponse},
			},
		})

		return aiResponse, nil
	}
	return "", fmt.Errorf("no text in response")
}

// FormatAIError formats AI errors for user display
func FormatAIError(err error) string {
    if err == nil {
        return "AI features enabled"  
    }

    e := err.Error()
    switch {
    case strings.Contains(e, "rate limit"):
        return "AI Error: Rate limit reached. Please wait and try again."
    case strings.Contains(e, "quota"):
        return "AI Error: Quota reached. Try later."
    case strings.Contains(e, "invalid prompt"):
        return "AI Error: Your prompt is invalid."
    default:
        return "AI Error: Please try again later."
    }
}