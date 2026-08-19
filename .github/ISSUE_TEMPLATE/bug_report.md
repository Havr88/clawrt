name: 🐛 Reporte de Error (Bug Report)
description: Reporta un error o comportamiento inesperado en ClawRT o LuCI
title: '[BUG]: '
labels: ['bug']
body:
  - type: markdown
    attributes:
      value: Gracias por tomarte el tiempo de reportar un problema en ClawRT.
  - type: textarea
    id: description
    attributes:
      label: Descripción del Problema
      description: Explica claramente qué ocurrió y cuál era el comportamiento esperado.
    validations:
      required: true
  - type: dropdown
    id: openwrt-version
    attributes:
      label: Versión de OpenWrt
      options:
        - 'OpenWrt 25.12'
        - 'OpenWrt 23.05'
        - 'OpenWrt 21.02'
        - 'Otra (Especificar)'
    validations:
      required: true
  - type: input
    id: router-model
    attributes:
      label: Modelo de Router y Arquitectura
      placeholder: Ej: Xiaomi Mi Router 4C (mipsle) / Linksys WHW01 (armv7)
    validations:
      required: true
  - type: textarea
    id: logs
    attributes:
      label: Registros del Sistema (logread -e clawrt)
      description: Pega las últimas líneas del registro del sistema.
      render: text
