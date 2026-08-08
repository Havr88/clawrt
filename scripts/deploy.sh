#!/usr/bin/env bash
# scripts/deploy.sh — Script de despliegue directo a router OpenWrt
set -euo pipefail

ROUTER_IP="${ROUTER_IP:-192.168.31.1}"
ROUTER_USER="${ROUTER_USER:-root}"
TARGET_ARCH="${TARGET_ARCH:-mipsle}"

BINARY="build/clawrt-${TARGET_ARCH}"

if [[ ! -f "${BINARY}" ]]; then
  echo "❌ El binario ${BINARY} no existe. Ejecuta primero bash scripts/build.sh"
  exit 1
fi

echo "🚀 Desplegando ClawRT (${TARGET_ARCH}) en ${ROUTER_USER}@${ROUTER_IP}..."

# 1. Copiar binario
echo "[1/4] Copiando binario a /usr/bin/clawrt..."
ssh "${ROUTER_USER}@${ROUTER_IP}" "/etc/init.d/clawrt stop 2>/dev/null || true"
scp -O "${BINARY}" "${ROUTER_USER}@${ROUTER_IP}:/usr/bin/clawrt"
ssh "${ROUTER_USER}@${ROUTER_IP}" "chmod +x /usr/bin/clawrt"

# 2. Copiar archivos de LuCI app
echo "[2/4] Copiando archivos de interfaz LuCI..."
scp -O luci-app-clawrt/root/etc/config/clawrt "${ROUTER_USER}@${ROUTER_IP}:/etc/config/clawrt" 2>/dev/null || true
scp -O luci-app-clawrt/root/etc/init.d/clawrt "${ROUTER_USER}@${ROUTER_IP}:/etc/init.d/clawrt"
ssh "${ROUTER_USER}@${ROUTER_IP}" "chmod +x /etc/init.d/clawrt"

ssh "${ROUTER_USER}@${ROUTER_IP}" "mkdir -p /usr/share/luci/menu.d /usr/share/rpcd/acl.d /www/luci-static/resources/view/clawrt /etc/hotplug.d/iface /etc/hotplug.d/dhcp /etc/hotplug.d/button"
scp -O luci-app-clawrt/root/usr/share/luci/menu.d/luci-app-clawrt.json "${ROUTER_USER}@${ROUTER_IP}:/usr/share/luci/menu.d/"
scp -O luci-app-clawrt/root/usr/share/rpcd/acl.d/luci-app-clawrt.json "${ROUTER_USER}@${ROUTER_IP}:/usr/share/rpcd/acl.d/"
scp -O luci-app-clawrt/htdocs/luci-static/resources/view/clawrt/config.js "${ROUTER_USER}@${ROUTER_IP}:/www/luci-static/resources/view/clawrt/"

echo "[3/4] Instalando scripts reactivos hotplug en /etc/hotplug.d/..."
scp -O luci-app-clawrt/root/etc/hotplug.d/iface/99-clawrt "${ROUTER_USER}@${ROUTER_IP}:/etc/hotplug.d/iface/99-clawrt"
scp -O luci-app-clawrt/root/etc/hotplug.d/dhcp/99-clawrt "${ROUTER_USER}@${ROUTER_IP}:/etc/hotplug.d/dhcp/99-clawrt"
scp -O luci-app-clawrt/root/etc/hotplug.d/button/99-clawrt "${ROUTER_USER}@${ROUTER_IP}:/etc/hotplug.d/button/99-clawrt"
ssh "${ROUTER_USER}@${ROUTER_IP}" "chmod +x /etc/hotplug.d/iface/99-clawrt /etc/hotplug.d/dhcp/99-clawrt /etc/hotplug.d/button/99-clawrt"

# 3. Reiniciar rpcd y servicio
echo "[3/4] Recargando servicios LuCI y rpcd..."
ssh "${ROUTER_USER}@${ROUTER_IP}" "/etc/init.d/rpcd restart; /etc/init.d/clawrt enable; /etc/init.d/clawrt restart"

echo "✅ Despliegue completado. Accede al router en http://${ROUTER_IP}/ para configurar ClawRT en Servicios -> ClawRT AI."
