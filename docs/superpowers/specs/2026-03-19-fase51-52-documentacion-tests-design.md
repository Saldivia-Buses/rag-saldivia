# Fase 5.1 + 5.2 — Documentación completa y Tests

**Fecha:** 2026-03-19
**Estado:** Aprobado
**Fases anteriores:** 5 (Crossdoc Pro)
**Próxima fase:** 6

---

## Contexto

El proyecto RAG Saldivia tiene un estado funcional avanzado (fases 1–5 completadas) pero carece de documentación navegable y cobertura de tests completa. Antes de avanzar a la fase 6, se consolida la base de conocimiento del proyecto y se establece una pirámide de tests robusta.

---

## Fase 5.1 — Documentación completa

### Objetivo

Documentar el proyecto de punta a punta de manera que:
- Un dev nuevo pueda entender la arquitectura y empezar a trabajar sin preguntar
- Enzo pueda retomar el proyecto después de meses y saber exactamente cómo está todo
- Cada carpeta de código se auto-explique con su propio README

### Estructura de archivos

```
README.md                              ← reescrito, puerta de entrada (~150 líneas)
docs/
  architecture.md                      ← qué es el sistema, cómo está compuesto
  development-workflow.md              ← metodología: Research→Brainstorm→Plan→Impl→Review
  testing.md                           ← cómo correr y escribir tests
  deployment.md                        ← perfiles, Brev, variables de entorno
  contributing.md                      ← convenciones de código, commits, PRs
```

### README.md global — secciones

1. Título + badges (CI, coverage)
2. Qué es en 3 líneas (overlay sobre NVIDIA RAG Blueprint, qué agrega)
3. Quick Start (5 comandos)
4. Tabla de contenidos con links a `docs/`
5. Diagrama ASCII de arquitectura (`Usuario → SDA → Gateway → RAG → LLM`)
6. Estado actual (fases completadas, en progreso, roadmap)
7. Links rápidos (tests, deploy, contributing)

### `docs/development-workflow.md` — el doc más importante

Explica la metodología de trabajo:
- Regla fundamental: ningún cambio no-trivial sin `Research → Brainstorm → Plan → Impl → Review`
- Qué es trivial vs no-trivial (con ejemplos concretos)
- Stack de herramientas: CGC + Repomix para explorar código, firecrawl para docs externos, superpowers skills
- Ciclo de una fase: cómo nace una spec, cómo se convierte en plan, cómo se ejecuta
- Cómo leer los docs en `docs/superpowers/` (specs, plans, su relación)

### READMEs de zona y subcarpeta (~20-25 archivos)

**Convención**: cada carpeta con código tiene un `README.md` con:
- Propósito de la carpeta (1-3 líneas)
- Tabla: `Archivo | Qué hace | Dependencias clave`
- Notas de diseño si aplica

**Carpetas cubiertas:**

```
services/sda-frontend/
  README.md
  src/lib/
    components/
      README.md
      chat/README.md          ← ChatInput, MessageList, CrossdocProgress, DecompositionView,
      │                          CrossdocSettingsPopover, HistoryPanel, MarkdownRenderer, SourcesPanel
      ui/README.md            ← Badge, Button, Card, Input, Modal, Skeleton, Toast, ToastContainer
      sidebar/README.md       ← Sidebar, SidebarItem
      layout/README.md        ← Sidebar (layout)
    stores/README.md          ← chat, collections, crossdoc, toast
    utils/README.md           ← markdown, scroll, chat-utils
    crossdoc/README.md        ← pipeline, tipos, lógica de descomposición
    server/README.md          ← BFF: auth, gateway client
    actions/README.md       ← clickOutside.ts: qué hace, cuándo usarlo
  src/routes/
    README.md
    (app)/README.md           ← rutas protegidas: chat, collections, admin, audit, settings, upload
    (auth)/README.md          ← login
    api/README.md             ← todos los endpoints BFF
    api/crossdoc/README.md    ← decompose, subquery, synthesize

saldivia/
  README.md                   ← SDK Python: gateway, auth, providers, config, etc.
  auth/README.md              ← database.py, models.py
  tests/README.md             ← qué testea cada archivo, cómo correrlos

config/
  README.md
  profiles/README.md          ← brev-2gpu, workstation-1gpu, full-cloud

patches/
  README.md
  frontend/README.md          ← archivos nuevos y modificados del blueprint

scripts/README.md             ← smart_ingest.py y otros scripts
cli/README.md                 ← comandos disponibles y ejemplos
```

### Convención de mantenimiento

Incorporada en `docs/contributing.md`: al terminar una feature que modifica código de una zona, actualizar el README de esa carpeta es un paso obligatorio del plan, no opcional.

---

## Fase 5.2 — Tests completos

### Objetivo

Cubrir el proyecto con una pirámide de tests robusta y mantenible:

```
          ┌─────────────────┐
          │  E2E Playwright  │  5 flows críticos, POM, corren en CI
          ├─────────────────┤
          │ Component Tests  │  @testing-library/svelte, componentes UI
          │  (@testing-lib)  │  críticos, jsdom, integrado a Vitest
          ├─────────────────┤
          │   Unit Tests     │  Vitest existente + expansión
          │   (Vitest)       │  stores, utils, pipeline crossdoc, API routes
          └─────────────────┘
```

### Nota sobre tablas y ejemplos

Los datos en las tablas de este spec (casos de test, estados de componentes, flows, módulos) son **ejemplos ilustrativos** que guían al implementador, no listas cerradas ni requisitos rígidos. El implementador debe:

- Expandir o ajustar los casos según lo que encuentre en el código real
- Agregar escenarios que no estén listados si detecta gaps de cobertura
- Omitir o reemplazar casos que no apliquen al estado actual del componente

Las tablas describen la intención, no la implementación exacta.

### Datos de test — convención global

**Todos los datos usados en los tests son ejemplos ilustrativos**, no datos de producción ni valores hardcodeados con significado real. Esto aplica a las tres capas:

- **Credenciales** (usuario, contraseña, JWT): valores de ejemplo como `test@example.com` / `password123`. Nunca usar credenciales reales del sistema Brev o del gateway de producción.
- **Fixtures JSON** (respuestas de API, colecciones, documentos): estructuras de ejemplo que representan el contrato de la API, con valores genéricos como `"Mi colección de prueba"`, `"Documento de ejemplo"`.
- **Queries y texto** (chat messages, crossdoc queries): frases de ejemplo como `"¿Cuál es la presión máxima?"` sin relación con datos reales del vault.
- **Archivos** (PDFs para upload): un PDF de ejemplo incluido en `tests/e2e/fixtures/sample.pdf` de contenido trivial.
- **Mocks de SSE / streaming**: tokens de ejemplo `"Este"`, `" es"`, `" un"`, `" ejemplo"` que simulan el stream sin ser una respuesta real del LLM.

Cuando el implementador defina valores concretos en los tests, debe elegir datos obvios, genéricos y sin información sensible. Si un valor específico parece "real", es una señal de que se está usando mal.

### Capa 1 — Unit Tests (Vitest)

Los archivos marcados como **existente** ya tienen test file — se expanden con los casos indicados. Los marcados **nuevo** requieren crear el archivo.

| Módulo | Estado | Tests a agregar |
|--------|--------|----------------|
| `crossdoc/pipeline.ts` | existente (`pipeline.test.ts`) | Casos edge: queries vacías, dedup agresivo (threshold), input con caracteres especiales |
| `stores/chat.svelte.ts` | existente (`chat.svelte.test.ts`) | Historial multi-turn, reset de estado, appendToken con markup |
| `stores/collections.svelte.ts` | existente (`collections.svelte.test.ts`) | CRUD completo: create, delete, select, error de red |
| `server/auth.ts` | existente (`auth.test.ts`) | JWT expirado, cookie malformada, refresh exitoso y fallido |
| `utils/markdown.ts` | existente (`markdown.test.ts`) | XSS sanitization, tablas, code blocks con highlight |
| `utils/scroll.ts` | existente (`scroll.test.ts`) | scroll-to-bottom edge cases, no-op si ya está abajo |
| `actions/clickOutside.ts` | **nuevo** (`clickOutside.test.ts`) | Listener adjuntado, cleanup en destroy, no dispara dentro del elemento |
| `routes/api/chat/+server.ts` | **nuevo** (`chat-server.test.ts`) | Proxy al gateway, error 401/403/500, streaming response |
| `routes/api/upload/+server.ts` | existente (`upload.test.ts`) | Expandir: archivo muy grande, tipo inválido, gateway caído |
| `routes/api/crossdoc/decompose` | existente (`decompose.test.ts`) | Expandir: error de timeout, response malformado |

**Meta de cobertura:** ≥80% en `src/lib/` y `src/routes/api/` medido sobre archivos `.ts` y `.svelte.ts`. Los archivos `.svelte` de componentes se cubren via Capa 2 y no entran en el threshold de cobertura automático de Vitest.

### Capa 2 — Component Tests (`@testing-library/svelte`)

Instalación: `@testing-library/svelte` + `@testing-library/user-event`, integrado al runner Vitest existente.

**Configuración de environment:** El `vitest.config.ts` actual usa `environment: 'node'` globalmente. Los component tests requieren `jsdom`. Se debe agregar `environmentMatchGlobs` para scoping por archivo sin romper los unit tests existentes:

```ts
test: {
  environment: 'node',
  environmentMatchGlobs: [
    ['src/**/*.svelte.test.ts', 'jsdom'],
    ['src/**/components/**/*.test.ts', 'jsdom'],
  ],
}
```

Los archivos de unit tests existentes (`*.test.ts` sin `.svelte.`) continúan en `node`. Los component tests nuevos deben nombrarse `ComponentName.component.test.ts` o estar en `src/lib/components/**` para recibir `jsdom` automáticamente.

| Componente | Estados a cubrir |
|-----------|-----------------|
| `ChatInput.svelte` | idle, escribiendo, enviando, crossdoc activado, disabled |
| `MessageList.svelte` | vacío, 1 mensaje, streaming en curso, error |
| `CrossdocProgress.svelte` | cada fase (decompose→search→synthesize→done), error |
| `Toast.svelte` | success, error, warning, auto-dismiss |
| `Modal.svelte` | abierto, cerrado, con/sin confirmación |
| `CollectionCard.svelte` (en `routes/(app)/collections/_components/`) | normal, sin docs, loading, selected |

### Capa 3 — E2E Playwright (Page Object Model)

**Estructura de archivos:**

```
services/sda-frontend/tests/e2e/
  pages/
    LoginPage.ts
    ChatPage.ts
    CollectionsPage.ts
    UploadPage.ts
    AdminPage.ts
  fixtures/
    auth.ts              ← fixture que loguea antes de cada test
  flows/
    auth.spec.ts         ← login válido/inválido, logout, redirect sin sesión
    chat.spec.ts         ← enviar mensaje, SSE streaming, fuentes, historial
    collections.spec.ts  ← listar, crear, eliminar, ver detalle
    upload.spec.ts       ← subir PDF, progreso, doc aparece en colección
    crossdoc.spec.ts     ← activar crossdoc, query, decompose→subqueries→síntesis
```

**5 flows críticos:**

1. **Auth** — login válido, login inválido, logout, redirect sin sesión
2. **Chat básico** — enviar mensaje, recibir respuesta SSE, ver fuentes, historial
3. **Colecciones** — listar, crear, eliminar, ver detalle de colección
4. **Upload** — subir PDF, ver progreso de ingesta, documento aparece en colección
5. **Crossdoc E2E** — activar modo crossdoc, enviar query, ver descomposición → subqueries → síntesis

**Estrategia de backend para E2E:**

Los flows se dividen en dos categorías según su dependencia de backend real:

| Flow | Backend | Estrategia |
|------|---------|-----------|
| `auth.spec.ts` | Gateway real (auth DB) | Corre contra app local con gateway mock o Brev |
| `chat.spec.ts` | SSE streaming + LLM | `page.route('/api/chat*', ...)` intercepta y devuelve SSE mock |
| `collections.spec.ts` | CRUD collections | `page.route('/api/collections*', ...)` devuelve fixtures JSON |
| `upload.spec.ts` | NV-Ingest + Milvus | Marcado `@slow`, solo corre con `PLAYWRIGHT_BASE_URL` apuntando a Brev |
| `crossdoc.spec.ts` | Pipeline 4 fases + LLM | `page.route('/api/crossdoc/*', ...)` intercepta cada fase, devuelve mocks |

**Regla:** Los flows con `page.route()` corren siempre (CI incluido). Los marcados `@slow` solo corren cuando `PLAYWRIGHT_BASE_URL` está definido (integración contra Brev).

**Configuración Playwright:**
- `playwright.config.ts` en `services/sda-frontend/`
- Base URL: `http://localhost:4173` (preview) por defecto, override con `PLAYWRIGHT_BASE_URL`
- Fixture de auth reutilizable (`fixtures/auth.ts`) para todos los flows que requieren sesión
- **La fixture de auth debe completarse y verificarse antes de escribir cualquier flow que la importe** (orden: POM → fixture → flows)

### Comandos CI

```bash
make test           # unit + component + e2e (todo)
make test-unit      # solo vitest (rápido, <5s)
make test-e2e       # solo playwright (necesita app corriendo)
make test-coverage  # vitest con --coverage
```

### Herramientas a instalar

| Herramienta | Propósito |
|-------------|-----------|
| `@testing-library/svelte` | Component tests |
| `@testing-library/user-event` | Simulación de interacciones de usuario |
| `@playwright/test` | E2E flows |
| `@vitest/coverage-v8` | Coverage reports |

---

## Criterios de éxito

### 5.1
- [ ] README global reescrito y navegable
- [ ] 5 docs temáticos en `docs/` completos
- [ ] ~20-25 READMEs de zona/subcarpeta con tabla de archivos
- [ ] Convención de mantenimiento documentada en `contributing.md`

### 5.2
- [ ] Vitest con cobertura ≥80% en `src/lib/` y `src/routes/api/`
- [ ] Component tests para 6 componentes UI críticos
- [ ] Playwright configurado con POM y 5 flows E2E pasando
- [ ] `make test` corre todo la pirámide en orden

---

## Orden de implementación

```
5.1.1  Escribir docs/ temáticos (architecture, workflow, testing, deployment, contributing)
5.1.2  Reescribir README.md global
5.1.3  Escribir READMEs de zona (services/sda-frontend, saldivia, config, patches, scripts, cli)
5.1.4  Escribir READMEs de subcarpeta (lib/components/*, lib/stores, lib/utils, lib/crossdoc, etc.)

5.2.1  Instalar @testing-library/svelte + @vitest/coverage-v8
5.2.2  Expandir unit tests Vitest (stores, utils, API routes)
5.2.3  Escribir component tests para 6 componentes
5.2.4  Instalar Playwright y configurar playwright.config.ts
5.2.5  Escribir Page Objects (Login, Chat, Collections, Upload, Admin)
5.2.6  Escribir fixture de auth
5.2.7  Escribir los 5 flows E2E
5.2.8  Integrar make test en Makefile
```
