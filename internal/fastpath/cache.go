package fastpath

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
)

type CacheItem struct {
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type QueryCache struct {
	mu        sync.RWMutex
	items     map[string]CacheItem
	cachePath string
	ttl       time.Duration
}

var (
	cacheInstance *QueryCache
	cacheOnce     sync.Once
)

func GetQueryCache() *QueryCache {
	cacheOnce.Do(func() {
		cacheInstance = &QueryCache{
			items:     make(map[string]CacheItem),
			cachePath: "/tmp/clawrt_query_cache.json",
			ttl:       15 * time.Minute,
		}
		cacheInstance.load()
	})
	return cacheInstance
}

func (qc *QueryCache) hashKey(query string) string {
	clean := strings.TrimSpace(strings.ToLower(query))
	h := sha256.Sum256([]byte(clean))
	return hex.EncodeToString(h[:])
}

func (qc *QueryCache) Get(query string) (string, bool) {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	key := qc.hashKey(query)
	item, ok := qc.items[key]
	if !ok {
		return "", false
	}

	if time.Now().After(item.ExpiresAt) {
		return "", false
	}

	return item.Response, true
}

func (qc *QueryCache) Set(query, response string) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	key := qc.hashKey(query)
	now := time.Now()
	qc.items[key] = CacheItem{
		Response:  response,
		CreatedAt: now,
		ExpiresAt: now.Add(qc.ttl),
	}
	_ = qc.saveLocked()
}

func (qc *QueryCache) load() {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	data, err := os.ReadFile(qc.cachePath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &qc.items)
}

func (qc *QueryCache) saveLocked() error {
	data, err := json.Marshal(qc.items)
	if err != nil {
		return err
	}
	return os.WriteFile(qc.cachePath, data, 0644)
}

// OfflineRescueAnswer provides local deterministic guidance when internet is down
func OfflineRescueAnswer(query string) string {
	q := strings.ToLower(query)
	if strings.Contains(q, "internet") || strings.Contains(q, "caida") || strings.Contains(q, "down") || strings.Contains(q, "conectar") {
		return "🌐 *Modo de Rescate Offline Activado:*\n• La conexión WAN externa no está disponible.\n• Se recomienda ejecutar `/diagnose` para auto-reparar la interfaz WAN o revisar el cable de fibra/módem."
	}
	if strings.Contains(q, "wifi") {
		return "📶 *Modo de Rescate Offline:*\n• La red WiFi local sigue operativa. Puedes consultar `/status` o generar el código QR con `/qrwifi`."
	}
	return "⚠️ *Modo de Rescate Offline (Sin Conexión WAN):*\nNo hay acceso a servidores de IA en la nube. Puedes ejecutar comandos directos de control: `/status`, `/diagnose`, `/wifi`, `/clients`, `/scan`, `/firewall`, `/reboot`."
}
