# 19 — SOTA Saldivia RAG: Vision para ser #1 de la industria

> **Fecha:** 2026-03-31
> **Proposito:** Definir que deberia tener Saldivia RAG para superar a todos los competidores
> y ser el producto RAG enterprise #1 indiscutido.
>
> **Fuentes:** Research SOTA 2026 (papers, productos, benchmarks), analisis de competidores
> (doc 18), comparacion con patrones de Cal.com/Clerk/Vercel/Perplexity/NotebookLM.
>
> **Nota:** Este es un documento de VISION — no un plan de implementacion.
> Los planes concretos se derivan de aca y van a `docs/plans/`.

---

## Principio rector

Ningun competidor combina hoy: **NVIDIA NIM inference local + auth/RBAC granular +
multi-coleccion + chat UI + admin panel + messaging interno + streaming + citations +
self-hosted + air-gapped + design system propio.**

El camino a #1 no es copiar a Glean ($60K/ano) — es construir sobre esta base unica
con tecnicas que los SaaS no pueden ofrecer (soberania de datos total, hardware
dedicado, cero latencia de red al LLM) y UX que los frameworks no incluyen.

---

## 1. CORE RAG ENGINE — lo que procesa queries

### 1.1 Agentic RAG (nadie lo tiene self-hosted)

**Que es:** El agente descompone queries complejas en sub-queries, retrieva por separado,
evalua si el contexto es suficiente, y hace rondas adicionales si no.

**Quien lo tiene:** LangGraph (framework), Perplexity Deep Research (SaaS), Google Gemini (SaaS).
Ningun producto self-hosted lo ofrece out-of-the-box.

**SOTA Saldivia RAG deberia:**
- Clasificar queries por complejidad: simple → retrieval directo, compleja → agentic loop
- Self-reflection: despues del primer retrieval, evaluar "tengo suficiente contexto?"
- Sub-query decomposition: "Compara la politica de vacaciones con la de licencias" →
  query 1: "politica de vacaciones", query 2: "politica de licencias", luego sintesis
- Tool use: el agente decide si buscar en Milvus, hacer web search, o consultar la DB

**Ventaja competitiva:** Agentic RAG corriendo en hardware local con Nemotron-49B.
Latencia de red = 0. Los SaaS tienen 100-500ms de overhead por cada hop al LLM.

---

### 1.2 RAG Fusion + HyDE (mejora de retrieval sin cambiar infra)

**RAG Fusion:** Generar 3-5 reformulaciones de la query, ejecutar retrieval para cada una,
fusionar resultados con Reciprocal Rank Fusion (RRF). Filtra topic drift naturalmente.

**HyDE:** En vez de embeddear la query del usuario, generar una "respuesta hipotetica" y
embeddear esa. El embedding de la respuesta esta mas cerca en vector space de los
documentos reales que el embedding de la pregunta.

**SOTA Saldivia RAG deberia:**
- RAG Fusion por defecto en queries normales (3 reformulaciones, RRF con k=60)
- HyDE como opcion para queries abstractas o ambiguas
- Adaptive: el sistema decide automaticamente si usar fusion, HyDE, o retrieval directo
  basado en la complejidad detectada de la query

**Ventaja competitiva:** Mejora calidad de retrieval sin cambiar Milvus ni el modelo.
Es una capa de software pura sobre la infra existente.

---

### 1.3 Graph RAG (knowledge graph + vector search)

**Que es:** Extraer un knowledge graph de los documentos (entidades + relaciones),
combinarlo con busqueda vectorial. Queries multi-hop como "Quien reporta al jefe de
finanzas y trabajo en el proyecto X?" son imposibles con vector search solo.

**Estado del arte:**
- Microsoft GraphRAG: costoso ($33K para indexar datasets grandes)
- **LightRAG** (HKUDS, 2024): dual-level retrieval (entidades + comunidades), mucho mas barato
- GraphRetriever (Meta, 2026): unifica embeddings + graph traversal

**SOTA Saldivia RAG deberia:**
- Dual retrieval: vector search para queries simples, graph traversal para multi-hop
- Router automatico que decide cuando usar graph vs vector
- LightRAG como implementacion (open-source, liviano)
- Indexado incremental del graph (no reindexar todo en cada ingesta)

**Ventaja competitiva:** Ningun competidor self-hosted tiene Graph RAG integrado.
Onyx no lo tiene. NVIDIA Blueprint no lo tiene. Seria una feature unica.

---

### 1.4 Corrective RAG (CRAG) — auto-correccion

**Que es:** Despues de cada retrieval, un evaluador scorea la confianza:
- **Correcto:** proceder con generacion
- **Ambiguo:** complementar con busqueda adicional
- **Incorrecto:** descartar documentos, buscar fuentes alternativas

**SOTA Saldivia RAG deberia:**
- Evaluador de confianza post-retrieval (puede ser el mismo LLM con un prompt especifico)
- Si confianza < umbral → re-query con reformulacion automatica
- Si 0 resultados relevantes → fallback a web search o respuesta "no tengo informacion
  sobre esto" con sugerencia de que documentos ingestar

**Ventaja competitiva:** El sistema nunca da respuestas de baja calidad sin avisar.
Se combina con knowledge gap detection (seccion 3.1).

---

### 1.5 Multimodal RAG (imagenes, tablas, charts en documentos)

**Que es:** Los PDFs empresariales tienen tablas, charts, diagramas, fotos. El RAG
tradicional los ignora o extrae texto roto.

**Estado del arte:**
- **ColPali/ColQwen:** Embeddea paginas completas como imagenes (no necesita OCR).
  Late interaction scoring sobre tokens visuales. "Supera ampliamente los pipelines
  modernos de document retrieval siendo drasticamente mas simple."
- **Unstructured.io:** Parseo universal de documentos con 4 estrategias (Fast, Hi-Res,
  VLM para charts/diagramas, Auto)

**SOTA Saldivia RAG deberia:**
- Parseo de documentos con Unstructured.io (estrategia Hi-Res o VLM para docs complejos)
- Extraccion de tablas como datos estructurados (no como texto plano)
- Charts/diagramas: describir con VLM y embeddear la descripcion
- Soporte para imagenes en respuestas (mostrar el chart original del documento)

**Ventaja competitiva:** La mayoria de los RAG enterprise ignoran contenido visual.
Con GPU local, el costo de VLM processing es $0 marginal.

---

### 1.6 Hallucination Detection + Faithfulness Scoring

**Que es:** Scorear cada respuesta del RAG en una escala de "groundedness" — que tan
fundamentada esta en los documentos fuente.

**Estado del arte:**
- Vectara HHEM-2.3: modelo de scoring (hay version open-source)
- NLI-based entailment: verificar si cada claim esta soportado por una citation
- Span-level attribution: mapear cada fragmento de respuesta a un fragmento fuente

**SOTA Saldivia RAG deberia:**
- Faithfulness score visible en cada respuesta (badge: "Alta confianza" / "Verificar fuentes")
- Si score < umbral → advertencia visual al usuario
- Citation verification: cada citation linkeada clickeable muestra el pasaje exacto
- Admin dashboard: distribucion de scores, queries con bajo faithfulness

**Ventaja competitiva:** Solo Vectara tiene HHEM integrado, y es SaaS de $100K/ano.
Un scoring similar self-hosted seria unico en el mercado.

---

### 1.7 Evaluacion automatica (RAGAS + custom metrics)

**Que es:** Medir la calidad del RAG automaticamente, no solo cuando un usuario reporta un problema.

**Metricas SOTA (framework RAGAS):**
- **Faithfulness:** La respuesta esta fundamentada en el contexto? (0-1)
- **Answer Relevancy:** La respuesta contesta la pregunta? (0-1)
- **Context Precision:** Los chunks top-ranked son relevantes?
- **Context Recall:** Se recuperaron todos los chunks relevantes?

**SOTA Saldivia RAG deberia:**
- Evaluacion automatica en background (muestreo de N queries/dia)
- Dashboard de calidad RAG: faithfulness promedio, relevancy trend, precision/recall
- Alertas cuando la calidad cae (ej: despues de ingestar documentos nuevos)
- Regression testing: set de queries gold con respuestas esperadas

**Ventaja competitiva:** La mayoria de los RAG no miden calidad automaticamente.
El admin se entera de problemas solo cuando un usuario se queja.

---

## 2. CONNECTORS & INGESTION — lo que entra

### 2.1 Plugin system de conectores

**Gap actual:** 0 conectores automaticos. Glean tiene 100+, Onyx 40+.

**SOTA Saldivia RAG deberia:**
- Arquitectura de plugins: interfaz estandar `IConnector` con metodos
  `authenticate()`, `listDocuments()`, `fetchDocument()`, `detectChanges()`
- Conectores prioritarios (cubren 80% de casos enterprise):
  1. Google Drive
  2. SharePoint / OneDrive
  3. Confluence
  4. Notion
  5. Slack (mensajes como conocimiento)
  6. Email (IMAP)
  7. Web crawler (URLs publicas)
  8. GitHub/GitLab (repos como documentacion)
- Cada conector corre como BullMQ job con retry y backoff
- Schedule configurable por conector (hourly/daily/weekly)
- Change detection: solo ingestar documentos nuevos o modificados

**Nota:** La tabla `external_sources` y el worker `external-sync.ts` ya existen en
el schema y en `_archive/`. La infraestructura esta preparada.

---

### 2.2 Parseo inteligente de documentos

**Estado del arte (SCORE-Bench, Unstructured.io 2025):**

| Estrategia | Velocidad | Calidad | Costo | Caso de uso |
|-----------|-----------|---------|-------|-------------|
| Fast (rule-based) | Rapida | Basica | Gratis | Texto plano, markdown |
| Hi-Res (layout model) | Media | Alta | Gratis | PDFs con columnas, headers |
| VLM (vision LLM) | Lenta | Maxima | Modelo | Charts, tablas complejas, forms |
| Auto | Adaptiva | Adaptiva | Variable | Mix de documentos |

**SOTA Saldivia RAG deberia:**
- Parseo auto-adaptivo: detectar tipo de documento y elegir estrategia
- Tablas extraidas como JSON estructurado (no texto plano)
- Charts/diagramas descritos por VLM y embeddeados
- Metadata enriquecida: autor, fecha, seccion, tipo de documento
- Preview del parseo antes de confirmar ingesta

---

### 2.3 Chunking inteligente

**El chunking es el factor #1 que diferencia un RAG bueno de uno malo** (consenso industria 2025).

**Estado del arte:**
- **Semantic chunking:** Splitear en boundaries de tema, no por token count
- **Hierarchical chunking:** Parent chunks (contexto amplio) + child chunks (precision).
  Search usa child, generation usa parent. Desacoplar "buscar" de "leer."
- **Late chunking (Jina, 2025):** Embeddear el documento completo primero, luego
  chunkear los embeddings — preserva contexto global
- **Document structure-aware:** Respetar headings, secciones, tablas como boundaries naturales

**SOTA Saldivia RAG deberia:**
- Chunking hibrido: semantic + structure-aware
- Parent-child chunks: busqueda granular, generacion con contexto
- Metadata por chunk: seccion, pagina, tipo de contenido
- Admin puede previsualizar chunks antes de confirmar ingesta

---

### 2.4 Incremental indexing + change detection

**SOTA Saldivia RAG deberia:**
- Content hashing por chunk: en re-ingesta, solo procesar chunks con hash diferente
- Timestamp + version tracking por documento fuente
- Webhooks de sistemas fuente (Drive, SharePoint) para push notifications de cambios
- Dashboard de freshness: "Ultima ingesta hace 3 dias, 12 documentos modificados desde entonces"

---

## 3. INTELLIGENCE LAYER — lo que aprende

### 3.1 Knowledge Gap Detection (ningun competidor lo tiene)

**Que es:** Detectar proactivamente que le falta al knowledge base.

**SOTA Saldivia RAG deberia:**
- Registrar queries con 0 citations como "gap"
- Clustering de gaps: agrupar queries similares sin respuesta
- Alertas al admin: "Esta semana 15 queries sobre 'politica de viajes' sin resultado"
- Sugerencias: "Documentos recomendados para ingestar basado en gaps detectados"
- Metricas: gap rate (% queries sin respuesta), trend temporal, gaps por area/coleccion

**Ventaja competitiva:** Vectara detecta alucinaciones. Nadie detecta gaps.
Es la diferencia entre "te digo si me equivoco" vs "te digo que me falta".

---

### 3.2 Predictive Cache Warming (nadie lo hace)

**Que es:** Usar patrones de uso historicos para pre-calentar caches.

**SOTA Saldivia RAG deberia:**
- Analizar `audit_log`: horarios de uso por usuario → pre-cargar colecciones antes del login
- Pattern detection: si 3+ usuarios hacen la misma query en 1 hora → cachear respuesta RAG
- Colecciones "siempre warm": areas que siempre consultan X → mantener X en cache
- Time-of-day warming: pre-cargar datos del admin dashboard antes del horario laboral

---

### 3.3 Learning from Feedback (feedback loop cerrado)

**SOTA Saldivia RAG deberia:**
- Feedback up/down ya existe en `message_feedback`
- Usar feedback negativo para: re-rankear chunks (bajar score de chunks que producen
  respuestas malas), ajustar RAG params por coleccion
- Usar feedback positivo para: identificar "respuestas gold" para regression testing
- Sugerir al admin: "Las respuestas sobre el tema X tienen 80% feedback negativo.
  Los documentos fuente podrian estar desactualizados."

---

### 3.4 Bidirectional RAG (knowledge base que crece con el uso)

**Estado del arte (Dec 2025):** El knowledge base crece durante el uso.
Respuestas verificadas (NLI entailment + source attribution + novelty detection)
se agregan al corpus de retrieval.

**SOTA Saldivia RAG deberia:**
- Respuestas con alto faithfulness score + feedback positivo → candidatas para write-back
- Admin aprueba antes de agregar (human-in-the-loop)
- "Curated answers" como documentos internos del knowledge base
- Casos de uso: FAQs que se generan solas, mejores practicas destiladas de queries repetidas

---

## 4. SECURITY & COMPLIANCE — lo que protege

### 4.1 SSO nativo (SAML/OIDC)

**Gap actual:** JWT custom con login/password. Campos SSO en schema pero no implementados.
**Blocker:** Ningun enterprise grande va a crear cuentas manuales para 500 empleados.

**SOTA Saldivia RAG deberia:**
- SAML 2.0 para enterprise (Azure AD, Okta, OneLogin)
- OIDC para cloud-first (Google, GitHub, Auth0)
- Provisioning automatico: primer login SSO crea usuario con rol default
- Herencia de grupos: grupo SSO → area en Saldivia RAG

---

### 4.2 JWT Edge Revocation (Upstash REST)

**SOTA Saldivia RAG deberia:**
- Access token: 15 min (corto)
- Refresh token: 7 dias (rotado en cada uso)
- Revocacion verificada en Edge via Upstash Redis REST
- Refresh token family tracking: si un refresh se usa dos veces → revocar toda la familia

---

### 4.3 Document-Level Access Control

**Estado del arte (AWS Bedrock pattern, 2025):**
Cada chunk lleva metadata de ACL. Pre-filtrado en retrieval: antes de buscar en Milvus,
aplicar permisos del usuario como filtro de metadata.

**SOTA Saldivia RAG deberia:**
- Metadata ACL por chunk en Milvus (owner, area, classification level)
- Pre-retrieval filter: chunks que el usuario no puede ver se excluyen de la busqueda
- Zero-trust: nunca pasar documentos restringidos al LLM si el usuario no tiene acceso
- Audit trail: que usuario accedio a que documento, cuando

**Ya existe parcialmente:** `area_collections` con permisos read/write/admin por area.
Falta llevar esto al nivel de chunk en Milvus.

---

### 4.4 PII Detection/Redaction

**SOTA Saldivia RAG deberia:**
- Deteccion de PII en respuestas del RAG (nombres, emails, DNI, telefonos)
- Redaccion por rol: admin ve todo, user ve redactado
- Configuracion por coleccion: "coleccion RRHH" siempre redacta PII para no-admins
- Herramientas: Presidio (Microsoft, open-source) o NER custom

**Cuidado:** Redaccion agresiva rompe coherencia del texto (problema documentado).
Implementar con granularidad — no todo es PII.

---

### 4.5 Security Headers completos + CSP

**SOTA Saldivia RAG deberia tener 6/6:**
- X-Content-Type-Options: nosniff (ya tiene)
- X-Frame-Options: DENY (ya tiene)
- Referrer-Policy: strict-origin-when-cross-origin (ya tiene)
- Strict-Transport-Security: max-age=63072000; includeSubDomains
- Permissions-Policy: camera=(), microphone=(), geolocation=()
- Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; ...

---

### 4.6 Compliance readiness

**SOTA Saldivia RAG deberia:**
- Audit trail completo (ya tiene `audit_log` + `events`)
- Data retention policies configurables por coleccion
- Export de datos de usuario (GDPR right to access)
- Delete de datos de usuario (GDPR right to erasure)
- Cifrado de credenciales en DB (external_sources.credentials hoy en texto plano)
- SOC 2 checklist documentado (no certificacion, pero si readiness)

---

## 5. UX & UI — lo que se ve

### 5.1 Sistema de Artifacts (como Claude)

**Que hacen Claude y ChatGPT:**
- Panel lateral dedicado que renderiza codigo, HTML, SVG, diagramas en vivo
- Version history con back/forward
- Highlight-to-edit: seleccionar parte del artifact y pedir cambios

**SOTA Saldivia RAG deberia:**
- `ArtifactPanel` ya existe (archivado). Soporta code, html, svg, mermaid, table, text
- Agregarle: version history, edicion inline, export (PDF, markdown)
- Artifacts como "resultados de investigacion" — el usuario arma un documento a partir
  de respuestas del RAG

---

### 5.2 Citations interactivas (como Perplexity)

**Que hace Perplexity:**
- Inline numbered citations [1][2][3] en cada claim
- Click en citation → popup con el pasaje exacto del documento fuente
- Sources panel con favicon, titulo, y dominio

**Saldivia RAG ya tiene:** Citations via `data-sources` en el stream, SourcesPanel.

**SOTA deberia agregar:**
- Click en citation → highlight del pasaje exacto en el documento
- Faithfulness badge por respuesta ("Alta confianza" / "Verificar fuentes")
- "Ver en contexto" — abrir el documento fuente con el pasaje resaltado

---

### 5.3 Deep Research mode

**Que hacen Perplexity/Gemini/ChatGPT:**
Multi-step agentic research: planifica sub-queries, busca, lee, sintetiza un reporte
estructurado con citations completas.

**SOTA Saldivia RAG deberia:**
- Modo "Investigacion profunda": el usuario hace una pregunta compleja
- El sistema muestra el plan de sub-queries en tiempo real ("Buscando X... Buscando Y...")
- Resultado: reporte estructurado con secciones, citations, y conclusiones
- Export como PDF o markdown

---

### 5.4 Audio Overviews (como NotebookLM)

**Que hace NotebookLM:**
Un click genera una conversacion tipo podcast entre dos voces AI que discuten
tus documentos. Fue la feature viral de 2025.

**SOTA Saldivia RAG deberia:**
- "Generar resumen en audio" de una coleccion o sesion de chat
- Dos voces AI discuten los hallazgos clave del knowledge base
- Util para: onboarding de nuevos empleados, resumen de cambios recientes,
  briefings ejecutivos sin leer documentos

**Ventaja competitiva:** Ningún RAG enterprise ofrece audio overviews hoy.
Con NVIDIA NIMs locales, text-to-speech es factible sin costo cloud.

---

### 5.5 Admin Analytics Dashboard (como ChatGPT Enterprise)

**Que tiene ChatGPT Enterprise (Marzo 2026):**
- Adoption funnel: Access → Activation → Weekly Usage → Depth
- Filtros por fecha y grupo/departamento
- Cohort analysis: comparar equipos
- Analytics viewer role (read-only para stakeholders)

**SOTA Saldivia RAG deberia:**
- **Uso:** Usuarios activos, sesiones/usuario, queries/sesion
- **Calidad:** Faithfulness promedio, queries sin resultado (gaps), feedback ratio
- **Performance:** Latencia P50/P95, tokens consumidos, tiempo de streaming
- **Costos:** GPU utilization, queries/dia, storage usado
- **Adoption:** Onboarding completion, feature adoption, user retention
- Filtros por area, coleccion, rol, fecha
- Export de reportes (PDF, CSV)
- Alertas configurables (ej: "faithfulness cayo abajo de 0.7")

---

### 5.6 Onboarding interactivo

**Estado del arte (Tandem case study 2025):**
Onboarding con AI agent contextual → completion rate de 11% a 64%.

**SOTA Saldivia RAG deberia:**
- First-run wizard: 5 pasos para el admin (crear areas, invitar usuarios, ingestar
  primer documento, hacer primera query, revisar resultados)
- Onboarding progresivo para usuarios: tooltips contextuales, no tutorial obligatorio
- Sandbox con datos de ejemplo para probar antes de ingestar datos reales
- Checklist visible con progreso ("3 de 5 pasos completados")

**Ya existe parcialmente:** `onboardingCompleted` en schema de users.

---

### 5.7 Mobile & Responsive

**SOTA Saldivia RAG deberia:**
- Responsive design (no app nativa — PWA)
- Thumb-first: acciones principales al alcance del pulgar
- Voice input en mobile (speech-to-text para queries)
- Dark mode adaptativo (ya en progreso, Plan 20)
- Streaming optimizado para conexiones lentas

---

### 5.8 Accessibility WCAG AAA

**Gap documentado en 09-testing.md:** Falta focus management, skip-to-content,
aria-live para streaming, labels en iconos.

**SOTA Saldivia RAG deberia:**
- aria-live="polite" para streaming (buffear parrafos completos antes de anunciar)
- Skip-to-content link
- Focus management en navegacion SPA
- Keyboard navigation completa en messaging (mentions, reactions)
- High contrast mode
- Screen reader testing como parte del CI

---

## 6. COLLABORATION — lo que conecta personas

### 6.1 RAG Multiplayer (nadie lo tiene)

**SOTA Saldivia RAG deberia:**
- Query RAG compartida en un canal de messaging → todos ven el stream
- Anotaciones colaborativas sobre respuestas del RAG
- "Research spaces" — proyectos de investigacion compartidos con contexto persistente
- "Mira lo que encontre" → share de sesion RAG en un canal con 1 click

**Infra existente:** WebSocket sidecar + channels + useChat. Falta conectar los flujos.

---

### 6.2 Workspaces / Projects (como Claude Projects)

**Que hace Claude:**
Projects con 200K context window. Documentos, instrucciones custom, actividad de equipo.

**SOTA Saldivia RAG deberia:**
- Projects que agrupan sesiones + documentos + instrucciones
- Context persistente por proyecto (el RAG "recuerda" el contexto del proyecto)
- Custom system prompt por proyecto ("para este proyecto, responder en formato ejecutivo")
- Actividad de equipo visible

**Ya existe parcialmente:** Tabla `projects` con `sessions` y `collections` en el schema.

---

## 7. OPERATIONS & OBSERVABILITY — lo que se mide

### 7.1 Pipeline Observability (Langfuse self-hosted)

**Estado del arte:** Langfuse (open-source, self-hostable). Traces de:
- Llamadas al LLM (tokens, latencia, costo)
- Operaciones de retrieval (queries a Milvus, chunks devueltos)
- Evaluacion de calidad por trace

**SOTA Saldivia RAG deberia:**
- Langfuse self-hosted como sidecar
- Trace completo de cada query: retrieval → generation → response
- Metricas: latencia por fase, tokens por query, retrieval precision
- UI para explorar traces individuales (debugging de queries malas)

---

### 7.2 Alerting

**SOTA Saldivia RAG deberia:**
- Alerta: RAG server caido
- Alerta: faithfulness promedio cayo abajo de umbral
- Alerta: gap rate subio (muchas queries sin respuesta)
- Alerta: Redis memory >80%
- Alerta: SQLite WAL file >100MB (necesita checkpoint)
- Canales: email, Slack webhook, notificacion en el admin panel

---

### 7.3 Backups (Litestream)

**SOTA Saldivia RAG deberia:**
- Litestream: streaming continuo del WAL a S3/MinIO
- Recovery point: <1 segundo de datos perdidos
- Restore testeado trimestralmente (backup que no se testea no existe)
- Backup de Redis: RDB snapshots cada hora

---

## 8. SCALE & ARCHITECTURE — lo que crece

### 8.1 Postgres migration path

**Cuando:** >50 usuarios concurrentes o >1M documentos indexados.

**SOTA Saldivia RAG deberia:**
- Drizzle ya soporta Postgres — migracion de 1 dia (ADR-001)
- Connection pooling (PgBouncer o Drizzle pool)
- Read replicas para queries de analytics (no bloquear writes)

---

### 8.2 Multi-tenant (si se vende como producto)

**SOTA Saldivia RAG deberia:**
- Database-per-tenant (SQLite por cliente, o schema-per-tenant en Postgres)
- Isolation: un tenant no puede ver datos de otro
- Billing: per-seat o per-query medido por tenant
- Admin super: gestionar todos los tenants

---

### 8.3 API publica + SDK

**SOTA Saldivia RAG deberia:**
- REST API documentada (OpenAPI spec)
- SDK TypeScript para integraciones
- Webhooks para eventos (nueva query, nuevo documento, gap detectado)
- MCP server: permitir que agentes AI externos consulten el RAG

---

## 9. INTEGRACIONES — lo que se conecta

### 9.1 Slack/Teams bots

**Ya existe parcialmente:** `bot_user_mappings` en schema, API routes en `_archive/`.

**SOTA Saldivia RAG deberia:**
- Bot de Slack: `/ask` query al RAG, respuesta con citations en el canal
- Bot de Teams: igual
- Respuestas con formatting nativo (Block Kit para Slack, Adaptive Cards para Teams)
- Permiso heredado: el bot respeta los permisos del usuario que pregunta

---

### 9.2 MCP Server (Saldivia RAG como tool para agentes)

**SOTA Saldivia RAG deberia:**
- MCP server que expone: `search(query, collection)`, `listCollections()`, `getDocument(id)`
- Agentes externos (Claude Code, Cursor, etc.) pueden consultar el RAG como tool
- Auth via SYSTEM_API_KEY o JWT

---

## 10. SCORECARD: Estado actual vs SOTA

| Categoria | Hoy | A la par | SOTA (#1) |
|-----------|-----|----------|-----------|
| **Core RAG** | | | |
| Query basica | Si | — | — |
| RAG Fusion | No | Si | — |
| Agentic RAG | No | — | Si |
| Graph RAG | No | — | Si |
| CRAG (auto-correccion) | No | — | Si |
| Multimodal (tablas, charts) | No | Si | — |
| Hallucination detection | No | Si | — |
| Evaluacion automatica (RAGAS) | No | — | Si |
| **Connectors** | | | |
| Conectores externos | 0 | 10+ | 40+ |
| Parseo inteligente docs | Basico | Hi-Res | VLM adaptive |
| Chunking inteligente | Token-based | Semantic | Hierarchical |
| Incremental indexing | No | Si | Si + webhooks |
| **Intelligence** | | | |
| Knowledge gap detection | No | — | Si |
| Predictive caching | No | — | Si |
| Feedback loop | Parcial (up/down) | Score-based reranking | Bidirectional RAG |
| **Security** | | | |
| Auth JWT | Si (24h) | Access+refresh | Edge revocation |
| SSO (SAML/OIDC) | No | Si | Si + group sync |
| Document-level ACL | Por area | Por chunk | Zero-trust pre-filter |
| PII redaction | No | Si | Role-based |
| Security headers | 3/6 | 6/6 | 6/6 + CSP nonces |
| Backups | No | Cron | Litestream streaming |
| Compliance | Audit log | GDPR ready | SOC 2 ready |
| **UX** | | | |
| Chat streaming | Si | — | — |
| Citations | Si (panel) | Inline + click | Highlight en doc |
| Artifacts | Archivado | Panel + preview | Versioning + export |
| Deep Research | No | — | Si |
| Audio Overviews | No | — | Si |
| Admin analytics | Basico | ChatGPT Ent level | + quality metrics |
| Onboarding | Parcial | Wizard | AI-guided adaptive |
| Mobile/responsive | Parcial | PWA | Thumb-first + voice |
| A11y | Parcial (52 attrs) | WCAG AA | WCAG AAA + streaming |
| Dark mode | Tokens (no refinado) | Refinado | Adaptive + OLED |
| **Collaboration** | | | |
| Messaging | Si (Plan 25) | — | — |
| RAG multiplayer | No | — | Si |
| Workspaces/Projects | Schema existe | Funcional | Persistent context |
| **Operations** | | | |
| Pipeline observability | Logger | Langfuse basic | Langfuse + alerting |
| Alerting | No | Email/Slack | Multi-canal + ML |
| Testing | 38% | 70% | 88%+ |
| **Scale** | | | |
| Postgres path | Drizzle ready | Migrated | + read replicas |
| Multi-tenant | No | DB-per-tenant | Full isolation |
| API publica | 18 routes | OpenAPI spec | + SDK + MCP |

---

## Hoja de ruta sugerida (de hoy a SOTA)

### Fase 0 — Hardening (1 semana)
Cero features nuevas. Solo config y coverage.
- SQLite PRAGMAs, Redis limits, security headers, standalone output
- JWT access+refresh rotation
- Backups (Litestream)
- Tests para ChatInterface, rbac.ts, ai-stream.ts
- Error handling con toast sistematico

### Fase 1 — Retrieval Quality (2-3 semanas)
Mejorar la calidad de las respuestas sin cambiar UI.
- RAG Fusion (multi-query + RRF)
- Chunking hibrido (semantic + structure-aware)
- Faithfulness scoring basico
- Knowledge gap detection (analytics sobre audit_log)
- RAGAS evaluation en background

### Fase 2 — UX Next Level (3-4 semanas)
La experiencia que ningún competidor tiene.
- Citations interactivas (click → pasaje exacto)
- Deep Research mode (agentic multi-step)
- Admin analytics dashboard completo
- RAG multiplayer (queries en canales de messaging)
- Onboarding wizard

### Fase 3 — Connectors & Ingestion (3-4 semanas)
Abrir la puerta a datos enterprise.
- Plugin system con interfaz IConnector
- Google Drive + Confluence + SharePoint (top 3)
- Parseo Hi-Res con Unstructured.io
- Incremental indexing con content hashing

### Fase 4 — Advanced Intelligence (4-6 semanas)
Lo que separa un RAG bueno de uno excepcional.
- Agentic RAG con self-reflection
- Graph RAG (LightRAG)
- CRAG (auto-correccion)
- Predictive cache warming
- Bidirectional RAG

### Fase 5 — Enterprise Ready (4-6 semanas)
Lo que necesita un cliente para comprar.
- SSO (SAML/OIDC)
- Document-level ACL (pre-filter en Milvus)
- PII detection/redaction
- API publica con OpenAPI spec
- MCP server
- Compliance checklist (SOC 2 readiness)

### Fase 6 — Diferenciadores (ongoing)
Features unicas que definen la marca.
- Audio Overviews (NotebookLM-style)
- Multimodal RAG (ColPali/ColQwen)
- Workspaces con persistent context
- Mobile PWA con voice input
- WCAG AAA

---

## Conclusion

El camino de Saldivia RAG a #1 no es construir todo lo que tienen los competidores.
Es construir lo que **nadie tiene** sobre una base que **nadie mas puede replicar**
facilmente: hardware dedicado con GPU local, cero latencia al LLM, soberania de datos
total, air-gapped capable, y un costo marginal de $0 por query.

Los SaaS cobran $20-60/usuario/mes porque pagan por compute en cada query.
Saldivia RAG no tiene ese costo. Eso permite features que para ellos son
economicamente inviables: Agentic RAG con 5 rondas de retrieval, Graph RAG con
traversal exhaustivo, Audio Overviews generados on-demand, evaluacion automatica
en cada query.

La GPU local no es una limitacion — es la ventaja competitiva fundamental.
