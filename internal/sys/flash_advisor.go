package sys

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type FlashPartitionInfo struct {
	MountPoint string  `json:"mount_point"`
	TotalMB    float64 `json:"total_mb"`
	FreeMB     float64 `json:"free_mb"`
	UsedMB     float64 `json:"used_mb"`
	UsedPct    float64 `json:"used_pct"`
}

type FlashAdvisorReport struct {
	Timestamp          time.Time          `json:"timestamp"`
	OverlayPartition   FlashPartitionInfo `json:"overlay_partition"`
	RootPartition      FlashPartitionInfo `json:"root_partition"`
	UpgradablePackages []string           `json:"upgradable_packages"`
	InstallationSafe   bool               `json:"installation_safe"`
	RiskLevel          string             `json:"risk_level"` // SAFE, WARNING, CRITICAL
	Recommendation     string             `json:"recommendation"`
}

func AuditFlashAndPackages() (*FlashAdvisorReport, error) {
	report := &FlashAdvisorReport{
		Timestamp:          time.Now(),
		UpgradablePackages: make([]string, 0),
		InstallationSafe:   true,
		RiskLevel:          "SAFE",
	}

	// 1. Inspect /overlay partition space
	report.OverlayPartition = getPartitionSpace("/overlay")
	if report.OverlayPartition.TotalMB == 0 {
		// Fallback to / if /overlay is not a separate mount
		report.OverlayPartition = getPartitionSpace("/")
	}
	report.RootPartition = getPartitionSpace("/")

	// 2. Query upgradable packages
	if _, err := exec.LookPath("apk"); err == nil {
		out, _ := exec.Command("apk", "list", "--upgradable").Output()
		for _, l := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(l) != "" {
				report.UpgradablePackages = append(report.UpgradablePackages, strings.TrimSpace(l))
			}
		}
	} else if _, err := exec.LookPath("opkg"); err == nil {
		out, _ := exec.Command("opkg", "list-upgradable").Output()
		for _, l := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(l) != "" {
				report.UpgradablePackages = append(report.UpgradablePackages, strings.TrimSpace(l))
			}
		}
	}

	// 3. Evaluate Flash filling risk
	freeOverlay := report.OverlayPartition.FreeMB
	if freeOverlay < 1.0 {
		report.InstallationSafe = false
		report.RiskLevel = "CRITICAL (Flash Llena)"
		report.Recommendation = fmt.Sprintf("🚨 ESPACIO CRÍTICO: Quedan solo %.2f MB libres en /overlay. NO instales paquetes para evitar bloquear el router.", freeOverlay)
	} else if freeOverlay < 3.0 {
		report.InstallationSafe = false
		report.RiskLevel = "WARNING (Espacio Reducido)"
		report.Recommendation = fmt.Sprintf("⚠️ Espacio en Flash limitado (%.2f MB libres). Se desaconseja instalar paquetes pesados sin verificar su tamaño.", freeOverlay)
	} else {
		report.InstallationSafe = true
		report.RiskLevel = "SAFE (Espacio Adecuado)"
		report.Recommendation = fmt.Sprintf("✅ Espacio en Flash adecuado (%.2f MB libres en /overlay). %d paquete(s) con actualizaciones disponibles.", freeOverlay, len(report.UpgradablePackages))
	}

	return report, nil
}

func getPartitionSpace(path string) FlashPartitionInfo {
	var stat syscall.Statfs_t
	info := FlashPartitionInfo{MountPoint: path}

	if err := syscall.Statfs(path, &stat); err == nil {
		total := float64(stat.Blocks) * float64(stat.Bsize) / (1024 * 1024)
		free := float64(stat.Bavail) * float64(stat.Bsize) / (1024 * 1024)
		used := total - free

		info.TotalMB = total
		info.FreeMB = free
		info.UsedMB = used
		if total > 0 {
			info.UsedPct = (used / total) * 100
		}
	}
	return info
}
