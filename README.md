# Piloto Wialon–SENER en Go

El piloto en Go recibe posiciones de Wialon Retranslator o las consulta con Remote API. La prueba integrada usa el UID ficticio `DEMO-CR13`, una ficha de ejemplo, SQLite y un simulador de SENER en el mismo equipo. CR-13 es la unidad elegida para una futura prueba con datos reales.

## Prueba completa de TCP a SENER simulado

Desde la raíz del repositorio, ejecuta en PowerShell:

```powershell
go test -count=1 -v ./cmd/tcp-sener -run '^TestR5TCPToSENERLocal$'
```

La prueba abre sus propios puertos de loopback y bases SQLite temporales. Entrega una posición ficticia al receptor TCP, procesa el intervalo del historial y comprueba el archivo JSON recibido por el simulador. Tras reiniciar y retransmitir, el resultado esperado es **una posición, un reporte aceptado y un solo POST**.

La ficha de captura identifica el UID y las señales históricas de ignición y conexión del GPS. Un cero medido es válido. Una lectura desconocida o ausente deja la posición en el historial y registra el motivo en la cola. Las posiciones antiguas que carecen de estas mediciones siguen siendo consultables. Si la trama original vuelve a llegar con una ficha que permite leer las señales faltantes, vuelve a procesar su intervalo. La cola podrá usar el mismo identificador del reporte para la primera entrega válida. [Ficha del ensayo](testdata/tcp-sener/profile.json) · [Requisitos de la prueba integrada](specs/tcp-sener-local/requirements.md)

Para procesar un intervalo con un simulador HTTP propio ya en marcha, inicia la captura con `-profile` y usa la misma ficha al procesar el historial. En una segunda terminal, ajusta el puerto del simulador y las dos fechas al ensayo. El ejemplo utiliza archivos distintos para la captura y para la cola.

```powershell
go run ./cmd/receptor listen -uid DEMO-CR13 -profile ./testdata/tcp-sener/profile.json -db ./state/tcp-sener-historial.db -listen 127.0.0.1:20163
```

```powershell
$simulador = 'http://127.0.0.1:18081/api/v1/telemetria/reporte/'
go run ./cmd/tcp-sener -uid DEMO-CR13 -profile ./testdata/tcp-sener/profile.json -history ./state/tcp-sener-historial.db -outbox ./state/tcp-sener-cola.db -from 2026-11-25T00:00:00Z -to 2026-11-26T00:00:00Z -endpoint $simulador
```

El comando muestra los conteos acumulados de la cola. Si llegan posiciones atrasadas, vuelve a procesar su intervalo. Los reportes aceptados conservan su identificador y una respuesta incierta queda pendiente de conciliación.

## Receptor TCP de CR-13

El receptor guarda las posiciones que llegan por Wialon Retranslator 1.0 en el archivo SQLite que indiques con `-db`. El siguiente ejemplo usa `state/cr13-demo.db`, el alias CR-13 y el UID ficticio `DEMO-CR13`. El receptor escucha en el propio equipo y el historial se puede consultar aun después de detenerlo. [Especificación del protocolo](https://extapi.wialon.com/hw/cfg/WialonRetranslator%201.0_en.pdf)

Desde la raíz del proyecto, inicia el receptor en una ventana de PowerShell.

```powershell
New-Item -ItemType Directory -Force ./state | Out-Null
go run ./cmd/receptor listen -uid DEMO-CR13 -db ./state/cr13-demo.db -listen 127.0.0.1:20163
```

En una segunda ventana, desde la misma carpeta, este emisor envía una posición **ficticia** del 18 de noviembre a las 12:00 UTC. Puedes ejecutarlo dos veces para comprobar que la retransmisión exacta produce una sola fila.

```powershell
$trama = [Convert]::FromBase64String('PwAAAERFTU8tQ1IxMwBq/ZNAAAAAAQu7AAAAJwACcG9zaW5mbwAAAAAAAAAEQAAAAAAAAPQ/AAAAAAAAJEAADABaBw==')
$cliente = [Net.Sockets.TcpClient]::new('127.0.0.1', 20163)
try {
    $cliente.GetStream().Write($trama, 0, $trama.Length)
} finally {
    $cliente.Dispose()
}
```

Consulta la última posición o el historial por hora del GPS. El intervalo incluye la hora inicial y excluye la final. La posición de ejemplo tiene latitud `1.25`, longitud `2.5`, velocidad de `12 km/h` y rumbo de `90°`.

```powershell
go run ./cmd/receptor latest -uid DEMO-CR13 -db ./state/cr13-demo.db
go run ./cmd/receptor history -uid DEMO-CR13 -db ./state/cr13-demo.db -from 2026-11-18T00:00:00Z -to 2026-11-19T00:00:00Z
```

Para detener la escucha, usa Ctrl+C en la primera ventana. Las consultas leen el mismo archivo cuando el receptor está detenido. La prueba automatizada usa un puerto de loopback y una base temporal propios.

```powershell
go test -count=1 -v ./cmd/receptor -run '^TestR6TCPHistoryLocal$'
```

Esta trama de ejemplo sirve para consultar el historial de posiciones. Para probar el envío local usa el ensayo integrado de la primera sección, que incluye las dos señales. [Requisitos del receptor](specs/receptor-tcp-cr13/requirements.md)

## Probar el envío a SENER en local

Desde la raíz del repositorio, con Go instalado:

```sh
go test -count=1 -v ./cmd/piloto -run '^TestR7SingleUnitEndToEnd$'
```

Este test abre dos servidores TLS de loopback y una base SQLite temporal. Lee una posición ficticia, inspecciona el archivo JSON adjunto, simula una aceptación y repite la ventana para verificar que la misma posición genera un solo envío. Los tests de almacenamiento y fallos se ejecutan con `go test -count=1 -v ./internal/delivery`. El propio Go puede descargar las dependencias del módulo en la primera ejecución.

Con Bash y Node o Python 3 puedes ejecutar `./init.sh` para compilar, probar y analizar el proyecto. En las sesiones de coordinación, `./init.sh --harness-only` comprueba el estado del harness.

## Datos y operación

Los tests usan el directorio temporal de cada prueba. El servicio del VPS está diseñado para guardar sus reportes, el avance de lectura y los resultados en `/var/lib/wialon-sener/pilot.db`. La ficha privada y las credenciales de ejecución pertenecen a `/etc/wialon-sener/`. Los reportes se forman en memoria al preparar el multipart y el cuerpo se conserva en la base para entrega y conciliación.

El modo de desarrollo requiere la ficha de una unidad y selección explícita de entorno. Las fechas del modo de una pasada delimitan el histórico a consultar. El modo periódico usa un cursor, con solapamiento en el último segundo leído. Si se incorporan mensajes históricos más antiguos tras un corte del GPS, se puede volver a consultar esa ventana con la misma base de datos.

[Configuración del VPS y almacenamiento](deploy/README.md) · [Ficha y señales de la unidad](internal/ingest/README.md) · [Análisis de la frecuencia de reporte](docs/frecuencia-sener.md) · [Especificación aprobada](specs/piloto-una-unidad/requirements.md)
