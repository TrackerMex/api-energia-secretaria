# Revisión independiente del receptor TCP de CR-13

**Veredicto aprobado para el piloto local.** Revisé la entrega `adf2882f0501e3da40be965e2535d7b1aff1c60b` del 18 de noviembre de 2026. El código de `cfdbef9` incorpora la corrección de lectura unidireccional y máscara de posición de `d69b921`.

Ejecuté `./init.sh` sin argumentos con Git Bash y comprobé el mismo HEAD antes y después. La compilación tardó 2 s y la suite sin caché, 4 s. Pasaron los siete paquetes con pruebas, incluidos el piloto previo, el comando TCP, el adaptador de protocolo y SQLite. `go vet ./...` terminó correctamente. El harness salió con código 0.

| Requisito | Comprobación independiente |
|---|---|
| R1 | El comando comprueba el UID y una IP literal de loopback antes de abrir la base. El test rechaza UID vacío, comodines IPv4 e IPv6 y una IP no local. El ensayo de integración acepta un puerto local efímero. |
| R2 | El vector de referencia comprueba hora GPS UTC, coordenadas, velocidad y rumbo. Se procesan tramas fragmentadas y consecutivas, incluso cuando el emisor solo escribe. Una trama con otro UID deja la consulta vacía. |
| R3 | La clave única combina UID y huella de toda la trama. Las pruebas conservan una fila ante inserciones simultáneas y retransmisión tras reabrir SQLite. Dos tramas distintas con el mismo segundo GPS conservan dos registros y la base mantiene su vinculación al UID. |
| R4 | La última posición se ordena por hora GPS y huella. Se comprobaron la llegada tardía, el inicio incluido, el fin excluido, las fracciones de segundo, el desempate estable y el resultado JSON vacío. El lector abre SQLite en modo de solo lectura. |
| R5 | La lectura acota a 1 MiB la trama antes de reservar el búfer. Las pruebas cubren longitud excesiva, cortes, coordenadas inválidas, discrepancia de máscara y telemetría sin posición. Otro cliente puede seguir enviando y la cancelación cierra conexiones inactivas. Una posición ya guardada sigue disponible. |
| R6 | El test abre el receptor por TCP, envía un vector sintético independiente del decodificador y consulta los valores por la interfaz del comando. Tras detener y reabrir el receptor, vuelve a enviar la trama y la consulta devuelve exactamente una posición. Usa loopback, base temporal y tokens vacíos. |

| Checkpoint | Resultado |
|---|---|
| C1 y C2 | El harness verifica estructura y coherencia. La lista tiene una sola feature activa, el estado indica 1/2 completadas y la sesión previa figura en el historial. |
| C3 | El dominio define la posición y el contrato de archivo. La aplicación filtra el UID y usa esa interfaz. El comando compone los adaptadores de protocolo y SQLite. |
| C4 y C5 | Cada R tiene pruebas y commits en la matriz. Los commits de pruebas `1cdb3ea`, `433f581` y `df06a87` preceden sus implementaciones. |
| C6 | Los seis requisitos figuran aprobados el 2026-11-18 en `46d0b33`. El historial del archivo termina en esa aprobación. |
| C7 | El comando TCP añade una vía de recepción y consulta. El piloto con Remote API y la cola de entrega a SENER conservan sus usuarios y pruebas. |

La persistencia usa el archivo indicado con `-db`. El historial contiene el UID, la identidad de CR-13, la huella y los campos de posición. La ruta de demostración `state/` está excluida de Git. El proceso solicita permisos de propietario al crear archivos y directorios en sistemas POSIX. En Windows el acceso al archivo depende de la ACL de la carpeta. La comparación del UID permite seleccionar la unidad del ensayo. La prueba corresponde a tráfico sintético en el equipo local. La conexión de la unidad real, el despliegue y la entrega de posiciones TCP a SENER quedan para otra fase.
