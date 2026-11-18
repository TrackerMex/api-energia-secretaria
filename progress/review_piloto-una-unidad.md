# Revisión independiente del piloto CR-13

Fecha de revisión: 2026-11-18. Alcance: ejecución local con una unidad sintética.

## Veredicto inicial

Revisión técnica del commit `109712a18f82ede9bc73999fec2945da33a98f09`. Las pruebas funcionales pasaron. Queda por verificar el checkpoint C3 sobre la dirección de dependencias del caso de uso. La firma de `runWindow` usa directamente las implementaciones concretas de ingesta, persistencia y envío.

## Evidencia

| Comprobación | Resultado |
|---|---|
| `./init.sh` sin argumentos, Git Bash | Código 0; pasaron build, los cuatro paquetes de pruebas y go vet. |
| R1–R3 | El adaptador consulta una unidad, pagina la ventana y valida catálogos, perfil y señales. La ignición digital requiere información de entrada en el propio mensaje. |
| R4–R5 | Las pruebas inspeccionan el archivo multipart y su nombre, MIME y Bearer, así como aceptación explícita, rechazos, respuestas ambiguas y espera tras HTTP 429. |
| R6 | La escritura del cursor comparte transacción con la de los resultados. La prueba cierra y vuelve a abrir SQLite, comprueba el mismo UUID en el reintento y conserva los estados aceptado e incierto. |
| R7 | Dos servidores de prueba TLS locales presentan una posición y aceptan un reporte. Al repetir la misma ventana, el receptor simulado recibe una sola entrega. |
| Validaciones adicionales de CLI | Se invocó el binario mediante go run con configuración sintética. Rechazó selección de producción, intervalo 0 en modo periódico y unidad incompatible con el perfil antes de conectar a la red. |
| C1, C2, C4–C6 | Harness en verde, una sola feature en curso, requisitos aprobados y siete filas de trazabilidad con pruebas y commits. |
| C7 | Primera feature del proyecto. |

La revisión usó SQLite temporal. La cadencia externa y el estado de conexión histórica por fuente verificable siguen pendientes para la eventual prueba real. La guía documenta el alcance de la recuperación de mensajes que llegan tarde y la ruta propuesta de almacenamiento en Linux.

## Veredicto final

**Aprobado para el piloto local de una unidad.** Los requisitos R1–R7 y los checkpoints aplicables quedan verificados en `5ba50b0a469567cc3fd7f36a82367f7395a43ce9`.

La corrección de C3 movió las interfaces a `internal/domain/ports.go`. El caso de uso en `internal/application/pilot.go` depende del dominio y de la biblioteca estándar. Los adaptadores implementan esos contratos y la CLI los compone. Verifiqué estas dependencias en el código.

Repetí la verificación completa por mi cuenta con `./init.sh` sobre el commit corregido. Dio código 0: build en 2 segundos, suite en 4 segundos y `go vet` en verde. El hash fue idéntico antes y después de la prueba. Las verificaciones locales cubren los tiempos, el mapeo, el envío simulado y la recuperación. La recepción por SENER desarrollo con datos reales sigue fuera de la evidencia de esta revisión.
