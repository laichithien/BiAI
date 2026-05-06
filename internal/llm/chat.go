package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type ChatConfig struct {
	BaseURL string
	Token   string
	Model   string
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type ToolDefinition struct {
	Type     string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"function"`
}

func Complete(ctx context.Context, cfg ChatConfig, messages []Message, tools []ToolDefinition) (Message, error) {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	token := strings.TrimSpace(cfg.Token)
	model := strings.TrimSpace(cfg.Model)
	if baseURL == "" {
		return Message{}, errors.New("LLM Base URL is not configured")
	}
	if token == "" {
		return Message{}, errors.New("API token is not configured")
	}
	if model == "" {
		return Message{}, errors.New("model is not selected")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return Message{}, err
	}
	u.Path = path.Join(strings.TrimRight(u.Path, "/"), "chat", "completions")
	body := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"temperature": 0.2,
	}
	if len(tools) > 0 {
		body["tools"] = tools
		body["tool_choice"] = "auto"
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return Message{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(payload))
	if err != nil {
		return Message{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Message{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1500))
		msg := strings.TrimSpace(string(b))
		if msg == "" {
			msg = resp.Status
		}
		return Message{}, errors.New("chat completion failed: " + resp.Status + ": " + msg)
	}
	var out struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Message{}, err
	}
	if len(out.Choices) == 0 {
		return Message{}, errors.New("chat completion returned no choices")
	}
	out.Choices[0].Message.Content = strings.TrimSpace(out.Choices[0].Message.Content)
	return out.Choices[0].Message, nil
}

func Chat(ctx context.Context, cfg ChatConfig, prompt string) (string, error) {
	msg, err := Complete(ctx, cfg, []Message{
		{Role: "system", Content: "You are BiAI AgentDesk, a concise local coding assistant."},
		{Role: "user", Content: prompt},
	}, nil)
	if err != nil {
		return "", err
	}
	return msg.Content, nil
}
