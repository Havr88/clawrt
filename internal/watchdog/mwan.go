package watchdog

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type MultiWANInterface struct {
	Name        string   `json:"name"`
	Status      string   `json:"status"` // online, offline, standby
	TrackingIPs []string `json:"tracking_ips,omitempty"`
	Score       int      `json:"score,omitempty"`
	LossPct     float64  `json:"loss_pct,omitempty"`
	LatencyMs   float64  `json:"latency_ms,omitempty"`
}

type MultiWANReport struct {
	Timestamp      time.Time           `json:"timestamp"`
	Installed      bool                `json:"installed"`
	Enabled        bool                `json:"enabled"`
	Interfaces     []MultiWANInterface `json:"interfaces"`
	ActiveWANCount int                 `json:"active_wan_count"`
	TotalWANCount  int                 `json:"total_wan_count"`
	CurrentPolicy  string              `json:"current_policy,omitempty"`
	Summary        string              `json:"summary"`
}

func CheckMultiWANStatus() (*MultiWANReport, error) {
	report := &MultiWANReport{
		Timestamp:  time.Now(),
		Interfaces: make([]MultiWANInterface, 0),
	}

	// 1. Check if mwan3 is installed
	if _, err := exec.LookPath("mwan3"); err != nil {
		report.Installed = false
		report.Summary = "ℹ️ Multi-WAN (mwan3) no está instalado en este router (configuración WAN simple)."
		return report, nil
	}
	report.Installed = true

	// 2. Query ubus call mwan3 status
	out, err := exec.Command("ubus", "call", "mwan3", "status").Output()
	if err == nil && len(out) > 0 {
		report.Enabled = true
		var ubusMWAN struct {
			Interfaces map[string]struct {
				Status  string  `json:"status"`
				Age     int64   `json:"age"`
				Score   int     `json:"score"`
				Loss    float64 `json:"loss"`
				Latency float64 `json:"latency"`
			} `json:"interfaces"`
		}
		if err := json.Unmarshal(out, &ubusMWAN); err == nil {
			for name, data := range ubusMWAN.Interfaces {
				report.TotalWANCount++
				if data.Status == "online" {
					report.ActiveWANCount++
				}
				report.Interfaces = append(report.Interfaces, MultiWANInterface{
					Name:      name,
					Status:    data.Status,
					Score:     data.Score,
					LossPct:   data.Loss,
					LatencyMs: data.Latency,
				})
			}
		}
	}

	// Fallback to mwan3 status CLI if ubus was empty
	if len(report.Interfaces) == 0 {
		outCli, _ := exec.Command("mwan3", "status").Output()
		cliStr := string(outCli)
		if strings.Contains(cliStr, "interface") {
			report.Enabled = true
			for _, line := range strings.Split(cliStr, "\n") {
				if strings.Contains(line, "is online") {
					parts := strings.Fields(line)
					if len(parts) > 1 {
						report.ActiveWANCount++
						report.TotalWANCount++
						report.Interfaces = append(report.Interfaces, MultiWANInterface{
							Name:   parts[1],
							Status: "online",
						})
					}
				}
			}
		}
	}

	if report.ActiveWANCount > 1 {
		report.Summary = fmt.Sprintf("✅ Multi-WAN Activo: %d interfaces WAN operativas con balanceo de carga/failover.", report.ActiveWANCount)
	} else if report.TotalWANCount > 1 && report.ActiveWANCount == 1 {
		report.Summary = fmt.Sprintf("⚠️ Multi-WAN en modo Failover: 1 interfaz activa de %d configuradas.", report.TotalWANCount)
	} else if report.Enabled {
		report.Summary = "ℹ️ Servicio mwan3 activo con 1 interfaz WAN."
	} else {
		report.Summary = "ℹ️ mwan3 está instalado pero inactivo."
	}

	return report, nil
}
