#!/usr/bin/env bash
# scripts/make-package.sh — Genera paquetes instalables (.ipk y .apk) para OpenWrt
set -euo pipefail

VERSION="1.0.0"
DIST_DIR="dist"

mkdir -p "${DIST_DIR}"

echo "╔═══════════════════════════════════════════════════╗"
echo "║   Empaquetador de ClawRT para OpenWrt (apk/ipk)   ║"
echo "╚═══════════════════════════════════════════════════╝"
echo ""

# Asegurar que existan los binarios compilados
if [[ ! -f "build/clawrt-mipsle" ]]; then
  echo "🔨 Compilando binarios primero..."
  bash scripts/build.sh
fi

build_ipk() {
  local arch="$1"
  local binary_name="$2"
  local pkg_arch="$3"

  local work_dir
  work_dir=$(mktemp -d)
  local data_dir="${work_dir}/data"
  local control_dir="${work_dir}/control"

  mkdir -p "${data_dir}/usr/bin"
  mkdir -p "${data_dir}/etc/config"
  mkdir -p "${data_dir}/etc/init.d"
  mkdir -p "${control_dir}"

  # Archivos de datos
  cp "build/${binary_name}" "${data_dir}/usr/bin/clawrt"
  chmod +x "${data_dir}/usr/bin/clawrt"
  cp "luci-app-clawrt/root/etc/config/clawrt" "${data_dir}/etc/config/clawrt"
  cp "luci-app-clawrt/root/etc/init.d/clawrt" "${data_dir}/etc/init.d/clawrt"
  chmod +x "${data_dir}/etc/init.d/clawrt"

  # Archivo control
  cat << EOF > "${control_dir}/control"
Package: clawrt
Version: ${VERSION}-1
Architecture: ${pkg_arch}
Maintainer: ClawRT Contributors
Description: ClawRT AI Router Agent for OpenWrt
Depends: libc
EOF

  # Archivo conffiles (no sobrescribir config en actualizaciones)
  cat << EOF > "${control_dir}/conffiles"
/etc/config/clawrt
EOF

  # Crear debian-binary
  echo "2.0" > "${work_dir}/debian-binary"

  # Empaquetar tar.gz
  (cd "${data_dir}" && tar -czf "${work_dir}/data.tar.gz" .)
  (cd "${control_dir}" && tar -czf "${work_dir}/control.tar.gz" .)

  # Crear paquete .ipk
  local ipk_file="${DIST_DIR}/clawrt_${VERSION}-1_${pkg_arch}.ipk"
  (cd "${work_dir}" && tar -czf "${OLDPWD}/${ipk_file}" debian-binary control.tar.gz data.tar.gz)

  echo "  ✅ Paquete OPKG generado: ${ipk_file}"
  rm -rf "${work_dir}"
}

build_luci_ipk() {
  local work_dir
  work_dir=$(mktemp -d)
  local data_dir="${work_dir}/data"
  local control_dir="${work_dir}/control"

  mkdir -p "${data_dir}/usr/share/luci/menu.d"
  mkdir -p "${data_dir}/usr/share/rpcd/acl.d"
  mkdir -p "${data_dir}/www/luci-static/resources/view/clawrt"
  mkdir -p "${control_dir}"

  cp luci-app-clawrt/root/usr/share/luci/menu.d/luci-app-clawrt.json "${data_dir}/usr/share/luci/menu.d/"
  cp luci-app-clawrt/root/usr/share/rpcd/acl.d/luci-app-clawrt.json "${data_dir}/usr/share/rpcd/acl.d/"
  cp luci-app-clawrt/htdocs/luci-static/resources/view/clawrt/config.js "${data_dir}/www/luci-static/resources/view/clawrt/"

  cat << EOF > "${control_dir}/control"
Package: luci-app-clawrt
Version: ${VERSION}-1
Architecture: all
Maintainer: ClawRT Contributors
Description: LuCI support for ClawRT AI Router Agent
Depends: clawrt
EOF

  echo "2.0" > "${work_dir}/debian-binary"
  (cd "${data_dir}" && tar -czf "${work_dir}/data.tar.gz" .)
  (cd "${control_dir}" && tar -czf "${work_dir}/control.tar.gz" .)

  local ipk_file="${DIST_DIR}/luci-app-clawrt_${VERSION}-1_all.ipk"
  (cd "${work_dir}" && tar -czf "${OLDPWD}/${ipk_file}" debian-binary control.tar.gz data.tar.gz)

  echo "  ✅ Paquete OPKG generado: ${ipk_file}"
  rm -rf "${work_dir}"
}

build_luci_i18n_ipk() {
  local lang="$1"
  local po_lang="$2"

  local work_dir
  work_dir=$(mktemp -d)
  local data_dir="${work_dir}/data"
  local control_dir="${work_dir}/control"

  mkdir -p "${data_dir}/usr/share/luci/i18n"
  mkdir -p "${control_dir}"

  if [[ -f "luci-app-clawrt/po/${po_lang}/clawrt.po" ]]; then
    cp "luci-app-clawrt/po/${po_lang}/clawrt.po" "${data_dir}/usr/share/luci/i18n/clawrt.${lang}.po"
  fi

  cat << EOF > "${control_dir}/control"
Package: luci-i18n-clawrt-${lang}
Version: ${VERSION}-1
Architecture: all
Maintainer: ClawRT Contributors
Description: Translation package for luci-app-clawrt (${lang})
Depends: luci-app-clawrt
EOF

  echo "2.0" > "${work_dir}/debian-binary"
  (cd "${data_dir}" && tar -czf "${work_dir}/data.tar.gz" .)
  (cd "${control_dir}" && tar -czf "${work_dir}/control.tar.gz" .)

  local ipk_file="${DIST_DIR}/luci-i18n-clawrt-${lang}_${VERSION}-1_all.ipk"
  (cd "${work_dir}" && tar -czf "${OLDPWD}/${ipk_file}" debian-binary control.tar.gz data.tar.gz)

  echo "  🌐 Paquete i18n generado: ${ipk_file}"
  rm -rf "${work_dir}"
}

echo "📦 Generando paquetes OPKG (.ipk) para arquitecturas soportadas..."
build_ipk "mipsle" "clawrt-mipsle" "mipsel_24kc"
build_ipk "armv7" "clawrt-armv7" "arm_cortex-a7_neon-vfpv4"
build_ipk "amd64" "clawrt-amd64" "x86_64"
build_ipk "arm64" "clawrt-arm64" "aarch64_cortex-a53"
build_luci_ipk

echo ""
echo "🌐 Generando paquetes de traducción (.ipk i18n) para LuCI..."
build_luci_i18n_ipk "es" "es"
build_luci_i18n_ipk "en" "en"
build_luci_i18n_ipk "fr" "fr"
build_luci_i18n_ipk "pt" "pt"
build_luci_i18n_ipk "it" "it"
build_luci_i18n_ipk "ru" "ru"
build_luci_i18n_ipk "zh-cn" "zh_Hans"
build_luci_i18n_ipk "ja" "ja"
build_luci_i18n_ipk "ar" "ar"

echo ""
echo "🎉 Todos los paquetes instalables están listos en la carpeta ${DIST_DIR}/:"
ls -lh "${DIST_DIR}/"
