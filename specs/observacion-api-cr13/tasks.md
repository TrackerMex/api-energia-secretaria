---
feature: observacion-api-cr13
status: approved
tags: [harness, spec]
---

# Tareas del modo de observación

| Requisito | (1) Test rojo | (2) Implementación mínima | (3) Refactor y comprobación |
|---|---|---|---|
| R1 ✓ | Verificar la selección de una unidad e intervalo, el resultado vacío y el fallo de una página. | Componer la fuente de posiciones con los parámetros UTC. | Revisar la propagación de errores y el cierre de sesión. |
| R2 ✓ | Cubrir encendido, apagado, cero medido, desacuerdos, entradas o RPM ausentes o inválidas y mensajes consecutivos que no comparten ambas señales. | Clasificar cada mensaje y asociar sus motivos. | Reutilizar la lógica de entradas digitales compatible con la distinción entre ausencia y cero. |
| R3 ✓ | Contar una ventana mixta y otra vacía, comprobar los motivos, el estado de conexión y el contenido de la salida y errores. | Agregar resultados y emitir un resumen determinista. | Verificar que el total coincide con la suma de las clases. |
| R4 ✓ | Ejecutar el punto de entrada con datos ficticios y un simulador local de Wialon. | Completar la composición del comando y su invocación de prueba. | Comprobar conteos esperados y que todas las peticiones corresponden a la sesión de lectura de Wialon. |

Las pruebas de implementación verifican la ventana inclusiva, las clasificaciones, la salida agregada y el acceso de solo lectura a la API simulada. [Cobertura por requisito](traceability.md)
