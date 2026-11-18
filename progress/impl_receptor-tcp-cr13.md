# Implementación local del receptor de CR-13

El comando `cmd/receptor` ofrece `listen`, `latest` e `history`. La captura compara el UID antes de guardar y vincula el archivo SQLite con el UID y el alias CR-13. Las consultas abren ese archivo en modo de solo lectura. La hora del GPS y la huella determinan el orden. Una restricción única conserva una sola copia de cada trama exacta.

El receptor acepta Wialon Retranslator 1.0 por loopback. El emisor sintético puede cerrar su lado de escritura y terminar el flujo sin esperar una respuesta. La prueba de extremo a extremo usa un paquete independiente, un puerto asignado por el sistema, una base temporal y credenciales vacías. Comprueba la recepción del UID permitido, la exclusión de otro UID, la consulta con el receptor en marcha, la recuperación al reabrir y la retransmisión. También preserva la ubicación previa al recibir un mensaje sin posición y una trama incompleta.

La verificación completa del implementer sobre el código de `cfdbef9`, que integra el decodificador de `d69b921` y los ejemplos de `7f4d13f`, terminó con código 0. Una ejecución de `./init.sh` compiló en 2 s, ejecutó la suite de siete paquetes con pruebas en 5 s y completó `go vet ./...` sin errores. Se deshabilitó la descarga de módulos para usar únicamente dependencias locales.

La prueba se puede repetir con `go test -count=1 -v ./cmd/receptor -run '^TestR6TCPHistoryLocal$'`. La guía aporta un segundo vector independiente en PowerShell. [Requisitos y evidencia por prueba](../specs/receptor-tcp-cr13/traceability.md)
