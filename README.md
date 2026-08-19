<p align="center">
  <img src="assets/logo/clawrt-logo-dark.svg" alt="ClawRT Fused Brand Identity" width="420"/>
</p>

<p align="center">
  <a href="https://github.com/Havr88/clawrt/actions/workflows/ci.yml"><img src="https://github.com/Havr88/clawrt/actions/workflows/ci.yml/badge.svg" alt="ClawRT CI"/></a>
  <a href="https://openwrt.org"><img src="https://img.shields.io/badge/OpenWrt-18.06%20%7C%2022.03%20%7C%2025.12-blue.svg" alt="OpenWrt Supported"/></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg" alt="Go Version"/></a>
  <a href="BRAND.md"><img src="https://img.shields.io/badge/Brand-Fused%20Identity-FF6F00.svg" alt="Brand Guide"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"/></a>
</p>

> **clawrt** es un agente autónomo de Inteligencia Artificial ultraligero, seguro y optimizado para routers con **OpenWrt**. Diseñado con la **Identidad Visual Fused** (`#FF6F00` ➔ `#C173FF` ➔ `#3AB5ED` ➔ `#00C7E2`), funciona con memoria RAM mínima (~2.2 MB en reposo) y ofrece control agéntico local mediante **LuCI Web Copilot**, mensajería por Telegram (CGNAT-safe), auto-sanación de red autónoma, guardia de tráfico Conntrack y configuración basada en intenciones declarativas.

---

## ⚙️ 1. Matriz de Requerimientos de Hardware y Viabilidad

| Nivel de Viabilidad | Memoria Flash | Memoria RAM | Versiones OpenWrt | Modelos de Ejemplo | Estado y Modo de Operación |
|:---|:---:|:---:|:---:|:---|:---|
| **Mínimo Viable (Extremo)** | **8 MB** | **32 MB** | OpenWrt 18.06 / 19.07 | TP-Link WR741ND, WR841N | ⚠️ `TierExtremeMinimal`: Descarga a `/tmp`, reposo RAM forzado, FastPath y Watchdog FSM ligero. |
| **Estándar Recomendado** | **16 MB** | **64 MB** | OpenWrt 22.03 / 23.05 / 25.12 | Xiaomi Mi Router 4C | ✅ `TierMinimal`: 100% Estable, soporte LuCI Copilot, bot Telegram, FastPath y Conntrack Guard. |
| **Alto Rendimiento (Ideal)** | **32 MB+** | **128 MB+** | OpenWrt 23.05 / 25.12+ | GL.iNet, Linksys, x86_64 | 🚀 `TierMedium`/`TierFull`: Streaming Realtime, Intent Engine, UBUS Listener y almacenamiento en la nube. |

---

## 🔥 2. Resumen de Funcionalidades del Agente Autónomo

| Categoría | Funcionalidad | Descripción Técnica |
|:---|:---|:---|
| 🩺 **Auto-Sanación Activa** | **Self-Healing Watchdog** | Diagnóstico en cascada continuo (Gateway ➔ DNS Local ➔ DNS Público ➔ WAN Ping) y auto-reparación en caliente (reinicio de `dnsmasq`, renegociación de WAN `ifup` y recarga de cortafuegos). |
| ⚡ **Eventos en Tiempo Real** | **UBUS Reactive Listener** | Daemon que escucha eventos del bus UBUS (`network.interface`, `hostapd`, `procd`) sin sobrecarga de sondeo ni subshells. |
| 🤖 **Copiloto Web Local** | **LuCI Web Copilot** | Asistente conversacional interactivo integrado directamente en la interfaz web de LuCI con botones de diagnóstico y optimización rápida. |
| 🛡️ **Guardia de Tráfico** | **Conntrack / NetFlow Guard** | Monitoreo de `/proc/net/nf_conntrack` para detectar dispositivos acaparadores de ancho de banda (Torrents/P2P), escaneos de puertos y ataques SYN flood, con bloqueo en firewall. |
| ✨ **Optimización Inalámbrica** | **WiFi Auto-Channel** | Análisis del espectro en 2.4 GHz y 5 GHz (`iwinfo scan`) para seleccionar y aplicar el canal con menor congestión e interferencia de redes vecinas. |
| 🚀 **Gestión de Latencia** | **SQM Bufferbloat QoS** | Evaluación y ajuste del algoritmo de cola Cake / FQ_Codel para erradicar el retardo y jitter en videollamadas y juegos. |
| 🎯 **Configuración por Intenciones**| **Intent-Based Config** | Traduce intenciones humanas de alto nivel (*"crea una red de invitados aislada"*, *"redirecciona el puerto 8080"*) en planes atómicos UCI con verificación `fw4 check` y rollback automático. |
| 🧠 **Motor Agéntico Híbrido** | **FastPath Cero-LLM** | Resuelve saludos, diagnósticos y consultas operativas en **0ms y 0 tokens** de IA. |
| 🛡️ **Seguridad ACP** | **Allowlist & Hard Denylist** | Bloquea comandos peligrosos (`rm -rf`, `sysupgrade`, `reboot`), previene ataques SSRF y sanitiza inyecciones en la shell. |
| ⚡ **Optimización RAM** | **Auto-Sleep (5 Minutos)** | Ejecuta `debug.FreeOSMemory()` tras 5 minutos de inactividad, manteniendo el consumo en ~2.2 MB. |

---

## 📲 3. Comandos de Telegram y LuCI Copilot (`/help`)

| Comando | Descripción | Ejemplo de Salida |
|:---|:---|:---|
| `/status` (o `/sysinfo`) | Estado del router, CPU, RAM y tiempo de actividad | `⚙️ OpenWrt 25.12.5 \| CPU: 0.15 \| RAM: 18/58 MB (31%) \| Uptime: 4d 12h` |
| `/diagnose` (o `/heal`) | Diagnóstico completo de conectividad y auto-sanación | `🩺 Diagnóstico: Gateway OK \| DNS OK \| WAN OK (Latencia: 14.2ms)` |
| `/optimize` (o `/wifiopt`)| Analiza canales WiFi y recomienda el menos congestionado | `✨ Canal actual: 6 \| Óptimo recomendado: Canal 1 (Interferencias: 1)` |
| `/conntrack` (o `/guard`) | Inspecciona tabla de conexiones en busca de anomalías | `🛡️ Conexiones activas: 84 \| Top Talker: 192.168.1.50 (32 sockets)` |
| `/sqm` (o `/bufferbloat`)| Diagnóstico y calidad de Bufferbloat QoS | `⚡ Calidad: Grado A \| Algoritmo Cake activo en WAN (90 Mbps)` |
| `/clients` (o `/dhcp`) | Lista dispositivos conectados con OUI y tipo de MAC | `💻 Xiaomi-Phone (192.168.1.145) - Apple Inc. [MAC Privada: No]` |
| `/wifi` | Estado de interfaces inalámbricas, SSID y clientes | `📶 Radio 0 (2.4GHz) \| SSID: ClawRT_Wi-Fi \| Clientes: 5` |
| `/qrwifi` (o `/qr`) | Genera código QR ASCII para conexión inmediata | `████████████ (Código QR ASCII escaneable en pantalla)` |
| `/scan [ip]` | Escaneo de 9 puertos críticos en dispositivo LAN | `🔍 Escaneo: 192.168.1.100 (SSH:22 ABIERTO, HTTP:80 ABIERTO)` |
| `/models` | Consulta modelos disponibles en vivo del proveedor LLM | `🤖 Modelos en vivo: deepseek-v4-flash-free, agnes-2.5-flash, grok-4.5-free` |
| `/firewall` | Muestra zonas de cortafuegos y reglas activas | `🔥 Cortafuegos UCI: WAN ➔ LAN (REJECT) \| Reglas: 12` |
| `/memory` (o `/gc`) | Muestra uso de memoria RAM y ejecuta Garbage Collection | `🧠 GC Ejecutado: Memoria liberada. Consumo en reposo: 2.2 MB` |
| `/clear` | Limpia el historial de hechos en `/tmp/clawrt_facts.json` | `🧹 Hechos aprendidos borrados correctamente.` |
| `/reboot` | Reinicia el servicio daemon de ClawRT en OpenWrt | `🔄 Servicio clawrt reiniciado con éxito.` |

---

## 🏗️ 4. Arquitectura del Proyecto

```text
clawrt/
├── cmd/clawrt/              ← Punto de entrada daemon, banderas CLI y despachador
├── internal/
│   ├── config/              ← Cargador de configuración UCI (/etc/config/clawrt) y entornos
│   ├── fastpath/            ← Motor Cero-LLM (0ms, 0 tokens)
│   ├── hotplug/             ← Eventos reactivos (/etc/hotplug.d/) con auto-sanación
│   ├── i18n/                ← Motor de traducción para 9 idiomas
│   ├── intent/              ← Motor de intenciones declarativas multi-paso con rollback
│   ├── knowledge/           ← Catálogo de programas OpenWrt & Table of Hardware (toh.json)
│   ├── llm/                 ← Cliente OpenAI-compatible con selector de modelos dinámico
│   ├── netintel/            ← Conntrack Guard, WiFi Optimizer, SQM QoS y Enriched Leases
│   ├── security/            ← Allowlist, SanitizeSecrets, SSRF guard y control de inyección
│   ├── skills/              ← Registro de herramientas agénticas (Tool Definitions)
│   ├── store/               ← Adaptador de base de datos híbrida (Supabase / Cloudflare / Upstash)
│   ├── sys/                 ← Información del sistema, Memory Manager y UCI Executor
│   ├── telegram/            ← Bot de Telegram (Long-polling CGNAT-safe y canal unificado)
│   ├── ubus/                ← Listener de eventos reactivos del sistema (ubus listen)
│   └── watchdog/            ← Motor de diagnóstico y auto-sanación autónoma (Self-Healing)
├── luci-app-clawrt/         ← Paquete LuCI (Web Copilot SPA + Configuración + i18n)
└── deploy/                  ← Scripts de instalación y esquemas de base de datos
```

---

## 📦 5. Métodos de Instalación

### 🚀 Despliegue Directo SSH/SCP
```bash
ROUTER_IP="192.168.1.1" bash scripts/deploy.sh
```

### 📦 Paquetes `.ipk` / `.apk`
```sh
scp -O *.ipk root@192.168.1.1:/tmp/
ssh root@192.168.1.1 "opkg install /tmp/*.ipk 2>/dev/null || apk add --allow-untrusted /tmp/*.ipk"
```

---

## 📄 Licencia
Este proyecto está bajo la Licencia **MIT**. Consulta el archivo [LICENSE](LICENSE) para más detalles.
