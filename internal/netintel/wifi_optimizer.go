package netintel

import (
	"clawrt/internal/sys"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type NeighborAP struct {
	SSID    string `json:"ssid"`
	BSSID   string `json:"bssid"`
	Channel int    `json:"channel"`
	Signal  int    `json:"signal_dbm"` // dBm
	Quality int    `json:"quality"`
}

type ChannelScore struct {
	Channel         int     `json:"channel"`
	Band            string  `json:"band"` // 2.4GHz or 5GHz
	OverlappingAPs  int     `json:"overlapping_aps"`
	CongestionScore float64 `json:"congestion_score"` // lower is better
	IsRecommended   bool    `json:"is_recommended"`
}

type WiFiOptimizationReport struct {
	Interface         string         `json:"interface"`
	CurrentChannel    int            `json:"current_channel"`
	Band              string         `json:"band"`
	DetectedAPs       []NeighborAP   `json:"detected_aps"`
	EvaluatedChannels []ChannelScore `json:"evaluated_channels"`
	OptimalChannel    int            `json:"optimal_channel"`
	Recommendation    string         `json:"recommendation"`
}

func OptimizeWiFiChannels(phyInterface string, applyChanges bool) (*WiFiOptimizationReport, error) {
	if phyInterface == "" {
		phyInterface = "wlan0"
	}

	report := &WiFiOptimizationReport{
		Interface:         phyInterface,
		DetectedAPs:       make([]NeighborAP, 0),
		EvaluatedChannels: make([]ChannelScore, 0),
	}

	// 1. Get current channel
	report.CurrentChannel = getCurrentChannel(phyInterface)

	// 2. Perform iwinfo scan
	outScan, err := exec.Command("iwinfo", phyInterface, "scan").Output()
	if err != nil || len(outScan) == 0 {
		// Fallback to iw scan
		outScan, _ = exec.Command("iw", "dev", phyInterface, "scan").Output()
	}

	report.DetectedAPs = parseNeighborAPs(string(outScan))

	// 3. Evaluate Channels
	is5GHz := report.CurrentChannel > 14
	var targetChannels []int

	if is5GHz {
		report.Band = "5 GHz"
		targetChannels = []int{36, 40, 44, 48, 149, 153, 157, 161}
	} else {
		report.Band = "2.4 GHz"
		targetChannels = []int{1, 6, 11}
	}

	for _, ch := range targetChannels {
		score := ChannelScore{
			Channel: ch,
			Band:    report.Band,
		}

		for _, ap := range report.DetectedAPs {
			diff := abs(ap.Channel - ch)
			if diff <= 2 {
				score.OverlappingAPs++
				signalWeight := 100.0 + float64(ap.Signal)
				if signalWeight < 5.0 {
					signalWeight = 5.0
				}
				if diff == 0 {
					score.CongestionScore += signalWeight * 1.5
				} else {
					score.CongestionScore += signalWeight * 0.8
				}
			}
		}
		report.EvaluatedChannels = append(report.EvaluatedChannels, score)
	}

	// Sort channels by lowest congestion score
	sort.Slice(report.EvaluatedChannels, func(i, j int) bool {
		return report.EvaluatedChannels[i].CongestionScore < report.EvaluatedChannels[j].CongestionScore
	})

	if len(report.EvaluatedChannels) > 0 {
		report.EvaluatedChannels[0].IsRecommended = true
		report.OptimalChannel = report.EvaluatedChannels[0].Channel
	}

	if report.OptimalChannel == report.CurrentChannel {
		report.Recommendation = fmt.Sprintf("✅ El canal actual (%d en %s) ya es el más limpio y óptimo.", report.CurrentChannel, report.Band)
	} else {
		report.Recommendation = fmt.Sprintf("📶 Se recomienda cambiar de canal %d a canal %d en %s (reduce interferencias de %d redes vecinas).",
			report.CurrentChannel, report.OptimalChannel, report.Band, len(report.DetectedAPs))
	}

	// 4. Apply changes if requested
	if applyChanges && report.OptimalChannel != 0 && report.OptimalChannel != report.CurrentChannel {
		radioName := "radio0"
		if is5GHz {
			radioName = "radio1"
		}
		setPath := fmt.Sprintf("wireless.%s.channel", radioName)
		_, errSet := sys.ExecuteTypedUCISet("wireless", setPath, fmt.Sprintf("%d", report.OptimalChannel))
		if errSet == nil {
			_, _ = sys.ExecuteTypedServiceRestart("network")
			report.Recommendation += " (Aplicado automáticamente y red reiniciada)."
		}
	}

	return report, nil
}

func parseNeighborAPs(output string) []NeighborAP {
	var aps []NeighborAP
	blocks := strings.Split(output, "Cell ")
	reSSID := regexp.MustCompile(`ESSID:\s*"(.*?)"`)
	reChan := regexp.MustCompile(`Channel:\s*([0-9]+)`)
	reSignal := regexp.MustCompile(`Signal:\s*(-?[0-9]+)\s*dBm`)
	reBSSID := regexp.MustCompile(`Address:\s*([0-9a-fA-F:]{17})`)

	for _, block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}
		var ap NeighborAP

		if m := reSSID.FindStringSubmatch(block); len(m) > 1 {
			ap.SSID = m[1]
		}
		if m := reBSSID.FindStringSubmatch(block); len(m) > 1 {
			ap.BSSID = m[1]
		}
		if m := reChan.FindStringSubmatch(block); len(m) > 1 {
			if c, err := strconv.Atoi(m[1]); err == nil {
				ap.Channel = c
			}
		}
		if m := reSignal.FindStringSubmatch(block); len(m) > 1 {
			if s, err := strconv.Atoi(m[1]); err == nil {
				ap.Signal = s
			}
		}

		if ap.Channel > 0 {
			aps = append(aps, ap)
		}
	}
	return aps
}

func getCurrentChannel(iface string) int {
	out, err := exec.Command("iwinfo", iface, "info").Output()
	if err == nil {
		re := regexp.MustCompile(`Channel:\s*([0-9]+)`)
		if m := re.FindStringSubmatch(string(out)); len(m) > 1 {
			if c, err := strconv.Atoi(m[1]); err == nil {
				return c
			}
		}
	}
	return 1
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
