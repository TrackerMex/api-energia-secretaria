# Revisión independiente de TCP a SENER en local

**Aprobada la entrega `819d5ca63a90458a2e7a5ccf64781ca7742adaaf` el 25 de noviembre de 2026.** El resultado corresponde al UID ficticio `DEMO-CR13`, bases SQLite temporales y un simulador HTTP de loopback.

## Resultado de las dos revisiones

En `fd91ce457852b85ac38cbf240e951da616cd6248`, el harness pasó. Una prueba adicional con los dos ejecutables detectó un fallo al retransmitir una posición de la versión anterior. Preparé una base con el esquema antiguo y la huella correcta de una trama con ignición 0 y conexión 1. Al recibir esos mismos bytes con la ficha, la fila siguió con dos señales NULL y selectores vacíos. El puente registró un reporte inválido y el simulador recibió cero POST. Rechacé esta candidata por R1 y R2.

En la entrega corregida, el historial incorpora cada medición que faltaba cuando la trama exacta se vuelve a recibir con un selector compatible. Los tests de regresión preceden a la corrección `481365d`. Repetí el ensayo fuera de la suite, con ambos ejecutables y la misma fila antigua. Primero procesé el intervalo y comprobé el UUID persistido en un reporte inválido con cero solicitudes. Luego ejecuté dos veces la secuencia de iniciar el receptor, reenviar los mismos bytes, cerrar el receptor y volver a ejecutar el puente. En ambos cierres hubo una posición con ignición 0 de `ign` y conexión 1 de `online`, un reporte aceptado con el UUID original y un solo POST acumulado. Leí ese mismo UUID del JSON multipart recibido.

## Evidencia por requisito

| Requisito | Resultado comprobado |
|---|---|
| R1 | La lectura de cada señal pertenece a la trama. La captura conserva el cero, el valor desconocido y su procedencia al reabrir SQLite. Las pruebas de recuperación por campo conservan un cero y su selector ante datos conflictivos, completan la medición que falta y rechazan reutilizar como fuente la ya asignada a la otra señal. |
| R2 | La migración conserva las posiciones antiguas. Una lectura ausente o un cambio de selector deja el motivo de validación en la cola. El simulador verifica la parte `gps`, el nombre del archivo, el token ficticio y los campos del reporte. |
| R3 | La identidad combina el UID y la huella de la trama. La selección por intervalo admite posiciones tardías y dos tramas distintas del mismo segundo. Al repetir una entrega aceptada se conserva el UUID. |
| R4 | Las pruebas comprueban aceptación, rechazo 422, respuesta ambigua y redirección. El 429 espera hasta el vencimiento persistido y reutiliza el UUID en el intervalo seleccionado. El test adelanta el vencimiento en su base temporal. Las entregas inciertas quedan disponibles para conciliación. |
| R5 | `TestR5TCPToSENERLocal` compila y lanza ambos comandos, reinicia el receptor, retransmite el mismo vector y vuelve a ejecutar el puente. Tras esos dos ciclos comprueba una posición, un reporte aceptado, un único POST y coincidencia del UUID y de la hora GPS en el multipart. |

## Harness, regresión y aislamiento

Ejecuté `./init.sh` una vez sobre cada candidata con `GOPROXY=off` y compilador local. Ambos recorridos finalizaron con código 0. En la candidata final el HEAD coincidió antes y después, la compilación tardó 1 s, las pruebas sin caché tardaron 5 s y `go vet ./...` terminó correctamente. Pasaron nueve paquetes con pruebas, entre ellos el piloto de Remote API, el receptor, el puente, la ingesta y la cola.

| Checkpoint | Verificación |
|---|---|
| C2 | La sesión activa corresponde al puente. El harness verificó una sola feature en curso, las dos previas cerradas y el contador 2/3 sincronizado. |
| C3 | El caso de uso del puente depende de interfaces de dominio para historial, cola y emisor. El comando compone los adaptadores. El reporte y la entrega multipart reutilizan la implementación compartida. |
| C4 y C5 | Los cinco requisitos tienen pruebas identificadas y commits en la matriz. Los casos iniciales `6aefb79` y `282c97c` y los de recuperación `f205b90` y `62b832e` preceden al código correspondiente. |
| C6 | La aprobación de los cinco requisitos tiene fecha 2026-11-25 y quedó registrada en `05d6288`. El contenido de requisitos se mantuvo desde esa aprobación. |
| C7 | El piloto de Remote API conserva su comando, su cursor y sus pruebas. El puente solo procesa las claves del intervalo solicitado. Las colas de las dos fuentes comprueban su identidad, incluidas las de Remote API anteriores a la migración. |

El cliente del simulador admite una IP literal de loopback, desactiva el proxy y corta las redirecciones. La operación con Wialon real, Hostinger o SENER requiere otra fase. [Matriz de trazabilidad](../specs/tcp-sener-local/traceability.md)
