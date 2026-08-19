package security

import (
	"regexp"
)

var SecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b[0-9]{8,10}:[A-Za-z0-9_-]{35,}\b`),                                         // Telegram Bot Token
	regexp.MustCompile(`(?i)\bsk-[A-Za-z0-9_-]{20,}\b`),                                                  // OpenAI / DeepSeek API Keys
	regexp.MustCompile(`(?i)\bgsk_[A-Za-z0-9_-]{20,}\b`),                                                 // Groq API Keys
	regexp.MustCompile(`(?i)\bbyn_[A-Za-z0-9_-]{20,}\b`),                                                 // Bynara API Keys
	regexp.MustCompile(`(?i)eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`),              // JWT Tokens (Supabase / Auth)
	regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._-]{20,}`),                                               // Bearer Header Tokens
	regexp.MustCompile(`(?i)(api_key|bot_token|db_token|password|secret|key)\s*[:=]\s*["']?[^"'\s]{8,}`), // UCI & Config Secrets
	regexp.MustCompile(`-----BEGIN [A-Z ]+ PRIVATE KEY-----[\s\S]*?-----END [A-Z ]+ PRIVATE KEY-----`),   // SSH/TLS Private Keys
}

func SanitizeSecrets(text string) string {
	result := text
	for _, pattern := range SecretPatterns {
		result = pattern.ReplaceAllString(result, "${1}: [SECRETO_REDACTADO]")
	}
	// Secondary sweep for plain tokens
	for i := 0; i < 4; i++ {
		result = SecretPatterns[i].ReplaceAllString(result, "[SECRETO_REDACTADO]")
	}
	return result
}
