# Perfil de una unidad

`profile.example.json` es una ficha ficticia. Copia su estructura a un archivo privado para la ejecución real. El servicio debe usar el ID interno de Wialon tanto en la consulta como en la ficha. IMEI, VIN, matrícula, permiso y los datos fiscales se comprueban con el permisionario.

`ignition_param` y `connection_param` nombran parámetros registrados en el objeto `p` de cada mensaje. En la muestra son `ign` y `online`. Confirma para la unidad elegida qué claves modelan realmente el encendido y la conectividad al instante de la lectura. El adaptador interpreta los booleanos, 0 y 1 (también sus equivalentes de texto `"false"`, `"true"`, `"0"`, `"1"`). Una lectura ausente o fuera de estos valores requiere corrección. Un equipo que solamente ofrezca una señal calculada o una consulta de estado presente necesita primero un mapeo comprobado con sus datos.

Para una entrada digital de Wialon se usa el selector explícito `digital:in1` a `digital:in32`. El adaptador extrae el bit correspondiente de `i` cuando la máscara `f` indica entradas digitales y `i` está presente, también si su valor es cero. Verifica el número y la polaridad de la entrada en cada ficha. [Entradas y salidas digitales de Wialon](https://help.wialon.com/en/wialon-hosting/expert-articles/unit/how-to-choose-the-number-of-digital-input-output-i-o).

Se consulta la ventana inclusiva en segundos UTC, seleccionando mensajes de datos con posición, y se consumen las páginas restantes en la misma sesión. La clave de cada lectura incluye la unidad y la huella de su contenido canónico. Las relecturas equivalentes conservan su identidad aunque cambie el orden de las propiedades JSON. El protocolo no expone una identidad adicional para dos mensajes idénticos en todo su contenido.

- [Plantilla de peticiones y centros de datos](https://help.wialon.com/en/api/user-guide)
- [Inicio de sesión](https://help.wialon.com/en/api/user-guide/api-reference/token/login)
- [Carga por intervalo](https://help.wialon.com/en/api/user-guide/api-reference/messages/load_interval)
- [Paginación por índice](https://help.wialon.com/en/api/user-guide/api-reference/messages/get_messages)

## Criterio de observación para CR-13

La fuente acordada es la API de Wialon. Para el motor se contrastan, en el mismo mensaje de la unidad, el estado de la entrada digital 1 (sensor `Ignicion`) y las revoluciones del motor (sensor `RPM`, parámetro `can_rpm / 10`). Este criterio guía la validación antes de usarlo en entregas reales.

| Evidencia del mismo mensaje | Resultado de observación |
|---|---|
| Entrada 1 activa y RPM válidas mayores que cero | Encendido según la telemetría |
| Entrada 1 inactiva y RPM válidas iguales a cero | Apagado según la telemetría |
| Las lecturas se contradicen, falta una de ellas o su valor es inválido | Indeterminado, con el motivo registrado al evaluar la muestra |

Para leer la entrada 1 en el mensaje de Remote API hay que comprobar la bandera `0x02` de `f` y luego el primer bit de `i`. Un intervalo calculado de ignición puede servir para contrastar tendencias. Por sí solo no demuestra qué valor contenía una trama individual. La referencia temporal es la hora GPS del mensaje. [Formato de mensajes de Wialon](https://help.wialon.com/en/api/user-guide/data-format/messages)

La conexión del GPS se evalúa por separado. Hasta identificar y validar una señal explícita en el mensaje histórico de CR-13, el estado de conexión queda indeterminado y la posición queda pendiente de entrega real a SENER.
