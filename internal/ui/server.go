package ui

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"biai/internal/agent"
	"biai/internal/config"
	"biai/internal/llm"
)

//go:embed assets/*
var embedded embed.FS

type Server struct {
	agent   *agent.Agent
	dataDir string
	token   string
	server  *http.Server
	ln      net.Listener
}

func NewServer(a *agent.Agent, dataDir string) (*Server, error) {
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	return &Server{agent: a, dataDir: dataDir, token: token}, nil
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.ln = ln
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/app.css", s.handleAsset("app.css", "text/css; charset=utf-8"))
	mux.HandleFunc("/app.js", s.handleAsset("app.js", "application/javascript; charset=utf-8"))
	mux.HandleFunc("/api/chat", s.requireToken(s.handleChat))
	mux.HandleFunc("/api/approval", s.requireToken(s.handleApproval))
	mux.HandleFunc("/api/settings", s.requireToken(s.handleSettings))
	mux.HandleFunc("/api/models", s.requireToken(s.handleModels))
	s.server = &http.Server{Handler: mux}
	go func() {
		_ = s.server.Serve(ln)
	}()
	return nil
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		LLMBaseURL string `json:"llm_base_url"`
		APIToken   string `json:"api_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}
	baseURL := strings.TrimSpace(req.LLMBaseURL)
	token := strings.TrimSpace(req.APIToken)
	if baseURL == "" || token == "" {
		cfg, _ := config.LoadUserConfig(s.dataDir)
		sec, _ := config.LoadUserSecrets(s.dataDir)
		if baseURL == "" {
			baseURL = cfg.LLMBaseURL
		}
		if token == "" {
			token = sec.APIToken
		}
	}
	models, err := llm.FetchModels(r.Context(), baseURL, token)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{"models": models}, http.StatusOK)
}

func (s *Server) Close() {
	if s.server != nil {
		_ = s.server.Close()
	}
}

func (s *Server) URL() string {
	if s.ln == nil {
		return ""
	}
	return "http://" + s.ln.Addr().String() + "/?token=" + s.token
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") != s.token {
		http.Error(w, "invalid token", http.StatusForbidden)
		return
	}
	s.serveEmbedded(w, "index.html", "text/html; charset=utf-8")
}

func (s *Server) handleAsset(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != s.token {
			http.Error(w, "invalid token", http.StatusForbidden)
			return
		}
		s.serveEmbedded(w, name, contentType)
	}
}

func (s *Server) serveEmbedded(w http.ResponseWriter, name, contentType string) {
	data, err := fs.ReadFile(embedded, filepath.ToSlash(filepath.Join("assets", name)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(data)
}

func (s *Server) requireToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-AgentDesk-Token") != s.token {
			http.Error(w, "invalid token", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req agent.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}
	resp := s.agent.Chat(r.Context(), req)
	writeJSON(w, resp, http.StatusOK)
}

func (s *Server) handleApproval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req agent.ApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
		return
	}
	resp := s.agent.DecideApproval(r.Context(), req)
	writeJSON(w, resp, http.StatusOK)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := config.LoadUserConfig(s.dataDir)
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		sec, err := config.LoadUserSecrets(s.dataDir)
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{
			"llm_base_url": cfg.LLMBaseURL,
			"model":        cfg.Model,
			"has_token":    sec.APIToken != "",
		}, http.StatusOK)
	case http.MethodPost:
		var req struct {
			LLMBaseURL string `json:"llm_base_url"`
			Model      string `json:"model"`
			APIToken   string `json:"api_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
			return
		}
		if err := config.SaveUserConfig(s.dataDir, config.UserConfig{LLMBaseURL: strings.TrimSpace(req.LLMBaseURL), Model: strings.TrimSpace(req.Model)}); err != nil {
			writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(req.APIToken) != "" {
			if err := config.SaveUserSecrets(s.dataDir, config.UserSecrets{APIToken: strings.TrimSpace(req.APIToken)}); err != nil {
				writeJSON(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
				return
			}
		}
		writeJSON(w, map[string]bool{"ok": true}, http.StatusOK)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func randomToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b[:])
	if strings.Trim(token, "0") == "" {
		return "", errors.New("bad random token")
	}
	return token, nil
}
