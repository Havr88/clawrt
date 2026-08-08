package skills

import (
	"bytes"
	"clawrt/internal/netintel"
	"clawrt/internal/security"
	"clawrt/internal/sys"
	"encoding/json"
	"fmt"
	"os/exec"
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
		if target != "" {
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
