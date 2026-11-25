---
feature: tcp-sener-local
status: draft
tags: [harness, spec]
---

# Tareas

| Requisito | (1) Test rojo | (2) Implementación mínima | (3) Refactor y comprobación |
|---|---|---|---|
| R1 | Enviar dos señales válidas, incluyendo cero, y una posición con lectura ausente. Reabrir SQLite. | Decodificar y conservar las lecturas por posición. | Revisar que recepción, almacenamiento y consulta histórica sigan funcionando. |
| R2 | Verificar los campos del multipart, el tratamiento de una señal ausente y el de una fila anterior a la migración. | Conectar historial y generador con ficha autorizada, registrando motivos en la cola. | Reutilizar validaciones y revisar que la consulta deje disponibles las posiciones incompletas. |
| R3 | Reprocesar el intervalo con un duplicado, dos tramas distintas del mismo segundo y una atrasada. | Encolar por identidad estable y conservar UUID. | Comprobar deduplicación y recuperación tras interrupción. |
| R4 | Simular aceptación, rechazo, 429 y respuesta ambigua, incluyendo reapertura. | Integrar el emisor con los resultados de la cola. | Comprobar qué estados habilitan un nuevo intento. |
| R5 | Ejecutar el caso completo por loopback y repetir la trama tras reiniciar. | Completar la composición local y documentar el comando reproducible. | Verificar una sola posición, aceptación y solicitud en el caso confirmado. Entregar a revisión independiente. |
