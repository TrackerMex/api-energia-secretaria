# Receptor Wialon Retranslator 1.0

El adaptador recibe tramas binarias por conexión TCP. Este piloto limita cada trama completa a **1 MiB**, incluidos los cuatro bytes de longitud. El prefijo anuncia el tamaño del cuerpo en little-endian. El receptor lee el tamaño antes de reservar el búfer de la trama. Los escalares enteros de la cabecera y del bloque de posición se codifican en big-endian y sus tres valores double en little-endian.

Para producir una posición histórica, la trama debe indicar ubicación en el bit `0x1` y contener un bloque `posinfo` válido. La huella SHA-256 cubre la trama completa, incluidos el prefijo y los bloques auxiliares. El flujo de Wialon a un servidor externo es unidireccional. El receptor procesa las tramas consecutivas aunque el emisor nunca lea de la conexión. Un emisor de pruebas puede cerrar su extremo de escritura y esperar EOF antes de consultar el historial.

Las tramas completas inválidas se descartan. Si una conexión termina con una trama parcial o anuncia un tamaño inválido, se cierra esa conexión. Un fallo de almacenamiento se devuelve al comando. La cancelación del comando cierra el listener y las conexiones pendientes.

Con `listen -profile`, la ficha enlaza `retranslator_uid` con el UID autorizado y define `ignition_param` y `connection_param`. La ficha del ensayo usa los bloques sintéticos `ign` y `online`. Los parámetros con nombre admiten 0 y 1 numéricos o las cadenas `0`, `1`, `false` y `true`. Un selector entre `digital:in1` y `digital:in32` lee un bit del bloque entero `avl_inputs` si está presente el indicador de entradas digitales `0x2`. Los valores ausentes, ambiguos o inválidos se guardan como desconocidos. Cada fila conserva los nombres de los selectores para verificar con qué ficha se capturó. Al omitir `-profile` solo se registran los datos de posición.

[Especificación de Wialon Retranslator 1.0](https://extapi.wialon.com/hw/cfg/WialonRetranslator%201.0_en.pdf)
