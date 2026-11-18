---
feature: piloto-una-unidad
status: approved
tags: [harness, spec]
---

# Requisitos del piloto de una unidad

- **R1**: Cuando se inicia el piloto con el ID de unidad y una ventana UTC, el servicio consulta por Remote API solo esa unidad y lee todos sus mensajes de posición disponibles en la ventana, incluidas páginas adicionales. La frecuencia de consulta se configura por separado y debe ser positiva.
- **R2**: Cuando llega una posición con sus datos obligatorios, el servicio genera un reporte con RFC, razón social y un elemento en `data` de tipo `posicion`, con su UUID, instante GPS en UTC, vehículo, geolocalización, estado del vehículo, estado del GPS y sensores adicionales. Usa los identificadores de carga y vehículo del Anexo 2.
- **R3**: Si falta un valor obligatorio, el catálogo es inválido o llega una lectura de sensor desconocida para un campo obligatorio, el servicio registra el mensaje pendiente de corrección con su motivo. Una velocidad de cero no determina el valor del encendido.
- **R4**: Cuando existe un reporte válido, el cliente envía un POST a SENER desarrollo con `Accept: application/json`, autorización Bearer y una parte `gps` de tipo `application/json` con nombre `<RFC>.json` y contenido UTF-8. En pruebas locales utiliza el mismo contrato contra un servidor simulado.
- **R5**: Cuando recibe HTTP 200 con un objeto JSON de respuesta cuyo campo entero `error` está presente y vale 0, el servicio registra el reporte como aceptado. Si recibe `error = 1` o un rechazo de datos, registra el motivo. Un 429 deja el reporte pendiente para un reintento diferido. Una respuesta ambigua conserva el reporte para conciliación.
- **R6**: Cuando se detiene el proceso entre la lectura y la confirmación, al reiniciar se recuperan los reportes guardados con el mismo UUID. El cursor solo avanza después de guardar el resultado de lectura. Los reportes ya aceptados no vuelven a enviarse y una entrega de resultado desconocido queda pendiente de conciliación.
- **R7**: Cuando se ejecuta la suite local, una unidad sintética recorre lectura, transformación, envío y aceptación con dos servidores simulados en loopback. La corrida real requiere selección explícita del ambiente de desarrollo, la unidad y sus credenciales de ejecución. Los registros de diagnóstico omiten secretos.

## Fuera de alcance

Esta primera entrega cubre posiciones de una unidad. Las alertas, la incorporación de las otras diez unidades, el envío a producción y la confirmación del plazo oficial de reporte quedan para fases posteriores.

## Aprobación

- [x] Aprobado por humano (fecha: 2026-11-18)
