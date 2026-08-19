package sys

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	Hostname      string  `json:"hostname"`
	OpenWrtVer    string  `json:"openwrt_version"`
	Architecture  string  `json:"architecture"`
	Uptime        string  `json:"uptime"`
	LoadAverage   string  `json:"load_average"`
	MemoryUsedMB  int     `json:"memory_used_mb"`
	MemoryTotalMB int     `json:"memory_total_mb"`
	MemoryUsedPct float64 `json:"memory_used_pct"`
}

func GetSystemInfo() *SystemInfo {
	info := &SystemInfo{
		Architecture: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	// Hostname & Uptime using gopsutil
	if hInfo, err := host.Info(); err == nil {
		info.Hostname = hInfo.Hostname
		d := time.Duration(hInfo.Uptime) * time.Second
		info.Uptime = fmt.Sprintf("%d d, %d h, %d m", int(d.Hours())/24, int(d.Hours())%24, int(d.Minutes())%60)
	} else {
		if hostName, err := os.Hostname(); err == nil {
			info.Hostname = hostName
		} else {
			info.Hostname = "openwrt"
		}
	}

	// OpenWrt Version from /etc/openwrt_release
	if data, err := os.ReadFile("/etc/openwrt_release"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "DISTRIB_DESCRIPTION=") {
				info.OpenWrtVer = strings.Trim(strings.TrimPrefix(line, "DISTRIB_DESCRIPTION="), "'\"")
				break
			}
		}
	}
	if info.OpenWrtVer == "" {
		info.OpenWrtVer = "OpenWrt (desconocido)"
	}

	// Load Avg
	if loadData, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(loadData))
		if len(fields) >= 3 {
			info.LoadAverage = strings.Join(fields[:3], ", ")
		}
	}

	// Memory info using gopsutil/v3/mem
	if vMem, err := mem.VirtualMemory(); err == nil {
		info.MemoryTotalMB = int(vMem.Total / (1024 * 1024))
		info.MemoryUsedMB = int(vMem.Used / (1024 * 1024))
		info.MemoryUsedPct = vMem.UsedPercent
	} else {
		// Fallback to /proc/meminfo
		if memData, err := os.ReadFile("/proc/meminfo"); err == nil {
			var total, free, buffers, cached int
			for _, line := range strings.Split(string(memData), "\n") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					val, _ := strconv.Atoi(fields[1])
					switch fields[0] {
					case "MemTotal:":
						total = val
					case "MemFree:":
						free = val
					case "Buffers:":
						buffers = val
					case "Cached:":
						cached = val
					}
				}
			}
			if total > 0 {
				used := total - (free + buffers + cached)
				info.MemoryTotalMB = total / 1024
				info.MemoryUsedMB = used / 1024
				info.MemoryUsedPct = (float64(used) / float64(total)) * 100
			}
		}
	}

	return info
}

func GetNetworkSummary() string {
	var sb strings.Builder

	// Read WAN/LAN interface IP
	out, err := exec.Command("ip", "-4", "addr", "show").Output()
	if err == nil {
		sb.WriteString("📡 Interfaces de red:\n")
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "inet ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					iface := "desconocida"
					if len(parts) >= 4 {
						iface = parts[len(parts)-1]
					}
					sb.WriteString(fmt.Sprintf(" • %s: %s\n", iface, parts[1]))
				}
			}
		}
	} else {
		sb.WriteString("📡 No se pudo ejecutar ip addr\n")
	}

	// WiFi clients via iwinfo
	wifiOut, err := exec.Command("iwinfo").Output()
	if err == nil && len(wifiOut) > 0 {
		sb.WriteString("\n📶 Estado WiFi:\n")
		lines := strings.Split(string(wifiOut), "\n")
		for _, l := range lines {
			if strings.Contains(l, "ESSID") || strings.Contains(l, "Access Point") {
				sb.WriteString(" " + strings.TrimSpace(l) + "\n")
			}
		}
	}

	return sb.String()
}
