package netintel

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type StationHealth struct {
	MAC         string  `json:"mac"`
	IP          string  `json:"ip,omitempty"`
	Hostname    string  `json:"hostname,omitempty"`
	Interface   string  `json:"interface"`
	SignalRSSI  int     `json:"signal_rssi"` // dBm (e.g. -78)
	TxRateMbps  float64 `json:"tx_rate_mbps,omitempty"`
	RxRateMbps  float64 `json:"rx_rate_mbps,omitempty"`
	IsSticky    bool    `json:"is_sticky"` // true if signal <= -80 dBm
	QualityDesc string  `json:"quality_desc"`
}

type StickyClientsReport struct {
	Timestamp      time.Time       `json:"timestamp"`
	TotalStations  int             `json:"total_stations"`
	StickyCount    int             `json:"sticky_count"`
	Stations       []StationHealth `json:"stations"`
	Recommendation string          `json:"recommendation"`
}

func DetectStickyClients(iface string, kickWeakest bool) (*StickyClientsReport, error) {
	if iface == "" {
		iface = "wlan0"
	}

	report := &StickyClientsReport{
		Timestamp: time.Now(),
		Stations:  make([]StationHealth, 0),
	}

	// 1. Try ubus call iwinfo assoclist
	out, err := exec.Command("ubus", "call", "iwinfo", "assoclist", fmt.Sprintf(`{"device":"%s"}`, iface)).Output()
	if err == nil && len(out) > 0 {
		var ubusRes struct {
			Results []struct {
				MAC    string `json:"mac"`
				Signal int    `json:"signal"`
				TxRate int    `json:"tx_rate"`
				RxRate int    `json:"rx_rate"`
			} `json:"results"`
		}
		if err := json.Unmarshal(out, &ubusRes); err == nil {
			for _, r := range ubusRes.Results {
				st := StationHealth{
					MAC:        r.MAC,
					Interface:  iface,
					SignalRSSI: r.Signal,
					TxRateMbps: float64(r.TxRate) / 1000.0,
					RxRateMbps: float64(r.RxRate) / 1000.0,
				}
				enrichStationInfo(&st)
				report.Stations = append(report.Stations, st)
			}
		}
	}

	// Fallback to iwinfo CLI if ubus was empty
	if len(report.Stations) == 0 {
		outCli, _ := exec.Command("iwinfo", iface, "assoclist").Output()
		parseIWInfoAssoclist(string(outCli), iface, report)
	}

	report.TotalStations = len(report.Stations)

	var stickyMACs []string
	for i := range report.Stations {
		st := &report.Stations[i]
		if st.SignalRSSI <= -80 && st.SignalRSSI != 0 {
			st.IsSticky = true
			st.QualityDesc = "🚨 Muy Débil (Cliente Pegajoso / Degrada celda)"
			report.StickyCount++
			stickyMACs = append(stickyMACs, st.MAC)
		} else if st.SignalRSSI <= -70 {
			st.QualityDesc = "⚠️ Regular"
		} else {
			st.QualityDesc = "✅ Óptima"
		}
	}

	if report.StickyCount > 0 {
		report.Recommendation = fmt.Sprintf("⚠️ Se detectaron %d dispositivos con señal muy débil (< -80 dBm). Forzar reconexión mejorará el rendimiento del resto de la celda WiFi.", report.StickyCount)
		if kickWeakest {
			for _, mac := range stickyMACs {
				_ = KickStation(iface, mac)
			}
			report.Recommendation += " (Se enviaron señales de desasociación suave para forzar Roaming)."
		}
	} else {
		report.Recommendation = "✅ Todos los clientes conectados mantienen una intensidad de señal adecuada."
	}

	return report, nil
}

func KickStation(iface, mac string) error {
	// Soft disassociate via ubus hostapd
	cmd := exec.Command("ubus", "call", "hostapd."+iface, "del_client", fmt.Sprintf(`{"addr":"%s", "reason":5, "deauth":true}`, mac))
	return cmd.Run()
}

func parseIWInfoAssoclist(output, iface string, report *StickyClientsReport) {
	lines := strings.Split(output, "\n")
	for _, l := range lines {
		f := strings.Fields(l)
		if len(f) >= 3 && strings.Contains(f[0], ":") {
			mac := f[0]
			sig, _ := strconv.Atoi(f[1])
			st := StationHealth{
				MAC:        mac,
				Interface:  iface,
				SignalRSSI: sig,
			}
			enrichStationInfo(&st)
			report.Stations = append(report.Stations, st)
		}
	}
}

func enrichStationInfo(st *StationHealth) {
	st.IP = resolveIPFromDHCPLeases(st.MAC)
	st.Hostname = resolveHostnameFromDHCP(st.IP)
}

func resolveIPFromDHCPLeases(mac string) string {
	leases, err := GetEnrichedLeases()
	if err != nil {
		return ""
	}
	macLower := strings.ToLower(mac)
	for _, l := range leases {
		if strings.ToLower(l.MAC) == macLower {
			return l.IP
		}
	}
	return ""
}
