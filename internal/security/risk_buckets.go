package security

import (
	"regexp"
	"strings"
	"sync"
)

type RiskBucket string
type Decision string

const (
	BucketReadonly     RiskBucket = "readonly"
	BucketMutating     RiskBucket = "mutating"
	BucketControlPlane RiskBucket = "control_plane"

	DecisionAllowOnce   Decision = "allow_once"
	DecisionAllowAlways Decision = "allow_always"
	DecisionDeny        Decision = "deny"
)

var validToolNameRegex = regexp.MustCompile(`^[a-z0-9._-]+$`)

type RiskClassifier struct {
	mu           sync.RWMutex
	SessionRules map[string]map[string]Decision // chatID -> toolName -> Decision
	DenialTallies map[string]int                // chatID -> consecutive denials
}

func NewRiskClassifier() *RiskClassifier {
	return &RiskClassifier{
		SessionRules:  make(map[string]map[string]Decision),
		DenialTallies: make(map[string]int),
	}
}

// SanitizeToolName validates and normalizes tool names against prompt injection
func SanitizeToolName(name string) (string, bool) {
	name = strings.TrimSpace(strings.ToLower(name))
	if len(name) > 128 || !validToolNameRegex.MatchString(name) {
		return "", false
	}
	return name, true
}

// ClassifyTool assigns a risk bucket to each tool
func ClassifyTool(name string) RiskBucket {
	sanitized, ok := SanitizeToolName(name)
	if !ok {
		return BucketControlPlane // Unknown/invalid tool names strictly treated as high risk
	}

	switch sanitized {
	case "get_system_info", "get_network_status", "read_uci_config":
		return BucketReadonly
	case "exec_safe_cmd":
		return BucketReadonly
	case "write_uci_config", "restart_service":
		return BucketMutating
	case "reboot", "sysupgrade", "change_model", "nanoclaw_update":
		return BucketControlPlane
	default:
		return BucketMutating
	}
}

func (rc *RiskClassifier) EvaluateToolPermission(chatID string, toolName string) (bool, Decision) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	bucket := ClassifyTool(toolName)
	if bucket == BucketReadonly {
		return true, DecisionAllowAlways
	}

	// Check if chat session has a saved "allow_always" rule
	if rules, exists := rc.SessionRules[chatID]; exists {
		if dec, found := rules[toolName]; found {
			if dec == DecisionAllowAlways {
				return true, DecisionAllowAlways
			}
			if dec == DecisionDeny {
				return false, DecisionDeny
			}
		}
	}

	// Default conservative fallback
	return false, DecisionAllowOnce
}

func (rc *RiskClassifier) RecordDecision(chatID string, toolName string, dec Decision) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, exists := rc.SessionRules[chatID]; !exists {
		rc.SessionRules[chatID] = make(map[string]Decision)
	}
	rc.SessionRules[chatID][toolName] = dec

	if dec == DecisionDeny {
		rc.DenialTallies[chatID]++
	} else {
		rc.DenialTallies[chatID] = 0 // Reset on approval
	}
}

func (rc *RiskClassifier) ShouldHardStop(chatID string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.DenialTallies[chatID] >= 3
}

func (rc *RiskClassifier) ResetDenialTally(chatID string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.DenialTallies[chatID] = 0
}
