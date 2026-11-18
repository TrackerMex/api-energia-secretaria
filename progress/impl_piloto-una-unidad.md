# Entrega local del piloto

El usuario aprobó los requisitos para CR-13 y pidió continuar únicamente en local. La prueba de punta a punta usa una unidad ficticia, servidores TLS de loopback y SQLite temporal.

## Archivos y cobertura

- `internal/ingest/` implementa R1–R3 con autenticación, lectura paginada, identificación de la lectura, validación del perfil y transformación al contrato SENER.
- `internal/delivery/` implementa R4–R6 con multipart, aceptación explícita, deduplicación local, persistencia de la ventana y resultados pendientes de corrección o conciliación.
- `cmd/piloto/` integra el flujo y ofrece pruebas repetibles de R7.
- `deploy/` describe una futura instalación en el VPS. No se ha instalado el servicio.

La evidencia por requisito está en [trazabilidad](../specs/piloto-una-unidad/traceability.md). El reporte final de revisión recoge las verificaciones y los límites de la prueba.
