# Piloto Wialon–SENER en Go

El piloto lee las posiciones de una unidad de Wialon Remote API, crea el reporte de SENER y registra cada entrega en SQLite. Se eligió CR-13 para una futura prueba real. La configuración incluida es sintética. El servidor de Hostinger es el destino previsto para una fase posterior.

## Probar en local

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
