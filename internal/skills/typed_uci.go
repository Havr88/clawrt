package skills

import (
	"bytes"
	"clawrt/internal/security"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ExecuteTypedUCISet(pkg, path, val string) (string, error) {
	pkg = strings.TrimSpace(strings.ToLower(pkg))
	path = strings.TrimSpace(path)
	val = strings.TrimSpace(val)

	if pkg == "" {
		if parts := strings.Split(path, "."); len(parts) > 0 {
			pkg = parts[0]
		}
	}

	if pkg == "" || path == "" {
		return "", fmt.Errorf("se requieren los parámetros 'package' o 'path' y 'value'")
	}

	configFile := filepath.Join("/etc/config", pkg)
	if err := security.ValidatePathSafety(configFile); err != nil {
		return "", err
	}

	// 1. Create Pre-execution Snapshot for Rollback
	snapshotFile := fmt.Sprintf("/tmp/clawrt_snapshot_%s.bak", pkg)
	if data, err := os.ReadFile(configFile); err == nil {
		_ = os.WriteFile(snapshotFile, data, 0644)
	}

	// 2. Perform uci set
	uciSetCmd := fmt.Sprintf("uci set %s=%s", path, val)
	if err := security.ValidateCommandSafety(uciSetCmd); err != nil {
		return "", err
	}

	cmdSet := exec.Command("uci", "set", fmt.Sprintf("%s=%s", path, val))
	if out, err := cmdSet.CombinedOutput(); err != nil {
		_ = exec.Command("uci", "revert", pkg).Run()
		return "", fmt.Errorf("fallo uci set: %v (%s)", err, string(out))
	}

	// 3. Test firewall syntax if firewall package is modified
	if pkg == "firewall" {
		cmdCheck := exec.Command("fw4", "check")
		if out, err := cmdCheck.CombinedOutput(); err != nil {
			_ = exec.Command("uci", "revert", pkg).Run()
			if snapData, snapErr := os.ReadFile(snapshotFile); snapErr == nil {
				_ = os.WriteFile(configFile, snapData, 0644)
			}
			return "", fmt.Errorf("🚫 ROLLBACK AUTOMÁTICO: Error de sintaxis en firewall4 (fw4 check: %s). Se revirtió la configuración", string(out))
		}
	}

	// 4. Commit changes
	cmdCommit := exec.Command("uci", "commit", pkg)
	if out, err := cmdCommit.CombinedOutput(); err != nil {
		_ = exec.Command("uci", "revert", pkg).Run()
		if snapData, snapErr := os.ReadFile(snapshotFile); snapErr == nil {
			_ = os.WriteFile(configFile, snapData, 0644)
		}
		return "", fmt.Errorf("fallo uci commit: %v (%s). Se realizó rollback automático", err, string(out))
	}

	_ = os.Remove(snapshotFile)
	return fmt.Sprintf("✅ Configuración UCI actualizada con éxito (Snapshot & Rollback verificado): %s = %s", path, val), nil
}

func ExecuteTypedServiceRestart(service string) (string, error) {
	service = strings.TrimSpace(strings.ToLower(service))
	allowedServices := map[string]bool{
		"network":  true,
		"dnsmasq":  true,
		"firewall": true,
		"clawrt":   true,
		"dropbear": true,
		"uhttpd":   true,
	}

	if !allowedServices[service] {
		return "", fmt.Errorf("🚫 SERVICIO NO PERMITIDO: El servicio '%s' no está en la lista segura (network, dnsmasq, firewall, clawrt, dropbear, uhttpd)", service)
	}

	cmdStr := fmt.Sprintf("/etc/init.d/%s restart", service)
	cmd := exec.Command("/etc/init.d/"+service, "restart")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error al reiniciar servicio %s: %v (%s)", service, err, out.String())
	}

	return fmt.Sprintf("🔄 Servicio '%s' reiniciado correctamente (%s)", service, cmdStr), nil
}
