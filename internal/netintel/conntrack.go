package netintel

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ConntrackEntry struct {
	Protocol string `json:"protocol"`
	SrcIP    string `json:"src_ip"`
	DstIP    string `json:"dst_ip"`
	SrcPort  int    `json:"src_port"`
	DstPort  int    `json:"dst_port"`
	State    string `json:"state"`
	Bytes    int64  `json:"bytes,omitempty"`
	Packets  int64  `json:"packets,omitempty"`
}

type HostTrafficSummary struct {
	IP               string `json:"ip"`
	Hostname         string `json:"hostname"`
	ActiveSockets    int    `json:"active_sockets"`
	TargetPortsCount int    `json:"target_ports_count"`
	TargetPorts      []int  `json:"target_ports"`
	IsHog            bool   `json:"is_hog"`
	PossiblePortScan bool   `json:"possible_port_scan"`
	SuspectedThreat  string `json:"suspected_threat,omitempty"`
}

type ConntrackReport struct {
	Timestamp         time.Time            `json:"timestamp"`
	TotalConnections  int                  `json:"total_connections"`
	MaxConnections    int                  `json:"max_connections"`
	TCPCount          int                  `json:"tcp_count"`
	UDPCount          int                  `json:"udp_count"`
	ICMPCount         int                  `json:"icmp_count"`
	TopTalkers        []HostTrafficSummary `json:"top_talkers"`
	SecurityAnomalies []string             `json:"security_anomalies"`
}

func AnalyzeConntrackTraffic() (*ConntrackReport, error) {
	// Try /proc/net/nf_conntrack first, fallback to /proc/net/ip_conntrack
	filePath := "/proc/net/nf_conntrack"
	file, err := os.Open(filePath)
	if err != nil {
		filePath = "/proc/net/ip_conntrack"
		file, err = os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("no se pudo abrir tabla conntrack (/proc/net/nf_conntrack o ip_conntrack): %w", err)
		}
	}
	defer file.Close()

	report := &ConntrackReport{
		Timestamp:         time.Now(),
		MaxConnections:    getConntrackMax(),
		TopTalkers:        make([]HostTrafficSummary, 0),
		SecurityAnomalies: make([]string, 0),
	}

	hostSockets := make(map[string]int)
	hostDstPorts := make(map[string]map[int]bool)
	synSentCount := make(map[string]int)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		report.TotalConnections++

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		proto := strings.ToLower(fields[0])
		switch proto {
		case "tcp", "ipv4":
			report.TCPCount++
		case "udp":
			report.UDPCount++
		case "icmp":
			report.ICMPCount++
		}

		var srcIP, dstIP string
		var dstPort int
		var isSynSent bool

		for _, f := range fields {
			if strings.HasPrefix(f, "src=") && srcIP == "" {
				srcIP = strings.TrimPrefix(f, "src=")
			} else if strings.HasPrefix(f, "dst=") && dstIP == "" {
				dstIP = strings.TrimPrefix(f, "dst=")
			} else if strings.HasPrefix(f, "dport=") && dstPort == 0 {
				if p, err := strconv.Atoi(strings.TrimPrefix(f, "dport=")); err == nil {
					dstPort = p
				}
			} else if f == "SYN_SENT" {
				isSynSent = true
			}
		}

		if srcIP != "" {
			hostSockets[srcIP]++
			if hostDstPorts[srcIP] == nil {
				hostDstPorts[srcIP] = make(map[int]bool)
			}
			if dstPort > 0 {
				hostDstPorts[srcIP][dstPort] = true
			}
			if isSynSent {
				synSentCount[srcIP]++
			}
		}
	}

	// Sort hosts by active socket count
	type hostPair struct {
		IP    string
		Count int
	}
	var pairs []hostPair
	for ip, count := range hostSockets {
		pairs = append(pairs, hostPair{IP: ip, Count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	// Top 10 Talkers & Anomaly Evaluation
	for i, p := range pairs {
		if i >= 10 {
			break
		}

		portsMap := hostDstPorts[p.IP]
		var portList []int
		for port := range portsMap {
			portList = append(portList, port)
		}
		sort.Ints(portList)

		summary := HostTrafficSummary{
			IP:               p.IP,
			Hostname:         resolveHostnameFromDHCP(p.IP),
			ActiveSockets:    p.Count,
			TargetPortsCount: len(portList),
			TargetPorts:      portList,
		}

		// Check for Connection Hog (> 300 active connections)
		if p.Count > 300 {
			summary.IsHog = true
			summary.SuspectedThreat = "Bandwidth / Connection Hog (P2P / Torrent o Saturación)"
			report.SecurityAnomalies = append(report.SecurityAnomalies,
				fmt.Sprintf("⚠️ IP %s (%s) mantiene %d conexiones activas (posible saturación P2P/Torrent).", p.IP, summary.Hostname, p.Count))
		}

		// Check for Port Scan (> 25 unique target ports)
		if len(portList) > 25 {
			summary.PossiblePortScan = true
			summary.SuspectedThreat = "Posible Escaneo de Puertos en Curso (Reconocimiento de Red)"
			report.SecurityAnomalies = append(report.SecurityAnomalies,
				fmt.Sprintf("🚨 IP %s ha intentado conectar a %d puertos destino distintos (posible Port Scan).", p.IP, len(portList)))
		}

		// Check for SYN Flood / DDoS
		if synSentCount[p.IP] > 50 {
			summary.SuspectedThreat = "Posible Ataque SYN Flood / DDoS saliente"
			report.SecurityAnomalies = append(report.SecurityAnomalies,
				fmt.Sprintf("🚨 IP %s tiene %d sockets en estado SYN_SENT sin respuesta (posible Flood/Malware).", p.IP, synSentCount[p.IP]))
		}

		report.TopTalkers = append(report.TopTalkers, summary)
	}

	return report, nil
}

func BlockAbuserIP(ip string, reason string) (string, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "127.0.0.1" || ip == "192.168.1.1" {
		return "", fmt.Errorf("dirección IP inválida o protegida de gateway: %s", ip)
	}

	// Insert temporary drop rule in nftables or iptables
	cmdNft := exec.Command("nft", "add", "rule", "inet", "fw4", "input", "ip", "saddr", ip, "drop")
	if out, err := cmdNft.CombinedOutput(); err == nil {
		return fmt.Sprintf("🛡️ IP %s bloqueada exitosamente en firewall (nftables fw4). Razón: %s", ip, reason), nil
	} else {
		// Fallback to iptables
		cmdIp := exec.Command("iptables", "-I", "FORWARD", "-s", ip, "-j", "DROP")
		if outIp, errIp := cmdIp.CombinedOutput(); errIp != nil {
			return "", fmt.Errorf("fallo al bloquear IP: nft: %s, iptables: %s", string(out), string(outIp))
		}
		return fmt.Sprintf("🛡️ IP %s bloqueada exitosamente en firewall (iptables FORWARD). Razón: %s", ip, reason), nil
	}
}

func getConntrackMax() int {
	data, err := os.ReadFile("/proc/sys/net/netfilter/nf_conntrack_max")
	if err == nil {
		if val, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			return val
		}
	}
	return 16384
}

func resolveHostnameFromDHCP(ip string) string {
	data, err := os.ReadFile("/tmp/dhcp.leases")
	if err != nil {
		return "desconocido"
	}
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		f := strings.Fields(l)
		if len(f) >= 4 && f[2] == ip {
			if f[3] != "*" {
				return f[3]
			}
		}
	}
	return "desconocido"
}
