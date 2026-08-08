package fastpath

import (
	"clawrt/internal/skills"
	"fmt"
	"regexp"
	"strings"
)

type Tier string

const (
	TierFast     Tier = "fast"
	TierBalanced Tier = "balanced"
	TierDeep     Tier = "deep"
)

type RoutingParams struct {
	Tier        Tier
	MaxTokens   int
	Temperature float64
}

type QuickRule struct {
	Pattern  *regexp.Regexp
	ToolName string
	Extract  func(matches []string) map[string]interface{}
}

type FastPathEngine struct {
	directAnswers map[string]string
	quickRules    []QuickRule
}

func NewFastPathEngine() *FastPathEngine {
	fp := &FastPathEngine{
		directAnswers: make(map[string]string),
	}
	fp.initDirectAnswers()
	fp.initQuickRules()
	return fp
}

func (fp *FastPathEngine) initDirectAnswers() {
	fp.directAnswers["hola"] = "👋 ¡Hola! Soy ClawRT. ¿En qué puedo ayudarte hoy con tu router?"
	fp.directAnswers["hello"] = "👋 Hello! I am ClawRT. How can I help you with your router today?"
	fp.directAnswers["hi"] = "👋 Hi! How can I assist you with your OpenWrt router?"
	fp.directAnswers["gracias"] = "😊 ¡De nada! Quedo a tu disposición."
	fp.directAnswers["thanks"] = "😊 You're welcome! Let me know if you need anything else."
	fp.directAnswers["ping"] = "🏓 pong"
	fp.directAnswers["test"] = "✅ ClawRT agent is running and operational."
}

func (fp *FastPathEngine) initQuickRules() {
	fp.quickRules = []QuickRule{
		{
			Pattern:  regexp.MustCompile(`(?i)^(estado|status|uptime|recursos|ram|memoria|cpu)$`),
			ToolName: "get_system_info",
			Extract:  func(m []string) map[string]interface{} { return map[string]interface{}{} },
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^(red|network|interfaces|ip|wifi|internet|clientes)$`),
			ToolName: "get_network_status",
			Extract:  func(m []string) map[string]interface{} { return map[string]interface{}{} },
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^(/leases|leases|dispositivos|equipos)$`),
			ToolName: "get_dhcp_leases",
			Extract:  func(m []string) map[string]interface{} { return map[string]interface{}{} },
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^(/qrwifi|qrwifi|qr|qr wifi)$`),
			ToolName: "get_wifi_qr",
			Extract:  func(m []string) map[string]interface{} { return map[string]interface{}{} },
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^(/scan|escanear|scan)\s+([0-9.]+)$`),
			ToolName: "scan_lan_ports",
			Extract: func(m []string) map[string]interface{} {
				return map[string]interface{}{"ip": m[2]}
			},
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^ping\s+([a-zA-Z0-9.-]+)$`),
			ToolName: "exec_safe_cmd",
			Extract: func(m []string) map[string]interface{} {
				return map[string]interface{}{"command": "ping", "target": m[1]}
			},
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^(traceroute|trace)\s+([a-zA-Z0-9.-]+)$`),
			ToolName: "exec_safe_cmd",
			Extract: func(m []string) map[string]interface{} {
				return map[string]interface{}{"command": "traceroute", "target": m[2]}
			},
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^(nslookup|dns)\s+([a-zA-Z0-9.-]+)$`),
			ToolName: "exec_safe_cmd",
			Extract: func(m []string) map[string]interface{} {
				return map[string]interface{}{"command": "nslookup", "target": m[2]}
			},
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^(logread|logs|log)$`),
			ToolName: "exec_safe_cmd",
			Extract: func(m []string) map[string]interface{} {
				return map[string]interface{}{"command": "logread"}
			},
		},
		{
			Pattern:  regexp.MustCompile(`(?i)^leer\s+uci\s+([a-zA-Z0-9._-]+)$`),
			ToolName: "read_uci_config",
			Extract: func(m []string) map[string]interface{} {
				return map[string]interface{}{"package": m[1]}
			},
		},
	}
}

func (fp *FastPathEngine) TryDirectAnswer(text string) (string, bool) {
	clean := strings.TrimSpace(strings.ToLower(text))
	if ans, ok := fp.directAnswers[clean]; ok {
		return ans, true
	}
	return "", false
}

func (fp *FastPathEngine) TryQuickRoute(text string, registry *skills.SkillRegistry) (string, bool) {
	clean := strings.TrimSpace(text)
	for _, rule := range fp.quickRules {
		matches := rule.Pattern.FindStringSubmatch(clean)
		if len(matches) > 0 {
			args := rule.Extract(matches)
			out, err := registry.ExecuteTool(rule.ToolName, args)
			if err != nil {
				return fmt.Sprintf("⚠️ FastPath error: %v", err), true
			}
			return fmt.Sprintf("⚡ *FastPath (Cero-LLM):*\n```\n%s\n```", out), true
		}
	}
	return "", false
}

func ClassifyTier(text string) RoutingParams {
	clean := strings.ToLower(text)

	deepKeywords := []string{"diseño", "design", "arquitectura", "architecture", "plan", "planificación", "diagnostico profundo", "análisis", "analysis", "refactor", "proposal"}
	for _, kw := range deepKeywords {
		if strings.Contains(clean, kw) || len(text) > 500 {
			return RoutingParams{
				Tier:        TierDeep,
				MaxTokens:   1024,
				Temperature: 0.7,
			}
		}
	}

	fastKeywords := []string{"status", "uptime", "ram", "memoria", "ping", "ip", "wifi", "hora", "time", "cpu", "leases", "qr", "scan"}
	for _, kw := range fastKeywords {
		if strings.Contains(clean, kw) {
			return RoutingParams{
				Tier:        TierFast,
				MaxTokens:   256,
				Temperature: 0.2,
			}
		}
	}

	return RoutingParams{
		Tier:        TierBalanced,
		MaxTokens:   384,
		Temperature: 0.3,
	}
}
