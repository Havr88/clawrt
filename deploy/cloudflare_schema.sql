-- ☁️ ClawRT — Cloudflare D1 SQL Schema
-- Ejecuta este script usando la CLI de Wrangler o el panel de Cloudflare D1.

CREATE TABLE IF NOT EXISTS clawrt_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp TEXT NOT NULL,
  event_type TEXT NOT NULL,
  severity TEXT NOT NULL,
  message TEXT NOT NULL,
  payload TEXT
);

CREATE TABLE IF NOT EXISTS clawrt_facts (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS clawrt_dhcp_clients (
  mac TEXT PRIMARY KEY,
  ip TEXT NOT NULL,
  hostname TEXT,
  vendor TEXT,
  is_random_mac INTEGER DEFAULT 0,
  last_seen TEXT NOT NULL
);
