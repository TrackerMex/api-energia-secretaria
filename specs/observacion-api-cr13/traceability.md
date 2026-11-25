---
feature: observacion-api-cr13
status: approved
tags: [harness, spec]
---

# Trazabilidad de la observación por Remote API

| Requisito | Test (archivo::nombre) | Commit |
|---|---|---|
| R1 | cmd/observar/main_test.go::TestR1InclusiveSecondAndFailedPage y TestR4ObservationReadOnly | 2b0be7e selecciona la unidad, valida segundos UTC inclusivos y espera todas las páginas |
| R2 | internal/domain/observation_test.go::TestR2ObservationClassifiesSameMessage | d618182 clasifica cada mensaje con la entrada digital y las RPM de esa posición |
| R3 | internal/domain/observation_test.go::TestR3ObservationCountsPositionsAndExplainsReasons y TestR3EmptyObservationKeepsStableJSON. cmd/observar/main_test.go::TestR3PrivateFailuresAndRedirect | d618182 y 2b0be7e agregan resultados y omiten datos individuales de las salidas |
| R4 | cmd/observar/main_test.go::TestR4ObservationReadOnly | 2b0be7e compone exclusivamente la lectura Wialon y la clasificación de dominio |

Los tests previos a la implementación están en `716648a` para el comando y `106996f` para el clasificador. [Informe de implementación](../../progress/impl_observacion-api-cr13.md)
