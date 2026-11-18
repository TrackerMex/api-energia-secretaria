# Convenciones del piloto

Go con `gofmt`, pruebas con `testing` y servidores locales con `httptest`. Los tests de funcionalidad se identifican por requisito, por ejemplo `TestR1UnaUnidad`. La validación inicial de los archivos sintéticos se ejecuta con `go test -v ./test/pilot`.

Usamos UTC para tiempos del GPS y de entrega. Los identificadores de unidad de Wialon y los IMEI tienen significado distinto. Conservamos VIN, RFC, IMEI y placas como texto. Cada señal debe tener una fuente documentada para la unidad piloto.

## Configuración prevista para la primera feature

| Variable | Uso |
|---|---|
| `WIALON_BASE_URL` | URL HTTPS del centro de datos que atiende la cuenta. |
| `WIALON_UNIT_ID` | ID interno de la única unidad autorizada para el piloto. |
| `WIALON_TOKEN` | Token de lectura para esa unidad. |
| `SENER_DEV_TOKEN` | JWT del ambiente de desarrollo. |
| `WIALON_POLL_INTERVAL` | Tiempo entre consultas. Su valor es una decisión operativa del piloto. |

El arranque local del harness y los tests usan datos ficticios. Para la prueba real se definen los secretos fuera del repositorio. Los logs de aplicación deben registrar el ID de transmisión y el resultado, con las credenciales y los datos personales excluidos.
