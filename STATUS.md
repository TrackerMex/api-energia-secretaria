# Piloto Wialon–SENER

**Última actualización**: 2026-11-25
**Features completadas**: 4/4 (`feature_list.json`)
**Pendientes**: validación de la conectividad histórica del GPS antes de una prueba con SENER desarrollo
**En producción**: no

## Estado actual

El piloto local recibe posiciones por TCP, guarda las señales históricas en SQLite y puede enviar un intervalo al simulador de SENER. En una retransmisión exacta ya aceptada conserva un único reporte. También cuenta con el flujo de lectura Remote API y con un comando independiente para observar el motor de CR-13 sin entregas externas.

## Última sesión

El usuario aprobó los cuatro requisitos de observación de la API de Wialon. `cmd/observar` contrasta la entrada digital 1 con las RPM del mismo mensaje, resume las clases de motor y explica los casos indeterminados. La prueba de extremo a extremo con Wialon simulado produjo una posición encendida, una apagada y dos indeterminadas. La consulta vacía y el fallo de lectura tienen resultados distintos. [Guía para ejecutar el comando](README.md) · [Revisión independiente](progress/review_observacion-api-cr13.md)

## Para retomar la prueba con SENER

El conector de Wialon permitió revisar la telemetría de la unidad en modo de lectura. Se ejecutó el observador en Go con la credencial privada del proyecto y el ID interno de CR-13. La consulta real del 24 al 25 de noviembre se contrastó con los campos crudos de la misma ventana. Los conteos agregados se guardaron en la carpeta local state/. El UID del retransmisor real permanece en la configuración local privada. La conectividad GPS seguirá pendiente hasta comprobar una señal explícita en el mensaje histórico. [Criterio para CR-13](internal/ingest/README.md#criterio-de-observación-para-cr-13)

El despliegue en el VPS, el acceso a SENER desarrollo, la confirmación de la frecuencia operativa y la ampliación a las once unidades corresponden a la fase externa. [Estado de la integración TCP](progress/review_tcp-sener-local.md) · [Revisión de los manuales](docs/frecuencia-sener.md)
