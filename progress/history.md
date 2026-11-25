# Historial de sesiones

## 2026-11-18

Se inició el repositorio a partir del harness `33ffd92` y se revisaron ambos manuales de SENER. No especifican una periodicidad de reporte de posiciones. El usuario eligió CR-13 y aprobó los requisitos de la prueba con una sola unidad.

Se implementó el cliente de Wialon Remote API, la validación de posiciones, el almacenamiento SQLite y el emisor para SENER desarrollo. Las pruebas usan datos sintéticos. La revisión independiente aprobó `5ba50b0` en local. El usuario difirió la conexión operativa al VPS y los envíos con datos reales.

## 2026-11-18 · Receptor TCP e historial de CR-13

El usuario eligió la recepción local de Wialon Retranslator 1.0 y aprobó los seis requisitos. Se implementaron el decodificador y la escucha TCP en Go, el filtro por UID, el historial en SQLite y las consultas por hora del GPS. La demostración dejó una posición sintética en `state/cr13-demo.db` y comprobó su persistencia y deduplicación entre dos ejecuciones del binario.

La revisión independiente aprobó la entrega local. [Trazabilidad R1–R6](../specs/receptor-tcp-cr13/traceability.md) · [Revisión](review_receptor-tcp-cr13.md).

## 2026-11-25 · Prueba local de TCP a SENER

El usuario aprobó el ensayo completo para una unidad ficticia. Se implementaron la captura de señales históricas y la conexión del historial SQLite con el emisor multipart local. La prueba comprobó una sola aceptación y un solo POST al reiniciar y repetir una posición. En un ensayo con dos ejecutables y un simulador independiente, una segunda posición sin señales quedó guardada con su motivo y sin otro envío.

La primera revisión identificó que una retransmisión debía poder completar señales antes desconocidas. La corrección conserva las mediciones previas y el UUID. La revisión independiente aprobó la entrega corregida `819d5ca`. Los detalles de R1–R5 se conservan en la [trazabilidad](../specs/tcp-sener-local/traceability.md) y el [informe de revisión](review_tcp-sener-local.md).
