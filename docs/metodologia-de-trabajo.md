# Metodología de trabajo — RAG Saldivia

> Documento de referencia portátil. Cubre el workflow completo para aplicar en cualquier repo.
> Última actualización: 2026-03-26

---

## Índice

1. [Loop OODA-SQ — el workflow central](#1-loop-ooda-sq)
2. [Brainstorming antes de implementar](#2-brainstorming)
3. [Escritura de planes](#3-escritura-de-planes)
4. [Ejecución de planes](#4-ejecución-de-planes)
5. [TDD — desarrollo guiado por tests](#5-tdd)
6. [Testing por capas](#6-testing-por-capas)
7. [Debugging sistemático](#7-debugging-sistemático)
8. [Git y CHANGELOG](#8-git-y-changelog)
9. [Verificación antes de declarar "terminado"](#9-verificación-antes-de-declarar-terminado)
10. [ADRs — decisiones de arquitectura](#10-adrs)
11. [Ciclo completo para una feature nueva](#11-ciclo-completo-para-una-feature-nueva)

---

## 1. Loop OODA-SQ

Todo cambio no trivial sigue este loop completo:

```
OBSERVE → ORIENT → DECIDE → ACT
                              └── Implement (TDD) → Simplify → Review → Docs
```

### Fases obligatorias en orden

| Gate | Fase | Qué hacer |
|------|------|-----------|
| 1 | **Observe** | Leer archivos críticos, entender estado actual |
| 2 | **Orient** | Brainstorming — explorar opciones, entender el problema |
| 3 | **Decide** | Elegir enfoque, escribir el plan |
| 4 | **Implement** | TDD directo. Si algo falla → debugger sistemático |
| 5 | **Simplify** | Dead code check, eliminar lo que sobra |
| 6 | **Review** | Code review del área afectada |
| 7 | **Docs** | Actualizar CHANGELOG.md y documentación relevante |

**Para cambios triviales (≤3 líneas):** solo gates 4 + 5 + 7.

---

## 2. Brainstorming

**Cuándo:** antes de cualquier trabajo creativo — feature nueva, componente, cambio de comportamiento, refactor. No hay excepción por "simplicidad".

### Proceso

1. **Explorar contexto del proyecto** — leer archivos, docs, commits recientes
2. **Hacer preguntas clarificadoras** — de a una por mensaje, no en bloque
3. **Proponer 2-3 enfoques** con trade-offs y recomendación
4. **Presentar el diseño** — en secciones, pedir aprobación después de cada sección
5. **Escribir el doc de diseño** — en `docs/plans/YYYY-MM-DD-<tema>-design.md`
6. **Transición a implementación** — invocar escritura de plan

### Regla dura

> No escribir código ni scaffolding hasta que el diseño esté presentado y aprobado.

### Principios

- **Una pregunta por mensaje** — no abrumar con múltiples preguntas
- **Preguntas de opción múltiple** cuando sea posible
- **YAGNI despiadado** — remover features innecesarias de todos los diseños
- **Validación incremental** — presentar diseño, obtener aprobación antes de seguir

---

## 3. Escritura de planes

**Cuándo:** cuando hay un spec o requerimientos para una tarea de múltiples pasos, antes de tocar código.

### Dónde guardar

```
docs/plans/YYYY-MM-DD-<nombre-feature>.md
```

### Header obligatorio de cada plan

```markdown
# [Nombre de la Feature] — Plan de implementación

> **Para el agente:** usar executing-plans para implementar tarea a tarea.

**Goal:** [Una oración describiendo qué construye]

**Architecture:** [2-3 oraciones sobre el enfoque]

**Tech Stack:** [Tecnologías clave]

---
```

### Estructura de cada tarea

```markdown
### Tarea N: [Nombre del componente]

**Archivos:**
- Crear: `ruta/exacta/archivo.ts`
- Modificar: `ruta/exacta/existente.ts:123-145`
- Test: `tests/ruta/exacta/test.ts`

**Paso 1: Escribir el test que falla**

```typescript
test("comportamiento esperado", () => {
  const result = funcion(input)
  expect(result).toBe(expected)
})
```

**Paso 2: Correr para verificar que falla**

Comando: `bun test ruta/test.ts`
Esperado: FAIL con "función no definida"

**Paso 3: Implementación mínima**

**Paso 4: Correr para verificar que pasa**

**Paso 5: Commit**

```bash
git commit -m "feat(scope): descripción"
```
```

### Granularidad — cada paso es 2-5 minutos

- "Escribir el test que falla" → un paso
- "Correrlo para verificar que falla" → otro paso
- "Implementar código mínimo" → otro paso
- "Correr tests para verificar que pasan" → otro paso
- "Commit" → otro paso

### Reglas

- Rutas de archivo exactas siempre
- Código completo en el plan (no "agregar validación")
- Comandos exactos con output esperado
- DRY, YAGNI, TDD, commits frecuentes

---

## 4. Ejecución de planes

**Cuándo:** cuando hay un plan escrito para ejecutar.

### Proceso

1. **Leer y revisar el plan** — identificar preguntas o problemas antes de empezar
2. **Ejecutar en lotes** — default: 3 tareas por lote
   - Para cada tarea: marcar en progreso → seguir cada paso exactamente → verificar → marcar completada
3. **Reportar** al terminar cada lote: qué se implementó + output de verificación
4. **Continuar** con el siguiente lote basado en feedback

### Cuándo parar y pedir ayuda

- Blocker a mitad de lote (dependencia faltante, test que falla, instrucción poco clara)
- El plan tiene gaps críticos
- No se entiende una instrucción
- La verificación falla repetidamente

**Preguntar en lugar de adivinar.**

### Regla importante

Nunca iniciar implementación en `main`/`master` sin consentimiento explícito.

---

## 5. TDD

### La ley de hierro

```
NINGÚN CÓDIGO DE PRODUCCIÓN SIN UN TEST QUE FALLA PRIMERO
```

Si escribiste código antes del test: borrarlo. Empezar de nuevo.

### Ciclo Red-Green-Refactor

```
RED: Escribir test que falla
  ↓
Verificar que falla por la razón correcta
  ↓
GREEN: Código mínimo para pasarlo
  ↓
Verificar que pasa (y no rompe otros tests)
  ↓
REFACTOR: Limpiar sin agregar comportamiento
  ↓
Repetir
```

### Red — escribir el test que falla

```typescript
// Bien: nombre descriptivo, testea comportamiento real, una cosa
test("reintenta operaciones fallidas 3 veces", async () => {
  let intentos = 0
  const op = () => {
    intentos++
    if (intentos < 3) throw new Error("fallo")
    return "éxito"
  }
  const result = await reintentar(op)
  expect(result).toBe("éxito")
  expect(intentos).toBe(3)
})
```

```typescript
// Mal: nombre vago, testea el mock no el código
test("reintento funciona", async () => {
  const mock = jest.fn()
    .mockRejectedValueOnce(new Error())
    .mockResolvedValueOnce("éxito")
  await reintentar(mock)
  expect(mock).toHaveBeenCalledTimes(2)
})
```

### Green — código mínimo

Escribir el código más simple posible para pasar el test. No agregar features, no refactorizar otro código, no "mejorar" más allá del test.

### Refactor — limpiar

Solo después de green: eliminar duplicación, mejorar nombres, extraer helpers. Mantener tests verdes.

### Cuándo aplica

- Features nuevas
- Bug fixes
- Refactors
- Cambios de comportamiento

**Excepciones (preguntar primero):** prototipos desechables, código generado, archivos de configuración.

### Checklist antes de marcar completado

- [ ] Cada función nueva tiene un test
- [ ] Se vio fallar cada test antes de implementar
- [ ] Cada test falló por la razón esperada (feature faltante, no typo)
- [ ] Se escribió código mínimo para pasar cada test
- [ ] Todos los tests pasan
- [ ] Output limpio (sin errores ni warnings)
- [ ] Tests usan código real (mocks solo si es inevitable)
- [ ] Edge cases y errores cubiertos

---

## 6. Testing por capas

### El principio central

> **Si el código no tiene test, no está terminado.**

Esto no es una aspiración — es un criterio de done que el CI enforcea.

### Metas de cobertura por capa

| Capa | Target | Enforced en CI |
|------|--------|----------------|
| `packages/*` (lógica pura) | **95%** | en cada PR |
| `src/lib/` (utilidades) | **95%** | en cada PR |
| `src/hooks/` (hooks) | **80%** | en cada PR |
| API routes | cobertura funcional (test manual) | revisión humana |
| React components | no requerido ahora | — |

### Matriz tipo de código → test requerido

| Tipo de código | Test requerido | En qué PR |
|----------------|----------------|-----------|
| Query nueva de DB | Test unitario en memoria | **mismo PR** |
| Función pura en `lib/` | Test unitario | **mismo PR** |
| Lógica pura de un hook | Extraer a `lib/` → test allí | **mismo PR** |
| Schema Zod nuevo | Test de validación | **mismo PR** |
| Server Action nueva | Test de la query subyacente | **mismo PR** |
| API route nueva | Test manual documentado en PR | **mismo PR** |
| Componente React nuevo | No requerido ahora | — |

### Patrón de test de DB (SQLite en memoria)

```typescript
import { describe, test, expect, beforeAll, afterEach } from "bun:test"

// CRÍTICO: debe ir ANTES de cualquier import que use getDb()
process.env["DATABASE_PATH"] = ":memory:"

import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import * as schema from "../schema"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

beforeAll(async () => {
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS users (...);
  `)
})

afterEach(async () => {
  await client.executeMultiple(`DELETE FROM users;`)
})

describe("nombreDelMódulo", () => {
  test("describe el comportamiento esperado", async () => {
    // arrange → act → assert
  })
})
```

**Reglas:**
- `process.env["DATABASE_PATH"] = ":memory:"` antes de cualquier import que use la DB
- DB en memoria siempre, nunca el archivo real
- Estado limpio entre tests con `afterEach` que borra las tablas
- Una assertion de comportamiento por test

### Tests de componentes React

```typescript
import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup, fireEvent } from "@testing-library/react"
import { MiComponente } from "@/components/MiComponente"

afterEach(cleanup)  // OBLIGATORIO en cada archivo

// Mock de server actions
mock.module("@/app/actions/dominio", () => ({
  actionHacerAlgo: mock(() => Promise.resolve()),
}))

describe("<MiComponente />", () => {
  test("renderiza el estado base", () => {
    const { getByText } = render(<MiComponente />)  // queries escopadas, no screen.*
    expect(getByText("Título")).toBeInTheDocument()
  })
})
```

**Reglas críticas:**
1. `afterEach(cleanup)` al inicio de cada archivo — sin esto los renders se acumulan
2. Queries escopadas: `const { getByRole } = render(...)` — nunca `screen.getByRole`
3. `fireEvent` en lugar de `userEvent` (incompatibilidades con happy-dom)
4. Mockear server actions y fetches con `mock.module`

---

## 7. Debugging sistemático

### La ley de hierro

```
NINGÚN FIX SIN INVESTIGACIÓN DE CAUSA RAÍZ PRIMERO
```

Si no completaste la Fase 1, no podés proponer un fix.

### Cuatro fases

#### Fase 1: Investigación de causa raíz

1. Leer mensajes de error completos (no saltear warnings ni stack traces)
2. Reproducir consistentemente — ¿se puede triggerear de forma confiable?
3. Revisar cambios recientes — git diff, commits, nuevas dependencias
4. En sistemas multi-componente: agregar instrumentación de diagnóstico antes de proponer fixes
5. Trazar el flujo de datos — ¿dónde se origina el valor malo?

#### Fase 2: Análisis de patrón

1. Encontrar ejemplos que funcionan en el mismo codebase
2. Comparar contra referencias (leer implementaciones de referencia completamente, no hojear)
3. Identificar diferencias — listar cada diferencia, por pequeña que parezca
4. Entender dependencias — qué necesita este componente para funcionar

#### Fase 3: Hipótesis y testing

1. Formular hipótesis única: "Creo que X es la causa raíz porque Y"
2. Testear con el cambio mínimo posible
3. Una variable a la vez — no hacer múltiples fixes simultáneos
4. Si no funciona: formular nueva hipótesis, no apilar más fixes

#### Fase 4: Implementación

1. Crear test que falla reproduciendo el bug
2. Implementar un único fix que ataca la causa raíz
3. Verificar que el test pasa y no rompe otros tests

**Si 3 o más fixes fallaron:** parar y cuestionar la arquitectura antes de intentar más.

### Red flags — parar y volver a Fase 1

- "Fix rápido por ahora, investigo después"
- "Solo probar cambiar X a ver qué pasa"
- "Agregar múltiples cambios y correr tests"
- "Probablemente sea X, voy a arreglarlo"
- "No entiendo bien pero esto podría funcionar"
- "Un intento más" (cuando ya se probaron 2+)

---

## 8. Git y CHANGELOG

### Formato de commits (Conventional Commits)

```
<tipo>(<scope>): <descripción en minúsculas>
```

**Tipos:** `feat` `fix` `refactor` `chore` `docs` `test` `ci` `perf` `revert`

**Límites:** subject ≤ 100 chars, header completo ≤ 120 chars.

**Ejemplos correctos:**
```bash
git commit -m "feat(web): agregar paginación en /admin/users"
git commit -m "fix(db): corregir removeAreaCollection para filtrar por collectionName"
git commit -m "test(logger): agregar tests de reconstructFromEvents"
git commit -m "docs: actualizar architecture.md con diagrama de auth"
git commit -m "refactor(web): extraer lógica SSE de ChatInterface a useRagStream"
git commit -m "chore(deps): actualizar drizzle-orm"
```

### CHANGELOG antes del commit

**Actualizar `CHANGELOG.md` antes de hacer commit**, no después.

```markdown
## [Unreleased]

### Plan N — Nombre del plan (YYYY-MM-DD)

#### Added
- `ruta/archivo.ts`: descripción de qué se agregó — YYYY-MM-DD *(Plan N F1.1)*

#### Fixed
- `ruta/archivo.ts`: descripción del bug y cómo se corrigió — YYYY-MM-DD

#### Changed
- `ruta/archivo.ts`: qué cambió y por qué — YYYY-MM-DD
```

**Categorías:** `Added` | `Changed` | `Deprecated` | `Removed` | `Fixed` | `Security`

**Reglas:**
- Cada plan nuevo abre su propia sección al tope de `[Unreleased]`
- El ID de tarea va al final: `*(Plan 4 F1.7)*`
- Las entradas del mismo plan se agrupan, nunca se intercalan

### Pre-push hook

El hook `pre-push` corre `type-check` automáticamente antes de cada push. Si falla, el push se cancela. No usar `--no-verify`.

### Crear un PR

1. CHANGELOG tiene la entrada correspondiente
2. Branch pusheada al remote
3. `gh pr create` con sección CHANGELOG obligatoria en el body
4. CI valida: commitlint + changelog check + type-check + tests + lint

---

## 9. Verificación antes de declarar "terminado"

### La ley de hierro

```
NINGUNA AFIRMACIÓN DE COMPLETITUD SIN EVIDENCIA DE VERIFICACIÓN FRESCA
```

Si no corriste el comando de verificación en este mensaje, no podés afirmar que pasa.

### El gate antes de cualquier claim

```
1. IDENTIFICAR: ¿qué comando prueba esta afirmación?
2. CORRER: ejecutar el comando completo (fresco, completo)
3. LEER: output completo, verificar exit code, contar failures
4. VERIFICAR: ¿el output confirma la afirmación?
   - Si NO: informar estado real con evidencia
   - Si SÍ: hacer la afirmación CON evidencia
5. RECIÉN ENTONCES: hacer la afirmación
```

### Claims comunes y qué requieren

| Afirmación | Requiere | No es suficiente |
|-----------|----------|------------------|
| Tests pasan | Output del comando: 0 failures | Corrida anterior, "debería pasar" |
| Lint limpio | Output: 0 errors | Verificación parcial |
| Build exitoso | Comando: exit 0 | Lint pasando |
| Bug corregido | Test del síntoma original: pasa | Código cambiado |
| Requerimientos cumplidos | Checklist línea a línea | Tests pasando |

### Red flags — parar

- Usar "debería", "probablemente", "parece"
- Expresar satisfacción antes de verificar ("¡Genial!", "¡Perfecto!", "¡Listo!")
- Estar por hacer commit/push/PR sin verificación
- Depender de verificación parcial

---

## 10. ADRs — decisiones de arquitectura

### Cuándo crear un ADR

- Se elige una tecnología sobre otra con trade-offs no obvios
- Se establece una convención que se desvía del default del ecosistema
- Se toma una decisión de arquitectura que afecta múltiples partes del sistema
- Se resuelve un bug que revela un patrón a seguir o evitar

**No crear ADR para:** decisiones triviales, preferencias de estilo, cosas que ya documenta el CHANGELOG.

### Formato

```markdown
# ADR-NNN: [Título]

**Fecha:** YYYY-MM-DD
**Estado:** Propuesto | Aceptado | Deprecado | Reemplazado por ADR-XXX

## Contexto
El problema y las restricciones.

## Opciones consideradas
- Opción A — pros/cons
- Opción B — pros/cons

## Decisión
Cuál se eligió y por qué.

## Consecuencias
Positivas y trade-offs.

## Referencias
Archivos o entradas de CHANGELOG relevantes.
```

### Convención de nombres

```
docs/decisions/NNN-kebab-case-titulo.md
```

El número es secuencial. Nunca reusar un número aunque se deprece el ADR.

---

## 11. Ciclo completo para una feature nueva

```
1. Leer contexto
   ↓
   Leer archivos críticos relevantes del área afectada

2. Brainstorming
   ↓
   Explorar enfoques, proponer 2-3 alternativas, acordar diseño
   Guardar en docs/plans/YYYY-MM-DD-<tema>-design.md

3. Escribir plan
   ↓
   docs/plans/YYYY-MM-DD-<nombre-feature>.md
   Con tareas granulares, rutas exactas, pasos TDD

4. Implementar (TDD)
   ↓
   Test rojo → verificar que falla → código mínimo → verificar que pasa → refactor
   Si algo falla inesperadamente → debugging sistemático (no adivinar)

5. Simplify
   ↓
   Dead code check, eliminar lo que sobra

6. Actualizar CHANGELOG.md
   ↓
   Antes del commit, bajo [Unreleased]

7. Commit
   ↓
   git commit -m "feat(scope): descripción"

8. Type-check (automático en pre-push)

9. Review (para cambios no triviales)
   ↓
   Code review del área afectada

10. Docs (si aplica)
    ↓
    Actualizar ADR, README, o documentación relevante si hubo decisión de arquitectura
```

### Dónde poner cada cosa

| Qué | Dónde |
|-----|-------|
| Página nueva | `src/app/(app)/[ruta]/page.tsx` (Server Component) |
| Mutación desde el servidor | `src/app/actions/[dominio].ts` (Server Action) |
| Endpoint REST | `src/app/api/[ruta]/route.ts` |
| Lógica de UI con estado | `src/components/[nombre].tsx` (Client Component) |
| Hook con lógica de fetch | `src/hooks/use[Nombre].ts` |
| Query de DB nueva | `packages/db/src/queries/[dominio].ts` |
| Tipo compartido | `packages/shared/src/schemas.ts` |

### Reglas de arquitectura

- **Server Components por defecto** — solo `"use client"` donde sea imprescindible
- **Validar con Zod** — para inputs de API
- **Timestamps en epoch ms** — `Date.now()` o `Temporal.Now.instant().epochMilliseconds`
- **Commits frecuentes** — no acumular trabajo sin commitear

---

## Resumen rápido

| Situación | Acción |
|-----------|--------|
| Voy a empezar una feature | Brainstorming primero → plan → TDD |
| Tengo un bug | Debugging sistemático 4 fases, no adivinar |
| Voy a commitear | CHANGELOG primero, luego commit con formato convencional |
| Creo que terminé | Correr verificación fresca, evidencia antes de afirmar |
| Hay una decisión de arquitectura importante | Crear ADR en `docs/decisions/` |
| La feature es compleja (>2-3 hs) | Plan en `docs/plans/` con tareas granulares |
| Es un cambio trivial (≤3 líneas) | TDD directo + CHANGELOG + commit |
