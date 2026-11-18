# Perfil de una unidad

`profile.example.json` es una ficha ficticia. Copia su estructura a un archivo privado para la ejecución real. El servicio debe usar el ID interno de Wialon tanto en la consulta como en la ficha. IMEI, VIN, matrícula, permiso y los datos fiscales se comprueban con el permisionario.

`ignition_param` y `connection_param` nombran parámetros registrados en el objeto `p` de cada mensaje. En la muestra son `ign` y `online`. Confirma para la unidad elegida qué claves modelan realmente el encendido y la conectividad al instante de la lectura. El adaptador interpreta los booleanos, 0 y 1 (también sus equivalentes de texto `"false"`, `"true"`, `"0"`, `"1"`). Una lectura ausente o fuera de estos valores requiere corrección. Un equipo que solamente ofrezca una señal calculada, una máscara de bits o una consulta de estado presente necesita primero un mapeo comprobado con sus datos.

Se consulta la ventana inclusiva en segundos UTC, seleccionando mensajes de datos con posición, y se consumen las páginas restantes en la misma sesión. La clave de cada lectura incluye la unidad y la huella de su contenido canónico. Las relecturas equivalentes conservan su identidad aunque cambie el orden de las propiedades JSON. El protocolo no expone una identidad adicional para dos mensajes idénticos en todo su contenido.

- [Plantilla de peticiones y centros de datos](https://help.wialon.com/en/api/user-guide)
- [Inicio de sesión](https://help.wialon.com/en/api/user-guide/api-reference/token/login)
- [Carga por intervalo](https://help.wialon.com/en/api/user-guide/api-reference/messages/load_interval)
- [Paginación por índice](https://help.wialon.com/en/api/user-guide/api-reference/messages/get_messages)
