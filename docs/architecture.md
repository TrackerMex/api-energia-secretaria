# Arquitectura del piloto

Un proceso Go en el VPS de Hostinger consultará una única unidad por Wialon Remote API. Las dos conexiones son HTTPS salientes. El piloto de posiciones tiene como destino SENER desarrollo.

| Capa | Responsabilidad |
|---|---|
| Dominio | Unidad, posición GPS y validación de campos y catálogos. |
| Aplicación | Elegir el intervalo, asociar los datos del permisionario, registrar el reporte y decidir su estado de entrega. |
| Infraestructura | Sesión Wialon, almacenamiento duradero y cliente HTTP multipart de SENER. |

El cursor y los reportes pendientes deben sobrevivir a un reinicio. La aceptación necesita una respuesta HTTP 200 con `error: 0`. Una respuesta ambigua conserva la identidad de la transmisión para conciliación.

Para Wialon, contrastar la paginación y el conteo devuelto con la documentación de [mensajes por intervalo](https://help.wialon.com/en/api/user-guide/api-reference/messages/load_interval). Para SENER, ver el [análisis de los manuales](frecuencia-sener.md).
