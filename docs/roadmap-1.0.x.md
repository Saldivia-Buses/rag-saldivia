# Roadmap 1.0.x — UI Foundation "Claude-style"

> **Sujeto a cambios.** Este documento refleja la intención actual, no un contrato.
> Cada release se decide y planifica en detalle antes de ejecutarse.
>
> **Decisión de arquitectura:** ADR-011 — `docs/decisions/011-ui-series-1.0.x-strategy.md`

---

## Objetivo de la serie

Transformar la UI de RAG Saldivia de "Warm Intelligence" (1.0.0) a un design system basado en los tokens oficiales de Anthropic/Claude. La serie 1.0.x completa este proceso en 4 releases incrementales, sin regresiones de funcionalidad.

La referencia es **Claude.ai**: ivory cálido, borders semitransparentes finísimos (0.5px), DM Sans como font, sidebar con labels, mucho whitespace, botón primario oscuro sobre ivory.

---

## Principios de la serie

- **Cimiento primero.** Los tokens CSS son la base. Una vez bien definidos en 1.0.1, las páginas siguientes solo consumen esos tokens.
- **Sin regresiones.** Los 413+ tests de lógica pasan en cada release. Los tests de componentes también.
- **Granular.** Cada release es un plan con commits atómicos. Un plan = un release = una versión en el CHANGELOG.
- **Solo visual.** No se toca lógica de negocio, auth, DB, ni API routes en esta serie.
- **Dark mode al final.** No tiene sentido rehacer dark mode hasta que el light mode esté estabilizado.

---

## 1.0.1 — UI Foundation *(Plan 13 — en curso)*

**Tagline:** "Establecer el cimiento"

El cambio más impactante de la serie: redefinir los tokens CSS y el shell. Beneficia todas las páginas simultáneamente.

### Scope

| Área | Cambio |
|---|---|
| `globals.css` | Tokens Claude oficiales. Ivory `#FAF9F5`, borders `rgba(31,30,29,0.12)`, accent azure `#0066cc` |
| Font | Instrument Sans → DM Sans (optical sizing, 1.65 line-height) |
| Button | Default variant: accent navy → dark `#141413` (como Claude) |
| NavRail | Icon-only 44px → sidebar 196px con labels visibles |
| Login | Card shadow sin border, logo oscuro |
| /chat | SessionList re-estilizada, empty state limpio |
| /chat/[id] | ChatInterface: user msg `bg-surface-2`, assistant sin fondo, input `0.5px border` |
| Visual baselines | 22 screenshots regenerados |

### Tests al cierre
- `bun run test` → 259 ✅
- `bun run test:components` → 153 ✅
- `bun run test:visual` → 22 ✅ (baselines nuevos)
- `bun run test:a11y` → 5 ✅ (contraste `#0066cc` sobre ivory = 5.2:1)

---

## 1.0.2 — Core Pages *(Plan 14 — pendiente)*

**Tagline:** "Extender a las páginas de uso diario"

Las páginas que un usuario regular ve después del chat. Construyen sobre los tokens de 1.0.1.

### Scope tentativo

| Página | Trabajo esperado |
|---|---|
| `/collections` | Tabla con DataTable re-estilizado. Empty state mejorado. Card de colección. |
| `/collections/[name]/graph` | Frame, controles del grafo |
| `/upload` | Drop zone, lista de archivos, progress bars |
| `/extract` | Wizard steps re-estilizados |
| `/saved` | Lista de respuestas guardadas |
| `/projects` | Lista + detail |
| Componentes UI | Badge, Input, Textarea, DataTable, StatCard, EmptyPlaceholder con nuevos tokens |

### Notas
- DataTable (`@tanstack/react-table`) es complejo — requiere revisión cuidadosa de estilos
- El grafo de documentos (`/collections/[name]/graph`) usa D3 — cambio mínimo (solo frame/controles)

---

## 1.0.3 — Admin Layer *(Plan 15 — pendiente)*

**Tagline:** "Completar la cobertura"

Las 12 páginas de admin tienen su propia densidad de información. Se trabajan juntas porque comparten patrones (tablas, filtros, formularios de edición).

### Scope tentativo

| Páginas | Notas |
|---|---|
| `/admin/users`, `/admin/areas`, `/admin/permissions` | CRUD con DataTable. Modals de edición. |
| `/admin/rag-config` | Sliders de parámetros del LLM |
| `/admin/system` | Status cards, badges de estado |
| `/admin/ingestion` | Kanban con SSE — el más complejo del admin |
| `/admin/analytics` | Recharts — cuidado con colores de gráficos |
| `/admin/knowledge-gaps`, `/admin/reports`, `/admin/webhooks`, `/admin/integrations`, `/admin/external-sources` | Tablas y formularios |
| `/audit` | Tabla con filtros y URL state |
| `/settings` | Profile, password, preferences |

### Notas
- Los componentes admin (`IngestionKanban`, `AnalyticsDashboard`) tienen alta complejidad ciclomática — cambios solo en CSS
- Los colores de Recharts deben adaptarse a la paleta claude (usar `var(--accent)` en lugar de colores hardcodeados)

---

## 1.0.4 — Dark Mode *(Plan 16 — pendiente)*

**Tagline:** "Completar la experiencia"

El dark mode se rehace desde cero con los dark tokens oficiales de Claude. El toggle existente se mantiene pero el resultado final es radicalmente mejor que el actual "warm dark".

### Scope tentativo

| Área | Cambio |
|---|---|
| `globals.css .dark {}` | Tokens dark oficiales Claude: `#141413` bg, `#FAF9F5` fg, `rgba(222,220,209,0.15)` border |
| Dark mode testing | Regenerar baselines con ambos modos. Agregar test visual de dark en Storybook |
| ThemeToggle | Revisar si el componente actual funciona bien o necesita ajuste |
| WCAG dark | Verificar contraste en dark mode con axe-playwright |

### Dark tokens oficiales Claude

```css
.dark {
  --bg:          #141413;
  --surface:     #1F1E1D;
  --surface-2:   #262521;
  --border:      rgba(222, 220, 209, 0.15);
  --border-strong: rgba(222, 220, 209, 0.25);
  --fg:          #FAF9F5;
  --fg-muted:    #C2C0B6;
  --fg-subtle:   #9C9A92;
  --accent:      #4d99e0;     /* azure más claro para contraste en dark */
  --accent-hover: #5aaae8;
  --accent-subtle: rgba(77, 153, 224, 0.12);
  --primary:     #FAF9F5;     /* botón primario = ivory sobre dark */
  --primary-foreground: #141413;
}
```

---

## Fuera del scope de 1.0.x

Las siguientes features están identificadas pero se posponen a versiones posteriores (2.x o rama separada):

| Feature | Razón |
|---|---|
| Animaciones y micro-interactions | Requiere Motion/Framer — nueva dependencia. No es cimiento. |
| Responsive / mobile | El producto se usa en desktop. Mobile es futuro. |
| `/share/[token]` rediseño | Página pública — baja frecuencia de uso |
| Onboarding tour (`driver.js`) | Feature completa, no rediseño |
| SSO (Google/Azure AD) | Fue removido como dead code en Plan 9 |
| Serif heading font (Playfair Display) | No se adopta. Solo DM Sans en la app. |

---

## Estado actual (post-1.0.0)

```
1.0.1  ████░░░░░░  en curso (Plan 13 — en ejecución)
1.0.2  ░░░░░░░░░░  pendiente
1.0.3  ░░░░░░░░░░  pendiente
1.0.4  ░░░░░░░░░░  pendiente
```

---

## Cómo se actualiza este roadmap

1. Al completar cada release, marcar el estado en la tabla y actualizar la barra de progreso
2. Si el scope de un release pendiente cambia, actualizar la sección correspondiente
3. Si aparece un release nuevo (1.0.5, etc.), agregar su sección antes de "Fuera del scope"
4. Este archivo se commitea con cada plan: `docs(roadmap): actualizar estado 1.0.x — planN cierre`
