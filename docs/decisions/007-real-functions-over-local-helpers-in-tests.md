# ADR-007: Llamar funciones reales en tests de DB (no helpers locales)

**Fecha:** 2026-03-26
**Estado:** Aceptado

---

## Contexto

Los tests de `packages/db/src/__tests__/` del Plan 5 F3 usaban el **patrón de local helpers**:
en lugar de importar y llamar las query functions del archivo a testear, cada test replicaba
la lógica usando `testDb` directamente.

```typescript
// Patrón anterior (local helpers) — EVITAR
async function save(data: ...) {
  const [row] = await testDb.insert(schema.savedResponses).values({...}).returning()
  return row!
}
// El test llama save() que es una copia de saveResponse() — no el código real
```

**El problema:** este patrón crea una falsa garantía. Si se cambia la firma o la lógica de
`saveResponse()` en `saved.ts`, el test sigue pasando porque llama al helper local, no a la
función de producción. Además, los query files no aparecen en el reporte de coverage porque
no son importados.

**Bug descubierto:** al implementar este ADR, `removeTag` en `tags.ts` fue identificado como
buggy — no filtraba por `tag` en el WHERE, borrando TODOS los tags de la sesión. El patrón
de helpers ocultó este bug porque el helper reimplementaba la lógica correctamente.

## Decisión

**Llamar las funciones reales importadas del query file en todos los tests de DB.**

Para resolver el problema de que `getDb()` es un singleton que los query files capturan a
nivel de módulo (haciendo imposible inyectar un testDb), se toman dos medidas:

### 1. Mover `getDb()` dentro de cada función de query

```typescript
// INCORRECTO (nivel módulo)
const db = getDb()
export async function setMemory(...) {
  await db.insert(...)
}

// CORRECTO (dentro de la función)
export async function setMemory(...) {
  const db = getDb()
  await db.insert(...)
}
```

Esto permite que `_injectDbForTesting()` (ADR-006) funcione: cuando el test inyecta `testDb`
antes de ejecutar, la función llama a `getDb()` y obtiene el `testDb` inyectado.

### 2. Patrón estándar de tests de DB

```typescript
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { setMemory, getMemory } from "../queries/memory"  // función real
import { createTestDb, initSchema, insertUser } from "./setup"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)   // ← singleton apunta a testDb
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM user_memory; DELETE FROM users;")
})

test("setMemory guarda y getMemory recupera", async () => {
  const user = await insertUser(db)
  await setMemory(user.id, "idioma", "español")  // función real
  const entries = await getMemory(user.id)         // función real
  expect(entries[0]!.value).toBe("español")
})
```

## Consecuencias

**Positivas:**
- Coverage real de los query files — si una función no está testeada, se ve en el reporte
- Si se cambia la firma de una función de producción, el test falla inmediatamente
- Tests más cortos — no hay helpers que dupliquen la lógica
- Bugs en el código de producción se detectan (como el de `removeTag`)

**Negativas / trade-offs:**
- Requiere que los query files no llamen a `getDb()` a nivel de módulo (4 archivos tenían
  este patrón: `areas.ts`, `users.ts`, `sessions.ts`, `events.ts` — todos corregidos)
- Los tests son ligeramente más acoplados a la implementación real — si cambia la API de
  una función, hay que actualizar tanto la implementación como los tests

## Archivos afectados por esta decisión

**Query files corregidos** (getDb movido dentro de funciones):
- `packages/db/src/queries/areas.ts`
- `packages/db/src/queries/users.ts`
- `packages/db/src/queries/sessions.ts`
- `packages/db/src/queries/events.ts`

**Bug corregido:**
- `packages/db/src/queries/tags.ts` — `removeTag` ahora filtra por `tag` además de `sessionId`

**Test infrastructure:**
- `packages/db/src/__tests__/setup.ts` — SQL completo del schema + helpers `insertUser`/`insertSession`/`insertMessage`

**Tests reescritos:**
- Los 17 archivos de `packages/db/src/__tests__/*.test.ts`

## Referencias

- ADR-006: estrategia de testing general
- `packages/db/src/connection.ts` — `_injectDbForTesting()`, `_resetDbForTesting()`
- `packages/db/src/__tests__/setup.ts` — helpers de test compartidos
