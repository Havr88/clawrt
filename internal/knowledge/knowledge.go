package knowledge

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

//go:embed embedded/*.json
var embeddedFS embed.FS

type HardwareTier string

const (
	TierMinimal HardwareTier = "minimal" // <= 64MB RAM (Xiaomi Mi Router 4C)
	TierMedium  HardwareTier = "medium"  // 64MB - 256MB RAM
	TierFull    HardwareTier = "full"    // > 256MB RAM (Linksys Velop WHW01)
)

type ProgramInfo struct {
	Name         string   `json:"name"`
	Package      string   `json:"package"`
	Config       string   `json:"config"`
	Control      string   `json:"control"`
	Logs         string   `json:"logs"`
	Dependencies []string `json:"dependencies"`
}

type Recipe struct {
	Name        string `json:"name"`
	Action      string `json:"action"`
	UCICache    string `json:"uci_cache"`
	ConfirmedOK bool   `json:"confirmed_ok"`
}

type KnowledgeEngine struct {
	mu            sync.RWMutex
	Tier          HardwareTier       `json:"tier"`
	PackageManager string            `json:"package_manager"`
	Programs      []ProgramInfo      `json:"programs"`
	Schemas       map[string]interface{} `json:"schemas"`
	ConfigWays    map[string]string  `json:"config_ways"`
	KeyFiles      map[string]string  `json:"key_files"`
	Recipes       map[string]Recipe  `json:"recipes"`
	MutableFacts  map[string]string  `json:"mutable_facts"`
	storagePath   string
}

func NewKnowledgeEngine(memTotalMB int, storagePath string) *KnowledgeEngine {
	if storagePath == "" {
		storagePath = "/tmp/clawrt_facts.json"
	}

	tier := TierMinimal
	if memTotalMB > 256 {
		tier = TierFull
	} else if memTotalMB > 64 {
		tier = TierMedium
	}

	ke := &KnowledgeEngine{
		Tier:         tier,
		Schemas:      make(map[string]interface{}),
		ConfigWays:   make(map[string]string),
		KeyFiles:     make(map[string]string),
		Recipes:      make(map[string]Recipe),
		MutableFacts: make(map[string]string),
		storagePath:  storagePath,
	}

	ke.detectPackageManager()
	ke.loadEmbeddedData()
	ke.loadConfigWays()
	ke.loadKeyFilesMap()
	ke.LoadMutableFacts()

	return ke
}

func (k *KnowledgeEngine) detectPackageManager() {
	if _, err := exec.LookPath("apk"); err == nil {
		k.PackageManager = "apk (OpenWrt 25.x+)"
	} else if _, err := exec.LookPath("opkg"); err == nil {
		k.PackageManager = "opkg (OpenWrt <=23.05/24.10)"
	} else {
		k.PackageManager = "desconocido"
	}
}

func (k *KnowledgeEngine) loadEmbeddedData() {
	// 1. Cargar catálogo de programas desde embedded/programs.json
	if progData, err := embeddedFS.ReadFile("embedded/programs.json"); err == nil {
		_ = json.Unmarshal(progData, &k.Programs)
	}

	// 2. Cargar esquemas UCI desde embedded/uci_schemas.json
	if schemaData, err := embeddedFS.ReadFile("embedded/uci_schemas.json"); err == nil {
		_ = json.Unmarshal(schemaData, &k.Schemas)
	}
}

func (k *KnowledgeEngine) loadConfigWays() {
	k.ConfigWays["1_uci_cli"] = "1. UCI CLI: uci set network.lan.ipaddr=... && uci commit network && /etc/init.d/network reload"
	k.ConfigWays["2_uci_api"] = "2. UCI API / C / Lua: acceso directo libuci (para desarrollo C/Lua)"
	k.ConfigWays["3_file_edit"] = "3. Edición directa de archivos: /etc/config/*, /etc/crontabs/root, /etc/firewall.user, /etc/rc.local"
	k.ConfigWays["4_luci_rpc"] = "4. LuCI RPC / ubus: llamadas ubus uci get/set vía RPC HTTP web"
	k.ConfigWays["5_init_scripts"] = "5. Init Scripts / procd: service <name> start|stop|restart|reload|enable|disable (/etc/init.d/*)"
	k.ConfigWays["6_hotplug"] = "6. Hotplug Scripts: /etc/hotplug.d/* (iface, button, usb, net)"
	k.ConfigWays["7_netifd_rpc"] = "7. netifd RPC: ubus call network.interface ... (cambios dinámicos en caliente)"
	k.ConfigWays["8_fw4_nftables"] = "8. fw4/nftables: nft list ruleset, reglas en /etc/nftables.d/ o /etc/config/firewall"
	k.ConfigWays["9_crontab"] = "9. Crontab / Tareas programadas: /etc/crontabs/root y service cron restart"
	k.ConfigWays["10_sysupgrade"] = "10. Sysupgrade: preservación de archivos de config en /etc/sysupgrade.conf"
}

func (k *KnowledgeEngine) loadKeyFilesMap() {
	k.KeyFiles["uci_configs"] = "/etc/config/* (network, wireless, firewall, dhcp, system, dropbear, luci)"
	k.KeyFiles["startup"] = "/etc/rc.local (comandos de inicio personalizado)"
	k.KeyFiles["crontab"] = "/etc/crontabs/root (tareas cron del sistema)"
	k.KeyFiles["hotplug"] = "/etc/hotplug.d/* (scripts reactivos de eventos)"
	k.KeyFiles["sysupgrade"] = "/etc/sysupgrade.conf (lista de archivos a preservar al actualizar firmware)"
	k.KeyFiles["dropbear_keys"] = "/etc/dropbear (llaves de host SSH)"
	k.KeyFiles["firewall_user"] = "/etc/firewall.user (reglas personalizadas de firewall)"
}

func (k *KnowledgeEngine) GetProgramInfo(name string) (*ProgramInfo, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	nameLower := strings.ToLower(name)
	for _, prog := range k.Programs {
		if strings.ToLower(prog.Name) == nameLower {
			return &prog, true
		}
	}
	return nil, false
}

func (k *KnowledgeEngine) SetFact(key, value string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.MutableFacts[key] = value
	_ = k.saveMutableFactsLocked()
}

func (k *KnowledgeEngine) GetFact(key string) (string, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	val, ok := k.MutableFacts[key]
	return val, ok
}

func (k *KnowledgeEngine) SaveRecipe(name, action, uciCache string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.Recipes[name] = Recipe{
		Name:        name,
		Action:      action,
		UCICache:    uciCache,
		ConfirmedOK: true,
	}
	_ = k.saveMutableFactsLocked()
}

func (k *KnowledgeEngine) GetContextSummary() string {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🧠 OpenWrt Knowledge Base (Hardware Tier: %s, Package Manager: %s):\n", k.Tier, k.PackageManager))
	sb.WriteString("Formas de Configuración Conocidas (10 Caminos):\n")
	for _, way := range k.ConfigWays {
		sb.WriteString(fmt.Sprintf(" • %s\n", way))
	}

	if len(k.MutableFacts) > 0 {
		sb.WriteString("\nHechos Dinámicos Aprendidos:\n")
		for key, val := range k.MutableFacts {
			sb.WriteString(fmt.Sprintf(" • %s = %s\n", key, val))
		}
	}

	return sb.String()
}

func (k *KnowledgeEngine) LoadMutableFacts() {
	k.mu.Lock()
	defer k.mu.Unlock()

	data, err := os.ReadFile(k.storagePath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &k.MutableFacts)
}

func (k *KnowledgeEngine) saveMutableFactsLocked() error {
	data, err := json.MarshalIndent(k.MutableFacts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(k.storagePath, data, 0644)
}
