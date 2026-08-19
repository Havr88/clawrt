# 📜 Historial de Cambios (CHANGELOG) — ClawRT

Todas las modificaciones notables a este proyecto serán documentadas en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0-openwrt] - 2026-08-18 (Autonomous Network Agent Upgrade)

### 🚀 Añadido
- **Self-Healing Watchdog Engine**: Motor de diagnóstico en cascada en segundo plano (Gateway ➔ DNS Local ➔ DNS Público ➔ Internet WAN) y auto-reparación en caliente (`dnsmasq`, `ifup wan`, `fw4/network`) sin requerir intervención humana.
- **UBUS Reactive Event Listener**: Listener reactivo conectado a `/var/run/ubus/ubus.sock` (`ubus listen`) para capturar eventos de red (`network.interface`), asociaciones WiFi (`hostapd`) y caídas de procesos (`procd`) en tiempo real con cero sobrecarga.
- **LuCI Web Copilot SPA**: Interfaz de chat flotante y conversacional dentro de LuCI (`clawrt/copilot`) con botones de acción rápida, diagnóstico en un clic y soporte de comandos directos (`-query`).
- **Conntrack / NetFlow Guard**: Analizador proactivo de `/proc/net/nf_conntrack` para detectar dispositivos que acaparan el ancho de banda (P2P/Torrents), escaneos de puertos LAN y ataques SYN flood, con capacidad de bloqueo en firewall.
- **Optimizador de Canales WiFi**: Escaneo del espectro inalámbrico en 2.4 GHz y 5 GHz (`iwinfo scan`), cálculo de congestión e interferencia de redes vecinas y recomendación/aplicación del canal más limpio.
- **Gestor de SQM / Bufferbloat QoS**: Diagnóstico y auto-ajuste de disciplinas de cola Cake / FQ_Codel con limitación de ancho de banda para erradicar el retardo y jitter.
- **Intent-Based Multi-step Config Engine**: Traductor de intenciones declarativas de alto nivel (red WiFi de invitados aislada, redirección de puertos, control de acceso por horario, reservas DHCP estáticas) con validación atómica y rollback automático.
- **Hotplug Inteligente Reactivo**: Integración de `/etc/hotplug.d/` con el motor de auto-sanación autónomo para mitigar caídas de WAN de inmediato.

---

## [1.0.0-openwrt] - 2026-08-08 (Release Inicial Oficial v1.0.0)

### 🚀 Añadido
- **Capa de Seguridad ACP & Hard Denylist**: Bloqueo absoluto de comandos destructivos (`reboot`, `rm -rf /`, `dd`, `sysupgrade`, `kill -9`, `echo > /dev/watchdog`).
- **Allowlist Estricta & Inyección de Shell**: Validación de binarios permitidos (`uci`, `ip`, `nft`, `service`, `ubus`, `logread`, `ping`, etc.) y filtro contra `;`, `&&`, `||`, backticks y tuberías peligrosas.
- **Herramientas UCI Tipadas con Rollback Automático**: Creación de snapshots pre-ejecución y validación `fw4 check` con reversión automática ante errores.
- **Alertas Reactivas del Sistema (`/etc/hotplug.d/`)**: Notificación proactiva por Telegram ante caída/restauración WAN (`iface`), conexiones/desconexiones de nuevos clientes LAN (`dhcp`) y botones físicos (`button`).
- **Escáner de Secretos en Tiempo Real & SSRF**: Redacción de tokens de Telegram, API keys y contraseñas (`[SECRETO_REDACTADO]`) y bloqueo de IP metadata `169.254.169.254`.
- **Inteligencia de Clientes DHCP (`/leases`)**: OUI MAC, detección de MAC Aleatoria privada (iOS/Android Privacy), señal RSSI, huella de SO y código QR WiFi (`/qrwifi`).
- **Model Registry & Auto-completado LuCI**: Soporte nativo para 9 proveedores LLM con auto-completado de URL base y modelos recomendados.
- **Detección de Modelos en Vivo**: Botón e interacción CLI `-fetch-models` para consultar el endpoint `/v1/models` en vivo.
- **Paquetes de Traducción LuCI (i18n)**: Catálogos GETTEXT PO y paquetes instaladores `.ipk` para 9 idiomas (Español, Inglés, Francés, Portugués, Italiano, Ruso, Chino Simplificado, Japonés, Árabe).
- **Custom Feed Remoto de Paquetes**: Generación de índices compatibles con `apk` (OpenWrt 25.12+) y `opkg` (OpenWrt <= 23.05).
- **Recetas de Empaquetado para OpenWrt**: Compatibilidad nativa con SDK de OpenWrt (`package/clawrt/Makefile` y `package/luci-app-clawrt/Makefile`).
- **Integración Continua & Publicación Híbrida (GitHub Actions)**: Workflows de CI (`ci.yml`), artefactos de feed (`feed.yml`) y publicación automática de Releases (`release.yml`).
