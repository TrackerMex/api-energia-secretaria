# Piloto Wialon–SENER en Go

Este repositorio prepara una prueba con una sola unidad. El servicio previsto consultará Wialon Remote API desde el VPS de Hostinger y enviará reportes de posición al ambiente de desarrollo de SENER mediante HTTPS saliente.

## Prueba local

Con Go, Bash y Node o Python 3 instalados, ejecuta desde la raíz:

```sh
./init.sh --harness-only
go test -v ./test/pilot
./init.sh
```

La prueba verifica las muestras de entrada y salida de una sola unidad sintética. Se ejecuta sin credenciales. La integración de extremo a extremo y su prueba de reinicio corresponden a la [primera especificación](specs/piloto-una-unidad/requirements.md).

## Prueba con una unidad real

1. Elegir su ID interno en Wialon y comprobar permisos de lectura de mensajes y disponibilidad de la señal de ignición. Asociar el ID con la ficha del vehículo y el permisionario.
2. Identificar el VPS y su IP pública de salida. Comprobar capacidad disponible y si SENER admite esa IP para el token de desarrollo. El servicio sólo requiere conexiones salientes.
3. Configurar los secretos en el entorno de ejecución. Probar un intervalo histórico acotado y revisar en desarrollo tanto el resultado HTTP como el campo `error`.
4. Antes de activar una consulta periódica, confirmar con el cliente o SENER la demora admisible y las reglas de reenvío. Registrar la hora GPS, la hora de lectura y la de aceptación permite medirla.

[Frecuencia y alcance de los manuales](docs/frecuencia-sener.md) · [Estado del proyecto](STATUS.md)
