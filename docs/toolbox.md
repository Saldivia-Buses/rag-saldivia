# Toolbox — Repos y herramientas externas

Registro de repos de GitHub, librerías y herramientas que usamos o evaluamos
para el desarrollo de RAG Saldivia. Este archivo lo mantiene Claude Code (Opus)
actualizado cada vez que se encuentra algo nuevo.

---

## En uso

| Repo/herramienta | Qué es | Para qué lo usamos |
|---|---|---|
| [nvidia/GenerativeAIExamples](https://github.com/NVIDIA/GenerativeAIExamples) | RAG Blueprint v2.5.0 | Base del sistema RAG (submodule en `vendor/rag-blueprint/`) |
| zod | Validación de schemas TypeScript | Compartido entre todos los paquetes (`packages/shared`) |
| shadcn/ui + Radix | Componentes UI + primitivos headless | Design system completo de la app |
| Tailwind CSS v4 | Utility-first CSS | Styling de toda la app |
| @tanstack/react-table | Tablas avanzadas | DataTable con sorting, filtro, paginación |
| Lucide React | Iconos | Iconografía consistente en toda la app |
| Drizzle ORM | ORM TypeScript | Todas las queries a SQLite |

## Por integrar (plan dedicado)

| Repo | Qué es | Para qué nos serviría | Plan |
|---|---|---|---|
| [ai (Vercel AI SDK)](https://sdk.vercel.ai) | Toolkit para apps de IA — streaming, tools, agents | Reemplaza SSE manual en `useRagStream`. Base para json-render. | Plan 14 |
| [vercel-labs/json-render](https://github.com/vercel-labs/json-render) | Generative UI — LLM produce JSON → renderiza React components. 36 shadcn/ui incluidos. Streaming con `SpecStreamCompiler`. | Respuestas ricas del RAG: tablas, cards, badges, gráficos. Packages: `@json-render/core` + `@json-render/react` + `@json-render/shadcn`. | Plan 17 |
| [JCodesMore/ai-website-cloner-template](https://github.com/JCodesMore/ai-website-cloner-template) | Clona websites usando AI agents. Pipeline: reconnaissance → foundation → component specs → parallel build → assembly + QA. | Extraer tokens/layout de claude.ai para nuestra UI. Modo reconnaissance, no clon literal. | Plan 15 |


## Por evaluar

| Repo | Qué es | Para qué nos serviría | Evaluación | Encontrado |
|---|---|---|---|---|
| [jamwithai/production-agentic-rag-course](https://github.com/jamwithai/production-agentic-rag-course) | Curso completo de RAG en producción. 7 semanas: infra → pipelines → BM25 → hybrid search → RAG → monitoring (Langfuse) → agentic RAG (LangGraph). Stack: FastAPI, PostgreSQL, OpenSearch, Redis, Airflow, Ollama. | **MUY relevante como referencia.** Patrones de producción: hybrid search (BM25+semantic), query rewriting, document grading, Langfuse monitoring, caching con Redis. La filosofía "BM25 primero, semantic después" es sólida. Usar como guía para mejorar nuestro RAG pipeline. | **Referencia de arquitectura** — leer cuando trabajemos en el RAG pipeline | Enzo, 2026-03-29 |
| [NousResearch/hermes-agent](https://github.com/NousResearch/hermes-agent) | Agente AI self-improving con skill system. Multi-plataforma: Terminal, Telegram, Discord, Slack, **WhatsApp**, Email. 40+ tools built-in. Subagent delegation. Scheduled automations. Soporta 200+ modelos. | **Relevante para integraciones futuras.** La visión incluye WhatsApp + Email — hermes-agent ya tiene gateways para ambos. El skill system y la memoria persistente son interesantes. No es para adoptar ahora pero sí para cuando lleguemos a integraciones. | **Referencia para integraciones WhatsApp/Email** | Enzo, 2026-03-29 |
| [vercel-labs/agent-browser](https://github.com/vercel-labs/agent-browser) | Browser automation CLI en Rust para AI agents. Headless, accessibility-first. Snapshot del DOM como accessibility tree. Batch execution. Multi-tab. Network interception. | Podría servir para scraping/extracción de contenido web para alimentar el RAG. También útil para E2E testing como alternativa a Playwright. Rápido (Rust nativo). | **Evaluar para content extraction** | Enzo, 2026-03-29 |
| [volcengine/OpenViking](https://github.com/volcengine/OpenViking) | Context database para AI agents. Almacena memorias, recursos, skills en paradigma filesystem. Tiered loading (L0/L1/L2). Retrieval jerárquico. Auto-compresión de conversaciones. Visualización de trayectorias. | Interesante para mejorar cómo el RAG almacena y recupera contexto. El tiered loading (L0 summary → L1 sections → L2 full) podría optimizar el consumo de tokens. | **Evaluar para optimización de context** | Enzo, 2026-03-29 |
| [donnemartin/system-design-primer](https://github.com/donnemartin/system-design-primer) | Referencia completa de diseño de sistemas. CAP theorem, caching, load balancing, DB scaling, microservices, message queues. Ejemplos: Twitter, URL shortener, web crawler. | **Referencia para el ADR-012** (stack evaluation). Los patrones de escalabilidad y trade-offs son directamente aplicables a nuestras decisiones de SaaS multi-tenant. | **Referencia de arquitectura** | Enzo, 2026-03-29 |

## Librerías frontend por evaluar

Librerías que no usamos pero podrían agregar valor:

| Librería | Qué hace | Relevancia para nosotros |
|---|---|---|
| [react-hook-form](https://react-hook-form.com) | Forms performantes con validación | Útil para settings, login, formularios admin. Combina bien con Zod. |
| [nuqs](https://nuqs.47ng.com) | URL search params management para Next.js | Útil para filtros en collections y futuras tablas. Type-safe. |
| [ai (Vercel AI SDK)](https://sdk.vercel.ai) | Toolkit para apps de IA — streaming, tools, agents | **Decidido adoptar (Plan 14).** Ver "Por integrar". |
| [zustand](https://zustand-demo.pmnd.rs) | State management minimalista | Alternativa a React context para estado global (theme, user, notifications). |
| [date-fns](https://date-fns.org) | Utilidades de fechas | Tree-shakeable, mejor que moment. Para timestamps en chat y audit. |
| [motion (Framer)](https://motion.dev) | Animaciones React | Para 2.x — no es prioridad ahora. |
| [dndkit](https://dndkit.com) | Drag & drop | Para 2.x — reordenar colecciones, kanban admin. |
| [react-window](https://react-window.vercel.app) | Listas virtualizadas | Útil si el chat tiene muchos mensajes. Performance. |
| [Slate](https://slatejs.org) | Editor de texto rico | Para 2.x — mejorar el input del chat con formatting. |
| tRPC + react-query | API type-safe end-to-end | Alternativa a fetch directo. Evaluar en ADR-012. |

## Plataformas de deploy

Para la visión SaaS (cada empresa = servidor), evaluar:

| Plataforma | Tipo | GPU support | Notas |
|---|---|---|---|
| [fly.io](https://fly.io) | Containers | Sí (GPU Machines) | Deploy con `flyctl`. Edge locations. **Candidato fuerte para SaaS.** |
| [render.com](https://render.com) | PaaS | No nativo | Simple, buen free tier. Solo para frontend sin GPU. |
| [koyeb.com](https://koyeb.com) | Serverless containers | Sí (GPU instances) | Auto-scaling. **Evaluar para SaaS.** |
| [vercel.com](https://vercel.com) | Serverless/Edge | No | Natural para Next.js. Solo para frontend. |
| [netlify.com](https://netlify.com) | JAMstack/Serverless | No | Solo para frontend estático. |
| NVIDIA Cloud (actual) | Workstation física | 1x RTX PRO 6000 96GB | Deploy actual. `make deploy PROFILE=workstation-1gpu`. |

**Nota:** la visión SaaS requiere GPU para el RAG (embeddings + LLM). Las plataformas
sin GPU solo sirven para el frontend si se separa de backend.

## Ideas adoptadas (concepto, no instalado)

| Repo | Qué adoptamos |
|---|---|
| [garagon/nanostack](https://github.com/garagon/nanostack) | Sprint sequence, scope drift detection, intensity modes, artifact persistence, WTF likelihood, conflict resolution, guard rules, ZEN principles. |
| [karpathy/autoresearch](https://github.com/karpathy/autoresearch) | **Mejora continua autónoma.** Loop: proponer cambio → testear → mejoró? → keep/discard → repeat. Aplicado a code quality, test coverage, dead code, performance, deps, security. Ver bible.md sección "Mejora continua". |

## Descartados (por ahora)

| Repo | Qué es | Por qué |
|---|---|---|
| [chenglou/pretext](https://github.com/chenglou/pretext) | Text layout en JS puro sin DOM | Sin uso directo. Reconsiderar para rendering custom. |
| [greensock/gsap-skills](https://github.com/greensock/gsap-skills) | Skills de GSAP para AI agents (animaciones) | Animaciones son 2.x. No es prioridad. |
| [0xSero/parchi](https://github.com/0xSero/parchi) | Browser automation extension con AI | No construimos un browser agent. Risks de seguridad altos. |
| Recoil (Facebook) | State management React | Deprecated. Usar zustand o jotai en su lugar. |
