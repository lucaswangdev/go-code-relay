package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var sessionsDir = filepath.Join(os.Getenv("HOME"), ".go-code", "sessions")

var safeSessionRe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

type Session struct {
	ID       string                   `json:"id"`
	Model    string                  `json:"model"`
	SavedAt  string                  `json:"saved_at"`
	Messages []map[string]interface{} `json:"messages"`
}

func normalizeSessionID(sessionID string) string {
	if sessionID == "" {
		return newSessionID()
	}
	name := sessionID
	name = safeSessionRe.ReplaceAllString(name, "-")
	return name
}

func newSessionID() string {
	return "session_" + time.Now().Format("20060102_150405") + "_" + uuid.New().String()[:8]
}

func SaveSession(messages []map[string]interface{}, model, sessionID string) (string, error) {
	sessionID = normalizeSessionID(sessionID)
	os.MkdirAll(sessionsDir, 0755)

	data := Session{
		ID:       sessionID,
		Model:    model,
		SavedAt:  time.Now().Format("2006-01-02 15:04:05"),
		Messages: messages,
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	path := filepath.Join(sessionsDir, sessionID+".json")
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func LoadSession(sessionID string) ([]map[string]interface{}, string, error) {
	path := filepath.Join(sessionsDir, normalizeSessionID(sessionID)+".json")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	var data Session
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, "", err
	}

	return data.Messages, data.Model, nil
}

func ListSessions() [](map[string]interface{}) {
	os.MkdirAll(sessionsDir, 0755)
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil
	}

	var sessions [](map[string]interface{})
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(sessionsDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var data Session
		if err := json.Unmarshal(content, &data); err != nil {
			continue
		}
		preview := ""
		for _, m := range data.Messages {
			if role, _ := m["role"].(string); role == "user" {
				if content, ok := m["content"].(string); ok {
					preview = content
					if len(preview) > 80 {
						preview = preview[:80]
					}
					break
				}
			}
		}
		sessions = append(sessions, map[string]interface{}{
			"id":      data.ID,
			"model":   data.Model,
			"saved_at": data.SavedAt,
			"preview":  preview,
		})
	}
	return sessions
}
