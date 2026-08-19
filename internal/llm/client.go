package llm

import (
	"bytes"
	"clawrt/internal/skills"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatMessage           `json:"messages"`
	Tools       []skills.ToolDefinition `json:"tools,omitempty"`
	Temperature float64                 `json:"temperature,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type ModelsListResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type Client struct {
	Provider      string
	BaseURL       string
	APIKey        string
	Model         string
	FallbackModel string
	HTTPClient    *http.Client
}

func NewClient(provider, baseURL, apiKey, model, fallbackModel string) *Client {
	info := GetProvider(provider)
	if baseURL == "" {
		baseURL = info.BaseURL
	}
	if model == "" {
		model = info.DefaultModel
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		Provider:      provider,
		BaseURL:       baseURL,
		APIKey:        apiKey,
		Model:         model,
		FallbackModel: fallbackModel,
		HTTPClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) FetchModels() ([]string, error) {
	url := fmt.Sprintf("%s/models", c.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear request GET /models: %w", err)
	}

	if strings.Contains(c.BaseURL, "openrouter.ai") {
		req.Header.Set("HTTP-Referer", "https://github.com/Havr88/clawrt")
		req.Header.Set("X-Title", "ClawRT OpenWrt Agent")
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de red al consultar /models: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var listResp ModelsListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("error parseando /models (HTTP %d): %w", resp.StatusCode, err)
	}

	var models []string
	for _, m := range listResp.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}

	if len(models) == 0 {
		info := GetProvider(c.Provider)
		models = info.PopularModels
	}

	return models, nil
}

func (c *Client) ChatCompletion(messages []ChatMessage, tools []skills.ToolDefinition) (*ChatMessage, error) {
	// 1. Try with primary model
	msg, err := c.doRequest(c.Model, messages, tools)
	if err == nil {
		return msg, nil
	}

	// 2. If primary model fails and fallback model is configured, attempt fallback
	if c.FallbackModel != "" && c.FallbackModel != c.Model {
		fallbackMsg, fallbackErr := c.doRequest(c.FallbackModel, messages, tools)
		if fallbackErr == nil {
			return fallbackMsg, nil
		}
		return nil, fmt.Errorf("modelo principal falló (%v) y modelo fallback '%s' también falló: %w", err, c.FallbackModel, fallbackErr)
	}

	return nil, err
}

func (c *Client) doRequest(targetModel string, messages []ChatMessage, tools []skills.ToolDefinition) (*ChatMessage, error) {
	url := fmt.Sprintf("%s/chat/completions", c.BaseURL)

	reqBody := ChatCompletionRequest{
		Model:       targetModel,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.3,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error al serializar petición LLM: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error al crear request HTTP LLM: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if strings.Contains(c.BaseURL, "openrouter.ai") {
		req.Header.Set("HTTP-Referer", "https://github.com/Havr88/clawrt")
		req.Header.Set("X-Title", "ClawRT OpenWrt Agent")
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de red con proveedor %s (%s): %w", c.Provider, c.BaseURL, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta LLM: %w", err)
	}

	var compResp ChatCompletionResponse
	if err := json.Unmarshal(respBytes, &compResp); err != nil {
		return nil, fmt.Errorf("error al deserializar respuesta LLM (HTTP %d): %w, raw: %s", resp.StatusCode, err, string(respBytes))
	}

	if compResp.Error != nil && compResp.Error.Message != "" {
		return nil, fmt.Errorf("error del proveedor LLM (%s): %s", c.Provider, compResp.Error.Message)
	}

	if len(compResp.Choices) == 0 {
		return nil, fmt.Errorf("respuesta de LLM vacía (HTTP %d)", resp.StatusCode)
	}

	return &compResp.Choices[0].Message, nil
}
