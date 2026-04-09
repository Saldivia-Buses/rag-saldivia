# Component Map — Source Selection

> Para cada tipo de componente, elegir de qué librería usamos.
> Formato: `componente — librería elegida`
> Opciones: shadcnblocks (sb), tailgrids (tg), fluid (fl), iconiqui (iq), abui (ab), magic-ui (mu), ai-elements (ai), supabase-ui (su), shadcn/ui base (base), reactbits (rb)

---

## Core UI Components

| Componente | sb | tg | fl | iq | ab | base | Elegido |
|---|---|---|---|---|---|---|---|
| accordion | 21 | 1 | 1 | - | 1 | 1 | |
| alert | 25 | 1 | - | 1 | - | 1 | sb ✅ |
| alert-dialog | 39 | 1 | - | - | - | - | sb ✅ |
| aspect-ratio | 7 | 1 | - | - | - | - | |
| avatar | 34 | 1 | - | - | - | 1 | sb ✅ |
| avatar-group | 9 | - | - | - | - | - | sb ✅ |
| badge | 20 | 1 | 1 | 1 | - | 1 | iq+tg+fl ✅ |
| breadcrumb | 14 | 1 | - | 1 | - | 1 | sb ✅ |
| button | 35 | 1 | 1 | 1 | 1 | 1 | sb+fl ✅ |
| button-group | 39 | 1 | - | - | - | 1 | sb ✅ |
| calendar | 16 | - | - | - | - | - | |
| card | 5 | 1 | - | - | - | 1 | sb ✅ |
| carousel | 4 | 1 | - | - | - | 1 | |
| chart | - | - | - | 1 | - | - | |
| checkbox | 12 | 1 | - | - | - | - | sb ✅ |
| collapsible | 23 | 1 | - | - | - | 1 | sb ✅ |
| color-picker | - | - | - | - | 1 | - | |
| combobox | 42 | 1 | - | - | - | - | |
| command | 21 | 1 | - | - | - | 1 | |
| context-menu | 27 | 1 | - | - | - | - | |
| data-table | 8 | - | - | - | - | - | |
| date-picker | 8 | - | - | - | - | - | |
| dialog | 17 | 1 | 1 | - | - | 1 | |
| drawer | 22 | 1 | - | - | - | 1 | |
| dropdown-menu | 30 | 1 | 1 | - | - | 1 | |
| empty | 22 | - | - | - | - | - | |
| error-boundary | - | - | - | - | - | - | |
| field | 38 | 1 | - | - | - | - | |
| file-tree | - | - | - | 1 | - | - | |
| file-upload | 44 | - | - | - | - | 1 | |
| form | 82 | - | - | - | - | - | |
| hover-card | 20 | 1 | - | - | - | 1 | |
| input | 24 | 1 | - | 1 | - | 1 | |
| input-group | 39 | 1 | - | 1 | - | 1 | |
| kbd | 39 | - | - | - | - | - | |
| label | 8 | 1 | - | 1 | - | 1 | |
| link | - | 1 | - | - | - | - | |
| list | - | 1 | - | - | - | - | |
| loading/spinner | 17 | - | - | - | 1 | 1 | |
| logo | - | - | - | - | - | - | |
| menubar | 10 | 1 | - | - | - | - | |
| navigation-menu | 20 | 1 | - | - | - | - | |
| otp-input | 20 | 1 | - | - | - | - | |
| pagination | 19 | 1 | - | - | - | - | |
| popover | 15 | 1 | - | - | - | 1 | |
| price | - | - | - | - | - | - | |
| progress | 20 | 1 | - | - | - | 1 | |
| radio-group | 9 | 1 | - | 1 | - | - | |
| rating | - | - | - | - | - | - | |
| resizable | - | 1 | - | - | - | - | |
| rich-text-editor | - | - | - | - | - | - | |
| scroll-area | 38 | 1 | - | - | - | 1 | |
| scrollable-tabs | - | - | - | - | - | - | |
| search/command-palette | 21 | - | - | - | - | 1 | |
| select | - | 1 | 1 | 1 | - | 1 | |
| separator | 18 | 1 | - | - | - | 1 | |
| sheet | 29 | 1 | - | - | - | 1 | |
| sidebar | - | 1 | - | - | - | 1 | |
| skeleton | 30 | 1 | - | - | - | 1 | |
| slider | 29 | 1 | - | 1 | - | - | |
| social-button | - | 1 | - | - | - | - | |
| sonner/toast | 24 | 1 | - | - | - | - | |
| stepper | - | - | - | - | - | - | |
| switch | 19 | - | 1 | - | - | 1 | |
| table | 8 | 1 | - | - | - | - | |
| tabs | 11 | 1 | - | 1 | - | 1 | |
| textarea | 13 | 1 | - | - | - | 1 | sb ✅ |
| time-picker | - | 1 | - | - | - | - | |
| toggle | 7 | 1 | - | - | - | - | sb+tg ✅ |
| toggle-group | 7 | - | - | - | - | - | sb+tg ✅ |
| toolbar | - | - | - | - | - | - | |
| tooltip | 8 | 1 | 1 | 1 | 1 | 1 | sb ✅ |

---

## Especiales por librería

### shadcnblocks (sb) — únicos
- empty states (22 variantes: data, search, actions)
- kbd shortcuts display (39 variantes)
- form patterns (82 variantes: validation, multi-field, layouts)
- combobox advanced (multi-select, grouped, custom-actions, rich-content)
- file-upload (44 variantes: special, validation, zones)
- data-table (8 variantes)

### ai-elements (ai) — AI/LLM UI
- conversation
- message
- reasoning
- chain-of-thought
- code-block
- prompt-input
- tool
- agent
- artifact
- sandbox
- terminal
- sources
- inline-citation
- attachments
- model-selector
- voice-selector / speech-input / mic-selector
- canvas / jsx-preview / web-preview

### magic-ui (mu) — efectos y animaciones
- marquee
- globe
- dock
- shimmer-button
- blur-fade
- border-beam
- animated-beam
- magic-card
- particles
- retro-grid
- dot-pattern / grid-pattern
- ripple
- confetti / cool-mode
- number-ticker
- animated-gradient-text
- typing-animation
- word-rotate

### abui (ab) — únicos
- accordion-multiselect
- animated-chart
- availability calendar
- calendar-year
- color-swatch-selector
- label-selector
- radio-tabs / radio-rows
- scroll-progress
- scroll-reveal-content
- text-gradient
- timeline / timeline-steps
- table of contents (toc)

### fluid (fl) — animaciones fluidas
- select (animated)
- dropdown (animated)
- dialog (animated)
- tooltip (animated)
- switch (animated)
- accordion (animated)
- badge (animated)
- button (animated)

### supabase-ui (su) — auth + realtime
- password-based-auth (login, signup, forgot-password, update-password)
- social-auth
- logout-button
- current-user-avatar
- realtime-chat
- realtime-cursor
- realtime-avatar-stack
- realtime-monaco (collaborative editor)
- dropzone (file upload)
- infinite-query hook

### iconiqui (iq) — animated icons
- 384 animated icons (Lucide-based, motion-powered)
- animated-badges
- file-tree
- chart
- slider
- tabs (animated)

### tailgrids (tg) — complete set
- 54 components (full shadcn-compatible set)
- social-button (único)
- time-picker (único)
- multi-combobox (único)
- range-date picker (único)

### reactbits (rb) — effects + creative (copy-paste)
- Text: split-text, blur-text, glitch-text, decrypted-text, ascii-text, scroll-velocity
- Cursor: blob-cursor, splash-cursor, ghost-cursor, target-cursor, crosshair
- Cards: pixel-card, tilted-card, spotlight-card, reflective-card, decay-card
- Nav: flowing-menu, gooey-nav, pill-nav, dock, bubble-menu, staggered-menu
- Layout: masonry, magic-bento, scroll-stack, stack, fluid-glass
- Backgrounds: silk, aurora, waves, particles, iridescence, liquid-chrome, hyperspeed
- Effects: electric-border, star-border, magnet, sticker-peel, pixel-trail

---

## Para decidir

Marcá la columna "Elegido" con la sigla de la librería.
Si querés ver un componente antes de elegir, decime el nombre.
