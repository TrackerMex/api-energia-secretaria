# Piloto Wialon–SENER en Go

El piloto lee las posiciones de una unidad de Wialon Remote API, crea el reporte de SENER y registra cada entrega en SQLite. Se eligió CR-13 para una futura prueba real. La configuración incluida es sintética. El servidor de Hostinger es el destino previsto para una fase posterior.

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

La fase TCP conserva el historial de una unidad. El emisor a SENER sigue tomando sus datos del piloto con Remote API. [Requisitos del receptor](specs/receptor-tcp-cr13/requirements.md)

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
