# ⚡ EXTREME_MINIMAL_HARDWARE.md — Guía para Routers de 4MB Flash / 32MB RAM

Esta guía detalla el despliegue de **ClawRT** en routers con especificaciones heredadas extremas (como el **TP-Link TL-WR741ND**, **TL-WR841N v8/v9**, **Atheros AR7240 / AR9331** con **4MB Flash / 32MB RAM** ejecutando OpenWrt 18.06 o 19.07).

---

## 🎯 1. Desafíos de Hardware en 4MB Flash / 32MB RAM

1. **Espacio en Flash (4 MB)**:
   - El sistema operativo OpenWrt (kernel + archivos de sistema) ocupa ~3.5 MB de la partición Flash.
   - **Solución**: No instalar el binario en la memoria Flash permanente. Ejecutar el binario en `/tmp` (memoria RAM volátil) o usar `extroot` (expansión por USB).
2. **Memoria RAM (32 MB)**:
   - El kernel de OpenWrt consumía ~18 MB de RAM al arrancar.
   - **Solución**: Activar **`TierExtremeMinimal`**:
     - Liberación forzada de memoria RAM (`FreeOSMemory()`) tras cada consulta.
     - Reducción del prompt de sistema a ~500 bytes.
     - Offload completo de tablas de hechos y logs hacia **Supabase / Upstash Redis**.

---

## 🚀 2. Métodos de Despliegue Extremo

### 📦 Método A: Ejecución desde `/tmp` (RAM Volátil - Sin Modificar Flash)

Ideal para routers de 4MB Flash sin puerto USB:

1. Edita el archivo `/etc/rc.local` en el router para descargar y ejecutar ClawRT al iniciar:

```sh
# Descargar binario comprimido mips_24kc directamente a RAM
curl -sL http://tu-servidor-local/clawrt-mips -o /tmp/clawrt
chmod +x /tmp/clawrt

# Ejecutar en segundo plano en modo TierExtremeMinimal
/tmp/clawrt -config /etc/config/clawrt &
exit 0
```

---

### 🔌 Método B: Expansión de Almacenamiento por USB (`extroot`)

Si el modelo cuenta con puerto USB (ej. TP-Link TL-WR842N o GL-iNet AR150):

1. Formatea un pendrive USB en `ext4`.
2. Instala `block-mount` y configura `extroot`:
   ```sh
   opkg update && opkg install kmod-usb-storage kmod-fs-ext4 block-mount
   ```
3. Monta el USB como `/` principal. El router pasará de 4 MB a 1 GB+ de memoria Flash y ClawRT se instalará normalmente con `opkg install clawrt_1.0.0-1_mips_24kc.ipk`.

---

## 🛡️ 3. Perfil de Ajuste Fino `TierExtremeMinimal`

Cuando ClawRT detecta 32 MB de RAM o 4 MB de Flash:
- **Límite de herramientas consecutivas**: 3 pasos max.
- **FastPath L1/L2 Prioritario**: Consultas de estado se responden en 0ms y 0 tokens.
- **Garbage Collection Inmediata**: Llama a `debug.FreeOSMemory()` inmediatamente al enviar cada respuesta en Telegram.
