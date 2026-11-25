---
feature: tcp-sener-local
status: approved
tags: [harness, spec]
---

# Trazabilidad de TCP a SENER local

| Requisito | Test (archivo::nombre) | Commit |
|---|---|---|
| R1 | internal/retranslator/signals_test.go::TestR1MappedSignalsBelongToThisFrame, TestR1DigitalBitRequiresMaskAndFlag y TestR1TCPStatesSurviveReopen. internal/unitprofile/profile_test.go::TestR1TCPProfileBindsUIDAndSelectors y cmd/tcp-sener/replay_test.go::TestR1R2ReplayRepairsPreviouslyInvalidReport | c0d6260 conserva las señales y sus selectores. 481365d completa las mediciones compatibles al recibir la misma trama |
| R2 | cmd/tcp-sener/main_test.go::TestR2MissingAndLegacySignals, TestR2ChangedSelectorKeepsInvalidUntilCorrected y TestR2RejectExternalEndpointAndForeignQueue. internal/historydb/signals_test.go::TestR1R2LegacyReaderThenMigrationPreservesUnknown. cmd/tcp-sener/replay_test.go::TestR1R2ReplayRepairsPreviouslyInvalidReport. TestR5TCPToSENERLocal comprueba el multipart recibido | c0d6260, ebd493b y 481365d |
| R3 | cmd/tcp-sener/main_test.go::TestR3ReprocessAndLateArrival y TestR5TCPToSENERLocal | ebd493b preserva las identidades de un mismo segundo, incorpora atrasadas y admite ventanas no contiguas |
| R4 | cmd/tcp-sener/main_test.go::TestR4DeliveryOutcomes | ebd493b conserva aceptación, rechazo, límite 429 y resultado incierto tras reabrir la cola |
| R5 | cmd/tcp-sener/main_test.go::TestR5TCPToSENERLocal | ebd493b y c0d6260 integran los ejecutables, la captura, la persistencia y el simulador |

Los casos iniciales del puente y de captura quedaron en rojo en `6aefb79` y `282c97c`. El caso de retransmisión de una posición heredada se reprodujo en `f205b90` y `62b832e` y quedó corregido en `481365d`. La prueba de operador está documentada en `5a7d12f`. [Informe de implementación](../../progress/impl_tcp-sener-local.md)
