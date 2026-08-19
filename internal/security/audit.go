package security

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"time"
)

type SecurityFinding struct {
	Level       string `json:"level"` // CRITICAL, WARNING, INFO, OK
	Title       string `json:"title"`
	Description string `json:"description"`
	Remediation string `json:"remediation,omitempty"`
}

type RouterSecurityAuditReport struct {
	Timestamp     time.Time         `json:"timestamp"`
	OverallGrade  string            `json:"overall_grade"` // A, B, C, D, F
	TotalScore    int               `json:"total_score"`   // 0 - 100
	Findings      []SecurityFinding `json:"findings"`
	CriticalCount int               `json:"critical_count"`
	WarningCount  int               `json:"warning_count"`
	Summary       string            `json:"summary"`
}

func AuditRouterSecurity() (*RouterSecurityAuditReport, error) {
	report := &RouterSecurityAuditReport{
		Timestamp: time.Now(),
		Findings:  make([]SecurityFinding, 0),
	}

	score := 100

	// 1. Audit Root Password in /etc/shadow
	rootPassAudit(&score, report)

	// 2. Audit SSH Dropbear WAN exposure
	dropbearAudit(&score, report)

	// 3. Audit LuCI uhttpd WAN exposure & HTTP/HTTPS
	uhttpdAudit(&score, report)

	// 4. Audit Insecure Legacy Services (Telnet, UPnP)
	legacyServicesAudit(&score, report)

	// 5. Audit Firewall Default Policies
	firewallPolicyAudit(&score, report)

	// Calculate Final Grade
	if score < 0 {
		score = 0
	}
	report.TotalScore = score

	if report.CriticalCount > 0 || score < 50 {
		report.OverallGrade = "F (Vulnerable / Riesgo Crítico)"
		report.Summary = "🚨 Se detectaron vulnerabilidades críticas de seguridad que comprometen la red del router."
	} else if report.WarningCount >= 2 || score < 75 {
		report.OverallGrade = "C (Mejorable)"
		report.Summary = "⚠️ El router tiene configuraciones potencialmente inseguras que deben ajustarse."
	} else if score < 90 {
		report.OverallGrade = "B (Seguro)"
		report.Summary = "🛡️ La postura de seguridad es buena, con recomendaciones menores."
	} else {
		report.OverallGrade = "A+ (Blindado / Óptimo)"
		report.Summary = "✅ Configuración de seguridad excelente y sin exposiciones indebidas."
	}

	return report, nil
}

func rootPassAudit(score *int, report *RouterSecurityAuditReport) {
	file, err := os.Open("/etc/shadow")
	if err != nil {
		report.Findings = append(report.Findings, SecurityFinding{
			Level:       "INFO",
			Title:       "Lectura de /etc/shadow restringida",
			Description: "No se pudo acceder directamente a /etc/shadow para verificar el hash de root.",
		})
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "root:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				hash := parts[1]
				if hash == "" || hash == "*" || hash == "!" {
					*score -= 35
					report.CriticalCount++
					report.Findings = append(report.Findings, SecurityFinding{
						Level:       "CRITICAL",
						Title:       "Contraseña de root vacía o sin configurar",
						Description: "El usuario root no tiene contraseña asignada en /etc/shadow.",
						Remediation: "Ejecuta 'passwd root' inmediatamente para establecer una contraseña segura.",
					})
				} else {
					report.Findings = append(report.Findings, SecurityFinding{
						Level:       "OK",
						Title:       "Contraseña de root configurada",
						Description: "El usuario root cuenta con hash de autenticación criptográfico shadow.",
					})
				}
			}
			break
		}
	}
}

func dropbearAudit(score *int, report *RouterSecurityAuditReport) {
	data, err := os.ReadFile("/etc/config/dropbear")
	if err != nil {
		return
	}
	content := string(data)

	// Check if PasswordAuth is enabled or RootPasswordAuth
	if strings.Contains(content, "PasswordAuth 'on'") || strings.Contains(content, "PasswordAuth '1'") {
		report.Findings = append(report.Findings, SecurityFinding{
			Level:       "WARNING",
			Title:       "SSH permite autenticación por contraseña",
			Description: "Dropbear permite iniciar sesión por contraseña en lugar de forzar llaves SSH públicas (ED25519/RSA).",
			Remediation: "Se recomienda habilitar autenticación por llaves SSH y deshabilitar PasswordAuth.",
		})
		*score -= 5
		report.WarningCount++
	}

	// Check if port 22 is open on WAN
	fwData, err := os.ReadFile("/etc/config/firewall")
	if err == nil && (strings.Contains(string(fwData), "dest_port '22'") || strings.Contains(string(fwData), "src_dport '22'")) {
		if strings.Contains(string(fwData), "src 'wan'") {
			*score -= 30
			report.CriticalCount++
			report.Findings = append(report.Findings, SecurityFinding{
				Level:       "CRITICAL",
				Title:       "Puerto SSH (22) expuesto hacia Internet (WAN)",
				Description: "Hay una regla de cortafuegos que permite conexiones entrantes al puerto SSH desde la red pública WAN.",
				Remediation: "Elimina la regla de apertura de puerto 22 en la zona WAN (/etc/config/firewall).",
			})
		}
	}
}

func uhttpdAudit(score *int, report *RouterSecurityAuditReport) {
	data, err := os.ReadFile("/etc/config/uhttpd")
	if err != nil {
		return
	}
	content := string(data)

	if strings.Contains(content, "list listen_http '0.0.0.0:80'") && !strings.Contains(content, "list listen_https") {
		*score -= 10
		report.WarningCount++
		report.Findings = append(report.Findings, SecurityFinding{
			Level:       "WARNING",
			Title:       "Panel web LuCI solo en HTTP sin cifrado TLS",
			Description: "La interfaz web uhttpd opera en texto plano (HTTP puerto 80) sin HTTPS activo.",
			Remediation: "Instala 'luci-ssl' o activa certificados TLS en /etc/config/uhttpd.",
		})
	}
}

func legacyServicesAudit(score *int, report *RouterSecurityAuditReport) {
	// Check if telnet is running
	out, err := exec.Command("pgrep", "telnetd").Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		*score -= 25
		report.CriticalCount++
		report.Findings = append(report.Findings, SecurityFinding{
			Level:       "CRITICAL",
			Title:       "Servicio Telnet (puerto 23) en ejecución",
			Description: "Telnet transmite tráfico y credenciales en texto plano sin cifrado.",
			Remediation: "Detén y deshabilita telnetd: '/etc/init.d/telnet stop && /etc/init.d/telnet disable'.",
		})
	} else {
		report.Findings = append(report.Findings, SecurityFinding{
			Level:       "OK",
			Title:       "Telnet inactivo",
			Description: "No hay servidores telnet inseguros ejecutándose.",
		})
	}
}

func firewallPolicyAudit(score *int, report *RouterSecurityAuditReport) {
	data, err := os.ReadFile("/etc/config/firewall")
	if err != nil {
		return
	}
	content := string(data)

	if strings.Contains(content, "name 'wan'") && strings.Contains(content, "input 'ACCEPT'") {
		*score -= 35
		report.CriticalCount++
		report.Findings = append(report.Findings, SecurityFinding{
			Level:       "CRITICAL",
			Title:       "Política de entrada WAN permisiva (input 'ACCEPT')",
			Description: "La zona WAN está configurada para aceptar todo el tráfico entrante por defecto.",
			Remediation: "Cambia la política de entrada de la zona WAN a 'REJECT' o 'DROP'.",
		})
	} else {
		report.Findings = append(report.Findings, SecurityFinding{
			Level:       "OK",
			Title:       "Política WAN segura",
			Description: "La zona WAN bloquea el tráfico entrante no solicitado (REJECT/DROP).",
		})
	}
}
