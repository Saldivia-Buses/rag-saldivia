# Frontend Review — Plan 29 Fases F2-F5 (Component Tests)

**Fecha:** 2026-04-01
**Tipo:** review
**Intensity:** thorough
**Scope:** 23 archivos de test, 115 tests (79 messaging + 36 admin), 1 setup file

## Resultado

**CAMBIOS REQUERIDOS** (0 bloqueantes, 5 correcciones, 6 sugerencias)

---

## Convenciones

| Regla | Status | Notas |
|---|---|---|
| `afterEach(cleanup)` | PASS | 100% de archivos lo tienen |
| Queries escopadas (no `screen`) | PASS | 0 imports de `screen` |
| `fireEvent` (no `userEvent`) | PASS | 0 imports de `userEvent` |
| Import de `bun:test` | PASS | Todos usan `describe/test/expect` de `bun:test` |
| Naming `ComponentName.test.tsx` | PASS | Todos siguen convenci&oacute;n |

---

## Hallazgos

### Bloqueantes

Ninguno.

### Debe corregirse

**1. [ThreadPanel.test.tsx:42] DOM traversal fragil con `parentElement?.querySelector`**

```ts
const closeBtn = heading.parentElement?.querySelector("button")
closeBtn?.click()
```

Navega el DOM con optional chaining -- si el markup cambia, el test pasa silenciosamente sin clickear nada (no falla, simplemente no ejecuta el assertion de forma confiable). El `?.click()` traga el null sin error.

**Fix:** Usar `getByRole("button", { name: /cerrar/i })` o agregar un `aria-label="Cerrar"` al boton del componente y buscarlo con `getByLabelText`. Si no es posible, al menos assert que `closeBtn` no es null antes de clickear.

**2. [AdminDashboard.test.tsx:5-9] Timer cleanup manual es fragil y peligroso**

```ts
afterEach(() => {
  cleanup()
  const id = setTimeout(() => {}, 0) as unknown as number
  for (let i = 0; i < id; i++) clearInterval(i)
})
```

Este patron crea un timeout para obtener el ID actual y luego limpia TODOS los timers desde 0 hasta ese ID. Esto puede limpiar timers de otros tests corriendo en paralelo, y el cast `as unknown as number` esconde que `setTimeout` en Bun retorna un `Timer`, no un numero. Es un footgun en suite paralela.

**Fix:** Usar `mock.module` para mockear `setInterval`/`setTimeout` en el componente, o usar `vi.useFakeTimers()` equivalent de bun:test si existe. Alternativa minima: extraer el interval logic del componente a un hook testeable aparte.

**3. [Inconsistencia de assertions entre archivos admin] AdminRagConfig, AdminAreas, AdminCollections usan `toBeInTheDocument()` mientras que AdminUsers, AdminRoles, AdminDashboard, AdminPermissions, small-components usan `toBeTruthy()`**

Mezclar estilos de assertion en el mismo directorio de tests es confuso. `toBeInTheDocument()` es mas semantico (verifica que el elemento esta en el document), `toBeTruthy()` solo verifica que no es null/undefined/false.

**Fix:** Estandarizar en `toBeInTheDocument()` para todos, o en `toBeTruthy()` para todos. Dado que el setup ya extiende expect con jest-dom matchers, `toBeInTheDocument()` es preferible.

**4. [AdminRagConfig.test.tsx:7-10, AdminCollections.test.tsx:7-10, AdminAreas.test.tsx:7-13] Mocks duplicados con component-test-setup.ts**

Estos 3 archivos definen sus propios `mock.module` para actions que ya estan mockeadas en `component-test-setup.ts` (lineas 86-113). El setup global ya mockea `@/app/actions/admin`, `@/app/actions/roles`, `@/app/actions/areas`. Las redefiniciones locales en estos archivos pueden causar confusion sobre cual mock esta activo.

**Fix:** Remover los `mock.module` locales de estos 3 archivos ya que el global los cubre. Si necesitan retornos distintos, usar `mock.mockImplementation()` en el test individual.

**5. [ChannelList.test.tsx] No testea unread counts**

El componente recibe `unreadCounts` como prop y renderiza `UnreadBadge` por canal, pero todos los tests pasan `unreadCounts={{}}`. No se verifica que los badges de unread aparezcan cuando hay mensajes sin leer.

**Fix:** Agregar test:
```ts
test("shows unread badge when channel has unreads", () => {
  const { getByText } = render(
    <ChannelList channels={CHANNELS} unreadCounts={{ "ch-1": 3 }} userId={1} />
  )
  expect(getByText("3")).toBeTruthy()
})
```

### Sugerencias

**1. Componentes sin tests: VoiceInput, PresenceIndicator, FileAttachment**

El comment en `small-components.test.tsx` dice "TypingIndicator, UnreadBadge, PresenceIndicator, FileAttachment, VoiceInput" pero solo testea TypingIndicator y UnreadBadge. Tres componentes mencionados no tienen tests:

- `VoiceInput` -- difícil de testear (SpeechRecognition API), pero se puede testear que renderiza null cuando no hay soporte
- `PresenceIndicator` -- trivial, 3 status visuales con aria-labels testeables
- `FileAttachment` -- image vs download link branch, facil de testear

**2. Admin components sin tests: UserRoleSelector, RoleCard, RoleForm, CreateUserForm, PasswordResetCell**

Son 5 componentes admin sin tests. Algunos (CreateUserForm, RoleForm) tienen logica de formulario y validacion que seria bueno cubrir. No es bloqueante pero es un gap.

**3. [MessageComposer.test.tsx] No testea Enter para enviar / Shift+Enter para newline**

El componente tiene `handleKeyDown` que intercepta Enter (send) y permite Shift+Enter (newline). Solo se testea click en boton de enviar. Agregar:
```ts
test("sends on Enter key", () => {
  const onOptimistic = mock(() => {})
  const { getByPlaceholderText } = render(
    <MessageComposer channelId="ch-1" currentUserId={1} members={MEMBERS} onOptimisticMessage={onOptimistic} />
  )
  const textarea = getByPlaceholderText("Escribí un mensaje...")
  fireEvent.change(textarea, { target: { value: "Hello" } })
  fireEvent.keyDown(textarea, { key: "Enter" })
  expect(onOptimistic).toHaveBeenCalled()
})
```

**4. [MessageActions.test.tsx] No testea callbacks (onEdit, onDelete, onReply, onPin)**

Los tests verifican que botones aparecen/no aparecen segun permisos (buen patron), pero nunca verifican que clickear los botones invoca los callbacks. Es mitad del test.

**5. [ChannelView.test.tsx] Solo 3 tests, 2 son smoke tests**

De 3 tests, "renders without crash" y "renders with empty messages" son smoke tests puros (solo verifican `container.firstChild`). El componente tiene state management (replyTo, threadMessage, optimisticMessages) que no se testea en absoluto. Es el componente que orquesta toda la vista de canal.

**6. Factor out helpers repetidos**

`mkMsg()`, `MEMBERS`, y `CHANNELS` se definen casi identicos en 5+ archivos. Considerar un `__tests__/fixtures.ts` compartido:
```ts
// __tests__/fixtures.ts
export const MEMBERS = [...]
export const mkMsg = (overrides = {}) => ({...})
export const CHANNELS = [...]
```

---

## Analisis de cobertura

| Categoria | Tests | % del total |
|---|---|---|
| Renderizado correcto de props/data | 62 | 54% |
| Conditional rendering (show/hide, permisos) | 25 | 22% |
| User interaction (click, input, filter) | 16 | 14% |
| Smoke tests (renders without crash) | 7 | 6% |
| Edge cases (empty, null, boundary) | 5 | 4% |

**Estimacion:** ~60% son tests de comportamiento real, ~34% son verificaciones de rendering (valiosos pero superficiales), ~6% son smoke tests puros.

---

### Lo que esta bien

- **Convenciones perfectas**: 100% compliance con afterEach(cleanup), queries escopadas, fireEvent. Cero violaciones.
- **MessageItem.test.tsx es el mejor archivo**: 11 tests cubriendo edited, deleted, system messages, reply count singular/plural, unknown user fallback. Es el modelo a seguir.
- **MessageActions.test.tsx tiene buen patron de permisos**: testea edit/delete visible para owner, oculto para otros, admin override. Exactamente como se debe testear RBAC en UI.
- **DirectMessageDialog.test.tsx cubre filtrado**: search filters, selection count, disabled state. Buen test de interactividad.
- **ReactionPicker.test.tsx cubre el ciclo completo**: open > select > callback. Conciso y suficiente.
- **Mocks en component-test-setup.ts bien estructurados**: WebSocket client, messaging actions, admin actions, areas, error-feedback, clipboard. Cobertura amplia sin over-mocking.
- **TypingIndicator tests cubren los 3 branches**: 1 user, 2 users, 3+ users. Completo.
- **UnreadBadge tests cubren edge cases**: 0, normal, 99+, negativo. Buen boundary testing.
- **Data fixtures usan `Date.now()`** en lugar de timestamps hardcodeados -- no se rompen con el tiempo.
