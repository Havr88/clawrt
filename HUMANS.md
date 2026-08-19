# 👤 HUMANS.md — Human Operator & Administrator Guide

This guide is designed for router administrators, human operators, and developers managing **ClawRT**.

---

## ⚡ 0. Hardware Viability & Minimum Requirements

Before deploying ClawRT, verify that your router hardware satisfies the minimum viable thresholds to prevent Out-Of-Memory (OOM) crashes:

- **Bare Minimum Viable**: **8 MB Flash** and **32 MB RAM** (TP-Link WR741ND / WR841N). Requires `TierExtremeMinimal` mode, executing from `/tmp` RAM or `extroot` USB, and using external cloud databases (Supabase / Upstash).
- **Recommended Standard**: **16 MB Flash** and **64 MB RAM** (Xiaomi Mi Router 4C). Fully stable with native LuCI, Telegram bot, and `TierMinimal`.
- **Ideal High Performance**: **32 MB+ Flash** and **128 MB+ RAM** (GL.iNet, Linksys, x86_64). Enables all features, `TierMedium`/`TierFull`, pgvector RAG, and real-time streaming.

---

## 📲 1. Telegram Slash Commands Reference

Interact with ClawRT on Telegram using any of the following built-in commands:

- `/status` (or `/sysinfo`): Router health, CPU load, RAM usage, Uptime, and network overview.
- `/clients` (or `/dhcp`): Active LAN devices, IP addresses, Hostnames, MAC Vendors, and Privacy MAC warnings.
- `/wifi`: Wireless network interfaces, SSIDs, security mode, channel, and connected station count.
- `/qrwifi` (or `/qr`): Generates a scannable ASCII QR code to connect mobile devices to WiFi instantly.
- `/scan`: Scans 9 critical ports on LAN devices for security vulnerabilities.
- `/models` (or `/v1models`): Queries active LLM models from Bynara AI or your configured provider.
- `/firewall`: Displays active firewall zones, forwarding rules, and open ports.
- `/routes`: Displays IPv4/IPv6 routing table and gateway information.
- `/ping [host]`: Tests latency to a host (default: 8.8.8.8).
- `/logs`: Displays the last 20 lines of system logs (`logread`).
- `/memory` (or `/gc`): Displays RAM metrics and manually triggers Go Garbage Collection & `FreeOSMemory()`.
- `/db`: Displays external database status (Supabase Realtime / REST API health).
- `/clear`: Empties the learned facts cache (`/tmp/clawrt_facts.json`).
- `/reboot`: Restarts the ClawRT daemon service.
- `/help`: Displays the interactive multi-language help menu.

---

## 🔑 2. Initial Setup Checklist

1. **Telegram Bot Token & Chat ID**:
   - Create a bot via `@BotFather`.
   - Obtain your Chat ID using `https://api.telegram.org/bot<TOKEN>/getUpdates`.
   - Set in OpenWrt UCI:
     ```sh
     uci set clawrt.telegram.bot_token='<TOKEN>'
     uci add_list clawrt.telegram.chat_id='<CHAT_ID>'
     uci commit clawrt
     service clawrt restart
     ```

2. **Bynara AI / LLM Provider Setup**:
   - Set in OpenWrt UCI:
     ```sh
     uci set clawrt.llm.provider='bynara'
     uci set clawrt.llm.base_url='https://router.bynara.id/v1'
     uci set clawrt.llm.api_key='<BYNARA_API_KEY>'
     uci set clawrt.llm.model='deepseek-v4-flash-free'
     uci set clawrt.llm.fallback_model='agnes-2.5-flash'
     uci commit clawrt
     service clawrt reload
     ```

3. **Supabase Integration Setup**:
   - Create a project on [Supabase](https://supabase.com).
   - Set in OpenWrt UCI:
     ```sh
     uci set clawrt.db.provider='supabase'
     uci set clawrt.db.url='https://your-project.supabase.co'
     uci set clawrt.db.token='Bearer eyJhbGci...'
     uci commit clawrt
     service clawrt reload
     ```

---

## 🔒 4. Seguridad en LuCI sin HTTPS (Servicio uhttpd sobre HTTP)

Muchos routers OpenWrt ejecutan el servidor web LuCI (`uhttpd`) sobre **HTTP sin HTTPS** para ahorrar espacio en memoria Flash y ciclos de CPU en handshakes TLS. ClawRT incorpora **4 protecciones automáticas** para este escenario:

1. **Redacción Automática de Secretos (`SanitizeSecrets`)**:
   - Todas las claves de API (`sk-...`), tokens de Telegram y credenciales de Supabase/Cloudflare son enmascaradas automáticamente en los registros del sistema (`logread`) y en las vistas del panel LuCI como `[SECRETO_REDACTADO]`.
2. **Ejecución Local IPC en el Router**:
   - Las llamadas diagnósticas desde LuCI (`-test-llm`, `-fetch-models`, `-test-telegram`) se ejecutan internamente vía IPC local (`fs.exec`), evitando que las credenciales viajen por paquetes de red en la LAN.
3. **Máscara HTML de Contraseñas**:
   - Los campos de entrada en LuCI (`bot_token`, `api_key`, `db_token`) utilizan protección de contraseña HTML5 `o.password = true`.
4. **Habilitación Opcional de HTTPS Autofirmado en uhttpd**:
   - Si deseas activar HTTPS seguro en tu router OpenWrt sin certificados de pago, ejecuta este comando en SSH:
     ```sh
     opkg update && opkg install px5g-mbedtls uhttpd-mod-ubox && /etc/init.d/uhttpd restart
     ```
     *uhttpd generará automáticamente un certificado TLS autofirmado y abrirá el puerto seguro 443 (`https://192.168.1.1`).*
