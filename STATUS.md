# Piloto Wialon–SENER

**Última actualización**: 2026-11-18
**Features completadas**: 1/2 (`feature_list.json`)
**Pendientes**: borrador de requisitos del receptor TCP para CR-13 y su historial local
**En producción**: no

## Estado actual

El piloto de una unidad está implementado y aprobado en local. La verificación usó Wialon y SENER simulados en loopback, SQLite temporal y datos ficticios. Se probaron paginación, mapeo, validación, multipart, recuperación tras reinicio y repetición de una ventana ya aceptada.

La siguiente fase elegida es recibir Wialon Retranslator por TCP y guardar el historial de CR-13. Se desarrollará primero con paquetes sintéticos en este equipo.

## Última sesión

El usuario eligió CR-13 y aprobó los requisitos. La revisión independiente verificó la entrega `5ba50b0`; los detalles están en [el informe](progress/review_piloto-una-unidad.md). El operador decidió continuar únicamente en local. La configuración de despliegue está preparada para revisión posterior.

## Para retomar la prueba real

Confirmar las credenciales y ficha privadas, la fuente verificable de estado de conexión GPS, la IP de salida autorizada y la frecuencia operativa acordada. En el VPS el almacenamiento previsto es SQLite en `/var/lib/wialon-sener/pilot.db`. La prueba local se reproduce desde el [README](README.md).
