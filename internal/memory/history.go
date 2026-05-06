package memory

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type History struct {
	path string
	mu   sync.Mutex
}

type HistoryEntry struct {
	Time      time.Time `json:"time"`
	SessionID string    `json:"session_id"`
	RunID     string    `json:"run_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
}

func NewHistory(dataDir string) *History {
	return &History{path: filepath.Join(dataDir, "sessions")}
}

func (h *History) Path() string {
	if h == nil {
		return ""
	}
	return h.path
}

func NewSessionID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "session_" + time.Now().Format("20060102_150405")
	}
	return "session_" + time.Now().Format("20060102_150405") + "_" + hex.EncodeToString(b[:])
}

func (h *History) SessionPath(sessionID string) string {
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join(h.path, sessionID+".jsonl")
}

func (h *History) Append(entry HistoryEntry) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	path := h.SessionPath(entry.SessionID)
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	if entry.Time.IsZero() {
		entry.Time = time.Now()
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(b, '\n'))
}

func (h *History) Recent(sessionID string, limit int) []HistoryEntry {
	if h == nil || limit <= 0 {
		return nil
	}
	f, err := os.Open(h.SessionPath(sessionID))
	if err != nil {
		return nil
	}
	defer f.Close()
	var entries []HistoryEntry
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		var e HistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
			entries = append(entries, e)
			if len(entries) > limit {
				entries = entries[len(entries)-limit:]
			}
		}
	}
	return entries
}
