# Security Audit — Plan 23 Phase 3: AI Review Gates — 2026-04-14

## Resumen ejecutivo

Los workflows de AI review y @claude assist estan bien estructurados en su
conception. DS1 (runner isolation) esta completamente satisfecho. DS6 (anti-
injection) esta parcialmente satisfecho: el AI output nunca se interpola en
shell, los heredoc delimiters son aleatorios, y los prompts tienen instrucciones
anti-injection. Sin embargo, hay dos vulnerabilidades de alta prioridad: la
logica de scoring falla silenciosamente en condiciones normales de uso (el
modelo con frecuencia produce output en code fences, lo que hace que jq no
parsee y el scoring se bypasea con exit 0), y las acciones de terceros estan
pinadas a tags mutables en lugar de SHAs inmutables. No hay findings criticos
que bloqueen deploy, pero los dos HIGH deben resolverse antes de que los gates
sean confiables.

---

## CRITICOS (bloquean deploy)

Ninguno.

---

## ALTOS (corregir antes de produccion)

### [ai-review.yml:66-76] Scoring se bypasea silenciosamente cuando el modelo produce code fences

**Problema:** Los tres prompts instruyen al modelo a responder con "ONLY valid
JSON (no markdown, no code fences)". Sin embargo, los ejemplos del output format
en los propios prompts estan dentro de code fences markdown (` ``` `). El modelo
frecuentemente produce salida como:

```
```json
{"findings": [...], "summary": "..."}
```
```

Cuando esto ocurre, `jq -e '.findings' "$EXEC_FILE"` falla porque el archivo
no es JSON valido. El codigo cae al branch `grep -oP '\{[\s\S]*\}'`, que en
grep sin `-z` trata `\s` y `\S` como caracteres de una sola linea (`.` no
matchea newlines por default). El JSON multilinea no es extraido. El archivo
`/tmp/review-output.json` queda vacio, se llega al check `[ ! -s ]`, y se hace
`exit 0`. El review "pasa" aunque haya findings criticos.

Este escenario no es hipotetico: es el comportamiento por defecto del modelo
cuando el prompt tiene un ejemplo en code fence. En produccion, este gate
*parece* funcionar pero no bloquea nada.

**Fix:**
1. Remover los code fences del ejemplo en los tres prompt files. El ejemplo
   debe ser texto plano, no markdown fenced.
2. Agregar `-z` a grep para habilitar multiline matching:
   ```bash
   grep -ozP '(?s)\{.*\}' "$EXEC_FILE" | tr -d '\0' > /tmp/review-output.json 2>/dev/null || true
   ```
3. Agregar un extractor de code fence como fallback adicional:
   ```bash
   # Si el modelo envuelve en ```json ... ```
   sed -n '/^```\(json\)\?$/,/^```$/{ /^```/d; p }' "$EXEC_FILE" > /tmp/review-output.json
   ```

**Afecta:** Los tres jobs de ai-review.yml. El gate de security review es el
mas critico — si este falla silenciosamente, ningun finding de seguridad bloquea
un PR.

---

### [ai-review.yml:46, claude-assist.yml:36] Acciones de terceros pinadas a tags mutables, no SHAs

**Problema:** Ambos workflows usan:
- `anthropics/claude-code-action@v1`
- `actions/checkout@v4`

Los tags como `@v1` son mutables: el propietario del repo puede mover el tag a
cualquier commit. Si `anthropics/claude-code-action` es comprometido o
introduce un cambio breaking (incluso involuntario), el workflow se ve afectado
en el proximo run sin ninguna notificacion.

El riesgo especifico de `claude-code-action` es mayor que el de `checkout`
porque tiene acceso a `ANTHROPIC_API_KEY` y puede leer el contenido del repo.
Un tag `@v1` comprometido podria exfiltrar secrets.

**Fix:** Pinear a commit SHA inmutable. Ejemplo:
```yaml
uses: anthropics/claude-code-action@<SHA-del-commit-de-v1>
uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2
```

Agregar un comentario con el tag correspondiente para legibilidad:
```yaml
uses: anthropics/claude-code-action@abc123def456  # v1.x.y
```

GitHub Dependabot puede mantener estos SHAs actualizados automaticamente.

---

## MEDIOS (backlog prioritario)

### [quality.md:88-103, security.md:108-126, dependencies.md:80-97] Output format examples usan code fences — instruction contradicts example

**Problema:** Los tres prompts dicen "Respond with ONLY valid JSON (no markdown,
no code fences)" pero inmediatamente muestran el formato dentro de code fences.
Este es el origen directo del HIGH anterior. Ademas de arreglar el scoring, los
ejemplos deben ser texto plano para evitar que el modelo copie el formato que ve.

**Fix:** Cambiar los ejemplos en los tres prompts de:

```
```
{"findings": [...]}
```
```

a:

```
Example output:
{"findings": [{"severity": "critical", "file": "...", "line": 42, "issue": "...", "fix": "..."}], "summary": "..."}
```

---

### [ai-review.yml:79, 139, 199] Fallback `|| echo "0"` en jq silencia errores de parse

**Problema:** La expresion:
```bash
BLOCKERS=$(jq '[.findings[] | select(.severity == "critical" or .severity == "high")] | length' /tmp/review-output.json 2>/dev/null || echo "0")
```
Redirige stderr a /dev/null y hace fallback a "0" si jq falla por cualquier
razon (JSON malformado, campo inesperado, error de I/O). El resultado es que
BLOCKERS=0 y el job pasa. Un atacante que logre hacer que el AI produzca JSON
invalido (via prompt injection en el diff) podria hacer que el security review
pase siempre.

**Fix:** Separar la verificacion de parseo de la logica de scoring:
```bash
if ! BLOCKERS=$(jq '[.findings[] | select(.severity == "critical" or .severity == "high")] | length' /tmp/review-output.json); then
  echo "::error::jq parse failed — treating as blocking to be safe"
  exit 1
fi
```
Si no se puede parsear el output, fail-closed (exit 1), no fail-open (exit 0).

---

### [claude-assist.yml:26] Concurrency group puede serializar todos los PR review comments

**Problema:**
```yaml
group: claude-assist-${{ github.event.issue.number || github.event.pull_request.number }}
```
Para eventos `pull_request_review_comment`, `github.event.issue.number` es null
y `github.event.pull_request.number` es el numero del PR. Esto funciona.

Sin embargo, para eventos `issue_comment` en issues (no PRs),
`github.event.pull_request.number` es null y `github.event.issue.number` es el
numero del issue. Esto tambien funciona.

El problema es cuando ambos son null (condicion de edge case en algunos
payloads malformados o eventos edge): el group se convierte en
`claude-assist-` (string vacio despues del prefijo), colisionando con todos
los demas eventos que tengan el mismo issue/null. En la practica esto es
improbable pero podria causar que una respuesta en curso sea encolada
inesperadamente.

**Fix:** Agregar un fallback explicito:
```yaml
group: claude-assist-${{ github.event.issue.number || github.event.pull_request.number || 'unknown' }}
```

---

### [ai-review.yml:56, 121, 181] execution_file path via env: es seguro pero requiere nota

**Observacion (no es hallazgo):** `EXEC_FILE: ${{ steps.review.outputs.execution_file }}`
aparece en `env:`, no en una expresion `run:` inline. Esto es el patron correcto
segun DS6 — el contenido del AI no se interpola en shell. Sin embargo, si
`claude-code-action@v1` fuera comprometido (ver HIGH anterior), podria retornar
un path malicioso como `../../etc/passwd` o `/tmp/../../root/.ssh/id_rsa`. El
`[ -f "$EXEC_FILE" ]` check no valida que el path este dentro de un directorio
seguro.

**Fix menor:** Agregar validacion de path si se pina el SHA de la action:
```bash
EXEC_FILE="${EXEC_FILE:-}"
if [[ "$EXEC_FILE" != /tmp/* ]] && [[ "$EXEC_FILE" != /home/runner/* ]]; then
  echo "::error::Unexpected execution_file path: $EXEC_FILE"
  exit 1
fi
```
Este fix es de menor urgencia si se resuelve el SHA pinning.

---

## BAJOS (nice to have)

### [claude-assist.yml:20-23] COLLABORATOR excluido del gate

El gate actual es `MEMBER || OWNER`. Un repo collaborator externo (con permisos
de write/maintain en el repo pero sin ser miembro de la org) no puede usar
`@claude`. Esto puede ser intencional si el repo es privado y todos los
colaboradores son internos. Si en algun momento se agregan colaboradores
externos de confianza, se debe revisar si agregar `COLLABORATOR` al gate.

**Fix (si aplica):** Agregar `COLLABORATOR` si se necesita:
```yaml
(github.event.comment.author_association == 'MEMBER' ||
 github.event.comment.author_association == 'OWNER' ||
 github.event.comment.author_association == 'COLLABORATOR')
```

---

### [ai-review.yml:50, 114, 175] `use_sticky_comment: "true"` — no blocking finding

Los tres reviews usan sticky comments. Cada sincronizacion del PR reemplaza el
comentario anterior. Esto es correcto para UX pero significa que si el primer
review encontro criticos y el segundo run (tras un push vacio) no puede parsear
el output, el comentario con los criticos desaparece. El bloqueo de merge debe
depender del job exit code (rama de status checks), no del contenido del
comentario.

**Verificar:** Que los tres jobs de `ai-review.yml` esten configurados como
required status checks on branch protection for `2.0.5`
y `2.0.x`. Si no estan como required checks, los gates no bloquean merge aunque
fallen.

---

### [ai-review.yml, claude-assist.yml] No hay rate limit de creditos de API

DS6 menciona "API key con spending limit mensual" como mitigacion. Esto no es
visible en el codigo (es una configuracion de Anthropic Console). Verificar que
el spending limit este configurado. Sin el, un miembro de la org podria spamear
`@claude` mentions o abrir muchos PRs en rapida sucesion para agotar el budget.

**Verificar externamente:** Anthropic Console → API Keys → Spending Limits.

---

## DS1 y DS6 — Verificacion de compliance

| Decision | Requisito | Estado |
|----------|-----------|--------|
| DS1 | CI en ubuntu-latest, nunca self-hosted | PASS — los 3 jobs + claude-assist usan `ubuntu-latest` |
| DS6 | AI output no interpolado en shell via `${{ }}` | PASS — output via `env: EXEC_FILE`, leido desde archivo |
| DS6 | Heredoc delimiters aleatorios (no EOF fijo) | PASS — `PROMPT_DELIM_7f3a`, `PROMPT_DELIM_9b2c`, `PROMPT_DELIM_4e1d` |
| DS6 | Anti-injection en system prompts | PASS — todos los prompts tienen Security Notice |
| DS6 | `@claude` restringido a MEMBER/OWNER | PASS — gate en if: condicional del job |
| DS6 | Output estructurado JSON con severity scores | PARTIAL — scoring implementado pero bypaseable (ver HIGH) |

---

## Tenant isolation audit

No aplica directamente a estos archivos. Los workflows de CI no tienen acceso
a bases de datos de tenants. El AI review solo lee el diff del PR (contenido
del repo), no datos de produccion.

El unico vector indirecto: si un PR con codigo malicioso pasa el review gate
(por el bypass del scoring), podria introducir vulnerabilidades de tenant
isolation en el codebase. Este riesgo es exactamente lo que el HIGH de scoring
hace concreto — un PR con SQL injection cross-tenant podria pasar el security
review si el modelo produce output en code fences.

---

## Faltantes respecto al spec

1. **SHA pinning de actions** — DS6 menciona seguridad de supply chain pero no
   especifica SHA pinning. Recomendado como extension de DS6.

2. **Required status checks** — el plan no documenta si los 3 jobs son required
   checks en branch protection. Sin esto, los gates son informativos pero no
   bloquean merge.

3. **Spending limit** — DS6 lo menciona como mitigacion de abuse pero no hay
   forma de verificarlo desde el codigo. Debe verificarse en Anthropic Console.

---

## CVEs

Ninguno aplicable a los archivos auditados. Las versiones `@v1` de
`claude-code-action` y `@v4` de `actions/checkout` no tienen CVEs conocidos
a la fecha de este audit. El riesgo es de supply chain (mutable tags), no CVE.

---

## Veredicto: APTO con condiciones

Los gates de AI review son funcionales como mecanismo de feedback pero NO son
confiables como mecanismo de bloqueo en su estado actual. El bypass de scoring
via code fences hace que el security gate falle silenciosamente en condiciones
normales de uso del modelo.

**Condiciones para APTO sin restricciones:**
1. Fix del scoring (remover code fences de ejemplos en prompts + fix del grep
   multiline + fail-closed en jq parse error)
2. SHA pinning de `anthropics/claude-code-action`
3. Confirmar que los 3 jobs son required status checks en branch protection

Hasta resolver estos dos HIGH, los gates deben tratarse como indicativos, no
como guardianes de merge.
