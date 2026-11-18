---
feature: piloto-una-unidad
status: draft
tags: [harness, spec]
---

# Diseño del piloto

## Recorrido de una posición

El proceso Go en el VPS abre conexiones HTTPS salientes. Lee mensajes de una unidad autorizada mediante `messages/load_interval` y completa las páginas de la ventana antes de consolidar el cursor. Un solapamiento controlado permite recuperar mensajes en los límites de ventana. Para identificar una lectura se conserva una clave estable que distingue mensajes distintos de la misma unidad con la misma marca temporal.

La ficha del permisionario y de la unidad aporta RFC, razón social, permiso, placa, VIN, IMEI y códigos de catálogo. La posición y el estado del vehículo proceden del mensaje y de sus sensores configurados. El origen de la ignición y del estado de conexión se valida con la unidad elegida antes del primer envío real. Se omiten mediciones opcionales ausentes y se conserva el objeto de sensores adicionales exigido por la tabla del manual. Los rechazos de validación deben poder rastrearse hasta la lectura.

## Almacenamiento y entrega

El almacén duradero guarda la lectura o su resultado de validación y el cursor de la ventana en una operación consistente. La primera inserción válida asigna un UUID y conserva el cuerpo que se enviará. Las relecturas y reintentos reutilizan ese registro. Antes del POST se persiste la intención de envío para detectar una interrupción durante la entrega. Tras un 429 se programa un nuevo intento respetando `Retry-After` si está presente. Ante un timeout después del envío, 5xx de efecto desconocido o una respuesta ilegible, se conserva evidencia para conciliación. Los manuales no especifican deduplicación remota ni un mecanismo de consulta de acuse.

## Responsabilidades previstas

| Ubicación propuesta | Responsabilidad |
|---|---|
| `internal/domain/` | Modelos de posición, ficha de unidad, validación y catálogos. |
| `internal/application/` | Lectura de la ventana, avance de cursor y transiciones de entrega. |
| `internal/infrastructure/` | Adaptadores Wialon, SENER y almacenamiento en disco. |
| `cmd/piloto/` | Configuración del modo de ejecución y composición del proceso. |

El archivo JSON se puede formar en memoria. El nombre del adjunto deriva del RFC configurado. Solo se usan secretos suministrados en tiempo de ejecución. El receptor de pruebas local comprueba el cuerpo multipart y devuelve respuestas deterministas. El receptor TCP y la capa intermedia de telemática quedan fuera del camino elegido por el cliente.

## Referencias

Manual técnico de integración, secciones 1–5. API de Integración GPS, secciones 4–8 y Anexo 2. El Anexo 1 documenta umbrales de alerta; los dos PDF examinados no fijan una cadencia general de envío de posiciones. Para el formato y la consulta de mensajes, véase la [referencia de Wialon](https://help.wialon.com/en/api/user-guide/api-reference/messages/load_interval).

La unidad elegida es CR-13. En Hostinger el proceso guardará el estado en /var/lib/wialon-sener/pilot.db con SQLite y un directorio privado del usuario de servicio. El archivo y sus archivos auxiliares persistirán entre despliegues. Los tests usarán directorios temporales.
