# Arquitectura del piloto

El comando Go sirve a una sola unidad. En desarrollo consulta Wialon y envía a SENER mediante HTTPS saliente. La prueba local usa dos servidores simulados y una base temporal.

| Código | Responsabilidad |
|---|---|
| `internal/ingest/source.go` | Autenticación, consulta de una ventana, lectura paginada y cierre de la sesión Wialon. |
| `internal/ingest/report.go` | Transformación y validación del mensaje, la ficha y los catálogos SENER. La función `Report` opera con datos en memoria. |
| `internal/delivery/` | Almacenamiento SQLite, transición de estados de entrega y preparación de multipart para SENER desarrollo. |
| `cmd/piloto/` | Selección de modo, configuración de una unidad y coordinación del recorrido completo. |

El cursor y la cola se actualizan en la misma transacción. Un reporte aceptado requiere HTTP 200 con `error: 0`. Antes del envío se conserva el UUID y se registra la intención de entrega. Si se pierde la respuesta, la transmisión queda identificada para conciliación.

[Formato Wialon](https://help.wialon.com/en/api/user-guide/api-reference/messages/load_interval) · [Manuales de SENER revisados](frecuencia-sener.md)
