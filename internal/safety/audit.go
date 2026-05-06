package safety

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AuditLog struct {
	path string
	mu   sync.Mutex
}

func NewAuditLog(dataDir string) *AuditLog {
	return &AuditLog{path: filepath.Join(dataDir, "audit.jsonl")}
}

func (a *AuditLog) Write(event map[string]interface{}) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_ = os.MkdirAll(filepath.Dir(a.path), 0o700)
	event["time"] = time.Now().UTC().Format(time.RFC3339)
	b, err := json.Marshal(event)
	if err != nil {
		return
	}
	f, err := os.OpenFile(a.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(b, '\n'))
}
