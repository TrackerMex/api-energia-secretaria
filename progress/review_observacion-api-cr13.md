# Revisión independiente de la observación por Wialon Remote API

**Aprobada la versión `c8aec80679e59499dd9c4a4a572b97af853264fa` para el ensayo de observación.** El 25 de noviembre de 2026 verifiqué los cuatro requisitos con telemetría sintética y un servidor HTTPS de loopback. La muestra de cuatro posiciones produjo un motor encendido, uno apagado y dos indeterminados. La conectividad del GPS quedó pendiente de validación. Estos resultados describen los mensajes del ensayo.

## Las dos revisiones

En `5bd9dcd0628ce1c4a260648ab203e3dafeadce91`, `./init.sh` pasó con once paquetes con pruebas y `go vet ./...` correcto. Al ejecutar el comando con inicio `2026-11-25T12:00:00.0000000001Z` y fin `2026-11-25T12:00:00Z`, encontré que aceptaba los límites hasta llegar a la validación del ID de la unidad. El truncamiento del lector de fechas convertía ese inicio posterior al fin en una ventana de un segundo. Rechacé esta versión por R1.

La prueba de regresión `3d78cb4` precede a la corrección `527f643`. En la versión final repetí la invocación con ambos límites originales y las variables de acceso vacías. Obtuve código 1 por fecha UTC inválida antes de la validación de credenciales. La prueba también comprueba que fracciones no nulas en cualquiera de los extremos se rechazan antes de abrir sesión. Una fracción íntegramente cero conserva el segundo y se admite dentro del rango de Wialon.

## Cobertura de la versión final

| Requisito | Evidencia |
|---|---|
| R1 | El servidor simulado verificó la unidad, el filtro de mensajes de posición y los segundos inicial y final inclusivos. Una ventana válida de un segundo sin mensajes devolvió conteos cero. Al fallar la segunda página, el comando terminó con error y la salida estándar quedó vacía. El adaptador solicitó el cierre de la sesión. Las pruebas cubren ambos límites del rango de Wialon. |
| R2 | La clasificación lee la bandera `0x02` de `f`, el primer bit de `i` y el parámetro numérico `p.can_rpm` de cada mensaje. Dividir `can_rpm` entre diez conserva el signo usado para la clasificación. Una entrada activa con RPM positivas cuenta como encendido y una entrada inactiva con RPM cero como apagado. Las discrepancias y las lecturas faltantes o inválidas tienen motivos propios. |
| R3 | El total coincide con la suma de las clases. Una posición con dos señales ausentes aparece una sola vez en indeterminados y aporta ambos motivos. El JSON explica esta convención y separa la conectividad histórica. Las pruebas de éxito, errores y redirección comprueban que las salidas omiten credenciales, identificadores y coordenadas de prueba. |
| R4 | En dos páginas, el servidor registró únicamente `token/login`, `messages/load_interval`, `messages/get_messages` y `core/logout`. El comando usó exclusivamente las tres variables de Wialon. La prueba observó cero solicitudes a SENER. |

En la candidata final ejecuté `./init.sh` una vez con `GOPROXY=off` y el compilador local. HEAD coincidió antes y después. La compilación tardó 2 s, las pruebas sin caché duraron 6 s y `go vet ./...` terminó correctamente. El harness salió con código 0. Pasaron once paquetes con pruebas, incluidos los pilotos Remote API y TCP, persistencia y entrega.

| Checkpoint | Comprobación |
|---|---|
| C2 | La sesión y el historial describen la observación activa y las fases anteriores cerradas. El harness verificó el contador 3/4 y una única feature en curso. |
| C3 | La aplicación depende de la interfaz de lectura del dominio. El clasificador trabaja con mensajes en memoria. El comando compone la fuente Wialon y el caso de uso. |
| C4 y C5 | Los cuatro requisitos tienen pruebas y commits en la matriz. Los commits de pruebas del comando, del clasificador y del defecto UTC preceden al código correspondiente. |
| C6 | Los requisitos aprobados están fechados el 2026-11-25. Su última modificación corresponde al registro de aprobación `20b720f`. |
| C7 | El comando de observación reutiliza la lectura de Remote API. Las vías existentes de captura y entrega conservan sus puntos de entrada y su cobertura de regresión. |

[Pruebas y revisiones por requisito](../specs/observacion-api-cr13/traceability.md)
