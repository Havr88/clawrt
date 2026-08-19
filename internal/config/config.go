package config

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Enabled             bool    `json:"enabled"`
	Language            string  `json:"language"`
	Provider            string  `json:"provider"`
	BotToken            string  `json:"bot_token"`
	ChatIDs             []int64 `json:"chat_ids"`
	LLMBaseURL          string  `json:"llm_base_url"`
	LLMAPIKey           string  `json:"llm_api_key"`
	LLMModel            string  `json:"llm_model"`
	FallbackModel       string  `json:"fallback_model"`
	ExternalDBProvider  string  `json:"external_db_provider"`
	ExternalDBURL       string  `json:"external_db_url"`
	ExternalDBToken     string  `json:"external_db_token"`
	ChatMaxSteps        int     `json:"chat_max_steps"`
	PlanMaxSteps        int     `json:"plan_max_steps"`
	HardMaxSteps        int     `json:"hard_max_steps"`
	MaxIterations       int     `json:"max_iterations"`
	EnableSmartApproval bool    `json:"enable_smart_approval"`
	EnableDebounce      bool    `json:"enable_debounce"`
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:             true,
		Language:            "auto",
		Provider:            "bynara",
		LLMBaseURL:          "https://router.bynara.id/v1",
		LLMModel:            "gpt-4o-mini",
		FallbackModel:       "deepseek-chat",
		ExternalDBProvider:  "none",
		ChatMaxSteps:        5,
		PlanMaxSteps:        15,
		HardMaxSteps:        20,
		MaxIterations:       5,
		EnableSmartApproval: true,
		EnableDebounce:      true,
		ChatIDs:             []int64{},
	}
}

// LoadConfig attempts to load configuration from UCI file (/etc/config/clawrt),
// JSON file if provided, and overrides with Environment Variables.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	// 1. Try reading from UCI file first (/etc/config/clawrt or custom path)
	uciPath := "/etc/config/clawrt"
	if path != "" && !strings.HasSuffix(path, ".json") {
		uciPath = path
	}

	if _, err := os.Stat(uciPath); err == nil {
		_ = parseUCIFile(uciPath, cfg)
	}

	// 2. Try JSON file if path ends with .json or if /tmp/clawrt.json exists
	jsonPath := path
	if jsonPath == "" || !strings.HasSuffix(jsonPath, ".json") {
		if _, err := os.Stat("/tmp/clawrt.json"); err == nil {
			jsonPath = "/tmp/clawrt.json"
		}
	}
	if jsonPath != "" && strings.HasSuffix(jsonPath, ".json") {
		if data, err := os.ReadFile(jsonPath); err == nil {
			_ = json.Unmarshal(data, cfg)
		}
	}

	// 3. Env var overrides
	if lang := os.Getenv("CLAWRT_LANG"); lang != "" {
		cfg.Language = lang
	}
	if provider := os.Getenv("LLM_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		cfg.BotToken = token
	}
	if chats := os.Getenv("TELEGRAM_CHAT_ID"); chats != "" {
		cfg.ChatIDs = parseChatIDs(chats)
	}
	if baseURL := os.Getenv("LLM_BASE_URL"); baseURL != "" {
		cfg.LLMBaseURL = baseURL
	}
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		cfg.LLMAPIKey = apiKey
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		cfg.LLMModel = model
	}
	if fallback := os.Getenv("LLM_FALLBACK_MODEL"); fallback != "" {
		cfg.FallbackModel = fallback
	}

	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 5
	}
	if cfg.ChatMaxSteps <= 0 {
		cfg.ChatMaxSteps = 5
	}

	return cfg, nil
}

func parseUCIFile(filePath string, cfg *Config) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == "config" {
			currentSection = parts[1]
			continue
		}

		if len(parts) >= 3 && parts[0] == "option" {
			key := parts[1]
			val := strings.Trim(strings.Join(parts[2:], " "), "'\"")

			switch currentSection {
			case "core", "main":
				if key == "enabled" {
					cfg.Enabled = val == "1" || val == "true"
				} else if key == "language" {
					cfg.Language = val
				}
			case "telegram":
				if key == "bot_token" {
					cfg.BotToken = val
				}
			case "db", "database":
				switch key {
				case "provider":
					cfg.ExternalDBProvider = val
				case "url":
					cfg.ExternalDBURL = val
				case "token":
					cfg.ExternalDBToken = val
				}
			case "llm":
				switch key {
				case "provider":
					cfg.Provider = val
				case "base_url":
					cfg.LLMBaseURL = val
				case "api_key":
					cfg.LLMAPIKey = val
				case "model":
					cfg.LLMModel = val
				case "fallback_model":
					cfg.FallbackModel = val
				case "chat_max_steps":
					if n, err := strconv.Atoi(val); err == nil {
						cfg.ChatMaxSteps = n
					}
				case "plan_max_steps":
					if n, err := strconv.Atoi(val); err == nil {
						cfg.PlanMaxSteps = n
					}
				case "hard_max_steps":
					if n, err := strconv.Atoi(val); err == nil {
						cfg.HardMaxSteps = n
					}
				case "max_iterations":
					if n, err := strconv.Atoi(val); err == nil {
						cfg.MaxIterations = n
					}
				}
			}
		} else if len(parts) >= 3 && parts[0] == "list" {
			key := parts[1]
			val := strings.Trim(strings.Join(parts[2:], " "), "'\"")

			if currentSection == "telegram" && key == "chat_id" {
				if id, err := strconv.ParseInt(val, 10, 64); err == nil {
					cfg.ChatIDs = append(cfg.ChatIDs, id)
				}
			}
		}
	}
	return nil
}

func parseChatIDs(raw string) []int64 {
	var ids []int64
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if id, err := strconv.ParseInt(part, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (c *Config) IsChatAuthorized(chatID int64) bool {
	if len(c.ChatIDs) == 0 {
		return true
	}
	for _, id := range c.ChatIDs {
		if id == chatID {
			return true
		}
	}
	return false
}
