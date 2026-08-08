#!/usr/bin/env bash
# scripts/make-feed.sh — Genera el Custom Feed de paquetes para OpenWrt (apk & opkg)
set -euo pipefail

VERSION="1.0.0"
DIST_DIR="dist"
FEED_DIR="${DIST_DIR}/feed"

mkdir -p "${FEED_DIR}"

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║  Generador de Feed de Paquetes OpenWrt (apk 25.x & opkg)  ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Asegurar que los paquetes estén construidos
if [[ ! -f "${DIST_DIR}/clawrt_${VERSION}-1_mipsel_24kc.ipk" ]]; then
  echo "📦 Generando paquetes instalables..."
  bash scripts/make-package.sh
fi

build_arch_feed() {
  local arch="$1"
  local target_dir="${FEED_DIR}/${arch}"
  mkdir -p "${target_dir}"

  echo "🌐 Creando índice de feed para arquitectura: ${arch}..."

  # Copiar paquetes correspondientes a la carpeta de arquitectura
  if [[ "${arch}" == "all" ]]; then
    cp "${DIST_DIR}"/*_all.ipk "${target_dir}/" 2>/dev/null || true
  else
    cp "${DIST_DIR}"/*_${arch}.ipk "${target_dir}/" 2>/dev/null || true
    cp "${DIST_DIR}"/*_all.ipk "${target_dir}/" 2>/dev/null || true
  fi

  # 1. Generar Packages & Packages.gz para opkg (OpenWrt <= 23.05)
  local pkg_file="${target_dir}/Packages"
  : > "${pkg_file}"

  for ipk in "${target_dir}"/*.ipk; do
    [[ -f "${ipk}" ]] || continue
    local tmp_extract
    tmp_extract=$(mktemp -d)
    tar -xzf "${ipk}" -C "${tmp_extract}" control.tar.gz 2>/dev/null || true
    if [[ -f "${tmp_extract}/control.tar.gz" ]]; then
      tar -xzf "${tmp_extract}/control.tar.gz" -C "${tmp_extract}" control 2>/dev/null || true
      if [[ -f "${tmp_extract}/control" ]]; then
        cat "${tmp_extract}/control" >> "${pkg_file}"
        echo "Filename: $(basename "${ipk}")" >> "${pkg_file}"
        echo "Size: $(stat -c%s "${ipk}")" >> "${pkg_file}"
        echo "SHA256sum: $(sha256sum "${ipk}" | awk '{print $1}')" >> "${pkg_file}"
        echo "" >> "${pkg_file}"
      fi
    fi
    rm -rf "${tmp_extract}"
  done

  gzip -9c "${pkg_file}" > "${target_dir}/Packages.gz"

  # 2. Generar APKINDEX.tar.gz para apk (OpenWrt 25.12+)
  local apk_dir
  apk_dir=$(mktemp -d)
  cp "${pkg_file}" "${apk_dir}/APKINDEX"
  (cd "${apk_dir}" && tar -czf "${OLDPWD}/${target_dir}/APKINDEX.tar.gz" APKINDEX)
  rm -rf "${apk_dir}"

  echo "  ✅ Índices generados para ${arch}: Packages.gz (opkg) y APKINDEX.tar.gz (apk)"
}

build_arch_feed "mipsel_24kc"
build_arch_feed "arm_cortex-a7_neon-vfpv4"
build_arch_feed "x86_64"
build_arch_feed "aarch64_cortex-a53"

# Generar index.html para la navegación del Feed en GitHub Pages
cat << 'EOF' > "${FEED_DIR}/index.html"
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>ClawRT AI — Repositorio Oficial de Paquetes para OpenWrt</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background: #0f172a; color: #f8fafc; padding: 40px; max-width: 900px; margin: 0 auto; }
        h1 { color: #38bdf8; }
        code { background: #1e293b; color: #f1f5f9; padding: 12px; display: block; border-radius: 8px; border-left: 4px solid #38bdf8; font-family: monospace; white-space: pre-wrap; }
        .badge { background: #0284c7; color: white; padding: 4px 8px; border-radius: 4px; font-size: 0.85em; font-weight: bold; }
        a { color: #38bdf8; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>🐾 ClawRT AI — Feed de Paquetes OpenWrt</h1>
    <p>Repositorio de paquetes compatible con <b>apk</b> (OpenWrt 25.12+) y <b>opkg</b> (OpenWrt 23.05 y anteriores).</p>

    <h2>📦 Configuración en OpenWrt 25.12+ (apk)</h2>
    <p>Agrega el repositorio a tu lista de fuentes ejecutando en el router:</p>
    <code>echo "https://havr88.github.io/clawrt/feed/mipsel_24kc" >> /etc/apk/repositories.d/clawrt.list
apk update
apk add clawrt luci-app-clawrt luci-i18n-clawrt-es</code>

    <h2>📦 Configuración en OpenWrt 23.05 / 21.02 (opkg)</h2>
    <code>echo "src/gz clawrt https://havr88.github.io/clawrt/feed/mipsel_24kc" >> /etc/opkg/customfeeds.conf
opkg update
opkg install clawrt luci-app-clawrt luci-i18n-clawrt-es</code>

    <h2>📂 Arquitecturas Disponibles</h2>
    <ul>
        <li><a href="mipsel_24kc/">mipsel_24kc (MIPS Little Endian — ej: Xiaomi Mi Router 4C)</a></li>
        <li><a href="arm_cortex-a7_neon-vfpv4/">arm_cortex-a7_neon-vfpv4 (ARMv7 — ej: Linksys Velop WHW01)</a></li>
        <li><a href="aarch64_cortex-a53/">aarch64_cortex-a53 (ARM64 — Raspberry Pi / Routers 64-bit)</a></li>
        <li><a href="x86_64/">x86_64 (x86 64-bit — PCs / Mini PCs / QEMU / Proxmox)</a></li>
    </ul>
</body>
</html>
EOF

echo ""
echo "🎉 Custom Feed generado con éxito en ${FEED_DIR}/"
