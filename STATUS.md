# Piloto Wialon–SENER

**Última actualización**: 2026-11-25
**Features completadas**: 3/3 (`feature_list.json`)
**Pendientes**: comprobar la captura real de CR-13 y preparar la prueba en SENER desarrollo
**En producción**: no

## Estado actual

El piloto local en Go recibe posiciones de Wialon Retranslator por TCP, conserva sus señales históricas en SQLite y puede procesar un intervalo UTC para enviar reportes a un simulador de SENER. Una retransmisión exacta ya aceptada conserva un único reporte. Las posiciones sin una señal obligatoria quedan registradas con el motivo. También está disponible el piloto anterior con Wialon Remote API.

## Última sesión

El usuario aprobó los cinco requisitos de integración para `DEMO-CR13`. El ensayo completo comprobó la recepción TCP, la persistencia, el multipart y la aceptación por un simulador en loopback. Tras reiniciar y repetir la trama, el resultado fue una posición, un reporte aceptado y un POST. Un ensayo adicional con una segunda posición sin señales confirmó que queda pendiente de corrección. La revisión también verificó que retransmitir una trama antigua con mediciones disponibles permite completar su reporte con el UUID original. [Guía para repetir la prueba](README.md) · [Revisión independiente](progress/review_tcp-sener-local.md)

## Para retomar la prueba real

El UID indicado por el operador se conservó en su configuración local privada. Al habilitar el retransmisor habrá que comprobar el identificador de salida, la procedencia y el significado de las señales históricas en el equipo real, la ficha del permisionario, el acceso al entorno de desarrollo de SENER y la frecuencia operativa. El despliegue en el VPS y las once unidades forman parte de una etapa posterior. [Configuración del piloto Remote API](deploy/README.md) · [Revisión de los manuales](docs/frecuencia-sener.md)
