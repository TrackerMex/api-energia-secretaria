# Frecuencia de reporte en los dos manuales entregados

Revisión del 18 de noviembre de 2026. Se revisaron las tres páginas del manual personalizado y las diecisiete páginas del documento de integración, incluidos sus anexos.

| Fuente | Qué documenta |
|---|---|
| Manual técnico de integración, secciones 2, 4 y 5, páginas PDF 1–3 | Forma del envío, resultado y código HTTP 429 cuando se supera el límite. La cifra de solicitudes por intervalo queda sin especificar. |
| API de Integración GPS, secciones 4–6, páginas PDF 3–7 | Uno o más registros en `data`, identificador por transmisión y hora GPS en UTC. |
| API de Integración GPS, anexo 1, página PDF 13 (hoja 12 de 16) | El evento `perdida_comunicacion` se evalúa tras 300 segundos o más sin comunicación durante un viaje activo. |

En los documentos entregados no aparece una periodicidad obligatoria de envío de posiciones, una demora máxima entre GPS y recepción ni una cuota numérica por minuto. Los umbrales temporales del anexo sirven para clasificar eventos.

## Decisión del piloto

La primera comprobación usará una sola unidad y un intervalo histórico explícito. Para el proceso periódico, el intervalo entre consultas a Wialon será configurable. El valor operativo de producción y el tratamiento de duplicados requieren confirmación del cliente o de SENER.
