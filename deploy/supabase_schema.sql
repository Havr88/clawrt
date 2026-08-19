-- 🗄️ ClawRT — Supabase Database Schema (PostgreSQL, Realtime, Storage & Vector)
-- Ejecuta este script en el "SQL Editor" de tu proyecto gratuito en Supabase.

-- 1. Tabla de Eventos y Notificaciones del Router
CREATE TABLE IF NOT EXISTS clawrt_events (
  id BIGSERIAL PRIMARY KEY,
  timestamp TIMESTAMPTZ DEFAULT NOW(),
  event_type TEXT NOT NULL,
  severity TEXT NOT NULL,
  message TEXT NOT NULL,
  payload JSONB
);

-- 2. Tabla de Memoria de Hechos Aprendidos (Fact Store)
CREATE TABLE IF NOT EXISTS clawrt_facts (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Tabla de Registro de Clientes DHCP y Dispositivos en LAN
CREATE TABLE IF NOT EXISTS clawrt_dhcp_clients (
  mac TEXT PRIMARY KEY,
  ip TEXT NOT NULL,
  hostname TEXT,
  vendor TEXT,
  is_random_mac BOOLEAN DEFAULT FALSE,
  last_seen TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Habilitar Supabase Realtime para la tabla de eventos
ALTER PUBLICATION supabase_realtime ADD TABLE clawrt_events;

-- 5. Crear el Bucket de Almacenamiento para Respaldos de OpenWrt (Storage)
INSERT INTO storage.buckets (id, name, public) 
VALUES ('router-backups', 'router-backups', false)
ON CONFLICT (id) DO NOTHING;

-- 6. Habilitar extensión Vector para Búsqueda Semántica RAG (pgvector)
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS clawrt_vector_docs (
  id BIGSERIAL PRIMARY KEY,
  content TEXT NOT NULL,
  metadata JSONB,
  embedding vector(1536)
);

-- Función de búsqueda por similitud de cosenos para pgvector
CREATE OR REPLACE FUNCTION match_openwrt_docs (
  query_embedding vector(1536),
  match_threshold float,
  match_count int
)
RETURNS TABLE (
  id bigint,
  content text,
  metadata jsonb,
  similarity float
)
LANGUAGE sql STABLE
AS $$
  SELECT
    clawrt_vector_docs.id,
    clawrt_vector_docs.content,
    clawrt_vector_docs.metadata,
    1 - (clawrt_vector_docs.embedding <=> query_embedding) AS similarity
  FROM clawrt_vector_docs
  WHERE 1 - (clawrt_vector_docs.embedding <=> query_embedding) > match_threshold
  ORDER BY similarity DESC
  LIMIT match_count;
$$;
