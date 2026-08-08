package fastpath

import (
	"clawrt/internal/skills"
	"testing"
)

func TestFastPathEngine(t *testing.T) {
	fp := NewFastPathEngine()
	registry := skills.NewRegistry()

	// 1. Test Direct Answer Layer 1
	ans, ok := fp.TryDirectAnswer("hola")
	if !ok || ans == "" {
		t.Errorf("DirectAnswer para 'hola' falló")
	}

	ansPing, ok := fp.TryDirectAnswer("ping")
	if !ok || ansPing != "🏓 pong" {
		t.Errorf("DirectAnswer para 'ping' falló, se obtuvo: %s", ansPing)
	}

	// 2. Test Quick Route Layer 2
	out, ok := fp.TryQuickRoute("status", registry)
	if !ok || out == "" {
		t.Errorf("QuickRoute para 'status' falló")
	}

	// 3. Test 3-Tier Classifier
	fastRouting := ClassifyTier("dame el status de la ram")
	if fastRouting.Tier != TierFast || fastRouting.MaxTokens != 256 || fastRouting.Temperature != 0.2 {
		t.Errorf("ClassifyTier 'status' falló, se obtuvo: %+v", fastRouting)
	}

	deepRouting := ClassifyTier("diseño de arquitectura de red para el hogar")
	if deepRouting.Tier != TierDeep || deepRouting.MaxTokens != 1024 || deepRouting.Temperature != 0.7 {
		t.Errorf("ClassifyTier 'diseño' falló, se obtuvo: %+v", deepRouting)
	}
}
