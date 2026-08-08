package security

import (
	"regexp"
)

var SecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b[0-9]{8,10}:[A-Za-z0-9_-]{35,}\b`), // Telegram Bot Token
	regexp.MustCompile(`(?i)\bsk-[A-Za-z0-9_-]{20,}\b`),           // OpenAI / DeepSeek API Keys
	regexp.MustCompile(`(?i)\bgsk_[A-Za-z0-9_-]{20,}\b`),          // Groq API Keys
	regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._-]{20,}`),        // Bearer Tokens
	regexp.MustCompile(`-----BEGIN [A-Z ]+ PRIVATE KEY-----[\s\S]*?-----END [A-Z ]+ PRIVATE KEY-----`), // SSH/TLS Private Keys
}

func SanitizeSecrets(text string) string {
	result := text
	for _, pattern := range SecretPatterns {
		result = pattern.ReplaceAllString(result, "[SECRETO_REDACTADO]")
	}
	return result
}
