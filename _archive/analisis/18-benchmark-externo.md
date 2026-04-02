# 18 — Benchmark Externo: Saldivia RAG vs La Industria

> **Fecha:** 2026-03-31
> **Fuentes:** Research de mercado (MarketsandMarkets, Yahoo Finance, Intel Market Research),
> documentacion oficial de cada competidor, pricing publico y estimado,
> patrones de Cal.com / Clerk / Supabase / Dub.co / next-forge / Vercel AI SDK
>
> **Contexto:** Saldivia RAG es un overlay sobre NVIDIA RAG Blueprint v2.5.0 que agrega
> auth, RBAC, multi-coleccion, chat UI, admin panel, y messaging interno.
> Deploy: workstation fisica con 1x RTX PRO 6000 Blackwell (96GB VRAM).

---

## El mercado RAG enterprise en numeros

| Metrica | Valor | Fuente |
|---------|-------|--------|
| Tamano del mercado RAG (2025) | USD 1.94 mil millones | MarketsandMarkets |
| Proyeccion a 2030 | USD 9.86 mil millones | MarketsandMarkets |
| CAGR | 38.4% | MarketsandMarkets |
| RAG-as-a-Service (2025) | USD 91.5 millones | Intel Market Research |
| RAG-as-a-Service (2032) | USD 187 millones | Intel Market Research |
| % implementaciones en grandes organizaciones | 73.3% | Roots Analysis |

**Takeaway:** No es nicho. Es un mercado de casi 2 mil millones creciendo al 38% anual.

---

## Competidores directos — productos RAG enterprise

### Tier 1: Plataformas all-in-one (SaaS)

#### Glean — lider del mercado
- **URL:** https://www.glean.com
- **Que hace:** AI search + asistentes + agentes sobre datos de la empresa. Indice persistente sobre 100+ apps (Slack, Drive, Confluence, Jira, etc.). Knowledge graph para resultados personalizados.
- **Precio:** ~$50/usuario/mes. Minimo 100 seats ($60K-$75K/ano). Add-on "Work AI" +$15/usuario/mes para features generativas. Contratos anuales.
- **Features:** 100+ conectores, SSO, admin controls, usage analytics, 15+ LLMs, indexacion en tiempo real.
- **Target:** Enterprise grande (1,000+ empleados). Tech, finanzas, servicios profesionales.
- **Deploy:** Cloud primario. On-prem via partnership con Dell (vendor-managed). "Cloud-Prem" en la nube del cliente. No self-hostable en hardware propio.
- **Debilidad:** No es open source. No se puede deployar en hardware propio. No air-gapped.

#### Microsoft Copilot — el gigante
- **URL:** https://microsoft.com/microsoft-365-copilot
- **Que hace:** AI assistant embebido en M365 (Word, Excel, Teams, Outlook). RAG via Microsoft Graph.
- **Precio:** $20-34/usuario/mes sobre una licencia M365 obligatoria. Costo real >$50/usuario/mes.
- **Features:** 100+ conectores (centrados en M365), Copilot Studio para agentes custom.
- **Target:** Organizaciones 100% en Microsoft 365.
- **Deploy:** Azure cloud unicamente.
- **Debilidad:** Lock-in total en ecosistema Microsoft. Sin opcion self-hosted.

#### ChatGPT Enterprise
- **URL:** https://openai.com/chatgpt/enterprise
- **Que hace:** Tier enterprise de ChatGPT con indexacion parcial de datos de la empresa.
- **Precio:** ~$50-60/usuario/mes. Contratos anuales. ChatGPT Team: $25-30/usuario/mes (menos features).
- **Features:** Modelos OpenAI (GPT-5.4, etc.), conectores limitados (Drive, SharePoint, GitHub, Confluence), custom GPTs, SOC 2 Type II.
- **Deploy:** Cloud unicamente.
- **Debilidad:** Pocos conectores. Solo modelos OpenAI. Sin self-hosted.

---

### Tier 2: Plataformas especializadas

#### Onyx (ex-Danswer) — el competidor open-source mas cercano
- **URL:** https://onyx.app
- **Que hace:** Plataforma AI completa: search + chat multi-modelo + deep research + agentes con MCP.
- **Precio:** Community edition GRATIS (funcional completa). Business: $20/usuario/mes (anual) o $25/mes. Enterprise: pricing custom con SLAs.
- **Features:** 40+ conectores, cualquier LLM (OpenAI, Anthropic, Gemini, DeepSeek, Llama, local via Ollama/vLLM), busqueda hibrida (keyword + semantica), herencia de permisos del sistema fuente, SOC 2 Type II, GDPR, code interpreter, air-gapped deploy. Licencia MIT.
- **Target:** Todos los tamanos. Fuerte en industrias reguladas (defensa, salud, finanzas, UE).
- **Deploy:** Self-hosted (Docker/K8s), managed cloud, o air-gapped con LLMs locales.
- **Relevancia:** Este es el competidor mas directo de Saldivia RAG. Hace practicamente lo mismo pero con mas conectores y sin NVIDIA NIMs.

#### Vectara — RAG-as-a-Service enterprise
- **URL:** https://www.vectara.com
- **Que hace:** Plataforma de agentes con RAG integrado. Deteccion de alucinaciones (HHEM score).
- **Precio:** SaaS desde $100K/ano. VPC desde $250K/ano. Trial de 30 dias.
- **Features:** Pipeline RAG end-to-end (sin vector DB separada), reranker built-in, multi-idioma, citations, hallucination scoring. SOC 2, GDPR.
- **Deploy:** SaaS o VPC (nube del cliente).

#### Contextual AI — agentes RAG especializados
- **URL:** https://contextual.ai
- **Que hace:** Plataforma para agentes RAG con retrieval activo y optimizacion en tiempo de inferencia.
- **Precio:** On-demand: pay-as-you-go ($25 creditos gratis). Enterprise: custom. Enterprise agrega SAML/SSO, RBAC, observabilidad de pipeline, retencion custom, ingesta continua, SLA de uptime.
- **Features:** Optimizacion de queries (reformulacion + descomposicion), soporte para docs complejos/charts/imagenes, datos estructurados y no estructurados, SOC 2 Type II, HIPAA.
- **Deploy:** Cloud SaaS unicamente.

---

### Tier 3: Knowledge management / Search

#### Guru
- **URL:** https://www.getguru.com
- **Que hace:** Knowledge management con "knowledge cards" buscables. Integracion Slack/Teams.
- **Precio:** $30/usuario/mes. Plan gratuito disponible.
- **Deploy:** SaaS unicamente.

#### GoSearch — alternativa economica a Glean
- **URL:** https://www.gosearch.ai
- **Que hace:** Enterprise AI search posicionado como alternativa a Glean a menor precio.
- **Precio:** Gratis (personal). Pro: $20/usuario/mes. Enterprise: contactar ventas.
- **Deploy:** Cloud unicamente.

#### Coveo — e-commerce y customer-facing
- **URL:** https://www.coveo.com
- **Que hace:** Plataforma de relevancia AI para search, recomendaciones, personalizacion.
- **Precio:** Desde ~$600/mes base. Modelo por consumo.
- **Deploy:** SaaS unicamente.

#### Kore.ai — search + virtual assistants
- **URL:** https://www.kore.ai
- **Que hace:** Enterprise search + chatbots + agentes con tool use.
- **Precio:** No publicado. Modelos flexibles (por request, sesion, seat, o pay-as-you-go).
- **Deploy:** Cloud, on-premises, hibrido, VPC.

---

## Frameworks open-source (toolkits, no productos)

Estos NO son competidores directos — son herramientas para construir RAG, no productos terminados.

### NVIDIA RAG Blueprint v2.5.0
- **URL:** https://github.com/NVIDIA-AI-Blueprints/rag
- **Que es:** Solucion de referencia para pipelines RAG con NVIDIA NIM microservices. Usa Milvus.
- **Licencia:** Blueprint gratis. NIMs en produccion requieren NVIDIA AI Enterprise: ~$4,500-$6,500/GPU/ano.
- **Que incluye:** NIM inference, Milvus vector DB, inferencia optimizada para GPU.
- **Que NO incluye:** Auth, RBAC, chat UI, admin panel, multi-tenant, messaging. Es infraestructura, no producto. **Esto es exactamente lo que Saldivia RAG agrega.**

### LangChain / LangGraph / LangSmith
- **URL:** https://langchain.com
- **Que es:** Framework para apps LLM. Modular. LangGraph para agentes. LangSmith para observabilidad.
- **Precio:** LangChain gratis. LangSmith: Developer gratis, Plus $39/seat/mes, Enterprise custom.
- **Fortaleza:** Composicion de agentes, workflows multi-paso, 300+ integraciones. Produccion: Klarna, Rakuten, Replit.
- **Debilidad:** No es un producto — es un toolkit. Curva de aprendizaje alta. Issues de latencia y mantenibilidad a escala.

### LlamaIndex / LlamaCloud
- **URL:** https://llamaindex.ai
- **Que es:** Framework de datos para conectar LLMs con datos privados. Especializado en ingestion, indexing, retrieval.
- **Precio:** Open source gratis. LlamaCloud: Free, Starter $50/mes (40K creditos, 5 users), Pro $500/mes (400K creditos, 15 users), Enterprise custom.
- **Fortaleza:** Retrieval-first. Mejor para RAG document-heavy.
- **Debilidad:** Solo framework, no producto turnkey.

### Haystack (deepset)
- **URL:** https://haystack.deepset.ai
- **Que es:** Framework Python para pipelines RAG + agentes en produccion.
- **Precio:** Haystack gratis (Apache 2.0). deepset Studio: gratis (1 workspace, 1 user, 100 hrs pipeline). Enterprise: custom.
- **Fortaleza:** Arquitectura de pipelines production-grade. Compliance enterprise (SOC 2, ISO 27001, HIPAA).
- **Debilidad:** Solo Python. Curva alta.

---

## Cloud provider RAG offerings (build-it-yourself con componentes)

### AWS Bedrock Knowledge Bases
- **Pricing:** $0.002/query + tokens por modelo + vector store (OpenSearch ~$0.24/OCU/hr)
- **Estimado 10K queries/mes:** ~$200-500/mes dependiendo del modelo
- **Auth/RBAC:** Via AWS IAM. Sin auth de usuario final built-in.
- **Chat UI:** No incluida. Construir propia.

### Google Vertex AI Search / RAG Engine
- **Pricing:** Free tier 10K queries/mes. Despues $14/1K queries + tokens por modelo.
- **Estimado 10K queries/mes:** ~$140+/mes solo queries + costos de modelo
- **Auth/RBAC:** Via Google Cloud IAM. Sin auth de usuario final.
- **Chat UI:** Vertex AI Studio para testing. Sin chat UI de produccion.

### Azure AI Search + Azure OpenAI
- **Pricing:** Azure AI Search Basic $74/mes, Standard S1 $245/mes + tokens por modelo.
- **Estimado 10K queries/mes:** ~$300-600/mes minimo
- **Auth/RBAC:** Via Azure AD/Entra. Document-level security disponible.
- **Chat UI:** Playground en Azure AI Foundry. Sin chat UI de produccion.

---

## Vector databases (capa de storage)

| Proveedor | Free tier | Desde | Enterprise | Deploy |
|-----------|-----------|-------|-----------|--------|
| Pinecone | Starter gratis (limitado) | $50/mes | $500/mes + custom | SaaS unicamente |
| Weaviate | Free cloud (1GB RAM) | $45/mes | $280/mes+, custom | Self-hosted (free) o cloud |
| Milvus | Gratis (open source) | ~$65/mes (Zilliz Cloud) | Custom | Self-hosted o Zilliz Cloud |
| Qdrant | Free (1GB RAM, 4GB disk) | ~$25/mes | Custom | Self-hosted (Apache 2.0) o cloud |

Saldivia RAG usa **Milvus** (open source, self-hosted) — costo $0 en hardware propio.

---

## Tabla de precios comparativa

### Productos all-in-one (por usuario/mes)

| Plataforma | Precio/usuario/mes | Minimo | Notas |
|-----------|-------------------|--------|-------|
| Glean | ~$50 | 100 seats ($60K/ano) | + $15 add-on Work AI |
| Microsoft Copilot | $20-34 + M365 | 1 seat | Total real >$50/usuario |
| ChatGPT Enterprise | ~$50-60 | Anual | Team: $25-30/usuario |
| Onyx Business | $20-25 | Ninguno | Community gratis |
| GoSearch Pro | $20 | Ninguno | Tier personal gratis |
| Guru | $30 | Ninguno | Plan gratis disponible |
| LangSmith Plus | $39/seat | 1 seat | Solo observabilidad |

### Plataformas enterprise (contratos anuales)

| Plataforma | Desde |
|-----------|-------|
| Vectara SaaS | $100K/ano |
| Vectara VPC | $250K/ano |
| Contextual AI Enterprise | Custom |
| Coveo | ~$7,200/ano base |
| Kore.ai | Custom |

### Self-hosted (Saldivia RAG model)

| Componente | Costo |
|-----------|-------|
| Hardware (RTX PRO 6000) | ~$5,000-7,000 (unica vez) |
| NVIDIA AI Enterprise | ~$4,500-6,500/GPU/ano |
| Electricidad/hosting | Variable |
| **Per-seat / per-query** | **$0** |
| **Saldivia RAG overlay** | **$0 (codigo propio)** |

**Para 50 usuarios en 3 anos:**
- Glean: 50 × $50 × 12 × 3 = **$90,000**
- Onyx Business: 50 × $20 × 12 × 3 = **$36,000**
- Saldivia RAG: ~$7,000 hardware + ~$13,500 licencia NVIDIA = **~$20,500 total**

---

## Matriz de features: Saldivia RAG vs competidores

| Feature | Saldivia RAG | Glean | Onyx | Vectara | NVIDIA Blueprint | ChatGPT Ent |
|---------|-------------|-------|------|---------|-----------------|-------------|
| Auth built-in | JWT + RBAC | SSO | SSO + RBAC | Si | **NO** | Via OpenAI |
| Multi-coleccion | Si | Si | Si | Si | Via Milvus (manual) | No |
| Chat UI | Si (Next.js 16) | Si | Si | No (API only) | **NO** | Si |
| Admin panel | Si (7 paginas) | Si | Si | Enterprise | **NO** | Basico |
| Streaming SSE | Si (AI SDK) | Si | Si | Si | Si (raw SSE) | Si |
| Citations/sources | Si | Si | Si (HHEM) | Si | Configurable | Parcial |
| Messaging interno | **Si (Plan 25)** | No | No | No | No | No |
| Custom branding | Si (design system) | No | Si (OSS) | No | Si (self-hosted) | No |
| Self-hosted | **Si (workstation)** | Dell partnership | Si (Docker/K8s) | VPC | Si (GPU req) | No |
| Air-gapped | **Si** | No | Si | No | Si | No |
| Modelo flexible | NVIDIA NIMs | 15+ LLMs | Cualquier LLM | Propio | NVIDIA NIMs | Solo OpenAI |
| Permisos por coleccion | Si (area × col) | Herencia | Herencia | N/A | No | No |
| Rate limiting | Si | Desconocido | Enterprise | Enterprise | No | No |
| Knowledge gaps detection | Planeado | No | No | Si (HHEM) | No | No |
| Deteccion de alucinaciones | No | No | No | **Si (HHEM)** | No | No |
| Conectores externos | 0 (manual) | **100+** | **40+** | N/A | 0 | Pocos |
| API publica | 18 routes | Si | Si | Si | Raw API | Si |
| Observabilidad pipeline | Logger propio | Si | Si | Enterprise | No | No |

---

## Gap analysis: donde Saldivia RAG esta atras

### 1. Conectores (gap grande)
**Competidores:** Glean tiene 100+, Onyx 40+. Conectores a Slack, Drive, Confluence, Jira, Notion, SharePoint, etc.
**Saldivia RAG:** 0 conectores automaticos. La ingesta es manual (upload o API).
**Impacto:** Sin conectores, el usuario tiene que exportar documentos de cada fuente y subirlos manualmente. Es el gap mas visible para un cliente enterprise.

### 2. Deteccion de alucinaciones
**Competidores:** Vectara tiene HHEM (Hallucination Evaluation Model) que scorea cada respuesta.
**Saldivia RAG:** No tiene scoring de alucinaciones.
**Impacto:** Medio. Para industrias reguladas (legal, salud, finanzas) es critico.

### 3. SSO nativo (SAML/OIDC)
**Competidores:** Todos los enterprise tienen SSO. Onyx, Glean, Contextual AI.
**Saldivia RAG:** JWT custom con login/password. SSO fields existen en schema (`ssoProvider`, `ssoSubject`) pero no implementado.
**Impacto:** Blocker para enterprise grande. Nadie quiere crear cuentas manuales para 500 empleados.

### 4. Observabilidad de pipeline
**Competidores:** LangSmith, Contextual AI Enterprise, Haystack Enterprise tienen dashboards de latencia, token usage, retrieval quality.
**Saldivia RAG:** Logger propio con rotacion de archivos. Sin dashboard de pipeline.

### 5. Multi-idioma
**Competidores:** Vectara tiene multi-idioma nativo. Glean indexa en cualquier idioma.
**Saldivia RAG:** `detectLanguageHint()` basico. La UI esta hardcoded en espanol.

---

## Donde Saldivia RAG ya esta a la par

| Area | Estado | Equivalente a |
|------|--------|---------------|
| Auth + RBAC granular | Roles + permisos por coleccion × area | Onyx Business |
| Chat UI con streaming | AI SDK + useChat + citations | Cualquier tier |
| Admin panel completo | 7 paginas, CRUD, config RAG | Onyx, Glean |
| Self-hosted en hardware propio | Workstation fisica | Onyx, NVIDIA Blueprint |
| Air-gapped capable | Sin dependencia cloud | Onyx, NVIDIA Blueprint |
| Costo per-seat | $0 | Solo open source |
| Design system | "Warm Intelligence" con Storybook | Diferenciador (los OSS no tienen) |

---

## Como ponerse ADELANTE de la industria (5 estrategias)

Estas no son mejoras incrementales — son features que ningun competidor ofrece hoy.

### 1. Revocacion JWT en Edge con Upstash REST

**Estado actual:** proxy.ts (Edge) no verifica revocacion en Redis. extractClaims() (Node.js) si. Hay una ventana de ataque post-logout. Clerk resuelve esto con tokens de 60 segundos (workaround, no solucion).

**Propuesta:** Usar Upstash Redis (HTTP REST) en el Edge middleware para verificar revocacion. Mantener ioredis para BullMQ (que requiere TCP).

```typescript
// proxy.ts — Edge Runtime
import { Redis } from "@upstash/redis"
const redis = Redis.fromEnv()
const isRevoked = await redis.get(`revoked:${jti}`)
if (isRevoked) return NextResponse.redirect("/login")
```

**Por que es adelante:** Ventana de ataque = cero. Ni Clerk, ni Auth0, ni Supabase verifican revocacion en Edge. Todos usan workarounds (tokens cortos, session cookies sin JWT).

**Esfuerzo:** Medio (agregar Upstash como dependencia, ajustar proxy.ts).

---

### 2. RAG multiplayer — collaborative research sessions

**Estado actual:** Chat RAG es single-player. Messaging (Plan 25) es independiente del RAG.

**Propuesta:** Combinar ambos. Un usuario hace una query RAG en un canal de messaging → todos los miembros ven el stream en tiempo real. Anotaciones colaborativas sobre las respuestas. "Mira lo que encontre" → share de sesion RAG en un canal.

**Por que es adelante:** Glean, Onyx, Vectara, ChatGPT Enterprise — todos son single-player. Nadie ofrece RAG colaborativo en tiempo real. El infra ya existe: WebSocket sidecar + canales + useChat.

**Esfuerzo:** Alto (integrar flujos de messaging + RAG streaming).

---

### 3. Knowledge gap detection automatica

**Estado actual:** `/admin/knowledge-gaps` esta en el roadmap. `audit_log` registra cada query.

**Propuesta:**
- Si el RAG responde con 0 citations → registrar la query como "gap"
- Si multiples usuarios preguntan variantes de lo mismo sin resultado → pattern de gap
- Dashboard: "Esta semana 12 queries sobre 'politica de vacaciones' no encontraron documentos. Sugerencia: ingestar el documento X."
- Alert automatico al admin cuando se detecta un tema con >5 queries sin respuesta.

**Por que es adelante:** Vectara tiene HHEM (detecta alucinaciones) pero no gaps. Ningun competidor detecta proactivamente QUE LE FALTA al knowledge base. Es la diferencia entre "te digo si me equivoco" vs "te digo que me falta".

**Esfuerzo:** Medio (analytics sobre audit_log + UI en admin).

---

### 4. Self-healing error UX con el propio RAG

**Estado actual:** `packages/logger/src/suggestions.ts` tiene `getSuggestion()` para errores comunes. Pero los errores se muestran como toasts genericos.

**Propuesta:**
- RAG falla → en vez de "Error 500", mostrar: "El servidor RAG no responde. Las colecciones X e Y tienen cache disponible. Queres consultar esas?"
- Ingesta falla → "El documento pesa 45MB. El limite es 25MB. Podes partirlo con la opcion de split en /upload."
- Rate limit alcanzado → "Alcanzaste el limite de 10 queries/hora. Tu proximo slot disponible es en 12 minutos. Mientras tanto, tus respuestas guardadas estan en /saved."

**Por que es adelante:** Todos los competidores muestran errores genericos. Nadie usa el propio sistema para ayudar al usuario a recuperarse del error.

**Esfuerzo:** Bajo-medio (mapear errores a mensajes contextuales).

---

### 5. Predictive cache warming basado en patrones de uso

**Estado actual:** Caches LRU reactivos. `collections-cache.ts` con TTL fijo.

**Propuesta:**
- Analizar `audit_log`: usuario se loguea a las 9am todos los dias → a las 8:55 pre-cargar sus colecciones en Redis
- Un area siempre consulta la misma coleccion → mantener esa coleccion siempre warm en Milvus
- Query clustering: si 3 usuarios preguntan lo mismo en 1 hora → cachear la respuesta RAG completa

**Por que es adelante:** Todos los caches en la industria son reactivos (LRU). Un cache proactivo que aprende de los patrones de uso seria unico.

**Esfuerzo:** Alto (analytics engine + cron jobs + cache strategy).

---

## Benchmark tecnico: Saldivia RAG vs patrones de produccion

Comparacion contra patrones usados por Cal.com (40K stars), Clerk, Supabase, Dub.co, next-forge, y las guias oficiales de Vercel/Next.js.

### Lo que Saldivia RAG hace IGUAL que la industria

| Patron | Saldivia RAG | Industria | Veredicto |
|--------|-------------|-----------|-----------|
| JWT en HttpOnly cookie | `makeAuthCookie()` en `jwt.ts` | Universal (Clerk, Auth0, Supabase) | Empate |
| Revocacion JWT en handler (no Edge) | `extractClaims()` en Node.js | Post-CVE-2025-29927 todos hacen esto | Empate |
| Server Components first | Default, `"use client"` solo donde necesario | Cal.com, Vercel, Next.js docs | Empate |
| Server Actions + Zod | `next-safe-action` con `authAction`/`adminAction` | Cal.com, next-forge | Empate |
| Monorepo structure | `apps/web` + `packages/{db,shared,config,logger}` | Cal.com, Dub.co, next-forge | Empate |
| State management (server-first) | useState local, sin store global | Cal.com (tRPC), Vercel patterns | Empate |
| Streaming adapter | NVIDIA SSE → AI SDK Data Stream | Patron estandar para LLMs no-estandar | Empate |
| React Compiler activo | `reactCompiler: true` en next.config | Vanguardia (no todos lo tienen) | Empate |
| Memoizacion | React.memo (9), useCallback (30+), useMemo (20+) | Industria standard | Empate |
| Promise.all sistematico | 31 usos en fetching paralelo | Cal.com, Vercel best practices | Empate |

### Lo que Saldivia RAG hace PEOR que la industria (actualizado post Plans 26-30)

| Patron | Saldivia RAG | Industria | Gap | Estado |
|--------|-------------|-----------|-----|--------|
| JWT lifetime | ~~24h unico~~ Access 15min + refresh 7d | Access 15-60min + refresh rotation | **CERRADO** (Plan 26) | Resuelto |
| SQLite config | ~~Sin PRAGMAs~~ WAL + foreign_keys + busy_timeout | WAL + synchronous=NORMAL + cache 20MB+ | **CERRADO** (Plan 26) | Resuelto |
| Redis limites | ~~Sin maxmemory~~ Configurado | Acotado + allkeys-lru | **CERRADO** (Plan 26) | Resuelto |
| Error handling UI | ~~Silencioso~~ Error feedback system | Toast sistematico | **CERRADO** (Plan 28) | Resuelto |
| Security headers | ~~3/6~~ 5/6 (falta CSP) | 6/6 en produccion | **PARCIAL** — CSP pendiente | Mejoro |
| Testing coverage | ~~38%~~ ~61% (1,059 tests) | ~88% | **PARCIAL** — subio pero falta | Mejoro |
| Next.js output | ~~Sin standalone~~ Standalone + compress | Standalone en produccion | **CERRADO** (Plan 26) | Resuelto |
| Backups | Script documentado (Plan 26) | Litestream a S3/MinIO | **PARCIAL** — script, no streaming | Mejoro |
| Redis degradation | Parcial (isRevoked fail-open) | Sistematico | Abierto | Sin cambio |
| BullMQ connections | N conexiones | Singleton compartido | Abierto | Sin cambio |
| Logging en prod | LOG_LEVEL configurable | LOG_LEVEL=WARN en produccion | Menor | Sin cambio |

### Hoja de ruta para cerrar los gaps

**~~Fase 1 — Config hardening~~** COMPLETADA (Plan 26)
- ~~SQLite PRAGMAs~~ HECHO — WAL, foreign_keys, busy_timeout
- ~~standalone + compress + optimizePackageImports~~ HECHO
- ~~Redis maxmemory~~ HECHO
- ~~Security headers~~ HECHO (5/6, falta CSP)

**~~Fase 2 — Auth hardening~~** COMPLETADA (Plan 26)
- ~~JWT access 15min + refresh 7d~~ HECHO
- ~~Backup script~~ HECHO (cron, no Litestream streaming)

**~~Fase 3 — UX y testing~~** COMPLETADA (Plans 27-29)
- ~~Error feedback system~~ HECHO (Plan 28)
- ~~Tests para ChatInterface, rbac.ts, messaging, ai-stream.ts~~ HECHO (Plans 27-29)
- ~~E2E tests~~ HECHO (Plan 29)

**Fase 4 — Diferenciadores (en progreso):**
- Self-healing error UX → **Plan 32 completado**
- SSO (SAML/OIDC) → **Plan 34 completado** (Google, Microsoft, GitHub, SAML 2.0)
- Conectores externos → **Plan 33 completado** (Google Drive, SharePoint, Confluence, Web Crawler)
- Knowledge gap detection → futuro
- Revocacion JWT en Edge (Upstash) → futuro
- RAG multiplayer → futuro
- Predictive cache warming → futuro

---

## Conclusion

Saldivia RAG esta construido sobre decisiones arquitecturales solidas — las mismas que usan Cal.com, Clerk, y los proyectos de referencia del ecosistema Next.js/Vercel. El gap no es de estructura sino de **hardening de configuracion** (SQLite, Redis, headers, backups) y **cobertura** (tests, error handling, i18n).

El diferenciador real es que **ningun competidor combina** NVIDIA NIM inference + auth/RBAC granular + multi-coleccion + chat UI + admin panel + messaging interno + streaming + citations + self-hosted en un solo producto. Los que mas se acercan (Onyx, Glean) o son cloud-only o no tienen messaging ni design system.

Las 5 estrategias de "adelante" (JWT Edge, RAG multiplayer, knowledge gaps, self-healing UX, predictive caching) aprovechan infra que ya existe en el proyecto — no requieren reescrituras, sino conectar piezas que ya estan ahi.
