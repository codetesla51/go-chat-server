package server 
import 

(
  "time"
"fmt"
  "net/http"
  "encoding/json"
  	"bytes"
  	
  "io"
  )
func handleAichatWithContext(apiKey, userPrompt, lobbyName, username string) (string, error) {
	// Get the appropriate AI guideline for this lobby
	guideline := getAIPromptForLobby(lobbyName)

	// Get or create conversation history for this lobby
	conversationsMutex.Lock()
	conv, exists := lobbyConversations[lobbyName]
	if !exists {
		conv = &ConversationHistory{
			messages:   []map[string]interface{}{},
			lastActive: time.Now(),
		}
		lobbyConversations[lobbyName] = conv
	}
	conversationsMutex.Unlock()

	conv.mu.Lock()
	defer conv.mu.Unlock()

	// Clear old conversations (after 30 min of inactivity)
	if time.Since(conv.lastActive) > 30*time.Minute {
		conv.messages = []map[string]interface{}{}
	}
	conv.lastActive = time.Now()

	if len(conv.messages) == 0 {
		// First interaction - send system prompt
		systemPrompt := guideline

		// Add lobby context if available
		lobbyContext := getLobbyContext(lobbyName)
		if lobbyContext != "" {
			systemPrompt += "\n\n" + lobbyContext
		}

		conv.messages = append(conv.messages, map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": systemPrompt},
			},
		})
		conv.messages = append(conv.messages, map[string]interface{}{
			"role": "model",
			"parts": []map[string]string{
				{"text": "Understood! I'm ready to assist. Beep-boop!"},
			},
		})
	}

	// Build the user prompt with context
	fullPrompt := userPrompt

	// Add recent lobby context for better awareness
	lobbyContext := getLobbyContext(lobbyName)
	if lobbyContext != "" {
		fullPrompt = lobbyContext + "\n" + username + " asked: " + userPrompt
	} else {
		fullPrompt = username + " asked: " + userPrompt
	}

	// Add user's message
	conv.messages = append(conv.messages, map[string]interface{}{
		"role": "user",
		"parts": []map[string]string{
			{"text": fullPrompt},
		},
	})

	// Limit history to last 20 messages to avoid token limits
	if len(conv.messages) > 20 {
		conv.messages = conv.messages[len(conv.messages)-20:]
	}

url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent?key=%s", apiKey)

	payload := map[string]interface{}{
		"contents": conv.messages,
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

		// Add AI's response to history
		conv.messages = append(conv.messages, map[string]interface{}{
			"role": "model",
			"parts": []map[string]string{
				{"text": aiResponse},
			},
		})

		return aiResponse, nil
	}
	return "", fmt.Errorf("no text in response")
}