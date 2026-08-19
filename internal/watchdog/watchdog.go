package watchdog

import (
	"bytes"
	"clawrt/internal/sys"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type StageStatus string

const (
	StatusHealthy  StageStatus = "HEALTHY"
	StatusDegraded StageStatus = "DEGRADED"
	StatusDown     StageStatus = "DOWN"
)

type DiagnosticResult struct {
	Timestamp          time.Time   `json:"timestamp"`
	OverallStatus      StageStatus `json:"overall_status"`
	GatewayReachable   bool        `json:"gateway_reachable"`
	GatewayIP          string      `json:"gateway_ip"`
	LocalDNSWorking    bool        `json:"local_dns_working"`
	PublicDNSWorking   bool        `json:"public_dns_working"`
	InternetPingOK     bool        `json:"internet_ping_ok"`
	LatencyMs          float64     `json:"latency_ms"`
	WANInterfaceStatus string      `json:"wan_interface_status"`
	DiagnosisSummary   string      `json:"diagnosis_summary"`
	RecoveryActions    []string    `json:"recovery_actions,omitempty"`
}

type WatchdogEngine struct {
	mu               sync.RWMutex
	checkInterval    time.Duration
	consecutiveFails int
	lastDiagnostic   *DiagnosticResult
	healingActive    bool
	onAlertFunc      func(msg string)
}

var (
	instance *WatchdogEngine
	once     sync.Once
)

func GetWatchdog() *WatchdogEngine {
	once.Do(func() {
		instance = &WatchdogEngine{
			checkInterval: 60 * time.Second,
			healingActive: true,
		}
	})
	return instance
}

func (w *WatchdogEngine) SetAlertCallback(fn func(msg string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onAlertFunc = fn
}

func (w *WatchdogEngine) Start(ctx context.Context) {
	log.Println("[WATCHDOG] Iniciando motor autónomo de auto-sanación (Self-Healing Watchdog)...")
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[WATCHDOG] Watchdog detenido.")
			return
		case <-ticker.C:
			res := w.RunDiagnostic()
			if res.OverallStatus != StatusHealthy {
				w.mu.Lock()
				w.consecutiveFails++
				fails := w.consecutiveFails
				w.mu.Unlock()

				log.Printf("[WATCHDOG] Falla de conectividad detectada (%s, fallos consecutivos: %d)", res.OverallStatus, fails)
				if w.healingActive && fails >= 1 {
					recovered, actions := w.AutoHeal(res)
					res.RecoveryActions = actions
					if recovered {
						w.mu.Lock()
						w.consecutiveFails = 0
						w.mu.Unlock()
						alert := fmt.Sprintf("🛡️ *Auto-Sanación Ejecutada con Éxito*\n• *Problema:* %s\n• *Acciones tomadas:* %s\n• *Estado final:* ✅ Restaurado",
							res.DiagnosisSummary, strings.Join(actions, " ➡️ "))
						w.notify(alert)
					} else if fails == 3 {
						alert := fmt.Sprintf("🚨 *Alerta Crítica de Red (Auto-Sanación no pudo restaurar conexión):*\n• *Diagnóstico:* %s\n• *Acciones intentadas:* %s",
							res.DiagnosisSummary, strings.Join(actions, " ➡️ "))
						w.notify(alert)
					}
				}
			} else {
				w.mu.Lock()
				w.consecutiveFails = 0
				w.mu.Unlock()
			}
		}
	}
}

func (w *WatchdogEngine) RunDiagnostic() *DiagnosticResult {
	res := &DiagnosticResult{
		Timestamp:          time.Now(),
		OverallStatus:      StatusHealthy,
		RecoveryActions:    make([]string, 0),
		GatewayIP:          getIPv4DefaultGateway(),
		WANInterfaceStatus: getWANStatus(),
	}

	// 1. Check Gateway Reachability
	if res.GatewayIP != "" {
		res.GatewayReachable = pingHost(res.GatewayIP, 1)
	} else {
		res.GatewayReachable = false
	}

	// 2. Check Local DNS (127.0.0.1 / dnsmasq)
	res.LocalDNSWorking = testDNSResolution("127.0.0.1:53", "openwrt.org")

	// 3. Check Public DNS (1.1.1.1 / 8.8.8.8)
	res.PublicDNSWorking = testDNSResolution("1.1.1.1:53", "google.com") || testDNSResolution("8.8.8.8:53", "cloudflare.com")

	// 4. Check Internet Reachability & Latency
	start := time.Now()
	res.InternetPingOK = pingHost("1.1.1.1", 2) || pingHost("8.8.8.8", 2)
	res.LatencyMs = float64(time.Since(start).Milliseconds()) / 2.0

	// 5. Evaluate Overall Status
	if !res.InternetPingOK && !res.PublicDNSWorking {
		if !res.GatewayReachable {
			res.OverallStatus = StatusDown
			res.DiagnosisSummary = "Enlace físico o puerta de enlace WAN desconectada"
		} else {
			res.OverallStatus = StatusDown
			res.DiagnosisSummary = "Puerta de enlace responde, pero no hay salida a Internet (posible corte ISP)"
		}
	} else if !res.LocalDNSWorking && res.PublicDNSWorking {
		res.OverallStatus = StatusDegraded
		res.DiagnosisSummary = "Fallo en servidor DNS local (dnsmasq colgado o caché corrupto)"
	} else if res.LatencyMs > 300 {
		res.OverallStatus = StatusDegraded
		res.DiagnosisSummary = "Latencia excesivamente alta (posible congestión o saturación de cola)"
	} else {
		res.OverallStatus = StatusHealthy
		res.DiagnosisSummary = "Conectividad WAN, DNS y Gateway en estado óptimo"
	}

	w.mu.Lock()
	w.lastDiagnostic = res
	w.mu.Unlock()

	return res
}

func (w *WatchdogEngine) AutoHeal(diag *DiagnosticResult) (bool, []string) {
	actions := make([]string, 0)

	// Stage 1: Fix Local DNS if local DNS is down but public ping is OK
	if !diag.LocalDNSWorking && diag.PublicDNSWorking {
		log.Println("[HEALING] Etapa 1: Reiniciando dnsmasq para recuperar resolución DNS local...")
		_, _ = sys.ExecuteTypedServiceRestart("dnsmasq")
		actions = append(actions, "Reiniciar servicio dnsmasq")
		time.Sleep(2 * time.Second)
		if testDNSResolution("127.0.0.1:53", "openwrt.org") {
			return true, actions
		}
	}

	// Stage 2: Renegotiate WAN Interface (ubus / ifup)
	if !diag.InternetPingOK {
		log.Println("[HEALING] Etapa 2: Renegociando interfaz WAN (ifdown / ifup)...")
		_ = exec.Command("ubus", "call", "network.interface.wan", "down").Run()
		time.Sleep(1 * time.Second)
		_ = exec.Command("ubus", "call", "network.interface.wan", "up").Run()
		_ = exec.Command("/sbin/ifup", "wan").Run()
		actions = append(actions, "Renegociar interfaz WAN (ifup wan)")
		time.Sleep(4 * time.Second)

		if pingHost("1.1.1.1", 2) || pingHost("8.8.8.8", 2) {
			return true, actions
		}
	}

	// Stage 3: Reload Network & Firewall Subsystem
	if !diag.InternetPingOK {
		log.Println("[HEALING] Etapa 3: Recargando subsistema de red y cortafuegos (fw4 / network)...")
		_, _ = sys.ExecuteTypedServiceRestart("firewall")
		_, _ = sys.ExecuteTypedServiceRestart("network")
		actions = append(actions, "Recargar red y firewall")
		time.Sleep(3 * time.Second)

		if pingHost("1.1.1.1", 2) {
			return true, actions
		}
	}

	return false, actions
}

func (w *WatchdogEngine) GetLastDiagnostic() *DiagnosticResult {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.lastDiagnostic == nil {
		return w.RunDiagnostic()
	}
	return w.lastDiagnostic
}

func (w *WatchdogEngine) notify(msg string) {
	if w.onAlertFunc != nil {
		w.onAlertFunc(msg)
	}
}

// Helpers
func pingHost(host string, count int) bool {
	cmd := exec.Command("/bin/ping", "-c", fmt.Sprintf("%d", count), "-W", "2", host)
	return cmd.Run() == nil
}

func testDNSResolution(serverAddr, domain string) bool {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 2 * time.Second}
			return d.DialContext(ctx, "udp", serverAddr)
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ips, err := r.LookupHost(ctx, domain)
	return err == nil && len(ips) > 0
}

func getIPv4DefaultGateway() string {
	out, err := exec.Command("/sbin/ip", "route").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "default via ") {
			parts := strings.Fields(l)
			if len(parts) >= 3 {
				return parts[2]
			}
		}
	}
	return ""
}

func getWANStatus() string {
	out, err := exec.Command("ubus", "call", "network.interface.wan", "status").Output()
	if err != nil || len(out) == 0 {
		return "desconocido"
	}
	if bytes.Contains(out, []byte(`"up": true`)) {
		return "up"
	}
	return "down"
}

func CheckHTTPReachability(targetURL string) bool {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
