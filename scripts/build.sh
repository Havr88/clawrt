#!/usr/bin/env bash
# scripts/build.sh — Compilador multiplataforma ClawRT para OpenWrt
set -euo pipefail

BUILD_DIR="build"
mkdir -p "${BUILD_DIR}"

echo "╔═══════════════════════════════════════════════════╗"
echo "║      ClawRT AI — Multi-Arch Build Script          ║"
echo "╚═══════════════════════════════════════════════════╝"
echo ""

build_target() {
  local os="$1"
  local arch="$2"
  local arm_or_mips="$3"
  local output_name="$4"

  echo "🔨 Compilando para ${output_name} (${os}/${arch} ${arm_or_mips})..."

  local env_vars=("GOOS=${os}" "GOARCH=${arch}" "CGO_ENABLED=0")
  if [[ "${arch}" == "mipsle" ]]; then
    env_vars+=("GOMIPS=${arm_or_mips}")
  elif [[ "${arch}" == "arm" ]]; then
    env_vars+=("GOARM=${arm_or_mips}")
  fi

  env "${env_vars[@]}" go build \
    -trimpath \
    -ldflags "-s -w -X main.Version=2.0.0-openwrt" \
    -o "${BUILD_DIR}/${output_name}" \
    ./cmd/clawrt

  local raw_size
  raw_size=$(du -h "${BUILD_DIR}/${output_name}" | cut -f1)
  echo "    └─ Tamaño original: ${raw_size}"

  if command -v upx &>/dev/null; then
    echo "    └─ Comprimiendo con UPX..."
    upx --best "${BUILD_DIR}/${output_name}" -q 2>/dev/null || echo "       (UPX skip)"
    local upx_size
    upx_size=$(du -h "${BUILD_DIR}/${output_name}" | cut -f1)
    echo "    └─ Tamaño comprimido: ${upx_size}"
  fi
  echo ""
}

# 1. Target MIPS LE softfloat (Xiaomi Mi Router 4C - ramips/mt76x8)
build_target "linux" "mipsle" "softfloat" "clawrt-mipsle"

# 2. Target ARMv7 (Linksys Velop WHW01 - ipq40xx)
build_target "linux" "arm" "7" "clawrt-armv7"

# 3. Target x86_64 (OpenWrt x86_64 / PC)
build_target "linux" "amd64" "" "clawrt-amd64"

# 4. Target ARM64 (Raspberry Pi 4 / OpenWrt arm64)
build_target "linux" "arm64" "" "clawrt-arm64"

echo "✅ Compilación completada con éxito. Archivos en ${BUILD_DIR}/:"
ls -lh "${BUILD_DIR}/"
