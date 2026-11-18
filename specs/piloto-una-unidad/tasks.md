---
feature: piloto-una-unidad
status: approved
tags: [harness, spec]
---

# Tareas del piloto

## R1 — Ámbito y consulta

- [x] (1) Escribir una prueba roja que nombre R1. Simular una ventana paginada. Comprobar ID, límites UTC y lectura completa. La revisión de CLI también verifica el rechazo de intervalo cero.
- [x] (2) Implementar el comportamiento mínimo y comprobar la prueba.
- [x] (3) Refactorizar en verde y registrar test y commit en [[traceability]].

## R2 — Reporte válido

- [x] (1) Escribir una prueba roja que nombre R2. Cubrir el JSON completo con datos ficticios, un VIN de 17 caracteres y códigos del catálogo.
- [x] (2) Implementar el comportamiento mínimo y comprobar la prueba.
- [x] (3) Refactorizar en verde y registrar test y commit en [[traceability]].

## R3 — Datos incompletos

- [x] (1) Escribir una prueba roja que nombre R3. Probar ignición desconocida, vehículo detenido con motor encendido, permiso ausente y códigos inválidos.
- [x] (2) Implementar el comportamiento mínimo y comprobar la prueba.
- [x] (3) Refactorizar en verde y registrar test y commit en [[traceability]].

## R4 — Contrato de envío

- [x] (1) Escribir una prueba roja que nombre R4. Inspeccionar método, ambiente, encabezados, nombre de la parte, archivo, MIME y contenido con httptest.
- [x] (2) Implementar el comportamiento mínimo y comprobar la prueba.
- [x] (3) Refactorizar en verde y registrar test y commit en [[traceability]].

## R5 — Estado de entrega

- [x] (1) Escribir una prueba roja que nombre R5. Probar aceptación, 200 sin error, 200 con error 1, 422, 429 y corte de conexión tras recibir la petición.
- [x] (2) Implementar el comportamiento mínimo y comprobar la prueba.
- [x] (3) Refactorizar en verde y registrar test y commit en [[traceability]].

## R6 — Reinicio y repetición

- [x] (1) Escribir una prueba roja que nombre R6. Detener y relanzar el proceso después de guardar y durante un envío. Releer una ventana con mensajes distintos del mismo segundo.
- [x] (2) Implementar el comportamiento mínimo y comprobar la prueba.
- [x] (3) Refactorizar en verde y registrar test y commit en [[traceability]].

## R7 — Prueba local y piloto real

- [x] (1) Escribir una prueba roja que nombre R7. Ejecutar sin credenciales externas. Comprobar que solo se contactan los simuladores y que producción no es un destino permitido en este piloto.
- [x] (2) Implementar el comportamiento mínimo y comprobar la prueba.
- [x] (3) Refactorizar en verde y registrar test y commit en [[traceability]].
