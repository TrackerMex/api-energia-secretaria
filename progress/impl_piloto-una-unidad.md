# Entrega local del piloto

El usuario aprobó los requisitos para CR-13 y pidió continuar únicamente en local. La prueba de punta a punta usa una unidad ficticia, servidores TLS de loopback y SQLite temporal.

## Archivos y cobertura

- `internal/ingest/` implementa la lectura paginada, autenticación e identificación de posiciones de R1. El dominio valida la ficha y transforma las posiciones para R2 y R3.
- `internal/delivery/` implementa los adaptadores SQLite y SENER para R4–R6. El caso de uso de ventana en internal/application coordina el origen y la entrega mediante los contratos de internal/domain.
- `cmd/piloto/` integra el flujo y ofrece pruebas repetibles de R7.
- `deploy/` describe una futura instalación en el VPS. No se ha instalado el servicio.

La evidencia por requisito está en [trazabilidad](../specs/piloto-una-unidad/traceability.md). El reporte final de revisión recoge las verificaciones y los límites de la prueba.

La verificación completa del implementer sobre c3a1ca0 terminó con código 0. El build tardó 1 s y las pruebas de los cuatro paquetes, 5 s. Go vet pasó. R5 comprueba que un 401 pausa toda la cola y conserva el UUID del próximo intento. R6 comprueba que una lectura inválida conserva su intervalo de origen después de un reinicio y puede corregirse con el mismo UUID.

Se corrigió el checkpoint C3 en 5ba50b0. Go list muestra que el caso de uso depende del dominio y de la biblioteca estándar. Los adaptadores Wialon, SQLite y SENER entran por los contratos de dominio y la prueba R7 los ejercita con dos servidores TLS locales. El init integrado de esta revisión terminó con código 0, build en 2 s, suite en 3 s y go vet correcto.
