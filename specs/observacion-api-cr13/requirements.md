---
feature: observacion-api-cr13
status: approved
tags: [harness, spec]
---

# Observación de CR-13 mediante Wialon Remote API

## Requisitos

- **R1**. Cuando el operador indica una unidad autorizada y una ventana UTC explícita en segundos, el comando debe leer sus mensajes de posición mediante el adaptador de Wialon Remote API, con ambos extremos incluidos. Si la consulta falla, debe terminar con error y distinguir el fallo de una ventana válida sin posiciones.
- **R2**. Para cada mensaje, el sistema debe comparar la ignición de la entrada digital 1 con las RPM del mismo mensaje. La entrada se obtiene del primer bit de `i` cuando `f` contiene `0x02`. Las RPM se obtienen de `p.can_rpm / 10`. Debe clasificar el motor como encendido cuando la entrada está activa y las RPM son positivas, o apagado cuando está inactiva y las RPM son cero. Ante discrepancia, lectura ausente o valor inválido debe clasificarlo como indeterminado e identificar el motivo.
- **R3**. Cuando completa la lectura, el comando debe mostrar un resumen reproducible con el total de posiciones, los conteos por las tres clases y los motivos de indeterminación. Cada posición contará una sola vez en el total y en su clase. La conectividad histórica del GPS figurará por separado como pendiente de validación, incluso cuando se conozca el estado del motor. La salida y los errores omitirán coordenadas, identificadores físicos y credenciales.
- **R4**. Cuando el ensayo local ejecuta el comando contra Wialon simulado con una unidad ficticia, debe producir los conteos esperados para una ventana mixta. El ensayo y el modo de observación deben operar independientemente de la cola de entregas y con cero solicitudes a SENER.

## Fuera de alcance

La validación de una señal de conectividad del equipo real, la autorización de entregas, el despliegue y la ampliación de flota quedan para fases posteriores.

## Aprobación

- [x] Aprobado por humano (fecha: 2026-11-25)
