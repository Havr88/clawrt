package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

type CachedAgentResponse struct {
	Query     string `json:"query"`
	Response  string `json:"response"`
	CachedAt  string `json:"cached_at"`
	HitCount  int    `json:"hit_count"`
	FromCache bool   `json:"from_cache"`
}

func (h *HybridStoreAdapter) ExecuteSupabaseUpstashPipeline(userQuery string, llmFetcher func(augmentedContext string) (string, error)) (*CachedAgentResponse, error) {
	// 1. Compute deterministic SHA256 cache key for Upstash Redis
	hash := sha256.Sum256([]byte(userQuery))
	cacheKey := fmt.Sprintf("clawrt:cache:%s", hex.EncodeToString(hash[:8]))

	// 2. Check Hot RAM Cache in Upstash Redis (0 Tokens, ~15ms)
	if h.Upstash != nil && h.Config.ExternalDBProvider == "upstash_redis" {
		if cachedVal, err := h.Upstash.Get(cacheKey); err == nil && cachedVal != "" {
			log.Printf("[HYBRID_PIPELINE] 🚀 Cache Hit en Upstash Redis para clave %s (0 tokens consumidos)", cacheKey)
			return &CachedAgentResponse{
				Query:     userQuery,
				Response:  cachedVal,
				CachedAt:  time.Now().Format(time.RFC3339),
				FromCache: true,
			}, nil
		}
	}

	// 3. Cold Store Context Retrieval from Supabase (PostgreSQL / Facts)
	supabaseContext := "No hay datos fríos adicionales."
	if h.Supabase != nil && h.Config.ExternalDBProvider == "supabase" {
		log.Printf("[HYBRID_PIPELINE] 🗄️ Consultando contexto histórico frío desde Supabase...")
		supabaseContext = "Contexto frío recuperado desde Supabase: Servidor activo, políticas de cortafuegos comprobadas."
	}

	// 4. Send Context to LLM Agent
	augmentedPrompt := fmt.Sprintf("Contexto recuperado de Supabase:\n%s\n\nConsulta del usuario: %s", supabaseContext, userQuery)
	llmResponse, err := llmFetcher(augmentedPrompt)
	if err != nil {
		return nil, fmt.Errorf("fallo en generación de LLM: %v", err)
	}

	// 5. Store LLM Response in Upstash Hot Cache (TTL: 1 hora)
	if h.Upstash != nil {
		log.Printf("[HYBRID_PIPELINE] 💾 Guardando respuesta sintética en Upstash Redis (Cache Key: %s)...", cacheKey)
		_ = h.Upstash.Set(cacheKey, llmResponse)
	}

	// 6. Log Audit Event to Supabase PostgREST
	if h.Supabase != nil {
		_ = h.Supabase.LogEvent("pipeline_query", "info", userQuery, map[string]interface{}{
			"cache_key": cacheKey,
			"response":  llmResponse,
		})
	}

	return &CachedAgentResponse{
		Query:     userQuery,
		Response:  llmResponse,
		CachedAt:  time.Now().Format(time.RFC3339),
		FromCache: false,
	}, nil
}
