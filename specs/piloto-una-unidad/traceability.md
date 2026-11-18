---
feature: piloto-una-unidad
status: approved
tags: [harness, spec]
---

# Trazabilidad del piloto

| Requisito | Test (archivo::nombre) | Commit |
|---|---|---|
| R1 | internal/ingest/ingest_test.go::TestR1FetchAllPositionsForOneUnit | 0211527 feat(ingest) consulta paginada |
| R2 | internal/ingest/ingest_test.go::TestR2BuildSENERPosition | 0211527 feat(ingest) estructura y catálogos |
| R3 | internal/ingest/ingest_test.go::TestR3RejectMissingAndUnknownRequiredValues y internal/ingest/digital_test.go::TestR3VerifiedDigitalInput | 0211527 y 8afa4d8 feat(ingest) validación y entradas digitales |
| R4 | internal/delivery/delivery_test.go::TestR4Multipart | 690542d feat(piloto) cliente multipart |
| R5 | internal/delivery/delivery_test.go::TestR5Classify | 690542d feat(piloto) clasificación de resultados |
| R6 | internal/delivery/delivery_test.go::TestR6RestartAndCursor | 690542d feat(piloto) persistencia y cursor |
| R7 | cmd/piloto/main_test.go::TestR7SingleUnitEndToEnd | 690542d feat(piloto) flujo local de punta a punta |

Los casos iniciales quedaron en rojo en `0e26879` y `aed4fff`. El caso específico de entrada digital se agregó en `76204de` antes del cambio de implementación.
