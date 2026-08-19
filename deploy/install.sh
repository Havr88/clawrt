#!/bin/sh
# deploy/install.sh — Script de instalación universal de ClawRT para OpenWrt (apk / opkg / standalone)
set -e

echo "╔═══════════════════════════════════════════════════╗"
echo "║     Instalador Universal de ClawRT para OpenWrt  ║"
echo "╚═══════════════════════════════════════════════════╝"
echo ""

# 1. Detectar gestor de paquetes (apk vs opkg)
PKG_MGR="unknown"
if command -v apk >/dev/null 2>&1; then
    PKG_MGR="apk"
    echo "📌 Gestor de paquetes detectado: apk (OpenWrt 25.x+)"
elif command -v opkg >/dev/null 2>&1; then
    PKG_MGR="opkg"
    echo "📌 Gestor de paquetes detectado: opkg (OpenWrt 23.x / 24.x)"
else
    echo "⚠️ No se detectó apk ni opkg. Realizando instalación manual..."
fi

# 2. Instalar binario ejecutable
if [ -f "/tmp/clawrt" ]; then
    echo "[1/4] Instalando binario ClawRT en /usr/bin/clawrt..."
    mv /tmp/clawrt /usr/bin/clawrt
    chmod +x /usr/bin/clawrt
elif [ ! -f "/usr/bin/clawrt" ]; then
    echo "❌ Error: Binario /tmp/clawrt no encontrado."
    exit 1
fi

# 3. Configuración UCI por defecto
if [ ! -f "/etc/config/clawrt" ]; then
    echo "[2/4] Creando archivo de configuración UCI en /etc/config/clawrt..."
    mkdir -p /etc/config
    cat << 'EOF' > /etc/config/clawrt
config core 'main'
	option enabled '1'

config telegram 'telegram'
	option bot_token ''

config llm 'llm'
	option base_url 'https://router.bynara.id/v1'
	option api_key ''
	option model 'gpt-4o-mini'
	option max_iterations '5'
EOF
fi

# 4. Servicio procd (/etc/init.d/clawrt)
echo "[3/4] Instalando servicio procd en /etc/init.d/clawrt..."
cat << 'EOF' > /etc/init.d/clawrt
#!/bin/sh /etc/rc.common

START=99
STOP=10
USE_PROCD=1

PROG=/usr/bin/clawrt

start_service() {
	config_load clawrt

	local enabled
	config_get_bool enabled main enabled 0

	[ "$enabled" -eq 1 ] || return 0

	procd_open_instance clawrt
	procd_set_param command "$PROG" -config /etc/config/clawrt
	procd_set_param respawn 3600 5 0
	procd_set_param stdout 1
	procd_set_param stderr 1
	procd_close_instance
}

reload_service() {
	stop
	start
}
EOF
chmod +x /etc/init.d/clawrt

# 5. Instalar archivos de interfaz web LuCI si están presentes en /tmp
if [ -d "/tmp/luci-app-clawrt" ] || [ -f "/tmp/config.js" ]; then
    echo "[4/4] Instalando interfaz Web LuCI (luci-app-clawrt)..."
    mkdir -p /usr/share/luci/menu.d /usr/share/rpcd/acl.d /www/luci-static/resources/view/clawrt

    if [ -f "/tmp/luci-app-clawrt.json" ]; then
        cp /tmp/luci-app-clawrt.json /usr/share/luci/menu.d/
    fi
    if [ -f "/tmp/acl-clawrt.json" ]; then
        cp /tmp/acl-clawrt.json /usr/share/rpcd/acl.d/luci-app-clawrt.json
    fi
    if [ -f "/tmp/config.js" ]; then
        cp /tmp/config.js /www/luci-static/resources/view/clawrt/
    fi

    /etc/init.d/rpcd restart 2>/dev/null || true
fi

# 6. Habilitar e iniciar servicio
/etc/init.d/clawrt enable 2>/dev/null || true
/etc/init.d/clawrt restart 2>/dev/null || true

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  ✅ ClawRT se ha instalado y activado correctamente."
echo ""
echo "  Configura el Bot Token de Telegram y la API Key en el"
echo "  panel de LuCI (Servicios -> ClawRT AI) o en /etc/config/clawrt"
echo "═══════════════════════════════════════════════════════════"
