---
feature: tcp-sener-local
status: draft
tags: [harness, spec]
---

# Prueba local de TCP a SENER

El ensayo usa una unidad ficticia con UID `DEMO-CR13`, una ficha de prueba y dos servicios en loopback.

## Requisitos

- **R1**: Cuando se recibe por TCP una posición de la unidad autorizada, el sistema debe guardar junto a ella las mediciones de ignición y conexión GPS de esa trama según el mapeo de la ficha. Debe distinguir un cero medido de una lectura ausente o desconocida después de reabrir SQLite.
- **R2**: Cuando el operador procesa un intervalo UTC del historial, el sistema debe preparar un reporte por posición válida de la unidad configurada con su hora GPS, coordenadas, rumbo, velocidad, señales y ficha. El simulador debe recibir el JSON mediante el contrato multipart de SENER. Si falta una señal obligatoria, debe conservarse la posición y su motivo de validación, con cero envíos de ese registro. Esto aplica también a bases creadas antes de esta fase.
- **R3**: Cuando se vuelve a procesar el mismo intervalo, el sistema debe reconocer cada trama por su identidad persistente, conservar el UUID asignado y omitir las entregas aceptadas. Una posición atrasada incorporada al intervalo debe poder procesarse al releerlo. Dos tramas diferentes del mismo segundo deben conservar identidades independientes.
- **R4**: Cuando el simulador confirma la aceptación, rechaza datos, limita el tráfico o deja un resultado ambiguo, el sistema debe persistir el estado correspondiente conforme a la política de la cola existente. Al reiniciar, un 429 permitirá reintentar con el mismo UUID al vencer la espera. Una entrega incierta debe quedar disponible para conciliación y exenta de reenvío automático.
- **R5**: Cuando se ejecuta la prueba integrada desde los puntos de entrada de la aplicación, una trama sintética debe recorrer TCP, SQLite, la cola y el simulador local. Después de la primera aceptación, se reinician los procesos, se retransmite la misma trama y se repite el procesamiento. El resultado debe ser una posición, un reporte aceptado y un único POST recibido en todo el ensayo. El ensayo debe poder repetirse con archivos temporales y credenciales ficticias.

## Fuera de alcance

La conexión con Wialon real, el despliegue en Hostinger, el envío a SENER desarrollo o producción, la frecuencia operativa y la incorporación de otras unidades corresponden a fases posteriores.

## Aprobación

- [ ] Aprobado por humano (fecha: ____-__-__)
