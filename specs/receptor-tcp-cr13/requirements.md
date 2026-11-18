---
feature: receptor-tcp-cr13
status: approved
tags: [harness, spec]
---

# Receptor TCP e historial local de CR-13

Este incremento permite comprobar en un equipo local que se recibe una posición por Wialon Retranslator 1.0 y que puede consultarse después de reiniciar. Se usará un emisor sintético con un UID de prueba asociado a CR-13.

## Requisitos funcionales

- **R1**: WHEN el operador inicia la recepción con un UID, la ruta de SQLite y una dirección de loopback, THE SYSTEM SHALL escuchar por TCP en esa dirección, sin credenciales externas. IF falta el UID o se solicita escuchar en una interfaz pública o comodín, THEN THE SYSTEM SHALL salir con error antes de aceptar conexiones.
- **R2**: WHEN llegan tramas binarias completas de Wialon Retranslator 1.0 del UID autorizado, THE SYSTEM SHALL extraer la hora del mensaje y las posiciones válidas de `posinfo`, con el mismo resultado si una trama llega en fragmentos o varias llegan juntas. IF llega una trama de otro UID, THEN THE SYSTEM SHALL excluirla del historial.
- **R3**: WHEN el receptor guarda una posición, THE SYSTEM SHALL conservar en SQLite la asociación con CR-13, la hora del GPS en UTC, latitud, longitud y los valores de velocidad y rumbo transmitidos. La retransmisión exacta debe producir un solo registro, incluso después de reiniciar. Dos tramas diferentes con posición en el mismo segundo deben conservarse como dos registros.
- **R4**: WHEN un operador consulta la base local, THE SYSTEM SHALL devolver mediante CLI la posición más reciente según la hora del GPS o las posiciones del intervalo UTC solicitado, incluyendo el inicio y excluyendo el fin. El historial se devuelve en orden cronológico con desempate estable. Si llegan datos atrasados o telemetría sin posición, debe poder seguir consultándose la posición válida más reciente. Una consulta sin coincidencias debe indicar que no hay registros.
- **R5**: IF la longitud anunciada excede el máximo documentado, una trama resulta inválida o incompleta al cerrar la conexión, o sus coordenadas exceden el rango geográfico, THEN THE SYSTEM SHALL descartar esa trama, preservar las ya guardadas y continuar atendiendo otras conexiones. Al detener el receptor, las conexiones inactivas deben cerrarse y las posiciones ya confirmadas en SQLite deben seguir disponibles.
- **R6**: WHEN se ejecuta la prueba local de extremo a extremo, THE SYSTEM SHALL aceptar una trama sintética por TCP, mostrar la posición por la CLI, recuperar la misma posición al reabrir la base y conservar un solo registro tras retransmitirla. El ensayo usa loopback y almacenamiento temporal, y debe funcionar con las credenciales de Wialon y SENER ausentes.

## Fuera de alcance

La conexión de CR-13 real, el despliegue en Hostinger, la activación para las once unidades, una política de retención, la API HTTP de consulta y el envío de las posiciones TCP a SENER corresponden a fases posteriores. La frecuencia de transmisión se acordará por separado.

## Aprobación

- [x] Aprobado por humano (fecha: 2026-11-18)
