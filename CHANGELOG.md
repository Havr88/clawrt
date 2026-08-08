# 📜 Historial de Cambios (CHANGELOG) — ClawRT

Todas las modificaciones notables a este proyecto serán documentadas en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
