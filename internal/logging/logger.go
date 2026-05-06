package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

type Logger struct {
	path string
	mu   sync.Mutex
}

func New(dataDir string) *Logger {
	return &Logger{path: filepath.Join(dataDir, "agentdesk.log")}
}

func (l *Logger) Path() string {
	return l.path
}

func (l *Logger) Printf(format string, args ...interface{}) {
	if l == nil {
		log.Printf(format, args...)
		return
	}
	msg := fmt.Sprintf(format, args...)
	log.Print(msg)
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = os.MkdirAll(filepath.Dir(l.path), 0o700)
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		log.Printf("open app log failed: %v", err)
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s %s\n", time.Now().Format(time.RFC3339), msg)
}

func (l *Logger) Recover(where string) {
	if v := recover(); v != nil {
		l.Printf("panic in %s: %v\n%s", where, v, string(debug.Stack()))
	}
}
