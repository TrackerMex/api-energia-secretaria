# Receptor Wialon Retranslator 1.0

El adaptador recibe tramas binarias por conexión TCP. Este piloto limita cada trama completa a **1 MiB**, incluidos los cuatro bytes de longitud. El prefijo anuncia el tamaño del cuerpo en little-endian. El receptor lee el tamaño antes de reservar el búfer de la trama. Los escalares enteros de la cabecera y del bloque de posición se codifican en big-endian y sus tres valores double en little-endian.

Una trama con un bloque `posinfo` produce una posición histórica. La huella SHA-256 cubre la trama completa, incluidos el prefijo y los bloques auxiliares. Después de procesar cada trama válida se devuelve el byte `0x11`. El emisor de pruebas puede esperar ese byte antes de consultar el historial. El protocolo admite la retransmisión a un servidor externo sin esperar respuesta.

Las tramas completas inválidas se descartan. Si una conexión termina con una trama parcial o anuncia un tamaño inválido, se cierra esa conexión. Un fallo de almacenamiento se devuelve al comando. La cancelación del comando cierra el listener y las conexiones pendientes.

[Especificación de Wialon Retranslator 1.0](https://extapi.wialon.com/hw/cfg/WialonRetranslator%201.0_en.pdf)
