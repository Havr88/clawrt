package netintel

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type DNSPrivacyReport struct {
	Timestamp           time.Time `json:"timestamp"`
	DoHActive           bool      `json:"doh_active"`
	DoHProvider         string    `json:"doh_provider,omitempty"`
	DoTActive           bool      `json:"dot_active"`
	AdblockActive       bool      `json:"adblock_active"`
	AdblockEngine       string    `json:"adblock_engine,omitempty"`
	BlockedDomainsCount int       `json:"blocked_domains_count,omitempty"`
	LocalResolver       string    `json:"local_resolver"`
	DNSLeakRisk         string    `json:"dns_leak_risk"` // LOW, MEDIUM, HIGH
	Recommendation      string    `json:"recommendation"`
}

func InspectDNSPrivacy() (*DNSPrivacyReport, error) {
	report := &DNSPrivacyReport{
		Timestamp:     time.Now(),
		LocalResolver: "dnsmasq",
		DNSLeakRisk:   "HIGH",
	}

	// 1. Check DoH (https-dns-proxy / AdGuardHome)
	if _, err := os.Stat("/etc/config/https_dns_proxy"); err == nil {
		if data, err := os.ReadFile("/etc/config/https_dns_proxy"); err == nil {
			content := string(data)
			if strings.Contains(content, "cloudflare") || strings.Contains(content, "google") || strings.Contains(content, "quad9") {
				report.DoHActive = true
				if strings.Contains(content, "cloudflare") {
					report.DoHProvider = "Cloudflare (1.1.1.1 DoH)"
				} else if strings.Contains(content, "quad9") {
					report.DoHProvider = "Quad9 (9.9.9.9 DoH Malware Blocking)"
				} else {
					report.DoHProvider = "DNS over HTTPS Activo"
				}
			}
		}
	}

	// Check stubby (DoT)
	if _, err := os.Stat("/etc/config/stubby"); err == nil {
		report.DoTActive = true
	}

	// 2. Check Adblock engines (adblock-fast, simple-adblock, adblock)
	if _, err := os.Stat("/etc/config/adblock-fast"); err == nil {
		report.AdblockActive = true
		report.AdblockEngine = "adblock-fast"
	} else if _, err := os.Stat("/etc/config/simple-adblock"); err == nil {
		report.AdblockActive = true
		report.AdblockEngine = "simple-adblock"
	} else if _, err := os.Stat("/etc/config/adblock"); err == nil {
		report.AdblockActive = true
		report.AdblockEngine = "luci-app-adblock"
	}

	// Check blocked domains count in dnsmasq adblock lists
	if adList, err := os.ReadFile("/tmp/adb_list.overall"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(adList)))
		count := 0
		for scanner.Scan() {
			count++
		}
		report.BlockedDomainsCount = count
	}

	// Check if AdGuard Home process is running
	if out, err := exec.Command("pgrep", "AdGuardHome").Output(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
		report.AdblockActive = true
		report.DoHActive = true
		report.AdblockEngine = "AdGuard Home"
	}

	// Evaluate Leak Risk & Recommendations
	if report.DoHActive || report.DoTActive {
		report.DNSLeakRisk = "LOW (Cifrado Activo)"
		if report.AdblockActive {
			report.Recommendation = fmt.Sprintf("✅ Privacidad DNS Blindada: Consultas cifradas vía %s con filtrado activo (%s).", report.DoHProvider, report.AdblockEngine)
		} else {
			report.Recommendation = fmt.Sprintf("🛡️ Consultas DNS cifradas vía %s. Se sugiere activar filtrado de anuncios (adblock-fast).", report.DoHProvider)
		}
	} else {
		report.DNSLeakRisk = "HIGH (Consultas en texto plano al ISP)"
		report.Recommendation = "⚠️ Tus consultas DNS se envían en texto plano sin cifrar al proveedor de Internet (ISP). Se recomienda instalar 'https-dns-proxy' o 'stubby'."
	}

	return report, nil
}
