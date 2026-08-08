package hotplug

import (
	"clawrt/internal/config"
	"clawrt/internal/netintel"
	"clawrt/internal/skills"
	"clawrt/internal/telegram"
	"fmt"
	"log"
	"strings"
)

type Event struct {
	Subsystem string   `json:"subsystem"`
	Action    string   `json:"action"`
	Params    []string `json:"params"`
}

func ProcessHotplugEvent(cfg *config.Config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("parámetros insuficientes para hotplug event")
	}

	subsystem := strings.ToLower(args[0])
	action := strings.ToLower(args[1])
	extraParams := args[2:]

	registry := skills.NewRegistry()
	bot := telegram.NewBot(cfg, registry)

	var alertMsg string

	switch subsystem {
	case "iface", "network":
		ifaceName := "wan"
		if len(extraParams) > 0 {
			ifaceName = extraParams[0]
		}
		ipAddr := ""
		if len(extraParams) > 2 {
			ipAddr = extraParams[2]
		}

		if action == "ifdown" {
			alertMsg = fmt.Sprintf("🚨 *Alerta de Red (WAN/Interfaz):*\nLa interfaz `%s` se ha **DESCONECTADO** (Caída de enlace).", ifaceName)
		} else if action == "ifup" {
			if ipAddr != "" {
				alertMsg = fmt.Sprintf("✅ *Alerta de Red (WAN/Interfaz):*\nLa interfaz `%s` se ha **RESTAURADO** correctamente.\n📍 Nueva IP: `%s`", ifaceName, ipAddr)
			} else {
				alertMsg = fmt.Sprintf("✅ *Alerta de Red (WAN/Interfaz):*\nLa interfaz `%s` se ha **RESTAURADO** correctamente.", ifaceName)
			}
		}

	case "dhcp":
		mac := ""
		ip := ""
		hostname := "desconocido"
		if len(extraParams) > 0 {
			mac = extraParams[0]
		}
		if len(extraParams) > 1 {
			ip = extraParams[1]
		}
		if len(extraParams) > 2 && extraParams[2] != "" {
			hostname = extraParams[2]
		}

		vendor := netintel.LookupVendor(mac)
		isRandom := netintel.IsRandomizedMAC(mac)
		macTypeStr := "MAC Fija de Hardware"
		if isRandom {
			macTypeStr = "⚠️ MAC Aleatoria / Privada (iOS/Android Privacy)"
		}

		if action == "add" {
			alertMsg = fmt.Sprintf("📱 *Nuevo Cliente Conectado a la LAN:*\n• **Equipo:** `%s`\n• **IP:** `%s`\n• **MAC:** `%s`\n• **Fabricante:** `%s`\n• **Tipo MAC:** %s", hostname, ip, mac, vendor, macTypeStr)
		} else if action == "del" {
			alertMsg = fmt.Sprintf("📴 *Cliente Desconectado de la LAN:*\n• **Equipo:** `%s`\n• **IP:** `%s`\n• **MAC:** `%s`", hostname, ip, mac)
		}

	case "button":
		buttonName := "WPS"
		seenSec := "0"
		if len(extraParams) > 0 {
			buttonName = extraParams[0]
		}
		if len(extraParams) > 1 {
			seenSec = extraParams[1]
		}

		alertMsg = fmt.Sprintf("🔘 *Evento de Botón Físico en Router:*\nSe presiono el botón `%s` (Acción: `%s`, Duración: %s s).", buttonName, action, seenSec)

	default:
		alertMsg = fmt.Sprintf("🔔 *Evento de Sistema Hotplug (%s):*\nAcción: `%s`, Parámetros: `%s`", subsystem, action, strings.Join(extraParams, " "))
	}

	if alertMsg == "" || len(cfg.ChatIDs) == 0 {
		return nil
	}

	// Send alert message to authorized Telegram chats
	for _, chatID := range cfg.ChatIDs {
		err := bot.SendMessage(chatID, alertMsg)
		if err != nil {
			log.Printf("[ERROR] Error al enviar alerta hotplug a chat %d: %v", chatID, err)
		}
	}

	return nil
}
