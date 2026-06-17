// Package session provides conversation session persistence.
package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/WuSuBuDuoMing/aidev/internal/storage"
	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// SessionSummary holds a session without messages (for listing).
type SessionSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	MessageCount int    `json:"message_count"`
}

// Manager handles session CRUD operations.
type Manager struct {
	db *storage.DB
}

// NewManager creates a new session manager.
func NewManager(db *storage.DB) *Manager {
	return &Manager{db: db}
}

// Save persists a conversation to the database.
func (m *Manager) Save(conv *types.Conversation) error {
	if conv.ID == "" {
		conv.ID = uuid.New().String()
	}
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}
	conv.UpdatedAt = time.Now()

	// Upsert session
	_, err := m.db.Exec(`
		INSERT INTO sessions (id, title, provider, model, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			provider = excluded.provider,
			model = excluded.model,
			updated_at = excluded.updated_at
	`, conv.ID, conv.Title, string(conv.Provider), conv.Model,
		conv.CreatedAt.Format(time.RFC3339),
		conv.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}

	// Delete existing messages (full replace)
	_, err = m.db.Exec("DELETE FROM messages WHERE session_id = ?", conv.ID)
	if err != nil {
		return fmt.Errorf("clear messages: %w", err)
	}

	// Insert all messages
	for i, msg := range conv.Messages {
		var toolCallsJSON *string
		if len(msg.ToolCalls) > 0 {
			data, _ := json.Marshal(msg.ToolCalls)
			s := string(data)
			toolCallsJSON = &s
		}

		_, err = m.db.Exec(`
			INSERT INTO messages (session_id, role, content, tool_calls, tool_call_id, tool_name, timestamp)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, conv.ID, string(msg.Role), msg.Content, toolCallsJSON,
			nullStr(msg.ToolCallID), nullStr(msg.ToolName), msg.Timestamp.UnixMilli())
		if err != nil {
			return fmt.Errorf("insert message %d: %w", i, err)
		}
	}

	return nil
}

// Load retrieves a full conversation by ID.
func (m *Manager) Load(sessionID string) (*types.Conversation, error) {
	var conv types.Conversation
	var provider, createdAt, updatedAt string

	err := m.db.QueryRow("SELECT id, title, provider, model, created_at, updated_at FROM sessions WHERE id = ?", sessionID).
		Scan(&conv.ID, &conv.Title, &provider, &conv.Model, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}

	conv.Provider = provider
	conv.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	conv.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	// Load messages
	rows, err := m.db.Query("SELECT role, content, tool_calls, tool_call_id, tool_name, timestamp FROM messages WHERE session_id = ? ORDER BY timestamp ASC", sessionID)
	if err != nil {
		return nil, fmt.Errorf("load messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg types.Message
		var toolCallsJSON, tcID, tn sql.NullString
		var ts int64

		if err := rows.Scan(&msg.Role, &msg.Content, &toolCallsJSON, &tcID, &tn, &ts); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		msg.Role = types.MessageRole(msg.Role)
		msg.Timestamp = time.UnixMilli(ts)
		msg.ToolCallID = tcID.String
		msg.ToolName = tn.String

		if toolCallsJSON.Valid {
			json.Unmarshal([]byte(toolCallsJSON.String), &msg.ToolCalls)
		}

		conv.Messages = append(conv.Messages, msg)
	}

	return &conv, nil
}

// List returns recent sessions (most recent first).
func (m *Manager) List(limit int) ([]SessionSummary, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := m.db.Query(`
		SELECT s.id, s.title, s.provider, s.model, s.created_at, s.updated_at, COUNT(m.id) as message_count
		FROM sessions s
		LEFT JOIN messages m ON m.session_id = s.id
		GROUP BY s.id
		ORDER BY s.updated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []SessionSummary
	for rows.Next() {
		var s SessionSummary
		if err := rows.Scan(&s.ID, &s.Title, &s.Provider, &s.Model, &s.CreatedAt, &s.UpdatedAt, &s.MessageCount); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// Search searches sessions by title or content.
func (m *Manager) Search(query string, limit int) ([]SessionSummary, error) {
	if limit <= 0 {
		limit = 10
	}
	like := "%" + query + "%"

	rows, err := m.db.Query(`
		SELECT DISTINCT s.id, s.title, s.provider, s.model, s.created_at, s.updated_at, COUNT(m2.id) as message_count
		FROM sessions s
		LEFT JOIN messages m ON m.session_id = s.id
		LEFT JOIN messages m2 ON m2.session_id = s.id
		WHERE s.title LIKE ? OR m.content LIKE ?
		GROUP BY s.id
		ORDER BY s.updated_at DESC
		LIMIT ?
	`, like, like, limit)
	if err != nil {
		return nil, fmt.Errorf("search sessions: %w", err)
	}
	defer rows.Close()

	var sessions []SessionSummary
	for rows.Next() {
		var s SessionSummary
		if err := rows.Scan(&s.ID, &s.Title, &s.Provider, &s.Model, &s.CreatedAt, &s.UpdatedAt, &s.MessageCount); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// Delete removes a session and its messages.
func (m *Manager) Delete(sessionID string) (bool, error) {
	result, err := m.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return false, fmt.Errorf("delete session: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// GenerateTitle creates a title from the first user message.
func GenerateTitle(content string) string {
	if len(content) > 60 {
		return content[:57] + "..."
	}
	if content == "" {
		return "New Conversation"
	}
	return content
}

// nullStr returns a sql.NullString.
func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
