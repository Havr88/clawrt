package skills

import (
	"bytes"
	"clawrt/internal/intent"
	"clawrt/internal/netintel"
	"clawrt/internal/security"
	"clawrt/internal/sys"
	"clawrt/internal/watchdog"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ToolDefinition struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type SkillRegistry struct {
	tools map[string]func(args map[string]interface{}) (string, error)
}

func NewRegistry() *SkillRegistry {
	r := &SkillRegistry{
		tools: make(map[string]func(args map[string]interface{}) (string, error)),
	}
	r.registerDefaults()
	return r
}

func (r *SkillRegistry) registerDefaults() {
	r.tools["get_system_info"] = func(args map[string]interface{}) (string, error) {
		info := sys.GetSystemInfo()
		b, _ := json.MarshalIndent(info, "", "  ")
		return security.SanitizeSecrets(string(b)), nil
	}

	r.tools["get_network_status"] = func(args map[string]interface{}) (string, error) {
		return security.SanitizeSecrets(sys.GetNetworkSummary()), nil
	}

	r.tools["get_dhcp_leases"] = func(args map[string]interface{}) (string, error) {
		leases, err := netintel.GetEnrichedLeases()
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(leases, "", "  ")
		return security.SanitizeSecrets(string(b)), nil
	}

	r.tools["scan_lan_ports"] = func(args map[string]interface{}) (string, error) {
		ip, _ := args["ip"].(string)
		if ip == "" {
			return "", fmt.Errorf("se requiere el parámetro 'ip' del dispositivo a escanear")
		}
		if err := security.ValidateSSRFURL("http://" + ip); err != nil {
			return "", err
		}
		result := netintel.ScanLANPorts(ip, nil)
		b, _ := json.MarshalIndent(result, "", "  ")
		return string(b), nil
	}

	r.tools["get_wifi_qr"] = func(args map[string]interface{}) (string, error) {
		qrResult, err := netintel.GenerateWiFiQRFromConfig()
		if err != nil {
			return "", err
		}
		resp := fmt.Sprintf("📶 *Código QR de red WiFi (%s)*\n```\n%s\n```\n📌 Payload: `%s`", qrResult.SSID, qrResult.ASCIIBlock, qrResult.QRPayload)
		return security.SanitizeSecrets(resp), nil
	}

	r.tools["read_uci_config"] = func(args map[string]interface{}) (string, error) {
		pkg, _ := args["package"].(string)
		if pkg == "" {
			pkg = "network"
		}

		if err := security.ValidatePathSafety("/etc/config/" + pkg); err != nil {
			return "", err
		}

		cmd := exec.Command("uci", "show", pkg)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		if err != nil {
			return "", fmt.Errorf("error al leer UCI (%s): %v, salida: %s", pkg, err, out.String())
		}
		return security.SanitizeSecrets(out.String()), nil
	}

	r.tools["write_uci_config"] = func(args map[string]interface{}) (string, error) {
		path, _ := args["path"].(string)
		val, _ := args["value"].(string)
		pkg, _ := args["package"].(string)

		res, err := ExecuteTypedUCISet(pkg, path, val)
		return security.SanitizeSecrets(res), err
	}

	r.tools["exec_safe_cmd"] = func(args map[string]interface{}) (string, error) {
		command, _ := args["command"].(string)
		target, _ := args["target"].(string)

		fullCmd := command
		if command == "ping" && !strings.Contains(target, "-c") {
			if target != "" {
				fullCmd = fmt.Sprintf("ping -c 3 %s", target)
			} else {
				fullCmd = "ping -c 3 1.1.1.1"
			}
		} else if target != "" {
			fullCmd = fmt.Sprintf("%s %s", command, target)
		}

		if err := security.ValidateCommandSafety(fullCmd); err != nil {
			return "", err
		}

		output, err := security.ExecSafeCommandWithTimeout(fullCmd, 15*time.Second)
		return security.SanitizeSecrets(output), err
	}

	r.tools["restart_service"] = func(args map[string]interface{}) (string, error) {
		service, _ := args["service"].(string)
		res, err := ExecuteTypedServiceRestart(service)
		return security.SanitizeSecrets(res), err
	}

	r.tools["backup_to_supabase"] = func(args map[string]interface{}) (string, error) {
		backupPath := "/tmp/backup-clawrt.tar.gz"
		cmd := exec.Command("sysupgrade", "-b", backupPath)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		if err != nil {
			return "", fmt.Errorf("error al generar respaldo sysupgrade: %v", err)
		}
		return fmt.Sprintf("✅ Respaldo de OpenWrt generado exitosamente en %s (listo para cargar en Supabase Storage).", backupPath), nil
	}

	// ── Nuevas Herramientas Agénticas Avanzadas ──

	// 1. Diagnóstico de Auto-Sanación y Watchdog
	r.tools["self_healing_diagnostic"] = func(args map[string]interface{}) (string, error) {
		wd := watchdog.GetWatchdog()
		diag := wd.RunDiagnostic()
		if diag.OverallStatus != watchdog.StatusHealthy {
			recovered, actions := wd.AutoHeal(diag)
			diag.RecoveryActions = actions
			if recovered {
				diag.OverallStatus = watchdog.StatusHealthy
				diag.DiagnosisSummary += " (✅ Recuperado automáticamente por Auto-Sanación)"
			}
		}
		b, _ := json.MarshalIndent(diag, "", "  ")
		return string(b), nil
	}

	// 2. Guardia de Tráfico Conntrack / NetFlow
	r.tools["analyze_conntrack_traffic"] = func(args map[string]interface{}) (string, error) {
		report, err := netintel.AnalyzeConntrackTraffic()
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(report, "", "  ")
		return string(b), nil
	}

	// 3. Bloqueo de Host / IP Abusiva
	r.tools["block_abuser_ip"] = func(args map[string]interface{}) (string, error) {
		ip, _ := args["ip"].(string)
		reason, _ := args["reason"].(string)
		if reason == "" {
			reason = "Tráfico anómalo detectado por ClawRT"
		}
		return netintel.BlockAbuserIP(ip, reason)
	}

	// 4. Optimizador de Canales WiFi
	r.tools["optimize_wifi_channels"] = func(args map[string]interface{}) (string, error) {
		iface, _ := args["interface"].(string)
		apply, _ := args["apply"].(bool)
		report, err := netintel.OptimizeWiFiChannels(iface, apply)
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(report, "", "  ")
		return string(b), nil
	}

	// 5. Gestor de SQM / Bufferbloat QoS
	r.tools["manage_sqm_qos"] = func(args map[string]interface{}) (string, error) {
		action, _ := args["action"].(string)
		if action == "configure" {
			down, _ := args["download_mbps"].(float64)
			up, _ := args["upload_mbps"].(float64)
			qdisc, _ := args["qdisc"].(string)
			return netintel.ConfigureSQM(int(down), int(up), qdisc)
		}
		st, err := netintel.InspectSQMStatus()
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(st, "", "  ")
		return string(b), nil
	}

	// 6. Ejecución de Intenciones Declarativas de Configuración
	r.tools["execute_intent_plan"] = func(args map[string]interface{}) (string, error) {
		intentType, _ := args["type"].(string)
		params, _ := args["parameters"].(map[string]interface{})
		if params == nil {
			params = make(map[string]interface{})
		}
		req := intent.IntentRequest{
			Type:       intent.IntentType(intentType),
			Parameters: params,
		}
		plan, err := intent.ExecuteIntent(req)
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(plan, "", "  ")
		return string(b), nil
	}

	// 7. Auditoría de Seguridad del Router
	r.tools["audit_router_security"] = func(args map[string]interface{}) (string, error) {
		report, err := security.AuditRouterSecurity()
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(report, "", "  ")
		return string(b), nil
	}

	// 8. Mitigación de Clientes Pegajosos (Sticky Clients Roaming)
	r.tools["manage_sticky_clients"] = func(args map[string]interface{}) (string, error) {
		iface, _ := args["interface"].(string)
		kick, _ := args["kick_weakest"].(bool)
		report, err := netintel.DetectStickyClients(iface, kick)
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(report, "", "  ")
		return string(b), nil
	}

	// 9. Respaldo Cifrado AES-256-GCM
	r.tools["generate_encrypted_backup"] = func(args map[string]interface{}) (string, error) {
		pass, _ := args["passphrase"].(string)
		outPath, _ := args["output_path"].(string)
		return security.GenerateEncryptedBackup(pass, outPath)
	}

	// 10. Guardián de WireGuard & VPN
	r.tools["manage_wireguard"] = func(args map[string]interface{}) (string, error) {
		iface, _ := args["interface"].(string)
		reconnect, _ := args["auto_reconnect"].(bool)
		res, err := netintel.ManageWireGuard(iface, reconnect)
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), nil
	}

	// 11. Monitor de Privacidad DNS & AdBlock
	r.tools["inspect_dns_privacy"] = func(args map[string]interface{}) (string, error) {
		res, err := netintel.InspectDNSPrivacy()
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), nil
	}

	// 12. Asesor de Espacio en Flash y Paquetes
	r.tools["audit_flash_and_packages"] = func(args map[string]interface{}) (string, error) {
		res, err := sys.AuditFlashAndPackages()
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), nil
	}

	// 13. Inspector de Multi-WAN (mwan3 Failover)
	r.tools["check_multiwan_status"] = func(args map[string]interface{}) (string, error) {
		res, err := watchdog.CheckMultiWANStatus()
		if err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), nil
	}
}

func (r *SkillRegistry) GetToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_system_info",
				Description: "Obtiene información detallada del hardware, procesador, memoria RAM, uptime y versión de OpenWrt del router.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_network_status",
				Description: "Obtiene el estado de las interfaces de red, IP WAN/LAN y clientes WiFi conectados.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_dhcp_leases",
				Description: "Obtiene la lista enriquecida de clientes DHCP conectados a la red LAN (con fabricante OUI, detección de MAC privada aleatoria, señal RSSI y huella de SO).",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "scan_lan_ports",
				Description: "Escanea los 9 puertos críticos e inseguros (SSH, Telnet, HTTP, SMB, ADB, Redis, UPnP) en una IP de la red local.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"ip": {"type": "string", "description": "Dirección IP del cliente a escanear (ej: 192.168.1.100)"}
					},
					"required": ["ip"]
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_wifi_qr",
				Description: "Genera el código QR de acceso rápido a la red WiFi (ASCII y payload) leyendo la configuración activa.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "read_uci_config",
				Description: "Lee la configuración de un paquete UCI de OpenWrt (ej: network, wireless, firewall, system).",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"package": {"type": "string", "description": "Nombre del paquete UCI (ej: network, wireless, firewall, system)"}
					},
					"required": ["package"]
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "write_uci_config",
				Description: "Establece una opción de configuración UCI en OpenWrt con snapshot previo y rollback automático si falla.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"path": {"type": "string", "description": "Ruta UCI completa (ej: wireless.default_radio0.ssid)"},
						"value": {"type": "string", "description": "Nuevo valor a asignar"}
					},
					"required": ["path", "value"]
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "exec_safe_cmd",
				Description: "Ejecuta comandos seguros de diagnóstico de red de la Allowlist con límite de 15s y salida truncada a 4KB.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"command": {"type": "string", "description": "Nombre del comando seguro (ping, traceroute, nslookup, logread, df, free, uptime)"},
						"target": {"type": "string", "description": "Host de destino si aplica (ej: 8.8.8.8 o google.com)"}
					},
					"required": ["command"]
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "restart_service",
				Description: "Reinicia un servicio seguro del sistema OpenWrt (network, dnsmasq, firewall, clawrt, dropbear, uhttpd).",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"service": {"type": "string", "description": "Servicio a reiniciar (network, dnsmasq, firewall, clawrt, dropbear, uhttpd)"}
					},
					"required": ["service"]
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "self_healing_diagnostic",
				Description: "Ejecuta un diagnóstico completo en cascada (Gateway, DNS local/público, Internet WAN) y ejecuta auto-sanación si detecta fallas.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "analyze_conntrack_traffic",
				Description: "Inspecciona la tabla de conexiones de red (nf_conntrack) para detectar IPs que saturan ancho de banda, escaneos de puertos o posibles ataques SYN flood.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "block_abuser_ip",
				Description: "Bloquea temporalmente una IP en el firewall por comportamiento abusivo o sospechoso.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"ip": {"type": "string", "description": "Dirección IP a bloquear"},
						"reason": {"type": "string", "description": "Motivo del bloqueo"}
					},
					"required": ["ip"]
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "optimize_wifi_channels",
				Description: "Analiza el espectro inalámbrico escaneando redes vecinas y calcula el canal con menor congestión e interferencia.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"interface": {"type": "string", "description": "Interfaz física WiFi (ej: wlan0)"},
						"apply": {"type": "boolean", "description": "Si es true, aplica automáticamente el canal óptimo"}
					}
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "manage_sqm_qos",
				Description: "Inspecciona o configura SQM (Smart Queue Management con Cake) para erradicar el Bufferbloat y latencia en llamadas/juegos.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"action": {"type": "string", "description": "'status' para diagnosticar o 'configure' para ajustar velocidades"},
						"download_mbps": {"type": "number", "description": "Velocidad de descarga en Mbps"},
						"upload_mbps": {"type": "number", "description": "Velocidad de subida en Mbps"},
						"qdisc": {"type": "string", "description": "Algoritmo de cola ('cake' o 'fq_codel')"}
					}
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "execute_intent_plan",
				Description: "Ejecuta intenciones complejas de configuración declarativa (guest_wifi, port_forward, block_client, static_lease) con validación atómica y rollback.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"type": {"type": "string", "description": "Tipo de intención: 'guest_wifi', 'port_forward', 'block_client', 'static_lease'"},
						"parameters": {"type": "object", "description": "Parámetros específicos de la intención"}
					},
					"required": ["type", "parameters"]
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "backup_to_supabase",
				Description: "Genera un respaldo completo de la configuración del router (/etc/config/*) para guardar en Supabase Storage.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "audit_router_security",
				Description: "Audita la seguridad general del router: verifica si root tiene clave, si SSH/LuCI están expuestos a Internet, servicios telnet y reglas de firewall.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "manage_sticky_clients",
				Description: "Detecta clientes WiFi conectados con señal débil (< -80 dBm) que ralentizan la celda y permite forzar su reconexión/roaming.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"interface": {"type": "string", "description": "Interfaz WiFi (ej: wlan0)"},
						"kick_weakest": {"type": "boolean", "description": "Si es true, desasocia suavemente a los clientes débiles para forzar roaming"}
					}
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "generate_encrypted_backup",
				Description: "Genera un respaldo de configuración de OpenWrt cifrado localmente con AES-256-GCM.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"passphrase": {"type": "string", "description": "Clave o contraseña para cifrar el respaldo"},
						"output_path": {"type": "string", "description": "Ruta destino (/tmp/backup.enc)"}
					}
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "manage_wireguard",
				Description: "Inspecciona el estado de túneles WireGuard, detecta peers con handshake congelado (> 180s) y permite reconectar automáticamente.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"interface": {"type": "string", "description": "Interfaz WireGuard (ej: wg0)"},
						"auto_reconnect": {"type": "boolean", "description": "Si es true, reinicia el túnel para forzar reconexión"}
					}
				}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "inspect_dns_privacy",
				Description: "Audita la privacidad DNS del router (DoH / DoT cifrado), filtrado de publicidad/malware (AdBlock) y riesgo de fuga de DNS (DNS leak).",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "audit_flash_and_packages",
				Description: "Audita el espacio libre en Flash (/overlay), evalúa el riesgo antes de instalar paquetes y lista actualizaciones disponibles.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "check_multiwan_status",
				Description: "Inspecciona el estado de múltiples enlaces WAN (mwan3), balanceo de carga y detección de failover.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
		},
	}
}

func (r *SkillRegistry) ExecuteTool(name string, args map[string]interface{}) (string, error) {
	fn, exists := r.tools[name]
	if !exists {
		return "", fmt.Errorf("herramienta no encontrada: %s", name)
	}
	res, err := fn(args)
	return security.SanitizeSecrets(res), err
}
