---
feature: receptor-tcp-cr13
status: draft
tags: [harness, spec]
---

# Diseño del receptor TCP de CR-13

## Entrada y alcance del protocolo

El operador inicia un comando Go con el UID que Wialon retransmite, el archivo SQLite y la dirección de loopback. El UID viaja en cada trama y puede diferir del nombre visible CR-13 o del identificador interno de Remote API. La comparación debe realizarse con el UID configurado antes de escribir una posición. R1–R2. [Especificación de Wialon Retranslator 1.0](https://extapi.wialon.com/hw/cfg/WialonRetranslator%201.0_en.pdf)

El adaptador TCP lee la longitud completa antes de reservar memoria, aplica un máximo documentado, lee exactamente la carga y recorre sus bloques dentro de los límites declarados. Una lectura TCP puede contener parte de una trama o varias. Interpretar el UID, el tiempo y el bloque de posición según la especificación. Un bloque de telemetría sin posición no produce una fila de coordenadas. El cierre del receptor interrumpe las lecturas pendientes. R2, R4–R5.

## Persistencia y consulta

El historial usa un archivo SQLite local con ruta explícita y el controlador Go ya usado en el piloto. La huella del contenido binario completo, con unicidad en SQLite, permite reconocer una retransmisión exacta incluso si dos conexiones la entregan a la vez. La hora del GPS se conserva con su precisión original. El índice de consulta cubre la unidad y la hora GPS e incluye un desempate estable para tramas distintas del mismo segundo. Cada inserción confirmada sobrevive al cierre y a la reapertura de la base. R3–R5.

La interfaz de consulta lee el archivo local y muestra la última posición o un intervalo semiabierto en UTC. El modo de consulta debe distinguir un intervalo vacío de un error de parámetros o de almacenamiento. El README de implementación incluirá las invocaciones reproducibles para escuchar y consultar. R4, R6.

## Capas previstas

| Ubicación | Responsabilidad |
|---|---|
| `internal/domain/` | Definición de la posición histórica y del contrato de almacenamiento necesario para su captura y lectura. |
| `internal/application/` | Reglas para aceptar posiciones de la unidad configurada y resolver consultas. |
| `internal/`, adaptadores TCP y SQLite | Lectura y validación de la trama, escritura transaccional y consultas a la base. |
| `cmd/`, comando local | Selección de modo, validación de configuración, escucha, consulta y cierre. |
| Tests del nuevo comando y sus adaptadores | Paquetes sintéticos con resultados conocidos, conexiones fragmentadas y consecutivas, consultas y reinicio. |

La composición del nuevo comando será independiente de las credenciales, el cliente de Remote API y la cola de envíos de SENER. La prueba de aceptación se ejecutará con una base nueva y un puerto de loopback asignado por el sistema. R1, R6.

## Alternativas consideradas

El piloto ya cuenta con el controlador SQLite. Un servicio PostgreSQL aparte exigiría administrar otro proceso para la prueba de una sola unidad. Guardar directamente en la cola SENER mezclaría la observación recibida con su estado de entrega y exigiría la ficha regulatoria para consultar una posición.
