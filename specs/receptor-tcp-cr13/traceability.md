---
feature: receptor-tcp-cr13
status: approved
tags: [harness, spec]
---

# Trazabilidad del receptor TCP

| Requisito | Test (archivo::nombre) | Commit |
|---|---|---|
| R1 | cmd/receptor/main_test.go::TestR1LocalOnly y cmd/receptor/main_test.go::TestR6TCPHistoryLocal | cfdbef9 valida la configuración y abre el puerto de loopback |
| R2 | internal/retranslator/retranslator_test.go::TestR2OfficialPacket, TestR2FragmentedAndCoalesced y TestR2OneWaySenderClosesWithoutReading. cmd/receptor/main_test.go::TestR6TCPHistoryLocal comprueba la exclusión de otro UID | ab903a0, d69b921 y cfdbef9 |
| R3 | internal/historydb/history_test.go::TestR3PersistAndDeduplicate y cmd/receptor/main_test.go::TestR6TCPHistoryLocal | cfdbef9 conserva la identidad y evita duplicados después de reabrir SQLite |
| R4 | internal/historydb/history_test.go::TestR4GPSOrderAndHalfOpenRanges, TestR3PersistAndDeduplicate y cmd/receptor/main_test.go::TestR6TCPHistoryLocal | cfdbef9 ordena por hora GPS y huella y consulta el intervalo UTC |
| R5 | internal/retranslator/retranslator_test.go::TestR5InvalidAndNonPositionFrames, TestR5PeerIsolationAndIdleShutdown y TestR5StorageFailureIsReturned. cmd/receptor/main_test.go::TestR6TCPHistoryLocal comprueba la persistencia del punto previo | ab903a0, d69b921 y cfdbef9 |
| R6 | cmd/receptor/main_test.go::TestR6TCPHistoryLocal | cfdbef9 conecta ambos adaptadores por TCP local, reabre el historial y verifica la retransmisión |

Las pruebas iniciales del protocolo quedaron en rojo en `1cdb3ea`. Las del comando, la persistencia y la integración quedaron en rojo en `433f581`. El flujo unidireccional y la comprobación de la máscara de ubicación se incorporaron primero en `df06a87`.
