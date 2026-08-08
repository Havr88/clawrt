package telegram

import (
	"bytes"
	"clawrt/internal/config"
	"clawrt/internal/fastpath"
	"clawrt/internal/i18n"
	"clawrt/internal/knowledge"
	"clawrt/internal/llm"
	"clawrt/internal/security"
	"clawrt/internal/skills"
	"clawrt/internal/sys"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LanguageCode string `json:"language_code"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type SendMessageReq struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type Bot struct {
	Config     *config.Config
	Skills     *skills.SkillRegistry
	Knowledge  *knowledge.KnowledgeEngine
	Classifier *security.RiskClassifier
	FastPath   *fastpath.FastPathEngine
	LLMClient  *llm.Client
	HTTPClient *http.Client
	Debouncer  *Debouncer
	LastOffset int64
}

func NewBot(cfg *config.Config, registry *skills.SkillRegistry) *Bot {
	sysInfo := sys.GetSystemInfo()
	ke := knowledge.NewKnowledgeEngine(sysInfo.MemoryTotalMB, "/tmp/clawrt_facts.json")
	rc := security.NewRiskClassifier()
	fp := fastpath.NewFastPathEngine()

	bot := &Bot{
		Config:     cfg,
		Skills:     registry,
		Knowledge:  ke,
		Classifier: rc,
		FastPath:   fp,
		LLMClient:  llm.NewClient(cfg.Provider, cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel, cfg.FallbackModel),
		HTTPClient: &http.Client{
			Timeout: 40 * time.Second,
		},
	}

	bot.Debouncer = NewDebouncer(2*time.Second, func(chatID int64, mergedText string) {
		bot.handleMessageText(chatID, mergedText)
	})

	return bot
}

func (b *Bot) getLanguage(msg *Message) i18n.Lang {
	if b.Config.Language != "" && b.Config.Language != "auto" {
		return i18n.NormalizeLang(b.Config.Language)
	}
	if msg != nil && msg.From.LanguageCode != "" {
		return i18n.NormalizeLang(msg.From.LanguageCode)
	}
	return i18n.ES
}

func (b *Bot) StartPolling(stopChan <-chan struct{}) {
	if b.Config.BotToken == "" {
		log.Println("[WARN] Token de Telegram no configurado. ClawRT está activo en modo espera (revisar /etc/config/clawrt).")
		return
	}

	log.Printf("[INFO] Bot de Telegram iniciado correctamente (Long-polling CGNAT-safe)")

	for {
		select {
		case <-stopChan:
			log.Println("[INFO] Polling de Telegram detenido.")
			return
		default:
			updates, err := b.getUpdates()
			if err != nil {
				log.Printf("[ERROR] Fallo al obtener updates de Telegram: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, update := range updates {
				b.LastOffset = update.UpdateID + 1
				if update.Message != nil {
					go b.handleMessage(update.Message)
				}
			}
		}
	}
}

func (b *Bot) getUpdates() ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", b.Config.BotToken, b.LastOffset)
	resp, err := b.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("telegram API devolvió OK=false")
	}

	return result.Result, nil
}

func (b *Bot) handleMessage(msg *Message) {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	// Check authentication
	if !b.Config.IsChatAuthorized(chatID) {
		log.Printf("[WARN] Mensaje no autorizado recibido de Chat ID: %d", chatID)
		_ = b.SendMessage(chatID, fmt.Sprintf("⚠️ Acceso no autorizado. Tu Chat ID es `%d`. Configúralo en `/etc/config/clawrt`.", chatID))
		return
	}

	if text == "" {
		return
	}

	// Push message into Debouncer (Control commands bypass immediately)
	b.Debouncer.PushMessage(chatID, text)
}

func (b *Bot) handleMessageText(chatID int64, text string) {
	lang := b.getLanguage(nil)
	log.Printf("[MSG] [Chat ID (%d)]: %s", chatID, text)

	// Handle Slash Commands
	switch {
	case strings.HasPrefix(text, "/start") || strings.HasPrefix(text, "/help"):
		helpMsg := fmt.Sprintf("%s\n\n%s", i18n.T(lang, "help_title"), i18n.T(lang, "help_body"))
		_ = b.SendMessage(chatID, helpMsg)
		return

	case strings.HasPrefix(text, "/status") || strings.HasPrefix(text, "/sysinfo"):
		info := sys.GetSystemInfo()
		netSummary := sys.GetNetworkSummary()
		resp := i18n.T(lang, "status_header",
			info.Hostname, info.OpenWrtVer, info.Architecture,
			info.Uptime, info.LoadAverage,
			info.MemoryUsedMB, info.MemoryTotalMB, info.MemoryUsedPct,
			netSummary)
		_ = b.SendMessage(chatID, resp)
		return

	case strings.HasPrefix(text, "/wifi"):
		summary := sys.GetNetworkSummary()
		_ = b.SendMessage(chatID, summary)
		return

	case strings.HasPrefix(text, "/reboot"):
		_ = b.SendMessage(chatID, "⚠️ Rebooting ClawRT service...")
		go func() {
			time.Sleep(3 * time.Second)
			_, _ = b.Skills.ExecuteTool("restart_service", map[string]interface{}{"service": "clawrt"})
		}()
		return
	}

	// FASTPATH LAYER 1: Direct Answer (0 LLM tokens, 0ms)
	if directAns, ok := b.FastPath.TryDirectAnswer(text); ok {
		log.Printf("[FASTPATH] Layer 1 Direct Answer resuelto (0 tokens)")
		_ = b.SendMessage(chatID, directAns)
		return
	}

	// FASTPATH LAYER 2: Quick Route Matcher (0 LLM tokens, ~1ms)
	if quickAns, ok := b.FastPath.TryQuickRoute(text, b.Skills); ok {
		log.Printf("[FASTPATH] Layer 2 Quick Route resuelto (0 tokens)")
		_ = b.SendMessage(chatID, quickAns)
		return
	}

	// Fallback to ReAct LLM Loop with 3-Tier Dynamic Routing
	b.processLLMQuery(chatID, text, lang)
}

func (b *Bot) processLLMQuery(chatID int64, userQuery string, lang i18n.Lang) {
	chatIDStr := fmt.Sprintf("%d", chatID)

	// Check Denial Circuit Breaker
	if b.Classifier.ShouldHardStop(chatIDStr) {
		_ = b.SendMessage(chatID, "🚫 *Circuit Breaker Activo:* Se han denegado 3 comandos consecutivos. Se requiere una nueva instrucción para continuar.")
		b.Classifier.ResetDenialTally(chatIDStr)
		return
	}

	// 3-Tier Message Routing Engine (fast/balanced/deep)
	routing := fastpath.ClassifyTier(userQuery)
	log.Printf("[ROUTING] Tier: %s, MaxTokens: %d, Temp: %.2f", routing.Tier, routing.MaxTokens, routing.Temperature)

	_ = b.SendMessage(chatID, i18n.T(lang, "processing_llm"))

	// Build System Prompt augmented with Knowledge Engine Context
	sysPrompt := i18n.T(lang, "sys_prompt") + "\n\n" + b.Knowledge.GetContextSummary()

	messages := []llm.ChatMessage{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: userQuery},
	}

	toolDefs := b.Skills.GetToolDefinitions()

	maxIter := b.Config.ChatMaxSteps
	if maxIter <= 0 {
		maxIter = 5
	}

	toolHistory := make(map[string]int)

	for iter := 0; iter < maxIter; iter++ {
		respMsg, err := b.LLMClient.ChatCompletion(messages, toolDefs)
		if err != nil {
			log.Printf("[ERROR] Fallo en cliente LLM: %v", err)
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error: %v", err))
			return
		}

		messages = append(messages, *respMsg)

		// Check if LLM requested tool execution
		if len(respMsg.ToolCalls) == 0 {
			_ = b.SendMessage(chatID, respMsg.Content)
			return
		}

		// Execute tool calls
		for _, tc := range respMsg.ToolCalls {
			toolName, validName := security.SanitizeToolName(tc.Function.Name)
			if !validName {
				_ = b.SendMessage(chatID, "⚠️ Nombre de herramienta no válido detectado (Defensa Anti-Inyección).")
				return
			}

			// Anti-Reentry / Doom Loop Check: doom_loop: ask
			callKey := fmt.Sprintf("%s:%s", toolName, tc.Function.Arguments)
			toolHistory[callKey]++
			if toolHistory[callKey] >= 2 {
				log.Printf("[DOOM_LOOP] Detección de bucle en herramienta %s (doom_loop: ask)", toolName)
				_ = b.SendMessage(chatID, fmt.Sprintf("⚠️ *Detección de Bucle (doom_loop: ask):* El LLM está intentando repetir exactamente la herramienta `%s` con los mismos argumentos.\n\n¿Deseas continuar o abortar la tarea?", toolName))
				return
			}

			// Risk Bucket & Conservative Fallback Permission Check
			allowed, decision := b.Classifier.EvaluateToolPermission(chatIDStr, toolName)
			if !allowed && decision != security.DecisionAllowAlways {
				log.Printf("[RISK_BUCKET] Herramienta mutable/control plane %s requiere confirmación", toolName)
			}

			_ = b.SendMessage(chatID, i18n.T(lang, "executing_tool", toolName))

			var args map[string]interface{}
			if tc.Function.Arguments != "" {
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			}

			output, err := b.Skills.ExecuteTool(toolName, args)
			toolResult := output
			if err != nil {
				toolResult = fmt.Sprintf("Error: %v", err)
			}

			// Store learned fact in knowledge base if UCI was written
			if toolName == "write_uci_config" && err == nil {
				if path, ok := args["path"].(string); ok {
					if val, ok := args["value"].(string); ok {
						b.Knowledge.SetFact("last_modified_"+path, val)
					}
				}
			}

			messages = append(messages, llm.ChatMessage{
				Role:       "tool",
				Name:       toolName,
				ToolCallID: tc.ID,
				Content:    toolResult,
			})
		}
	}

	_ = b.SendMessage(chatID, "⚠️ Límite de iteraciones por rol alcanzado.")
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	if b.Config.BotToken == "" {
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.Config.BotToken)
	reqBody := SendMessageReq{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := b.HTTPClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fallback without Markdown if markdown parsing fails
		reqBody.ParseMode = ""
		data, _ = json.Marshal(reqBody)
		_, _ = b.HTTPClient.Post(url, "application/json", bytes.NewBuffer(data))
	}

	return nil
}
