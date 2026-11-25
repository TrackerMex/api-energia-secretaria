---
feature: observacion-api-cr13
status: draft
tags: [harness, spec]
---

# Diseño del modo de observación

El comando compone la fuente existente `internal/ingest.Source` y comunica sus resultados mediante la interfaz de lectura de posiciones del dominio. Usa el ID interno autorizado para la consulta, la paginación y el cierre de la sesión que ya soporta ese adaptador. Los límites son segundos UTC inclusivos. El comando valida la configuración antes de abrir la sesión y publica el resumen al completar la consulta.

El clasificador toma `f`, `i` y `p.can_rpm` del mismo mensaje de posición. Un valor de `i` igual a cero es una lectura válida si existe la bandera de entradas y la máscara es válida. Las RPM requieren un valor numérico finito y no negativo. Cualquier valor reservado como desconocido por el perfil técnico documentado del equipo se considera inválido. La velocidad, las lecturas de otro instante y los intervalos calculados de ignición no forman parte del cálculo. La representación de las señales conserva la diferencia entre ausencia y cero. [Formato de mensajes de Wialon](https://help.wialon.com/en/api/user-guide/data-format/messages)

El caso de uso reúne los conteos en memoria y utiliza motivos distinguibles para desacuerdo y para señales ausentes o inválidas. Si un mensaje tiene varios defectos, la salida debe explicar la convención de conteo de los motivos. Una ventana vacía conserva los conteos en cero. El comando compone exclusivamente lectura y observación. [Arquitectura](../../docs/architecture.md) · [Criterio acordado para CR-13](../../internal/ingest/README.md)
