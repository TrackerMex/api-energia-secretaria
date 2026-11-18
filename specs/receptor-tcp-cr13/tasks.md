---
feature: receptor-tcp-cr13
status: draft
tags: [harness, spec]
---

# Tareas del receptor TCP

## R1 — Arranque local

- [ ] (1) Test rojo. Probar una dirección de loopback, un UID ausente y una dirección comodín.
- [ ] (2) Implementación mínima. Añadir composición del comando y validación previa a la escucha.
- [ ] (3) Refactor. Revisar los errores de configuración con los tests verdes.

## R2 — Lectura del protocolo

- [ ] (1) Test rojo. Transmitir una trama en varios fragmentos, dos tramas juntas y otra de UID diferente.
- [ ] (2) Implementación mínima. Decodificar los campos necesarios y seleccionar el UID configurado.
- [ ] (3) Refactor. Consolidar el decodificador con los tests verdes.

## R3 — Historial y retransmisión

- [ ] (1) Test rojo. Reabrir SQLite y repetir una trama. Añadir dos posiciones distintas del mismo segundo.
- [ ] (2) Implementación mínima. Persistir la posición con unicidad y hora GPS.
- [ ] (3) Refactor. Revisar las operaciones de escritura con los tests verdes.

## R4 — Consultas por hora GPS

- [ ] (1) Test rojo. Probar límites de intervalo, empate, datos atrasados, telemetría sin coordenadas y consulta vacía.
- [ ] (2) Implementación mínima. Implementar consultas locales de última posición y por intervalo.
- [ ] (3) Refactor. Verificar el orden y la salida del comando con los tests verdes.

## R5 — Errores y cierre

- [ ] (1) Test rojo. Enviar tramas excesivas, incompletas o con coordenadas inválidas. Abrir otra conexión y detener el receptor con un cliente inactivo.
- [ ] (2) Implementación mínima. Aplicar límites, aislar errores y cerrar las conexiones en curso.
- [ ] (3) Refactor. Revisar la liberación de recursos con los tests verdes.

## R6 — Ensayo reproducible

- [ ] (1) Test rojo. Ejecutar con emisor sintético, puerto local, base temporal, reapertura y retransmisión.
- [ ] (2) Implementación mínima. Conectar los adaptadores y documentar los comandos de prueba y consulta.
- [ ] (3) Refactor. Correr la verificación del proyecto y preparar la revisión independiente.
