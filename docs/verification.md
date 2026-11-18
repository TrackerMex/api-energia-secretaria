# Verification — Cómo demostrar que una feature funciona

> El implementer entrega una feature verificada. El reviewer comprueba por su
> cuenta la misma versión antes de aprobarla.

---

## Qué ejecutar en cada momento

| Momento | Responsable | Verificación |
|---|---|---|
| Arranque de sesión | Leader | `./init.sh --harness-only` comprueba el entorno, las specs, la trazabilidad y el conteo de features. |
| Rojo, verde y refactor | Implementer | Ejecuta los tests del requisito y de los módulos afectados. |
| Entrega de la feature integrada | Implementer | Ejecuta una vez `./init.sh` sin argumentos. En el reporte registra el resultado del build y la suite de esa ejecución. |
| Revisión | Reviewer | Ejecuta personalmente `./init.sh` sin argumentos y registra su propio resultado antes de aprobar. |
| Cierre administrativo | Leader | Tras actualizar `feature_list.json` y `STATUS.md`, ejecuta `./init.sh --harness-only`. |

Configura en `init.config.sh` el build y la suite completa del proyecto. El
init completo mide instalación, build y tests por separado. Si faltan
dependencias para empezar el ciclo TDD, instálalas con el comando del proyecto.

Por ejemplo, con Jest puedes ejecutar `pnpm exec jest --runTestsByPath <archivo>`
desde el paquete de Nest o el de Expo. Si cambia el contrato de un endpoint,
incluye tanto las pruebas del servidor como las del cliente que lo consume.
Si hay cambios de código, tests o configuración después de la revisión,
verifica y revisa de nuevo la versión que se va a entregar.

---

## Disciplina TDD

Esta es la regla central de verificación de este harness, no una sugerencia:

1. **El implementer escribe el test ANTES que el código.** Por cada requisito
   `R<n>` de `specs/<feature>/requirements.md`: primero un test que falla
   (rojo), después la implementación mínima que lo pasa (verde), después
   refactor manteniendo los tests en verde. Esto es exactamente el checklist
   de `specs/<feature>/tasks.md`.
2. **Cada test nombra su requisito.** Convención: `describe('R1: ...', ...)`
   o el equivalente idiomático del framework de test de `Go`
   (documentado en `docs/conventions.md`).
3. **El reviewer verifica en el diff/historia** que los tests existen y
   nombran los R-ids correctos — no le basta con que "algo" tenga cobertura.
   Si una feature no tiene tests que nombren sus requisitos, el reviewer
   **rechaza**, sin excepción.
4. **`specs/<feature>/traceability.md` se actualiza en cada commit**, nunca
   al final: cada fila pasa de "pendiente" a `test::nombre` + `hash commit`
   tan pronto ese requisito queda verde.

---

## Verificación por feature (a rellenar por el proyecto)

<!-- Cuando el proyecto tenga features concretas, añade aquí un bloque de
     verificación manual por feature (curl, comandos de CLI, pasos de UI),
     análogo a:

### Feature <id> — <nombre>

```bash
# pasos de verificación manual específicos de esta feature
```
-->

---

## Notas para el implementer

- No declares done solo porque el build pasa. Prueba los casos edge (errores,
  permisos, condiciones límite).
- Los scripts de verificación manual (curl, CLI, etc.) no sustituyen a los
  tests automáticos exigidos por la disciplina TDD.
- Si alguna verificación falla, documenta por qué en `progress/impl_<feature>.md`
  antes de reportar al leader.
