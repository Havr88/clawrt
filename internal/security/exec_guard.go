package security

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var AllowedCommands = map[string]bool{
	"uci": true, "ip": true, "nft": true, "service": true, "wifi": true,
	"logread": true, "iwinfo": true, "ubus": true, "cat": true, "ls": true,
	"ps": true, "top": true, "free": true, "df": true, "du": true,
	"mount": true, "uptime": true, "date": true, "nslookup": true, "ping": true,
	"traceroute": true, "sysctl": true, "swconfig": true, "curl": true, "wget": true,
	"apk": true, "opkg": true, "hostname": true, "uname": true, "logger": true,
	"md5sum": true, "sha256sum": true, "cmp": true, "stat": true,
}

var HardDenylistPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(reboot|poweroff|shutdown|halt)\b`),
	regexp.MustCompile(`(?i)rm\s+-[rRf]*\s+/`),
	regexp.MustCompile(`(?i)dd\s+if=.*of=/dev/mtd`),
	regexp.MustCompile(`(?i)\b(format|mkfs|fdisk|mtd\s+erase)\b`),
	regexp.MustCompile(`(?i)\bsysupgrade\b`),
	regexp.MustCompile(`(?i)kill\s+-9`),
	regexp.MustCompile(`(?i)iptables\s+-F`),
	regexp.MustCompile(`(?i)echo.*>/dev/watchdog`),
}

var DangerousShellOperators = regexp.MustCompile(`[;&|` + "`" + `$><]`)

var AllowedFilePaths = []string{
	"/etc/config/",
	"/etc/rc.local",
	"/etc/crontabs/root",
	"/etc/firewall.user",
}

type CommandGuardResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

func ValidateCommandSafety(rawCmd string) error {
	trimmed := strings.TrimSpace(rawCmd)
	if trimmed == "" {
		return fmt.Errorf("comando vacío")
	}

	// 1. Check Hard Denylist
	for _, pattern := range HardDenylistPatterns {
		if pattern.MatchString(trimmed) {
			return fmt.Errorf("🚫 COMANDO DENEGADO (Hard Denylist): El comando '%s' está clasificado como peligroso y no se permite su ejecución", trimmed)
		}
	}

	// 2. Check Dangerous Shell Operators (Injection Defense)
	if DangerousShellOperators.MatchString(trimmed) {
		return fmt.Errorf("⚠️ RECHAZADO (Defensa Inyección Shell): Se detectaron operadores de shell no permitidos (;, &&, ||, `, $(...), redirects o pipes)")
	}

	// 3. Extract primary binary name
	fields := strings.Fields(trimmed)
	binary := strings.ToLower(filepath.Base(fields[0]))

	if !AllowedCommands[binary] {
		return fmt.Errorf("🚫 COMANDO NO PERMITIDO: El binario '%s' no está en la Allowlist de comandos seguros", binary)
	}

	// 4. Validate package manager restriction (read-only for apk/opkg)
	if (binary == "apk" || binary == "opkg") && len(fields) > 1 {
		subcmd := strings.ToLower(fields[1])
		if subcmd != "list" && subcmd != "search" && subcmd != "info" && subcmd != "version" {
			return fmt.Errorf("🚫 OPERACIÓN DENEGADA: Gestor de paquetes (%s) sólo permite comandos de consulta (list, search, info, version)", binary)
		}
	}

	return nil
}

func ValidatePathSafety(filePath string) error {
	cleanPath := filepath.Clean(filePath)
	for _, allowed := range AllowedFilePaths {
		if strings.HasPrefix(cleanPath, allowed) {
			return nil
		}
	}
	return fmt.Errorf("🚫 RUTA NO PERMITIDA: La edición directa está restringida únicamente a /etc/config/*, /etc/rc.local, /etc/crontabs/root y /etc/firewall.user")
}

func ExecSafeCommandWithTimeout(cmdStr string, timeoutSec time.Duration) (string, error) {
	if err := ValidateCommandSafety(cmdStr); err != nil {
		return "", err
	}

	if timeoutSec <= 0 {
		timeoutSec = 15 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSec)
	defer cancel()

	fields := strings.Fields(cmdStr)
	cmd := exec.CommandContext(ctx, fields[0], fields[1:]...)

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()

	outputStr := outBuf.String()

	// Truncate output to <= 4 KB (4096 bytes) for LLM context optimization
	if len(outputStr) > 4096 {
		outputStr = outputStr[:4096] + "\n... [Salida truncada a 4 KB por seguridad y optimización de contexto]"
	}

	if ctx.Err() == context.DeadlineExceeded {
		return outputStr, fmt.Errorf("⏳ TIMEOUT: El comando '%s' excedió el tiempo límite de %v y fue terminado", fields[0], timeoutSec)
	}

	if err != nil {
		return outputStr, fmt.Errorf("error al ejecutar %s: %w", fields[0], err)
	}

	return outputStr, nil
}
