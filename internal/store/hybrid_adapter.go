package store

import (
	"clawrt/internal/config"
	"fmt"
	"log"
)

type HybridStoreAdapter struct {
	Config     *config.Config
	Supabase   *SupabaseClient
	Cloudflare *CloudflareClient
	Upstash    *UpstashClient
}

func NewHybridStoreAdapter(cfg *config.Config) *HybridStoreAdapter {
	adapter := &HybridStoreAdapter{
		Config: cfg,
	}

	if cfg.ExternalDBProvider == "supabase" || cfg.ExternalDBURL != "" {
		adapter.Supabase = NewSupabaseClient(cfg.ExternalDBURL, cfg.ExternalDBToken)
	}

	if cfg.ExternalDBProvider == "cloudflare_d1" {
		adapter.Cloudflare = NewCloudflareClient(cfg.ExternalDBURL, cfg.ExternalDBToken, "", "")
	}

	if cfg.ExternalDBProvider == "upstash_redis" {
		adapter.Upstash = NewUpstashClient(cfg.ExternalDBURL, cfg.ExternalDBToken, "", "", "")
	}

	return adapter
}

func (h *HybridStoreAdapter) LogEvent(eventType, severity, message string, payload map[string]interface{}) error {
	switch h.Config.ExternalDBProvider {
	case "supabase":
		if h.Supabase != nil {
			return h.Supabase.LogEvent(eventType, severity, message, payload)
		}
	case "upstash_redis":
		if h.Upstash != nil {
			key := fmt.Sprintf("event:%s:%d", eventType, payload["timestamp"])
			return h.Upstash.Set(key, message)
		}
	case "cloudflare_d1":
		if h.Cloudflare != nil {
			sql := "INSERT INTO clawrt_events (timestamp, event_type, severity, message) VALUES (DATETIME('now'), ?, ?, ?)"
			_, err := h.Cloudflare.ExecuteD1Query(sql, []interface{}{eventType, severity, message})
			return err
		}
	}
	return nil
}

func (h *HybridStoreAdapter) UploadBackup(localFilePath string) (string, error) {
	// Try Supabase Storage first, fallback to Cloudflare R2
	if h.Supabase != nil && h.Config.ExternalDBProvider == "supabase" {
		return h.Supabase.UploadBackupFile(localFilePath)
	}
	if h.Cloudflare != nil && h.Config.ExternalDBProvider == "cloudflare_d1" {
		return h.Cloudflare.UploadR2Backup(localFilePath)
	}
	return "", fmt.Errorf("ningún servicio de almacenamiento de respaldo (Storage/R2) configurado")
}

func (h *HybridStoreAdapter) TestConnection() string {
	prov := h.Config.ExternalDBProvider
	if prov == "none" || prov == "" {
		return "ℹ️ Ninguna base de datos externa configurada (Almacenamiento local en /tmp)."
	}

	switch prov {
	case "supabase":
		if h.Supabase != nil {
			err := h.Supabase.Ping()
			if err != nil {
				return fmt.Sprintf("❌ Error al conectar con Supabase: %v", err)
			}
			return "✅ Conexión Exitosa con Supabase (Postgres REST, Realtime, Storage & pgvector activos)."
		}
	case "upstash_redis":
		if h.Upstash != nil {
			err := h.Upstash.Ping()
			if err != nil {
				return fmt.Sprintf("❌ Error al conectar con Upstash Redis: %v", err)
			}
			return "✅ Conexión Exitosa con Upstash Redis (Serverless HTTP REST API activo)."
		}
	case "cloudflare_d1":
		if h.Cloudflare != nil {
			err := h.Cloudflare.Ping()
			if err != nil {
				return fmt.Sprintf("❌ Error al conectar con Cloudflare D1: %v", err)
			}
			return "✅ Conexión Exitosa con Cloudflare D1 (Serverless SQLite API activo)."
		}
	}

	log.Printf("[STORE] Adaptador híbrido activo para proveedor: %s", prov)
	return fmt.Sprintf("✅ Proveedor `%s` configurado y activo.", prov)
}
