package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type UserConfig struct {
	LLMBaseURL string `json:"llm_base_url"`
	Model      string `json:"model"`
}

type UserSecrets struct {
	APIToken string `json:"api_token"`
}

func LoadUserConfig(dataDir string) (UserConfig, error) {
	cfg := UserConfig{
		LLMBaseURL: "https://api.openai.com/v1",
		Model:      "gpt-4.1-mini",
	}
	b, err := os.ReadFile(filepath.Join(dataDir, "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func SaveUserConfig(dataDir string, cfg UserConfig) error {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "config.json"), b, 0o600)
}

func LoadUserSecrets(dataDir string) (UserSecrets, error) {
	var sec UserSecrets
	b, err := os.ReadFile(filepath.Join(dataDir, "secrets.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return sec, nil
		}
		return sec, err
	}
	if err := json.Unmarshal(b, &sec); err != nil {
		return sec, err
	}
	return sec, nil
}

func SaveUserSecrets(dataDir string, sec UserSecrets) error {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(sec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "secrets.json"), b, 0o600)
}
