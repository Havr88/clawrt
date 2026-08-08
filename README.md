# ClawRT para OpenWrt 🐾⚡

[![ClawRT CI](https://github.com/Havr88/clawrt/actions/workflows/ci.yml/badge.svg)](https://github.com/Havr88/clawrt/actions/workflows/ci.yml)
[![OpenWrt Supported](https://img.shields.io/badge/OpenWrt-21.02%20%7C%2023.05%20%7C%2025.12-blue.svg)](https://openwrt.org)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **ClawRT** es un agente autónomo de IA ultraligero y seguro adaptado específicamente para routers con **OpenWrt** (probado en plataformas `mipsle` como Xiaomi Mi Router 4C y `armv7` como Linksys Velop WHW01). Inspirado en arquitecturas agénticas compactas y optimizado a partir del aprendizaje de proyectos libres como **RouterClaw**, **NanoClaw**, **OpenClaw**, **Hermes** y **AgentWRT**.

---

## 📦 Métodos de Instalación y Despliegue

### 🚀 Método 1: Despliegue Directo vía SSH/SCP (Recomendado para Laboratorio)

Si tienes acceso SSH a tu router en la red local:

```bash
ROUTER_IP="192.168.1.1" bash scripts/deploy.sh
```

---

### 📦 Método 2: Instalación de Paquetes `.ipk` (GitHub Releases / Artifacts)

1. Descarga el paquete de tu arquitectura desde la sección **Releases** o **Actions Artifacts** en GitHub:
   - `clawrt_0.5.0-1_mipsel_24kc.ipk` (para Xiaomi Mi Router 4C y MIPS LE)
   - `luci-app-clawrt_0.5.0-1_all.ipk` (Interfaz de usuario LuCI)
   - `luci-i18n-clawrt-es_0.5.0-1_all.ipk` (Traducción al español)
2. Transfiere los archivos al router e instala:
   ```sh
   scp -O *.ipk root@192.168.1.1:/tmp/
   ssh root@192.168.1.1 "opkg install /tmp/*.ipk 2>/dev/null || apk add --allow-untrusted /tmp/*.ipk"
   ```

---

### 🌐 Método 3: Feed en Red Local LAN (Opcional)

Si deseas compartir el repositorio de paquetes en tu red local desde tu computadora:

```bash
# En tu PC de desarrollo:
bash scripts/make-feed.sh
cd dist/feed && python3 -m http.server 8080
```

Y en el router agrega la IP de tu computadora (ej: `192.168.1.50`):

```sh
# En OpenWrt 25.12+ (apk):
echo "http://192.168.1.50:8080/mipsel_24kc" >> /etc/apk/repositories.d/clawrt.list
apk update && apk add clawrt luci-app-clawrt luci-i18n-clawrt-es
```

---

## 💡 Filosofía e Inspiración

ClawRT busca ser el núcleo agéntico **más ligero, rápido y seguro** para routers de recursos limitados (64 MB RAM, 16 MB Flash NOR):
- **Bajo consumo de memoria**: ~3–5 MB de RAM en ejecución.
- **Sin dependencias dinámicas**: Compilado estáticamente en **Go** (`CGO_ENABLED=0`) con compresión **UPX**.
- **Capa de Seguridad ACP, Allowlist & Hard Denylist**: Bloqueo absoluto de comandos destructivos (`reboot`, `rm -rf /`, `sysupgrade`), validación de inyección de shell y ejecuciones limitadas a 15s / 4KB.
- **Herramientas UCI Tipadas con Snapshot & Rollback Automático**: Creación de respaldos previos a cambios UCI y verificación de sintaxis del cortafuegos (`fw4 check`) con reversión automática si ocurren errores.
- **Escáner de Secretos en Tiempo Real & SSRF**: Enmascaramiento de tokens y claves secretas (`[SECRETO_REDACTADO]`) y bloqueo de IP metadata `169.254.169.254`.
- **Alertas Reactivas del Sistema (`/etc/hotplug.d/`)**: Notificación proactiva e instantánea por Telegram ante caídas/restauraciones de enlace WAN (`iface`), conexiones/desconexiones de nuevos clientes LAN (`dhcp`) y eventos de botones físicos (`button`).
- **Inteligencia Enriquecida de Clientes DHCP (`/leases`)**: Cruce multifuente de concesiones DHCP, reservas estáticas UCI, señal WiFi RSSI (dBm), PHY rate (Mbps), huella de SO y fabricante por OUI MAC.
- **Detección de MAC Privada / Aleatoria (Randomized MAC)**: Alerta automática cuando un dispositivo utiliza MAC aleatoria privada (iOS/Android Privacy).
- **Generación de Código QR WiFi (`/qrwifi`)**: Genera el código QR de conexión rápida a la red WiFi en formato bloques ASCII y como payload estándar.
- **Escáner Pasivo de Puertos LAN (`/scan`)**: Inspección ultraligera de 9 puertos críticos e inseguros (SSH 22, Telnet 23, HTTP 80, HTTPS 443, SMB 445, UPnP 1900, ADB 5555, Redis 6379, MQTT 1883).
- **Base de Datos Externa Optativa (Cloudflare D1 / Upstash Redis / Supabase)**: Activación condicional por UCI para almacenar eventos e historial en la nube sin consumir la memoria RAM del router.
- **Cascada FastPath Cero-LLM**: Respuestas directas e instantáneas (0ms, 0 tokens) para saludos y consultas típicas de red/sistema.
- **Enrutador de 3 Tiers (fast / balanced / deep)**: Presupuestos dinámicos de tokens y ajuste de temperatura (0.2 para comandos, 0.7 para análisis).
- **Canal Telegram bidireccional con Debouncer**: Long-polling (100% compatible con CGNAT) con agrupación inteligente de mensajes e interactividad `doom_loop: ask`.
- **Soporte Multilingüe (i18n)**: Compatible con 9 idiomas (Español, Inglés, Francés, Portugués, Italiano, Ruso, Chino, Japonés, Árabe) con paquetes `.ipk` dedicados.
- **Model Registry & Enrutamiento**: Integración nativa con 9 proveedores comerciales y gratuitos (Groq, OpenRouter, DeepSeek, Gemini, NaraRouter, OpenAI, Mistral, Ollama local, Custom) con modelo fallback automático y detección de modelos en vivo (`/v1/models`).
- **Capa de Conocimiento OpenWrt (`go:embed`)**: Catálogo embebido de programas (`embedded/programs.json`), esquemas UCI (`embedded/uci_schemas.json`), detección runtime de `apk` (25.x) / `opkg` y 10 vías de configuración.
- **Interfaz Web LuCI integrada**: Paquete `luci-app-clawrt` para configuración directa desde el panel de OpenWrt con visor de registros en vivo, métricas de RAM y botones interactivos.

---

## 🏗️ Arquitectura de la Aplicación

```text
clawrt/
├── .github/
│   ├── ISSUE_TEMPLATE/      ← Plantillas estructuradas de reporte de bugs y sugerencias
│   └── workflows/           ← CI/CD (compilación, tests, artifact builds & releases)
├── cmd/
│   └── clawrt/              ← Punto de entrada principal (Daemon)
├── internal/
│   ├── config/              ← Cargador de configuración (UCI / JSON / Env)
│   ├── fastpath/            ← Cascada de respuesta Cero-LLM (Capa 1 Direct, Capa 2 QuickRoute) + Router 3-Tiers
│   ├── hotplug/             ← Procesador de eventos reactivos del sistema (WAN, DHCP, botones)
│   ├── i18n/                ← Motor multilingüe para 9 idiomas (ES, EN, FR, PT, IT, RU, ZH, JA, AR)
│   ├── knowledge/           ← Base de conocimiento OpenWrt con go:embed (programs.json, uci_schemas.json)
│   ├── llm/                 ← Model Registry (9 proveedores) + Enrutador con Modelo Fallback y consulta /v1/models
│   ├── netintel/            ← Inteligencia de red: DHCP leases, OUI MAC, MAC Aleatoria, QR WiFi, Escáner Puertos
│   ├── security/            ← Allowlist, Hard Denylist, Inyección Shell, Secret Scanner, SSRF, Criptografía
│   ├── store/               ← Base de Datos Externa Optativa (Upstash Redis / Cloudflare D1 / Supabase)
│   ├── sys/                 ← Recopilador de métricas y estado del sistema OpenWrt
│   ├── skills/              ← Herramientas (Skills) expuestas al LLM y Telegram con rollback UCI
│   └── telegram/            ← Bot de Telegram + Debouncer de mensajes + doom_loop ask + Circuit Breaker
├── luci-app-clawrt/         ← Interfaz LuCI para OpenWrt (Menú Servicios → ClawRT AI)
│   ├── po/                  ← Catálogos de traducción GETTEXT PO (9 idiomas)
│   ├── root/etc/config/     ← Archivo de configuración UCI por defecto (/etc/config/clawrt)
│   ├── root/etc/hotplug.d/  ← Scripts reactivos del sistema (iface, dhcp, button)
│   ├── root/etc/init.d/     ← Script de servicio procd (/etc/init.d/clawrt)
│   └── htdocs/              ← Vista JavaScript LuCI (JS SPA render)
├── package/                 ← Recetas Makefile para el SDK y compilador oficial de OpenWrt
└── scripts/
    ├── build.sh             ← Compilador multiplataforma (mipsle, armv7, amd64, arm64) + UPX
    ├── make-package.sh      ← Generador de paquetes instalables .ipk (binarios, LuCI e i18n)
    ├── make-feed.sh         ← Generador de Custom Feed (APKINDEX.tar.gz para apk y Packages.gz para opkg)
    └── deploy.sh            ← Script de despliegue directo por SCP/SSH
```

---

## 📜 Licencia

MIT License — Libre y de código abierto. Ver el archivo [`LICENSE`](LICENSE) para más detalles.
