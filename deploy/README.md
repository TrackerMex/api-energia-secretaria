# VPS de Hostinger

El proceso usa HTTPS saliente para consultar Wialon y para enviar a SENER desarrollo. El receptor y los demás proyectos del VPS no comparten una base de datos.

| Ruta en el VPS | Contenido y acceso previsto |
|---|---|
| `/opt/wialon-sener/piloto` | Binario Go compilado para la arquitectura del VPS. |
| `/etc/wialon-sener/profile.json` | Ficha y parámetros de la unidad piloto. Propietario root, grupo del servicio, modo 0640. |
| `/etc/wialon-sener/pilot.env` | Tokens de desarrollo, URL de Wialon, ID de la unidad, instante inicial y frecuencia del piloto. Modo 0600, propietario root. El gestor systemd lee este archivo. |
| `/var/lib/wialon-sener/pilot.db` | Base SQLite con reportes, estado de entrega y cursor. El servicio escribe en un directorio dedicado de modo 0700. Los archivos auxiliares de SQLite viven junto a la base. |
| Journal de systemd | Resultado operativo. Los registros de diagnóstico deben excluir coordenadas, cuerpos de reportes, JWT y tokens de sesión. |

El usuario local de servicio y los administradores autorizados del VPS tienen acceso al estado. SQLite por sí solo no cifra el archivo. Una copia o instantánea del VPS puede incluir esos datos según la configuración de Hostinger. Antes de operar con la flota se acuerdan el acceso a respaldos y la retención de datos. Mantener juntos el estado y el cursor en cada copia consistente de SQLite.

## Primera ejecución en desarrollo

1. Confirmar el host y usuario SSH, arquitectura Linux, soporte de systemd, espacio libre, memoria disponible e IP de salida. Confirmar la autorización de esa IP para SENER desarrollo y el acceso de lectura de la unidad en Wialon.
2. Instalar el binario en la ruta indicada y preparar el usuario de sistema `wialon-sener`. Crear un directorio de configuración privado y el directorio de estado propiedad de ese usuario.
3. Completar en el servidor el perfil de CR-13 a partir de `internal/ingest/profile.example.json`. Verificar contra la ficha del permisionario el RFC, razón social, permiso, VIN, placa, IMEI y códigos de carga y vehículo. Confirmar las señales de encendido y conexión para el intervalo elegido.
4. Definir `WIALON_TOKEN`, `WIALON_UNIT_ID`, `WIALON_BASE_URL` y `SENER_DEV_TOKEN` en el entorno del proceso. Para el servicio periódico, definir también `WIALON_START_UTC` y `WIALON_POLL_INTERVAL`. Guardar los valores reales únicamente en la configuración privada del servidor.
5. Ejecutar una pasada acotada con el usuario de servicio, la configuración de desarrollo y fechas UTC explícitas. Ejemplo de la invocación con variables de ventana del operador:

```sh
/opt/wialon-sener/piloto -environment development \
  -profile /etc/wialon-sener/profile.json \
  -db /var/lib/wialon-sener/pilot.db \
  -from "$INICIO_UTC" -to "$FIN_UTC"
```

6. Revisar el resultado por transmisión y los errores de validación antes de habilitar la ejecución periódica. Los reportes de resultado incierto requieren conciliación. La unidad systemd de este directorio usa el mismo perfil y la misma base que la corrida acotada.

La conexión SSH permite administrar el servidor. La lectura de telemetría empieza al ejecutar el piloto con autorización y las credenciales disponibles. Una entrega aceptada conserva el registro local y deja una copia del reporte en SENER desarrollo.
