---
feature: tcp-sener-local
status: draft
tags: [harness, spec]
---

# Diseño de la prueba integrada

**Captura.** El adaptador TCP extrae las dos lecturas del mismo mensaje que la posición. Para el ensayo, la ficha documenta los dos parámetros ficticios y sus valores binarios. La velocidad y la existencia de una conexión TCP carecen de información suficiente sobre las señales históricas. El historial necesita mantener el valor desconocido tras una migración compatible con sus filas anteriores. Las consultas de posición siguen aceptando esas filas.

**Preparación.** Un caso de uso lee el intervalo del historial de una sola unidad y reutiliza el generador del reporte y las validaciones de la ficha. La selección usa los límites publicados del historial, con inicio incluido y fin excluido. La clave de entrada combina origen TCP, UID y huella de la trama. Releer una ventana después de un corte recupera lo que llegó tarde y permite identificar lo que ya se procesó. La ficha establece de forma explícita la asociación entre el UID retransmitido y la unidad de prueba.

**Entrega.** Componer la fuente de historial y la cola SQLite existente desde un punto de entrada documentado. Para la prueba local, usar una base de historial y una cola específicas del ensayo, el cliente multipart existente y un destino HTTP de loopback. La cola conserva el UUID, el estado y el motivo de cada resultado. Una interrupción entre enviar y confirmar se resuelve con la política de conciliación ya implementada.

**Verificación.** El simulador examina el nombre del archivo, la parte `gps`, el cuerpo JSON, el encabezado de autorización ficticio y el número de solicitudes. Se comprueba el estado en almacenamiento y lo que recibió el simulador. Los tests del receptor y del piloto de Remote API siguen siendo comprobaciones de regresión. [Arquitectura del proyecto](../../docs/architecture.md)
