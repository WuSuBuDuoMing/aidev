// Package storage provides tests for the SQLite database layer.
package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Verify the file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestOpenCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "subdir", "deep", "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open should create parent directories: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created in nested directory")
	}
}

func TestClose(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestMigrate(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Verify sessions table exists
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions'")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("sessions table should exist after migration")
	}

	// Verify messages table exists
	rows2, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='messages'")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows2.Close()

	if !rows2.Next() {
		t.Fatal("messages table should exist after migration")
	}
}

func TestExec(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO sessions (id, title) VALUES (?, ?)", "test-1", "Test Session")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("RowsAffected failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 row affected, got %d", n)
	}
}

func TestQuery(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Insert test data
	_, err = db.Exec("INSERT INTO sessions (id, title, provider, model) VALUES (?, ?, ?, ?)",
		"s1", "Session One", "deepseek", "deepseek-v4")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	_, err = db.Exec("INSERT INTO sessions (id, title, provider, model) VALUES (?, ?, ?, ?)",
		"s2", "Session Two", "claude", "claude-sonnet-4-6")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	rows, err := db.Query("SELECT id, title FROM sessions ORDER BY id")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, title string
		if err := rows.Scan(&id, &title); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		count++
	}
	if count != 2 {
		t.Fatalf("expected 2 rows, got %d", count)
	}
}

func TestQueryRow(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO sessions (id, title) VALUES (?, ?)", "s1", "My Session")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	var title string
	err = db.QueryRow("SELECT title FROM sessions WHERE id = ?", "s1").Scan(&title)
	if err != nil {
		t.Fatalf("QueryRow failed: %v", err)
	}
	if title != "My Session" {
		t.Fatalf("expected 'My Session', got %q", title)
	}
}

func TestForeignKeyCascade(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Insert session and messages
	_, err = db.Exec("INSERT INTO sessions (id, title) VALUES (?, ?)", "s1", "Test")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	_, err = db.Exec("INSERT INTO messages (session_id, role, content, timestamp) VALUES (?, ?, ?, ?)",
		"s1", "user", "hello", 1000)
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	// Delete session — should cascade to messages
	_, err = db.Exec("DELETE FROM sessions WHERE id = ?", "s1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id = ?", "s1").Scan(&count)
	if err != nil {
		t.Fatalf("QueryRow failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 messages after cascade delete, got %d", count)
	}
}

func TestDoubleClose(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// First close should succeed
	if err := db.Close(); err != nil {
		t.Fatalf("First Close failed: %v", err)
	}

	// Second close will error (expected behavior — just verify no panic)
	_ = db.Close()
}
