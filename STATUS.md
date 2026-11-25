# Piloto Wialon–SENER

**Última actualización**: 2026-11-25
**Features completadas**: 2/3 (`feature_list.json`)
**Pendientes**: aprobación de la prueba integrada con SENER simulado y, en una fase posterior, prueba real con CR-13
**En producción**: no

## Estado actual

El proyecto tiene dos flujos locales. El receptor TCP guarda posiciones en SQLite y permite consultar la última o un intervalo por hora del GPS. El piloto de Remote API tiene su propia cola SQLite de envíos y contempla SENER desarrollo. Ambos se validaron con datos sintéticos.

## Última sesión

El usuario aprobó los seis requisitos del receptor de CR-13. La demostración con el ejecutable recibió una posición ficticia, reinició el proceso, reprodujo la misma trama y recuperó una sola posición con el servicio detenido. El ejemplo usa `state/cr13-demo.db`. La revisión independiente verificó la entrega TCP. [Informe de revisión](progress/review_receptor-tcp-cr13.md) · [Guía para repetir el ensayo](README.md)

## Para retomar la prueba real

El piloto permanece en este equipo. Para la etapa externa quedan por confirmar el UID que retransmite CR-13, la ficha privada y las señales requeridas por SENER, la IP de salida autorizada y la frecuencia operativa. El servicio de Remote API tiene prevista su base de reportes en `/var/lib/wialon-sener/pilot.db` en el VPS. [Datos y despliegue](deploy/README.md) · [Revisión de los dos manuales](docs/frecuencia-sener.md)
