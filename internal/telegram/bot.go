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
	"clawrt/internal/store"
	"clawrt/internal/sys"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type SetMyCommandsReq struct {
	Commands []BotCommand `json:"commands"`
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

func (b *Bot) RegisterTelegramCommands() {
	if b.Config.BotToken == "" {
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", b.Config.BotToken)
	reqObj := SetMyCommandsReq{
		Commands: []BotCommand{
			{Command: "status", Description: "📊 Estado del router, CPU, RAM y red"},
			{Command: "diagnose", Description: "🩺 Diagnóstico de red y Auto-Sanación"},
			{Command: "audit", Description: "🛡️ Auditoría de seguridad y contraseñas del router"},
			{Command: "wireguard", Description: "🔒 Estado de túneles WireGuard & VPN"},
			{Command: "dns", Description: "🌐 Privacidad DNS, DoH & AdBlock"},
			{Command: "flash", Description: "💾 Espacio en Flash (/overlay) y paquetes"},
			{Command: "mwan", Description: "🔀 Estado de enlaces Multi-WAN (mwan3)"},
			{Command: "clients", Description: "📱 Dispositivos conectados en la LAN"},
			{Command: "sticky", Description: "📶 Detectar y reconectar clientes WiFi débiles"},
			{Command: "wifi", Description: "📶 Estado de la red WiFi"},
			{Command: "optimize", Description: "✨ Optimizar canales WiFi (Menor interferencia)"},
			{Command: "conntrack", Description: "🛡️ Guardia de tráfico y anomalías de red"},
			{Command: "sqm", Description: "⚡ Bufferbloat y calidad de servicio QoS"},
			{Command: "qrwifi", Description: "📷 Código QR de conexión a WiFi"},
			{Command: "scan", Description: "🛡️ Escáner de puertos en la LAN"},
			{Command: "models", Description: "🔍 Modelos de IA disponibles (Bynara/Provider)"},
			{Command: "firewall", Description: "🔥 Zonas y reglas del cortafuegos"},
			{Command: "routes", Description: "🌐 Tabla de enrutamiento IPv4/IPv6"},
			{Command: "ping", Description: "⚡ Prueba de latencia y conectividad"},
			{Command: "logs", Description: "📜 Últimos registros del sistema"},
			{Command: "memory", Description: "🧠 Estado de RAM y liberación manual (GC)"},
			{Command: "db", Description: "🗄️ Estado de la Base de Datos Externa (Supabase)"},
			{Command: "clear", Description: "🧹 Vaciar memoria de hechos aprendidos"},
			{Command: "reboot", Description: "🔄 Reiniciar servicio ClawRT"},
			{Command: "help", Description: "❓ Ayuda y lista de comandos"},
		},
	}

	data, err := json.Marshal(reqObj)
	if err != nil {
		return
	}

	resp, err := b.HTTPClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err == nil && resp != nil {
		_ = resp.Body.Close()
		log.Println("[INFO] Comandos de Telegram registrados exitosamente en Telegram Bot API (/setMyCommands)")
	}
}

func (b *Bot) StartPolling(stopChan <-chan struct{}) {
	if b.Config.BotToken == "" {
		log.Println("[WARN] Token de Telegram no configurado. ClawRT está activo en modo espera (revisar /etc/config/clawrt).")
		return
	}

	log.Printf("[INFO] Bot de Telegram iniciado correctamente (Long-polling CGNAT-safe)")
	b.RegisterTelegramCommands()

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
	sys.GetMemoryManager().RecordActivity()

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

	cmd := strings.ToLower(strings.Fields(text)[0])

	// Handle Slash Commands
	switch {
	case cmd == "/start" || cmd == "/help":
		helpMsg := fmt.Sprintf("%s\n\n%s", i18n.T(lang, "help_title"), i18n.T(lang, "help_body"))
		_ = b.SendMessage(chatID, helpMsg)
		return

	case cmd == "/status" || cmd == "/sysinfo":
		info := sys.GetSystemInfo()
		netSummary := sys.GetNetworkSummary()
		resp := i18n.T(lang, "status_header",
			info.Hostname, info.OpenWrtVer, info.Architecture,
			info.Uptime, info.LoadAverage,
			info.MemoryUsedMB, info.MemoryTotalMB, info.MemoryUsedPct,
			netSummary)
		_ = b.SendMessage(chatID, resp)
		return

	case cmd == "/wifi":
		summary := sys.GetNetworkSummary()
		_ = b.SendMessage(chatID, summary)
		return

	case cmd == "/diagnose" || cmd == "/heal" || cmd == "/selfheal":
		_ = b.SendMessage(chatID, "🩺 *Ejecutando diagnóstico de conectividad y protocolo de auto-sanación...*")
		out, err := b.Skills.ExecuteTool("self_healing_diagnostic", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error en diagnóstico: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("🩺 *Resultado de Diagnóstico & Auto-Sanación:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/audit" || cmd == "/security" || cmd == "/hardening":
		_ = b.SendMessage(chatID, "🛡️ *Ejecutando auditoría de seguridad y contraseñas del router...*")
		out, err := b.Skills.ExecuteTool("audit_router_security", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error en auditoría: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("🛡️ *Informe de Seguridad del Router:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/wireguard" || cmd == "/wg" || cmd == "/vpn":
		_ = b.SendMessage(chatID, "🔒 *Consultando estado de túneles WireGuard & VPN...*")
		out, err := b.Skills.ExecuteTool("manage_wireguard", map[string]interface{}{"interface": "wg0", "auto_reconnect": false})
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error en WireGuard: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("🔒 *Estado de WireGuard:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/dns" || cmd == "/doh" || cmd == "/adblock":
		_ = b.SendMessage(chatID, "🌐 *Inspeccionando privacidad DNS, cifrado DoH/DoT y AdBlock...*")
		out, err := b.Skills.ExecuteTool("inspect_dns_privacy", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al consultar privacidad DNS: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("🌐 *Informe de Privacidad DNS & AdBlock:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/flash" || cmd == "/disk" || cmd == "/overlay" || cmd == "/packages":
		_ = b.SendMessage(chatID, "💾 *Analizando espacio en Flash (/overlay) y paquetes actualizables...*")
		out, err := b.Skills.ExecuteTool("audit_flash_and_packages", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al auditar Flash: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("💾 *Informe de Espacio en Flash:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/mwan" || cmd == "/multiwan" || cmd == "/failover":
		out, err := b.Skills.ExecuteTool("check_multiwan_status", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error en Multi-WAN: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("🔀 *Estado de Enlaces Multi-WAN:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/sticky" || cmd == "/roaming":
		_ = b.SendMessage(chatID, "📶 *Detectando clientes WiFi pegajosos con señal débil (< -80 dBm)...*")
		out, err := b.Skills.ExecuteTool("manage_sticky_clients", map[string]interface{}{"interface": "wlan0", "kick_weakest": false})
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al inspeccionar clientes WiFi: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("📶 *Estado de Señal & Clientes Pegajosos:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/optimize" || cmd == "/wifiopt":
		_ = b.SendMessage(chatID, "✨ *Analizando espectro inalámbrico y congestión de canales vecinos...*")
		out, err := b.Skills.ExecuteTool("optimize_wifi_channels", map[string]interface{}{"interface": "wlan0", "apply": false})
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al analizar canales WiFi: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("📶 *Informe de Optimización WiFi:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/conntrack" || cmd == "/guard" || cmd == "/traffic":
		_ = b.SendMessage(chatID, "🛡️ *Inspeccionando tabla conntrack y evaluando posibles anomalías de tráfico...*")
		out, err := b.Skills.ExecuteTool("analyze_conntrack_traffic", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al analizar tráfico: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("🛡️ *Informe de Tráfico & Conexiones Activas:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/sqm" || cmd == "/qos" || cmd == "/bufferbloat":
		out, err := b.Skills.ExecuteTool("manage_sqm_qos", map[string]interface{}{"action": "status"})
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al consultar SQM: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("⚡ *Estado de SQM & Calidad Bufferbloat:*\n```json\n%s\n```", out))
		}
		return

	case cmd == "/clients" || cmd == "/leases" || cmd == "/dhcp":
		out, err := b.Skills.ExecuteTool("get_dhcp_leases", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al obtener clientes DHCP: %v", err))
		} else {
			_ = b.SendMessage(chatID, out)
		}
		return

	case cmd == "/qrwifi" || cmd == "/qr":
		out, err := b.Skills.ExecuteTool("get_wifi_qr", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al generar código QR WiFi: %v", err))
		} else {
			_ = b.SendMessage(chatID, out)
		}
		return

	case cmd == "/scan":
		_ = b.SendMessage(chatID, "🛡️ *Iniciando escáner de puertos inseguros en la LAN...*")
		out, err := b.Skills.ExecuteTool("scan_lan_ports", nil)
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error en el escáner de puertos: %v", err))
		} else {
			_ = b.SendMessage(chatID, out)
		}
		return

	case cmd == "/firewall":
		out, err := b.Skills.ExecuteTool("read_uci_config", map[string]interface{}{"package": "firewall"})
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al leer cortafuegos: %v", err))
		} else {
			if len(out) > 3000 {
				out = out[:3000] + "\n... [salida truncada]"
			}
			_ = b.SendMessage(chatID, fmt.Sprintf("🔥 *Configuración de Cortafuegos UCI:*\n```text\n%s\n```", out))
		}
		return

	case cmd == "/routes":
		outCmd, err := exec.Command("/sbin/ip", "route").Output()
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al consultar tabla de rutas: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("🌐 *Tabla de Enrutamiento IPv4:*\n```text\n%s\n```", string(outCmd)))
		}
		return

	case cmd == "/ping":
		host := "8.8.8.8"
		fields := strings.Fields(text)
		if len(fields) > 1 {
			host = fields[1]
		}
		outCmd, err := exec.Command("/bin/ping", "-c", "3", host).Output()
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al hacer ping a %s: %v", host, err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("⚡ *Prueba de Latencia (Ping a %s):*\n```text\n%s\n```", host, string(outCmd)))
		}
		return

	case cmd == "/logs":
		outCmd, err := exec.Command("/sbin/logread", "-e", "clawrt", "-l", "20").Output()
		if err != nil {
			_ = b.SendMessage(chatID, fmt.Sprintf("❌ Error al leer registros: %v", err))
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("📜 *Registros del Sistema (Últimas 20 líneas):*\n```text\n%s\n```", string(outCmd)))
		}
		return

	case cmd == "/memory" || cmd == "/gc":
		info := sys.GetSystemInfo()
		bBefore, bAfter := sys.GetMemoryManager().ForceGC()
		_ = b.SendMessage(chatID, fmt.Sprintf("🧠 *Optimización de Memoria RAM*\n\n• *Hardware Tier:* `%s`\n• *RAM Utilizada:* %d MB / %d MB (%.1f%%)\n• *Memoria Alloc Antes de GC:* %d KB\n• *Memoria Alloc Después de GC:* %d KB\n\n✅ *Garbage Collector & FreeOSMemory() ejecutados con éxito.*",
			b.Knowledge.Tier, info.MemoryUsedMB, info.MemoryTotalMB, info.MemoryUsedPct, bBefore/1024, bAfter/1024))
		return

	case cmd == "/db":
		if b.Config.ExternalDBProvider == "supabase" {
			sc := store.NewSupabaseClient(b.Config.ExternalDBURL, b.Config.ExternalDBToken)
			err := sc.Ping()
			if err != nil {
				_ = b.SendMessage(chatID, fmt.Sprintf("❌ *Fallo al conectar con Supabase:* %v", err))
			} else {
				_ = b.SendMessage(chatID, "✅ *Conexión Exitosa con Supabase (PostgreSQL REST + Realtime API).*")
			}
		} else {
			_ = b.SendMessage(chatID, fmt.Sprintf("ℹ️ *Proveedor de Base de Datos Externa:* `%s` (Configura Supabase en `/etc/config/clawrt` para almacenamiento persistente).", b.Config.ExternalDBProvider))
		}
		return

	case cmd == "/models" || cmd == "/v1models":
		_ = b.SendMessage(chatID, "🔍 *Consultando modelos disponibles en el proveedor LLM...*")
		models, err := b.LLMClient.FetchModels()
		if err != nil || len(models) == 0 {
			_ = b.SendMessage(chatID, fmt.Sprintf("⚠️ No se pudieron obtener modelos del proveedor configurado (%s).\n\nModelos sugeridos para Bynara AI:\n• `deepseek-v4-flash-free`\n• `agnes-2.5-flash`\n• `agnes-2.0-flash`\n• `grok-4.5-free`\n• `laguna-s-2.1`\n• `ling-3.0-flash-free`\n• `mimo-v2.5-free`\n• `mistral-large`\n• `tencent-hy3-free`", b.Config.Provider))
		} else {
			resp := fmt.Sprintf("🔍 *Modelos Disponibles en %s (%d encontrados):*\n\n", b.Config.Provider, len(models))
			for i, m := range models {
				if i >= 15 {
					resp += fmt.Sprintf("... y %d modelos más.", len(models)-i)
					break
				}
				resp += fmt.Sprintf("• `%s`\n", m)
			}
			_ = b.SendMessage(chatID, resp)
		}
		return

	case cmd == "/clear":
		_ = os.Remove("/tmp/clawrt_facts.json")
		b.Knowledge.ClearFacts()
		_ = b.SendMessage(chatID, "🧹 *Memoria de hechos aprendidos vaciada correctamente.*")
		return

	case cmd == "/reboot":
		_ = b.SendMessage(chatID, "🔄 *Reiniciando servicio ClawRT...*")
		go func() {
			time.Sleep(2 * time.Second)
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

// ProcessDirectQuery processes a user instruction directly (e.g. from LuCI Web Copilot or CLI)
func (b *Bot) ProcessDirectQuery(userQuery string) string {
	userQuery = strings.TrimSpace(userQuery)
	if userQuery == "" {
		return "Instrucción vacía."
	}

	// 1. Check FastPath L1
	if directAns, ok := b.FastPath.TryDirectAnswer(userQuery); ok {
		return directAns
	}

	// 2. Check FastPath L2
	if quickAns, ok := b.FastPath.TryQuickRoute(userQuery, b.Skills); ok {
		return quickAns
	}

	// 3. Check Local Query Cache (0 tokens, 0ms)
	cache := fastpath.GetQueryCache()
	if cachedAns, ok := cache.Get(userQuery); ok {
		return fmt.Sprintf("⚡ *Caché Local (0 tokens):*\n%s", cachedAns)
	}

	// 4. Fallback to LLM ReAct loop
	lang := i18n.ES
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

	for iter := 0; iter < maxIter; iter++ {
		respMsg, err := b.LLMClient.ChatCompletion(messages, toolDefs)
		if err != nil {
			// Check if network/offline failure -> return Offline Rescue Answer
			return fastpath.OfflineRescueAnswer(userQuery)
		}
		messages = append(messages, *respMsg)

		if len(respMsg.ToolCalls) == 0 {
			cache.Set(userQuery, respMsg.Content)
			return respMsg.Content
		}

		for _, tc := range respMsg.ToolCalls {
			toolName, valid := security.SanitizeToolName(tc.Function.Name)
			if !valid {
				return "⚠️ Herramienta no válida detectada."
			}
			var args map[string]interface{}
			if tc.Function.Arguments != "" {
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			}
			output, err := b.Skills.ExecuteTool(toolName, args)
			toolRes := output
			if err != nil {
				toolRes = fmt.Sprintf("Error: %v", err)
			}
			messages = append(messages, llm.ChatMessage{
				Role:       "tool",
				Name:       toolName,
				ToolCallID: tc.ID,
				Content:    toolRes,
			})
		}
	}
	return "⚠️ Límite de iteraciones alcanzado."
}
