package llm

import "strings"

type ProviderInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	BaseURL     string   `json:"base_url"`
	DefaultModel string  `json:"default_model"`
	PopularModels []string `json:"popular_models"`
	Description string   `json:"description"`
	IsFreeTier  bool     `json:"is_free_tier"`
}

var ProviderRegistry = map[string]ProviderInfo{
	"bynara": {
		ID:           "bynara",
		Name:         "NaraRouter Gateway (Bynara)",
		BaseURL:      "https://router.bynara.id/v1",
		DefaultModel: "gpt-4o-mini",
		PopularModels: []string{"gpt-4o-mini", "deepseek-chat", "claude-3-5-haiku"},
		Description:  "Gateway unificado NaraRouter (Recomendado)",
		IsFreeTier:   false,
	},
	"groq": {
		ID:           "groq",
		Name:         "Groq Cloud (Ultra Rápido / Tier Gratuito)",
		BaseURL:      "https://api.groq.com/openai/v1",
		DefaultModel: "llama-3.3-70b-versatile",
		PopularModels: []string{"llama-3.3-70b-versatile", "llama-3.1-8b-instant", "mixtral-8x7b-32768"},
		Description:  "Inferencia por hardware LPU (Súper rápida, incluye capa gratuita)",
		IsFreeTier:   true,
	},
	"deepseek": {
		ID:           "deepseek",
		Name:         "DeepSeek API Directo",
		BaseURL:      "https://api.deepseek.com/v1",
		DefaultModel: "deepseek-chat",
		PopularModels: []string{"deepseek-chat", "deepseek-reasoner"},
		Description:  "Modelos DeepSeek V3 y R1 de alto rendimiento y bajo costo",
		IsFreeTier:   false,
	},
	"openrouter": {
		ID:           "openrouter",
		Name:         "OpenRouter (Multi-Proveedor / Modelos Gratis)",
		BaseURL:      "https://openrouter.ai/api/v1",
		DefaultModel: "meta-llama/llama-3.3-70b-instruct:free",
		PopularModels: []string{
			"meta-llama/llama-3.3-70b-instruct:free",
			"google/gemini-2.0-flash-exp:free",
			"deepseek/deepseek-chat",
			"openai/gpt-4o-mini",
		},
		Description:  "Acceso unificado a cientos de modelos (incluye modelos 100% gratuitos)",
		IsFreeTier:   true,
	},
	"openai": {
		ID:           "openai",
		Name:         "OpenAI Directo",
		BaseURL:      "https://api.openai.com/v1",
		DefaultModel: "gpt-4o-mini",
		PopularModels: []string{"gpt-4o-mini", "gpt-4o", "o3-mini"},
		Description:  "API oficial de OpenAI",
		IsFreeTier:   false,
	},
	"gemini": {
		ID:           "gemini",
		Name:         "Google Gemini (API OpenAI Compatible)",
		BaseURL:      "https://generativelanguage.googleapis.com/v1beta/openai/",
		DefaultModel: "gemini-1.5-flash",
		PopularModels: []string{"gemini-1.5-flash", "gemini-1.5-pro", "gemini-2.0-flash-exp"},
		Description:  "API de Google AI Studio (Tier gratuito muy amplio)",
		IsFreeTier:   true,
	},
	"mistral": {
		ID:           "mistral",
		Name:         "Mistral AI",
		BaseURL:      "https://api.mistral.ai/v1",
		DefaultModel: "mistral-small-latest",
		PopularModels: []string{"mistral-small-latest", "open-mistral-7b", "mistral-large-latest"},
		Description:  "Modelos europeos de alta eficiencia",
		IsFreeTier:   false,
	},
	"ollama": {
		ID:           "ollama",
		Name:         "Ollama (Servidor Local LAN / PC)",
		BaseURL:      "http://192.168.1.100:11434/v1",
		DefaultModel: "llama3.2",
		PopularModels: []string{"llama3.2", "qwen2.5-coder", "mistral"},
		Description:  "Inferencia 100% privada ejecutada en tu PC/Servidor local",
		IsFreeTier:   true,
	},
	"custom": {
		ID:           "custom",
		Name:         "Personalizado (Custom API)",
		BaseURL:      "",
		DefaultModel: "",
		PopularModels: []string{},
		Description:  "Endpoint compatible con OpenAI personalizado",
		IsFreeTier:   false,
	},
}

func GetProvider(providerID string) ProviderInfo {
	if p, ok := ProviderRegistry[strings.ToLower(providerID)]; ok {
		return p
	}
	return ProviderRegistry["custom"]
}
