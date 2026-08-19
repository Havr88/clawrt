package knowledge

import (
	"fmt"
	"strconv"
	"strings"
)

type DeviceHardwareProfile struct {
	Brand         string       `json:"brand"`
	Model         string       `json:"model"`
	Version       string       `json:"version"`
	CPU           string       `json:"cpu"`
	CPUMHz        string       `json:"cpu_mhz"`
	FlashMB       int          `json:"flash_mb"`
	RAMMB         int          `json:"ram_mb"`
	Architecture  string       `json:"architecture"`
	Target        string       `json:"target"`
	Subtarget     string       `json:"subtarget"`
	OpenWrtVer    string       `json:"openwrt_ver"`
	Tier          HardwareTier `json:"tier"`
	ClawRTPackage string       `json:"clawrt_package"`
}

func AnalyzeHardwareProfile(brand, model string, flashMB, ramMB int, arch, target string) *DeviceHardwareProfile {
	tier := TierMinimal
	if ramMB <= 32 || flashMB <= 4 {
		tier = TierExtremeMinimal
	} else if ramMB >= 256 || flashMB >= 32 {
		tier = TierFull
	} else if ramMB >= 128 || flashMB >= 16 {
		tier = TierMedium
	}

	normArch := normalizeArch(arch, target)
	clawPackage := fmt.Sprintf("clawrt_1.0.0-1_%s.ipk", normArch)

	return &DeviceHardwareProfile{
		Brand:         brand,
		Model:         model,
		FlashMB:       flashMB,
		RAMMB:         ramMB,
		Architecture:  normArch,
		Target:        target,
		Tier:          tier,
		ClawRTPackage: clawPackage,
	}
}

func normalizeArch(arch, target string) string {
	arch = strings.ToLower(strings.TrimSpace(arch))
	target = strings.ToLower(strings.TrimSpace(target))

	if strings.Contains(arch, "aarch64") || strings.Contains(arch, "cortex-a53") {
		return "aarch64_cortex-a53"
	}
	if strings.Contains(arch, "arm") || strings.Contains(arch, "cortex-a7") {
		return "arm_cortex-a7_neon-vfpv4"
	}
	if strings.Contains(arch, "x86_64") || strings.Contains(target, "x86") {
		return "x86_64"
	}
	if strings.Contains(arch, "mipsel") || strings.Contains(target, "ramips") {
		return "mipsel_24kc"
	}
	if arch != "" {
		return arch
	}
	return "mipsel_24kc"
}

func ParseRAMMB(raw string) int {
	raw = strings.TrimSpace(raw)
	if val, err := strconv.Atoi(raw); err == nil {
		return val
	}
	return 64 // Fallback 64MB
}

func ParseFlashMB(raw string) int {
	raw = strings.TrimSpace(raw)
	// Truncate non-digit characters like "128NAND" -> 128
	var digits strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		} else {
			break
		}
	}
	if val, err := strconv.Atoi(digits.String()); err == nil {
		return val
	}
	return 8 // Fallback 8MB
}
