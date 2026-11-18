# Receptor Wialon Retranslator 1.0

El adaptador recibe tramas binarias por conexión TCP. Este piloto limita cada trama completa a **1 MiB**, incluidos los cuatro bytes de longitud. El prefijo anuncia el tamaño del cuerpo en little-endian. El receptor lee el tamaño antes de reservar el búfer de la trama. Los escalares enteros de la cabecera y del bloque de posición se codifican en big-endian y sus tres valores double en little-endian.

Para producir una posición histórica, la trama debe indicar ubicación en el bit `0x1` y contener un bloque `posinfo` válido. La huella SHA-256 cubre la trama completa, incluidos el prefijo y los bloques auxiliares. El flujo de Wialon a un servidor externo es unidireccional. El receptor procesa las tramas consecutivas aunque el emisor nunca lea de la conexión. Un emisor de pruebas puede cerrar su extremo de escritura y esperar EOF antes de consultar el historial.

Las tramas completas inválidas se descartan. Si una conexión termina con una trama parcial o anuncia un tamaño inválido, se cierra esa conexión. Un fallo de almacenamiento se devuelve al comando. La cancelación del comando cierra el listener y las conexiones pendientes.

[Especificación de Wialon Retranslator 1.0](https://extapi.wialon.com/hw/cfg/WialonRetranslator%201.0_en.pdf)
