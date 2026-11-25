# Arquitectura del piloto

La prueba local recorre Wialon simulado, el caso de uso de una unidad, SQLite temporal y SENER simulado. El comando de desarrollo compone los mismos adaptadores para conexiones salientes HTTPS.

| Código | Responsabilidad |
|---|---|
| `internal/domain/ports.go` | Tipos de posición y entrega. Contratos `PositionSource`, `Outbox` y `ReportSender`. |
| `internal/domain/report.go` | Ficha de unidad, validación de señales y catálogos, transformación de una posición al JSON de SENER en memoria. |
| `internal/application/pilot.go` | `RunWindow` lee la unidad, prepara los reportes con identidades duraderas y ordena el envío. `DeliverPending` procesa los resultados mediante los contratos de dominio. |
| `internal/ingest/source.go` | Adaptador HTTP de Wialon Remote API con autenticación, paginación y cierre de sesión. |
| `internal/delivery/store.go` | Adaptador SQLite que actualiza el cursor y la cola en una transacción y persiste el resultado de cada intento. |
| `internal/delivery/sener.go` | Adaptador multipart HTTPS que traduce la respuesta de SENER a un resultado de entrega. |
| `cmd/piloto/` | Carga el perfil privado, configura el modo, compone los adaptadores y programa las ventanas UTC. |

La aplicación importa únicamente el dominio y la biblioteca estándar de Go. Los adaptadores implementan los contratos del dominio. El piloto conserva el UUID antes del POST. Un HTTP 200 con `error: 0` confirma la aceptación. Una entrega de resultado incierto queda pendiente de conciliación.

[Formato Wialon](https://help.wialon.com/en/api/user-guide/api-reference/messages/load_interval) · [Manuales de SENER revisados](frecuencia-sener.md)

## Recepción local por TCP

Una conexión TCP entrega tramas de Wialon Retranslator 1.0 al adaptador `internal/retranslator`. El caso de uso en `internal/application` selecciona el UID configurado. El adaptador `internal/historydb` persiste cada posición aceptada en el archivo SQLite del receptor. El comando `cmd/receptor` conecta estas piezas y sirve también para consultar la última posición y el historial. La cola de entrega a SENER conserva su propio ciclo de vida.

La hora del GPS determina qué posición es la más reciente. Una huella de la trama completa identifica las retransmisiones exactas, incluso tras un reinicio. [Diseño y requisitos de la fase TCP](../specs/receptor-tcp-cr13/design.md)

## Del historial TCP al simulador de SENER

El receptor asocia las lecturas de ignición y conexión con la posición de su misma trama. El historial SQLite distingue el valor falso de la ausencia de medición. El comando `cmd/tcp-sener` procesa las posiciones de la unidad incluidas en un intervalo UTC, con inicio incluido y fin excluido. La ficha sintética vincula el UID del retransmisor con los datos usados para construir el reporte.

La cola de este ensayo se guarda en un archivo independiente del historial. Al volver a procesar un intervalo, el UID y la huella de la trama permiten reconocer los registros ya incorporados, conservar su identificador de transmisión y recuperar posiciones que llegaron tarde. El cliente multipart entrega el archivo JSON a un simulador de loopback. Una respuesta ambigua queda registrada para conciliación. [Diseño de la integración](../specs/tcp-sener-local/design.md)
