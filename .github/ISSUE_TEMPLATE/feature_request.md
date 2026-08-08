name: 💡 Sugerencia de Funcionalidad (Feature Request)
description: Propón una nueva idea o mejora para ClawRT
title: '[FEAT]: '
labels: ['enhancement']
body:
  - type: textarea
    id: feature-description
    attributes:
      label: Descripción de la Funcionalidad
      description: Describe detalladamente qué te gustaría que haga el agente o la interfaz web.
    validations:
      required: true
  - type: textarea
    id: use-case
    attributes:
      label: Casos de Uso y Beneficio para OpenWrt
      description: ¿Cómo beneficia esta mejora a la gestión del router o a la seguridad de la red?
    validations:
      required: true
