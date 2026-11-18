---
feature: receptor-tcp-cr13
status: approved
tags: [harness, spec]
---

# Tareas del receptor TCP

## R1 — Arranque local

- [x] (1) Test rojo. Probar una dirección de loopback, un UID ausente y una dirección comodín.
- [x] (2) Implementación mínima. Añadir composición del comando y validación previa a la escucha.
- [x] (3) Refactor. Revisar los errores de configuración con los tests verdes.

## R2 — Lectura del protocolo

- [x] (1) Test rojo. Transmitir una trama en varios fragmentos, dos tramas juntas y otra de UID diferente.
- [x] (2) Implementación mínima. Decodificar los campos necesarios y seleccionar el UID configurado.
- [x] (3) Refactor. Consolidar el decodificador con los tests verdes.

## R3 — Historial y retransmisión

- [x] (1) Test rojo. Reabrir SQLite y repetir una trama. Añadir dos posiciones distintas del mismo segundo.
- [x] (2) Implementación mínima. Persistir la posición con unicidad y hora GPS.
- [x] (3) Refactor. Revisar las operaciones de escritura con los tests verdes.

## R4 — Consultas por hora GPS

- [x] (1) Test rojo. Probar límites de intervalo, empate, datos atrasados, telemetría sin coordenadas y consulta vacía.
- [x] (2) Implementación mínima. Implementar consultas locales de última posición y por intervalo.
- [x] (3) Refactor. Verificar el orden y la salida del comando con los tests verdes.

## R5 — Errores y cierre

- [x] (1) Test rojo. Enviar tramas excesivas, incompletas o con coordenadas inválidas. Abrir otra conexión y detener el receptor con un cliente inactivo.
- [x] (2) Implementación mínima. Aplicar límites, aislar errores y cerrar las conexiones en curso.
- [x] (3) Refactor. Revisar la liberación de recursos con los tests verdes.

## R6 — Ensayo reproducible

- [x] (1) Test rojo. Ejecutar con emisor sintético, puerto local, base temporal, reapertura y retransmisión.
- [x] (2) Implementación mínima. Conectar los adaptadores y documentar los comandos de prueba y consulta.
- [x] (3) Refactor. Correr la verificación del proyecto y preparar la revisión independiente.
