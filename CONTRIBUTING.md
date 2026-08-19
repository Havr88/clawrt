# 🤝 Guía de Contribución a ClawRT

¡Gracias por tu interés en contribuir a **ClawRT**! Este proyecto es libre y de código abierto para toda la comunidad de OpenWrt.

---

## 🛠️ Cómo Empezar

1. **Haz un Fork del Repositorio** en GitHub.
2. **Clona tu Fork localmente**:
   ```bash
   git clone https://github.com/TU_USUARIO/clawrt.git
   cd clawrt
   ```
3. **Crea una Rama para tu Funcionalidad o Corrección**:
   ```bash
   git checkout -b feat/nueva-funcionalidad
   ```

---

## 🧪 Pruebas y Compilación

Antes de enviar un Pull Request, asegúrate de que todos los tests unitarios pasen y que los binarios compilen correctamente:

```bash
# Ejecutar tests unitarios
go test -v ./...

# Verificar formato e inspección de código
go vet ./...

# Probar compilación multiplataforma
bash scripts/build.sh
```

---

## 📝 Reglas de Código y Seguridad

- **Mantener bajo consumo de memoria**: Evita librerías pesadas en Go que incrementen el tamaño del ejecutable (mantenlo comprimido por debajo de 3 MB con UPX).
- **Seguridad ante todo**: Toda nueva herramienta debe pasar por la validación de `security.ValidateCommandSafety` o el filtro de rutas.
- **Traducción**: Si añades textos en la interfaz LuCI, actualiza los archivos de plantilla `.po` en `luci-app-clawrt/po/`.
