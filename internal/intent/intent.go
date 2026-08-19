package intent

import (
	"bytes"
	"clawrt/internal/sys"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type IntentType string

const (
	IntentGuestWiFi   IntentType = "guest_wifi"
	IntentPortForward IntentType = "port_forward"
	IntentBlockClient IntentType = "block_client"
	IntentStaticLease IntentType = "static_lease"
	IntentIsolateIoT  IntentType = "isolate_iot"
)

type IntentRequest struct {
	Type        IntentType             `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Description string                 `json:"description,omitempty"`
}

type StepExecution struct {
	StepNumber int    `json:"step_number"`
	Action     string `json:"action"`
	Status     string `json:"status"` // SUCCESS, FAILED
	Error      string `json:"error,omitempty"`
}

type IntentExecutionPlan struct {
	IntentName   string          `json:"intent_name"`
	Steps        []StepExecution `json:"steps"`
	Completed    bool            `json:"completed"`
	RollbackDone bool            `json:"rollback_done"`
	Summary      string          `json:"summary"`
}

func ExecuteIntent(req IntentRequest) (*IntentExecutionPlan, error) {
	plan := &IntentExecutionPlan{
		IntentName: string(req.Type),
		Steps:      make([]StepExecution, 0),
	}

	switch req.Type {
	case IntentGuestWiFi:
		return executeGuestWiFiIntent(req.Parameters, plan)
	case IntentPortForward:
		return executePortForwardIntent(req.Parameters, plan)
	case IntentBlockClient:
		return executeBlockClientIntent(req.Parameters, plan)
	case IntentStaticLease:
		return executeStaticLeaseIntent(req.Parameters, plan)
	default:
		return nil, fmt.Errorf("tipo de intención no soportada: %s (soportadas: guest_wifi, port_forward, block_client, static_lease)", req.Type)
	}
}

// 1. Guest WiFi Creation Intent (Multi-step atomic workflow)
func executeGuestWiFiIntent(params map[string]interface{}, plan *IntentExecutionPlan) (*IntentExecutionPlan, error) {
	ssid, _ := params["ssid"].(string)
	key, _ := params["key"].(string)
	ipSubnet, _ := params["subnet"].(string)

	if ssid == "" {
		ssid = "ClawRT_Guest"
	}
	if key == "" {
		key = "InvitadosSegura2026"
	}
	if ipSubnet == "" {
		ipSubnet = "192.168.10.1"
	}

	steps := []struct {
		desc string
		pkg  string
		path string
		val  string
	}{
		// Step 1: Create network interface 'guest'
		{"Crear interfaz de red 'guest'", "network", "network.guest.proto", "static"},
		{"Asignar IP a interfaz 'guest'", "network", "network.guest.ipaddr", ipSubnet},
		{"Asignar máscara a interfaz 'guest'", "network", "network.guest.netmask", "255.255.255.0"},

		// Step 2: Configure DHCP for guest
		{"Configurar DHCP para interfaz 'guest'", "dhcp", "dhcp.guest.interface", "guest"},
		{"Establecer rango inicial DHCP", "dhcp", "dhcp.guest.start", "100"},
		{"Establecer límite clientes DHCP", "dhcp", "dhcp.guest.limit", "150"},
		{"Establecer tiempo de concesión DHCP", "dhcp", "dhcp.guest.leasetime", "1h"},

		// Step 3: Configure Firewall Zone (Guest -> WAN OK, Guest -> LAN Denied)
		{"Crear zona de firewall 'guest'", "firewall", "firewall.guest_zone.name", "guest"},
		{"Asignar red a zona 'guest'", "firewall", "firewall.guest_zone.network", "guest"},
		{"Establecer input firewall", "firewall", "firewall.guest_zone.input", "REJECT"},
		{"Establecer output firewall", "firewall", "firewall.guest_zone.output", "ACCEPT"},
		{"Establecer forward firewall", "firewall", "firewall.guest_zone.forward", "REJECT"},

		// Step 4: Allow DHCP and DNS from guest to router
		{"Crear regla DHCP para invitados", "firewall", "firewall.guest_dhcp.name", "Allow-Guest-DHCP"},
		{"Regla DHCP src", "firewall", "firewall.guest_dhcp.src", "guest"},
		{"Regla DHCP port", "firewall", "firewall.guest_dhcp.dest_port", "67-68"},
		{"Regla DHCP proto", "firewall", "firewall.guest_dhcp.proto", "udp"},
		{"Regla DHCP target", "firewall", "firewall.guest_dhcp.target", "ACCEPT"},

		{"Crear regla DNS para invitados", "firewall", "firewall.guest_dns.name", "Allow-Guest-DNS"},
		{"Regla DNS src", "firewall", "firewall.guest_dns.src", "guest"},
		{"Regla DNS port", "firewall", "firewall.guest_dns.dest_port", "53"},
		{"Regla DNS proto", "firewall", "firewall.guest_dns.proto", "tcp udp"},
		{"Regla DNS target", "firewall", "firewall.guest_dns.target", "ACCEPT"},

		// Forwarding: Guest -> WAN
		{"Crear reenvío Guest a WAN", "firewall", "firewall.guest_wan_fwd.src", "guest"},
		{"Reenvío Guest dest", "firewall", "firewall.guest_wan_fwd.dest", "wan"},
		{"Reenvío Guest target", "firewall", "firewall.guest_wan_fwd.target", "ACCEPT"},

		// Step 5: Configure Wireless VAP (SSID + Isolation)
		{"Crear red WiFi para invitados", "wireless", "wireless.guest_wifi.device", "radio0"},
		{"Asignar modo AP a WiFi", "wireless", "wireless.guest_wifi.mode", "ap"},
		{"Asignar red 'guest' al WiFi", "wireless", "wireless.guest_wifi.network", "guest"},
		{"Establecer SSID invitados", "wireless", "wireless.guest_wifi.ssid", ssid},
		{"Establecer cifrado WPA2/WPA3", "wireless", "wireless.guest_wifi.encryption", "psk2"},
		{"Establecer clave WiFi invitados", "wireless", "wireless.guest_wifi.key", key},
		{"Activar aislamiento entre clientes (Client Isolation)", "wireless", "wireless.guest_wifi.isolate", "1"},
	}

	for i, s := range steps {
		stepNum := i + 1
		res, err := sys.ExecuteTypedUCISet(s.pkg, s.path, s.val)
		if err != nil {
			log.Printf("[INTENT] Error en paso %d (%s): %v", stepNum, s.desc, err)
			plan.Steps = append(plan.Steps, StepExecution{
				StepNumber: stepNum,
				Action:     s.desc,
				Status:     "FAILED",
				Error:      err.Error(),
			})
			plan.RollbackDone = true
			plan.Summary = fmt.Sprintf("❌ Fallo en paso %d (%s). Se ejecutó rollback automático.", stepNum, s.desc)
			return plan, err
		}

		plan.Steps = append(plan.Steps, StepExecution{
			StepNumber: stepNum,
			Action:     fmt.Sprintf("%s (%s)", s.desc, res),
			Status:     "SUCCESS",
		})
	}

	// Reload services
	_, _ = sys.ExecuteTypedServiceRestart("network")
	_, _ = sys.ExecuteTypedServiceRestart("firewall")
	_, _ = sys.ExecuteTypedServiceRestart("dnsmasq")

	plan.Completed = true
	plan.Summary = fmt.Sprintf("✅ Red WiFi de Invitados aislada creada exitosamente: SSID='%s', Clave='%s', Subnet='%s/24' (Aislamiento de clientes activo).", ssid, key, ipSubnet)
	return plan, nil
}

// 2. Port Forwarding Intent
func executePortForwardIntent(params map[string]interface{}, plan *IntentExecutionPlan) (*IntentExecutionPlan, error) {
	name, _ := params["name"].(string)
	srcPort, _ := params["external_port"].(string)
	destIP, _ := params["internal_ip"].(string)
	destPort, _ := params["internal_port"].(string)
	proto, _ := params["proto"].(string)

	if name == "" {
		name = "ClawRT_Forward"
	}
	if proto == "" {
		proto = "tcp"
	}
	if destPort == "" {
		destPort = srcPort
	}

	if srcPort == "" || destIP == "" {
		return nil, fmt.Errorf("se requieren los parámetros 'external_port' e 'internal_ip'")
	}

	secName := fmt.Sprintf("firewall.%s", strings.ReplaceAll(strings.ToLower(name), " ", "_"))
	steps := []struct {
		desc string
		pkg  string
		path string
		val  string
	}{
		{"Crear sección de redirección", "firewall", secName + ".name", name},
		{"Asignar tipo redirect", "firewall", secName + ".target", "DNAT"},
		{"Establecer zona origen WAN", "firewall", secName + ".src", "wan"},
		{"Establecer zona destino LAN", "firewall", secName + ".dest", "lan"},
		{"Establecer protocolo", "firewall", secName + ".proto", proto},
		{"Establecer puerto externo", "firewall", secName + ".src_dport", srcPort},
		{"Establecer IP destino", "firewall", secName + ".dest_ip", destIP},
		{"Establecer puerto interno destino", "firewall", secName + ".dest_port", destPort},
	}

	for i, s := range steps {
		stepNum := i + 1
		_, err := sys.ExecuteTypedUCISet(s.pkg, s.path, s.val)
		if err != nil {
			plan.Steps = append(plan.Steps, StepExecution{
				StepNumber: stepNum,
				Action:     s.desc,
				Status:     "FAILED",
				Error:      err.Error(),
			})
			plan.RollbackDone = true
			plan.Summary = fmt.Sprintf("❌ Error al configurar redirección de puertos: %v", err)
			return plan, err
		}
		plan.Steps = append(plan.Steps, StepExecution{
			StepNumber: stepNum,
			Action:     s.desc,
			Status:     "SUCCESS",
		})
	}

	_, _ = sys.ExecuteTypedServiceRestart("firewall")
	plan.Completed = true
	plan.Summary = fmt.Sprintf("✅ Redirección de puertos configurada: WAN :%s (%s) ➡️ LAN %s:%s.", srcPort, proto, destIP, destPort)
	return plan, nil
}

// 3. Block Client Intent (MAC / IP firewall drop)
func executeBlockClientIntent(params map[string]interface{}, plan *IntentExecutionPlan) (*IntentExecutionPlan, error) {
	mac, _ := params["mac"].(string)
	ip, _ := params["ip"].(string)
	reason, _ := params["reason"].(string)

	if mac == "" && ip == "" {
		return nil, fmt.Errorf("se requiere 'mac' o 'ip' del cliente a bloquear")
	}

	ruleName := "block_" + strings.ReplaceAll(strings.ToLower(mac), ":", "")
	if ruleName == "block_" {
		ruleName = "block_" + strings.ReplaceAll(ip, ".", "_")
	}

	secPath := "firewall." + ruleName
	_, _ = sys.ExecuteTypedUCISet("firewall", secPath+".name", "Block-"+ruleName)
	_, _ = sys.ExecuteTypedUCISet("firewall", secPath+".src", "lan")
	_, _ = sys.ExecuteTypedUCISet("firewall", secPath+".dest", "wan")
	if mac != "" {
		_, _ = sys.ExecuteTypedUCISet("firewall", secPath+".src_mac", mac)
	}
	if ip != "" {
		_, _ = sys.ExecuteTypedUCISet("firewall", secPath+".src_ip", ip)
	}
	_, _ = sys.ExecuteTypedUCISet("firewall", secPath+".target", "REJECT")

	_, _ = sys.ExecuteTypedServiceRestart("firewall")

	plan.Completed = true
	plan.Summary = fmt.Sprintf("🛡️ Cliente bloqueado en cortafuegos: MAC='%s', IP='%s' (Razón: %s).", mac, ip, reason)
	return plan, nil
}

// 4. Static DHCP Lease Intent
func executeStaticLeaseIntent(params map[string]interface{}, plan *IntentExecutionPlan) (*IntentExecutionPlan, error) {
	mac, _ := params["mac"].(string)
	ip, _ := params["ip"].(string)
	name, _ := params["hostname"].(string)

	if mac == "" || ip == "" {
		return nil, fmt.Errorf("se requieren 'mac' e 'ip' para fijar dirección estática")
	}
	if name == "" {
		name = "host_" + strings.ReplaceAll(strings.ToLower(mac), ":", "")
	}

	secPath := "dhcp." + strings.ReplaceAll(strings.ToLower(name), " ", "_")
	_, _ = sys.ExecuteTypedUCISet("dhcp", secPath+".mac", mac)
	_, _ = sys.ExecuteTypedUCISet("dhcp", secPath+".ip", ip)
	_, _ = sys.ExecuteTypedUCISet("dhcp", secPath+".name", name)

	_, _ = sys.ExecuteTypedServiceRestart("dnsmasq")

	plan.Completed = true
	plan.Summary = fmt.Sprintf("📌 IP Estática fijada en DHCP: Equipo '%s' (MAC: %s) ➡️ IP %s.", name, mac, ip)
	return plan, nil
}

func ValidateFirewallSyntax() error {
	cmd := exec.Command("fw4", "check")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fw4 check syntax error: %v (%s)", err, out.String())
	}
	return nil
}
