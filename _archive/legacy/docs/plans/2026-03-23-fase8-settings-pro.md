# Fase 8 — Settings Pro Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Personalización completa del usuario: perfil con avatar/color, cambio de contraseña, preferencias RAG persistidas en DB, idioma UI, y notificaciones in-app de ingesta y alertas.

**Architecture:** Columna JSON `preferences` en la tabla `users` de SQLite. El gateway expone 4 nuevos endpoints REST. El BFF carga las prefs en `+layout.server.ts` e hidrata un store reactivo. La página de settings se reescribe con 5 secciones usando SvelteKit form actions.

**Tech Stack:** Python/FastAPI + SQLite + bcrypt, SvelteKit 5 runes, Vitest, pytest

---

### Task 1: DB — columna `preferences` + métodos CRUD

**Files:**
- Modify: `saldivia/auth/database.py`
- Create: `saldivia/tests/test_fase8.py`

**Contexto:** `AuthDB` vive en `saldivia/auth/database.py`. La función `init_db_conn(conn)` corre todas las DDL migrations — agregar el ALTER TABLE ahí. Los métodos nuevos van en la clase `AuthDB`. Los tests usan `AuthDB(":memory:")`.

**Step 1: Escribir tests que fallan**

Crear `saldivia/tests/test_fase8.py`:

```python
"""Tests para Fase 8 — Settings Pro (backend)."""
import pytest
from saldivia.auth.database import AuthDB


@pytest.fixture
def db():
    import bcrypt
    d = AuthDB(":memory:")
    pw = bcrypt.hashpw(b"pass", bcrypt.gensalt()).decode()
    with d._conn() as conn:
        conn.execute("INSERT INTO areas (id, name) VALUES (1, 'Test')")
        conn.execute(
            "INSERT INTO users (id, email, name, password_hash, role, area_id, api_key_hash) "
            "VALUES (1, 'u@test.com', 'U', ?, 'user', 1, 'dummy')", (pw,)
        )
    return d


def test_get_user_preferences_defaults(db):
    """Usuario sin prefs retorna dict con defaults."""
    prefs = db.get_user_preferences(1)
    assert prefs["default_query_mode"] == "standard"
    assert prefs["vdb_top_k"] == 10
    assert prefs["avatar_color"] == "#6366f1"
    assert prefs["ui_language"] == "es"
    assert prefs["notify_ingestion_done"] is True


def test_update_user_preferences_merge(db):
    """update_user_preferences hace merge parcial — preserva campos no incluidos."""
    db.update_user_preferences(1, {"vdb_top_k": 20})
    prefs = db.get_user_preferences(1)
    assert prefs["vdb_top_k"] == 20
    assert prefs["reranker_top_k"] == 5  # campo no tocado conservado


def test_update_user_name(db):
    """update_user_name actualiza el nombre."""
    db.update_user_name(1, "Nuevo Nombre")
    with db._conn() as conn:
        row = conn.execute("SELECT name FROM users WHERE id=1").fetchone()
    assert row[0] == "Nuevo Nombre"


def test_update_user_password_correct(db):
    """update_user_password retorna True cuando la contraseña actual es correcta."""
    result = db.update_user_password(1, "pass", "nueva123")
    assert result is True
    # Verificar que la nueva contraseña funciona
    import bcrypt
    with db._conn() as conn:
        row = conn.execute("SELECT password_hash FROM users WHERE id=1").fetchone()
    assert bcrypt.checkpw(b"nueva123", row[0].encode())


def test_update_user_password_wrong(db):
    """update_user_password retorna False cuando la contraseña actual es incorrecta."""
    result = db.update_user_password(1, "incorrecta", "nueva123")
    assert result is False
    # La contraseña original sigue siendo válida
    import bcrypt
    with db._conn() as conn:
        row = conn.execute("SELECT password_hash FROM users WHERE id=1").fetchone()
    assert bcrypt.checkpw(b"pass", row[0].encode())
```

**Step 2: Correr tests para verificar que fallan**

```bash
cd /Users/enzo/rag-saldivia/.worktrees/fase8-settings-pro
uv run pytest saldivia/tests/test_fase8.py -v
```

Expected: FAIL — `AttributeError: 'AuthDB' object has no attribute 'get_user_preferences'`

**Step 3: Implementar**

En `saldivia/auth/database.py`, dentro de `init_db_conn(conn)`, agregar migration al final (antes del `conn.close()` en `init_db`):

```python
    # Migration: add preferences column to users
    try:
        conn.execute("ALTER TABLE users ADD COLUMN preferences TEXT NOT NULL DEFAULT '{}'")
    except Exception:
        pass  # Column already exists
```

Luego, en la clase `AuthDB`, agregar estos métodos (después de `update_api_key`):

```python
_DEFAULT_PREFERENCES = {
    "default_collection": "",
    "default_query_mode": "standard",
    "vdb_top_k": 10,
    "reranker_top_k": 5,
    "max_sub_queries": 4,
    "follow_up_retries": True,
    "show_decomposition": False,
    "avatar_color": "#6366f1",
    "ui_language": "es",
    "notify_ingestion_done": True,
    "notify_system_alerts": True,
}

def get_user_preferences(self, user_id: int) -> dict:
    with self._conn() as conn:
        row = conn.execute(
            "SELECT preferences FROM users WHERE id = ?", (user_id,)
        ).fetchone()
    if not row:
        return dict(self._DEFAULT_PREFERENCES)
    stored = json.loads(row[0] or "{}")
    return {**self._DEFAULT_PREFERENCES, **stored}

def update_user_preferences(self, user_id: int, prefs: dict):
    current = self.get_user_preferences(user_id)
    merged = {**current, **prefs}
    with self._conn() as conn:
        conn.execute(
            "UPDATE users SET preferences = ? WHERE id = ?",
            (json.dumps(merged), user_id)
        )

def update_user_name(self, user_id: int, name: str):
    with self._conn() as conn:
        conn.execute(
            "UPDATE users SET name = ? WHERE id = ?", (name, user_id)
        )

def update_user_password(self, user_id: int, current_pw: str, new_pw: str) -> bool:
    import bcrypt
    with self._conn() as conn:
        row = conn.execute(
            "SELECT password_hash FROM users WHERE id = ?", (user_id,)
        ).fetchone()
    if not row or not bcrypt.checkpw(current_pw.encode(), row[0].encode()):
        return False
    new_hash = bcrypt.hashpw(new_pw.encode(), bcrypt.gensalt()).decode()
    with self._conn() as conn:
        conn.execute(
            "UPDATE users SET password_hash = ? WHERE id = ?", (new_hash, user_id)
        )
    return True
```

**Step 4: Correr tests para verificar que pasan**

```bash
uv run pytest saldivia/tests/test_fase8.py -v
```

Expected: 5 tests PASS

**Step 5: Correr todos los tests**

```bash
uv run pytest saldivia/tests/ -q --tb=short
```

Expected: 149 passed

**Step 6: Commit**

```bash
git add saldivia/auth/database.py saldivia/tests/test_fase8.py
git commit -m "feat(fase8): preferences column + CRUD methods in AuthDB"
```

---

### Task 2: Gateway — 4 endpoints de settings

**Files:**
- Modify: `saldivia/gateway.py`
- Modify: `saldivia/tests/test_fase8.py`

**Contexto:** El gateway usa `Depends(get_user_from_token)` para auth. El patrón es idéntico al de los endpoints de chat. Buscar cerca de la línea 650 (sección `/auth/me`).

**Step 1: Agregar tests al archivo existente**

Agregar al final de `saldivia/tests/test_fase8.py`:

```python
from fastapi.testclient import TestClient
from unittest.mock import patch
from saldivia.gateway import app
from saldivia.auth.models import User, Role, generate_api_key


@pytest.fixture
def gw_client():
    return TestClient(app, raise_server_exceptions=False)


@pytest.fixture
def user_gw():
    key, hash_val = generate_api_key()
    return User(id=1, email="u@test.com", name="U", area_id=1,
                role=Role.USER, api_key_hash=hash_val)


def test_get_preferences_endpoint(gw_client, user_gw):
    """GET /auth/me/preferences retorna las prefs del usuario."""
    prefs = {"vdb_top_k": 15, "ui_language": "en"}
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.get_user_preferences.return_value = prefs
        resp = gw_client.get(
            "/auth/me/preferences?user_id=1",
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
    assert resp.json()["vdb_top_k"] == 15


def test_patch_preferences_endpoint(gw_client, user_gw):
    """PATCH /auth/me/preferences actualiza prefs."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.update_user_preferences.return_value = None
        resp = gw_client.patch(
            "/auth/me/preferences?user_id=1",
            json={"vdb_top_k": 20},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
    assert resp.json()["ok"] is True
    mock_db.update_user_preferences.assert_called_once_with(1, {"vdb_top_k": 20})


def test_patch_profile_endpoint(gw_client, user_gw):
    """PATCH /auth/me/profile actualiza el nombre."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.update_user_name.return_value = None
        resp = gw_client.patch(
            "/auth/me/profile?user_id=1",
            json={"name": "Nuevo Nombre"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
    assert resp.json()["ok"] is True


def test_patch_password_wrong(gw_client, user_gw):
    """PATCH /auth/me/password retorna 400 si contraseña incorrecta."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.update_user_password.return_value = False
        resp = gw_client.patch(
            "/auth/me/password?user_id=1",
            json={"current_password": "wrong", "new_password": "nueva123"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 400


def test_patch_password_correct(gw_client, user_gw):
    """PATCH /auth/me/password retorna 200 si contraseña correcta."""
    with patch("saldivia.gateway.BYPASS_AUTH", False), \
         patch("saldivia.gateway.db") as mock_db:
        mock_db.get_user_by_api_key_hash.return_value = user_gw
        mock_db.update_user_password.return_value = True
        resp = gw_client.patch(
            "/auth/me/password?user_id=1",
            json={"current_password": "pass", "new_password": "nueva123"},
            headers={"Authorization": "Bearer rsk_dummy"}
        )
    assert resp.status_code == 200
```

**Step 2: Correr tests para verificar que fallan**

```bash
uv run pytest saldivia/tests/test_fase8.py -v -k "endpoint"
```

Expected: FAIL — 404 Not Found (endpoints no existen aún)

**Step 3: Implementar en `saldivia/gateway.py`**

Agregar después del endpoint `GET /auth/me` (cerca de línea 665):

```python
class UpdatePreferencesRequest(BaseModel):
    preferences: dict


class UpdateProfileRequest(BaseModel):
    name: str


class UpdatePasswordRequest(BaseModel):
    current_password: str
    new_password: str


@app.get("/auth/me/preferences")
def get_preferences(user_id: int, user: User = Depends(get_user_from_token)):
    """Get user preferences."""
    if user is None:
        raise HTTPException(status_code=401, detail="Not authenticated")
    if user.role != Role.ADMIN and user.id != user_id:
        raise HTTPException(status_code=403, detail="Access denied")
    return db.get_user_preferences(user_id)


@app.patch("/auth/me/preferences")
def update_preferences(body: dict, user_id: int,
                        user: User = Depends(get_user_from_token)):
    """Update user preferences (partial merge)."""
    if user is None:
        raise HTTPException(status_code=401, detail="Not authenticated")
    if user.role != Role.ADMIN and user.id != user_id:
        raise HTTPException(status_code=403, detail="Access denied")
    db.update_user_preferences(user_id, body)
    return {"ok": True}


@app.patch("/auth/me/profile")
def update_profile(body: UpdateProfileRequest, user_id: int,
                    user: User = Depends(get_user_from_token)):
    """Update user display name."""
    if user is None:
        raise HTTPException(status_code=401, detail="Not authenticated")
    if user.role != Role.ADMIN and user.id != user_id:
        raise HTTPException(status_code=403, detail="Access denied")
    name = body.name.strip()
    if not name:
        raise HTTPException(status_code=400, detail="El nombre no puede estar vacío")
    db.update_user_name(user_id, name)
    return {"ok": True}


@app.patch("/auth/me/password")
def update_password(body: UpdatePasswordRequest, user_id: int,
                     user: User = Depends(get_user_from_token)):
    """Change user password. Returns 400 if current password is wrong."""
    if user is None:
        raise HTTPException(status_code=401, detail="Not authenticated")
    if user.role != Role.ADMIN and user.id != user_id:
        raise HTTPException(status_code=403, detail="Access denied")
    if len(body.new_password) < 8:
        raise HTTPException(status_code=400, detail="La contraseña debe tener al menos 8 caracteres")
    ok = db.update_user_password(user_id, body.current_password, body.new_password)
    if not ok:
        raise HTTPException(status_code=400, detail="Contraseña actual incorrecta")
    return {"ok": True}
```

**Nota:** FastAPI no acepta `body: dict` directamente en PATCH — usar un modelo genérico o `Request`. Alternativa limpia:

```python
from fastapi import Request

@app.patch("/auth/me/preferences")
async def update_preferences(request: Request, user_id: int,
                              user: User = Depends(get_user_from_token)):
    if user is None:
        raise HTTPException(status_code=401, detail="Not authenticated")
    if user.role != Role.ADMIN and user.id != user_id:
        raise HTTPException(status_code=403, detail="Access denied")
    body = await request.json()
    db.update_user_preferences(user_id, body)
    return {"ok": True}
```

**Step 4: Correr tests**

```bash
uv run pytest saldivia/tests/test_fase8.py -v
```

Expected: todos los tests PASS (10 total)

**Step 5: Correr todos los tests**

```bash
uv run pytest saldivia/tests/ -q --tb=short
```

Expected: 154 passed

**Step 6: Commit**

```bash
git add saldivia/gateway.py saldivia/tests/test_fase8.py
git commit -m "feat(fase8): gateway endpoints — get/patch preferences, profile, password"
```

---

### Task 3: BFF — tipos + funciones gateway client

**Files:**
- Create: `services/sda-frontend/src/lib/types/preferences.ts`
- Modify: `services/sda-frontend/src/lib/server/gateway.ts`
- Create: `services/sda-frontend/src/lib/server/gateway.test.ts` (agregar tests al existente)

**Step 1: Crear tipos**

Crear `services/sda-frontend/src/lib/types/preferences.ts`:

```typescript
export interface UserPreferences {
    default_collection: string;
    default_query_mode: 'standard' | 'crossdoc';
    vdb_top_k: number;
    reranker_top_k: number;
    max_sub_queries: number;
    follow_up_retries: boolean;
    show_decomposition: boolean;
    avatar_color: string;
    ui_language: 'es' | 'en';
    notify_ingestion_done: boolean;
    notify_system_alerts: boolean;
}

export const DEFAULT_PREFERENCES: UserPreferences = {
    default_collection: '',
    default_query_mode: 'standard',
    vdb_top_k: 10,
    reranker_top_k: 5,
    max_sub_queries: 4,
    follow_up_retries: true,
    show_decomposition: false,
    avatar_color: '#6366f1',
    ui_language: 'es',
    notify_ingestion_done: true,
    notify_system_alerts: true,
};
```

**Step 2: Escribir tests que fallan**

Agregar al final de `services/sda-frontend/src/lib/server/gateway.test.ts`:

```typescript
describe('gatewayGetPreferences', () => {
    it('llama a GET /auth/me/preferences con user_id', async () => {
        vi.resetModules();
        vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
        vi.stubEnv('SYSTEM_API_KEY', 'test-key');
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ vdb_top_k: 15, ui_language: 'en' }),
        }));
        const { gatewayGetPreferences } = await import('./gateway.js');
        const prefs = await gatewayGetPreferences(7);
        const call = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
        expect(call[0]).toContain('/auth/me/preferences');
        expect(call[0]).toContain('user_id=7');
        expect(prefs.vdb_top_k).toBe(15);
        vi.unstubAllEnvs(); vi.unstubAllGlobals();
    });
});

describe('gatewayUpdatePreferences', () => {
    it('llama a PATCH /auth/me/preferences con body parcial', async () => {
        vi.resetModules();
        vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
        vi.stubEnv('SYSTEM_API_KEY', 'test-key');
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({ ok: true }),
        }));
        const { gatewayUpdatePreferences } = await import('./gateway.js');
        await gatewayUpdatePreferences(7, { vdb_top_k: 20 });
        const call = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
        expect(call[1].method).toBe('PATCH');
        expect(JSON.parse(call[1].body).vdb_top_k).toBe(20);
        vi.unstubAllEnvs(); vi.unstubAllGlobals();
    });
});
```

**Step 3: Correr para verificar que fallan**

```bash
cd services/sda-frontend && npm test -- --run --reporter=verbose 2>&1 | grep -A3 "gatewayGetPreferences"
```

Expected: FAIL — `gatewayGetPreferences is not a function`

**Step 4: Implementar en `src/lib/server/gateway.ts`**

Agregar después de `gatewayDeleteSession`:

```typescript
import type { UserPreferences } from '$lib/types/preferences';

export async function gatewayGetPreferences(userId: number): Promise<UserPreferences> {
    return gw<UserPreferences>(`/auth/me/preferences?user_id=${userId}`);
}

export async function gatewayUpdatePreferences(
    userId: number, prefs: Partial<UserPreferences>
): Promise<void> {
    await gw<{ ok: boolean }>(
        `/auth/me/preferences?user_id=${userId}`,
        { method: 'PATCH', body: JSON.stringify(prefs) }
    );
}

export async function gatewayUpdateProfile(userId: number, name: string): Promise<void> {
    await gw<{ ok: boolean }>(
        `/auth/me/profile?user_id=${userId}`,
        { method: 'PATCH', body: JSON.stringify({ name }) }
    );
}

export async function gatewayUpdatePassword(
    userId: number, currentPw: string, newPw: string
): Promise<void> {
    await gw<{ ok: boolean }>(
        `/auth/me/password?user_id=${userId}`,
        { method: 'PATCH', body: JSON.stringify({ current_password: currentPw, new_password: newPw }) }
    );
}
```

**Step 5: Correr tests**

```bash
npm test -- --run 2>&1 | tail -4
```

Expected: 355+ tests PASS

**Step 6: Commit**

```bash
git add src/lib/types/preferences.ts src/lib/server/gateway.ts src/lib/server/gateway.test.ts
git commit -m "feat(fase8): UserPreferences type + 4 gateway client functions"
```

---

### Task 4: BFF — `+layout.server.ts` carga prefs + notificaciones

**Files:**
- Modify: `services/sda-frontend/src/routes/(app)/+layout.server.ts`

**Contexto:** El layout server load ya retorna `user` y `sessions`. Ver el archivo actual antes de modificar.

**Step 1: Leer el archivo actual**

```bash
cat services/sda-frontend/src/routes/\(app\)/+layout.server.ts
```

**Step 2: Implementar**

Agregar al `load()`:

```typescript
import { gatewayGetPreferences, gatewayListSessions, gatewayListAlerts } from '$lib/server/gateway';
import { DEFAULT_PREFERENCES } from '$lib/types/preferences';

export const load: LayoutServerLoad = async ({ locals }) => {
    if (!locals.user) redirect(302, '/login');

    const [sessions, preferences, alerts] = await Promise.all([
        gatewayListSessions(locals.user.id).then(r => r.sessions).catch(() => []),
        gatewayGetPreferences(locals.user.id).catch(() => DEFAULT_PREFERENCES),
        locals.user.role === 'admin'
            ? gatewayListAlerts(false).then(r => r.alerts).catch(() => [])
            : Promise.resolve([]),
    ]);

    // Notificaciones pendientes (in-app toasts)
    const notifications: { type: 'ingestion' | 'alert'; message: string }[] = [];
    if (preferences.notify_system_alerts && alerts.length > 0) {
        notifications.push({ type: 'alert', message: `${alerts.length} alerta${alerts.length > 1 ? 's' : ''} de sistema sin resolver` });
    }

    return { user: locals.user, sessions, preferences, notifications };
};
```

**Nota:** `notify_ingestion_done` para jobs completados se omite en esta fase — requeriría estado de "ya notificado" para no repetir el toast en cada carga. Se puede agregar en una fase futura con un campo `notified_at` en `ingestion_jobs`.

**Step 3: Correr todos los tests frontend**

```bash
npm test -- --run 2>&1 | tail -4
```

Expected: todos pasan

**Step 4: Commit**

```bash
git add src/routes/\(app\)/+layout.server.ts
git commit -m "feat(fase8): layout carga preferences + notificaciones de alertas"
```

---

### Task 5: BFF — `+layout.svelte` init stores + toasts

**Files:**
- Create: `services/sda-frontend/src/lib/stores/preferences.svelte.ts`
- Modify: `services/sda-frontend/src/routes/(app)/+layout.svelte`

**Step 1: Crear el store de preferencias**

Crear `services/sda-frontend/src/lib/stores/preferences.svelte.ts`:

```typescript
import type { UserPreferences } from '$lib/types/preferences';
import { DEFAULT_PREFERENCES } from '$lib/types/preferences';

export class PreferencesStore {
    current = $state<UserPreferences>({ ...DEFAULT_PREFERENCES });

    init(prefs: UserPreferences) {
        this.current = prefs;
    }

    get avatarColor() { return this.current.avatar_color; }
    get language() { return this.current.ui_language; }
}

export const preferencesStore = new PreferencesStore();
```

**Step 2: Modificar `+layout.svelte`**

Agregar en el `<script>` del layout:

```typescript
import { preferencesStore } from '$lib/stores/preferences.svelte';
import { crossdoc } from '$lib/stores/crossdoc.svelte';
import { toast } from '$lib/stores/toast.svelte';

$effect(() => {
    if (data.preferences) {
        preferencesStore.init(data.preferences);
        // Hidratar CrossdocStore con preferencias del usuario
        crossdoc.options.vdbTopK = data.preferences.vdb_top_k;
        crossdoc.options.rerankerTopK = data.preferences.reranker_top_k;
        crossdoc.options.maxSubQueries = data.preferences.max_sub_queries;
        crossdoc.options.followUpRetries = data.preferences.follow_up_retries;
        crossdoc.options.showDecomposition = data.preferences.show_decomposition;
    }
    // Mostrar notificaciones pendientes como toasts
    for (const notif of data.notifications ?? []) {
        toast.add({ type: notif.type === 'alert' ? 'warning' : 'info', message: notif.message });
    }
});
```

**Nota:** Ver cómo está implementado el `toastStore` antes de llamarlo — puede tener una API diferente. Leer `src/lib/stores/toast.svelte.ts` primero.

**Step 3: Correr tests frontend**

```bash
npm test -- --run 2>&1 | tail -4
```

Expected: todos pasan

**Step 4: Commit**

```bash
git add src/lib/stores/preferences.svelte.ts src/routes/\(app\)/+layout.svelte
git commit -m "feat(fase8): PreferencesStore + hidratación de CrossdocStore en layout"
```

---

### Task 6: Settings page — server (load + actions)

**Files:**
- Modify: `services/sda-frontend/src/routes/(app)/settings/+page.server.ts`

**Contexto:** Leer el archivo actual antes de modificar. Actualmente tiene `load()` y la action `refresh_key`.

**Step 1: Leer archivo actual**

```bash
cat services/sda-frontend/src/routes/\(app\)/settings/+page.server.ts
```

**Step 2: Implementar**

Reemplazar con:

```typescript
import type { Actions, PageServerLoad } from './$types';
import {
    gatewayGetMe, gatewayRefreshKey, gatewayListCollections,
    gatewayGetPreferences, gatewayUpdatePreferences,
    gatewayUpdateProfile, gatewayUpdatePassword,
    GatewayError
} from '$lib/server/gateway';
import { fail, redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) redirect(302, '/login');
    const [preferences, collectionsRes] = await Promise.all([
        gatewayGetPreferences(locals.user.id),
        gatewayListCollections(),
    ]);
    return {
        user: locals.user,
        preferences,
        collections: collectionsRes.collections,
    };
};

export const actions: Actions = {
    refresh_key: async ({ locals }) => {
        if (!locals.user) return fail(401);
        const { api_key } = await gatewayRefreshKey();
        return { api_key };
    },

    update_profile: async ({ request, locals }) => {
        if (!locals.user) return fail(401);
        const data = await request.formData();
        const name = (data.get('name') as string ?? '').trim();
        const avatar_color = data.get('avatar_color') as string ?? '#6366f1';
        const ui_language = data.get('ui_language') as string ?? 'es';

        if (!name) return fail(400, { error: 'El nombre no puede estar vacío', field: 'name' });

        await gatewayUpdateProfile(locals.user.id, name);
        await gatewayUpdatePreferences(locals.user.id, { avatar_color, ui_language });
        return { success: true, section: 'profile' };
    },

    update_password: async ({ request, locals }) => {
        if (!locals.user) return fail(401);
        const data = await request.formData();
        const current_password = data.get('current_password') as string;
        const new_password = data.get('new_password') as string;
        const confirm_password = data.get('confirm_password') as string;

        if (new_password !== confirm_password)
            return fail(400, { error: 'Las contraseñas no coinciden', field: 'confirm_password' });
        if (new_password.length < 8)
            return fail(400, { error: 'La contraseña debe tener al menos 8 caracteres', field: 'new_password' });

        try {
            await gatewayUpdatePassword(locals.user.id, current_password, new_password);
            return { success: true, section: 'password' };
        } catch (err) {
            if (err instanceof GatewayError && err.status === 400)
                return fail(400, { error: 'Contraseña actual incorrecta', field: 'current_password' });
            throw err;
        }
    },

    update_preferences: async ({ request, locals }) => {
        if (!locals.user) return fail(401);
        const data = await request.formData();
        await gatewayUpdatePreferences(locals.user.id, {
            default_collection: data.get('default_collection') as string ?? '',
            default_query_mode: (data.get('default_query_mode') as 'standard' | 'crossdoc') ?? 'standard',
            vdb_top_k: Number(data.get('vdb_top_k') ?? 10),
            reranker_top_k: Number(data.get('reranker_top_k') ?? 5),
            max_sub_queries: Number(data.get('max_sub_queries') ?? 4),
            follow_up_retries: data.get('follow_up_retries') === 'on',
            show_decomposition: data.get('show_decomposition') === 'on',
        });
        return { success: true, section: 'preferences' };
    },

    update_notifications: async ({ request, locals }) => {
        if (!locals.user) return fail(401);
        const data = await request.formData();
        await gatewayUpdatePreferences(locals.user.id, {
            notify_ingestion_done: data.get('notify_ingestion_done') === 'on',
            notify_system_alerts: data.get('notify_system_alerts') === 'on',
        });
        return { success: true, section: 'notifications' };
    },
};
```

**Step 3: Correr tests frontend**

```bash
npm test -- --run 2>&1 | tail -4
```

Expected: todos pasan

**Step 4: Commit**

```bash
git add src/routes/\(app\)/settings/+page.server.ts
git commit -m "feat(fase8): settings server — 4 actions (profile, password, preferences, notifications)"
```

---

### Task 7: Settings page — UI completa

**Files:**
- Modify: `services/sda-frontend/src/routes/(app)/settings/+page.svelte`

**Contexto:** Leer el archivo actual COMPLETO antes de modificar. Mantener la sección "API Key + Sesión" sin cambios y respetar el design system existente (`var(--bg-surface)`, `var(--accent)`, etc.).

**Step 1: Leer archivo actual**

```bash
cat services/sda-frontend/src/routes/\(app\)/settings/+page.svelte
```

**Step 2: Reescribir con las 5 secciones**

La nueva página usa `enhance` de SvelteKit para submissions progresivas. Estructura:

```svelte
<script lang="ts">
    import { enhance } from '$app/forms';
    import type { PageData, ActionData } from './$types';

    let { data, form }: { data: PageData; form: ActionData } = $props();

    // Inicial del nombre para el avatar
    let nameInitial = $derived(data.user?.name?.charAt(0)?.toUpperCase() ?? '?');

    // Estado local para password fields
    let currentPw = $state('');
    let newPw = $state('');
    let confirmPw = $state('');
    let pwMatch = $derived(newPw === confirmPw || confirmPw === '');
</script>

<!-- Sección 1: Perfil -->
<section>
  <h2>Perfil</h2>
  <form method="POST" action="?/update_profile" use:enhance>
    <!-- Avatar preview -->
    <div style="background: {data.preferences.avatar_color}"
         class="w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-xl">
      {nameInitial}
    </div>
    <label>Color del avatar
      <input type="color" name="avatar_color" value={data.preferences.avatar_color} />
    </label>
    <label>Nombre
      <input type="text" name="name" value={data.user.name} required />
    </label>
    <p>Email: {data.user.email} · Rol: {data.user.role}</p>
    <label>Idioma
      <select name="ui_language">
        <option value="es" selected={data.preferences.ui_language === 'es'}>Español</option>
        <option value="en" selected={data.preferences.ui_language === 'en'}>English</option>
      </select>
    </label>
    {#if form?.field === 'name'}<p class="text-red-500">{form.error}</p>{/if}
    {#if form?.success && form.section === 'profile'}<p class="text-green-500">Guardado</p>{/if}
    <button type="submit">Guardar perfil</button>
  </form>
</section>

<!-- Sección 2: Contraseña -->
<section>
  <h2>Contraseña</h2>
  <form method="POST" action="?/update_password" use:enhance>
    <input type="password" name="current_password" bind:value={currentPw} placeholder="Contraseña actual" required />
    <input type="password" name="new_password" bind:value={newPw} placeholder="Nueva contraseña" required minlength="8" />
    <input type="password" name="confirm_password" bind:value={confirmPw} placeholder="Confirmar nueva" required />
    {#if !pwMatch}<p class="text-red-500">Las contraseñas no coinciden</p>{/if}
    {#if form?.field === 'current_password'}<p class="text-red-500">{form.error}</p>{/if}
    {#if form?.success && form.section === 'password'}<p class="text-green-500">Contraseña actualizada</p>{/if}
    <button type="submit" disabled={!pwMatch}>Cambiar contraseña</button>
  </form>
</section>

<!-- Sección 3: Preferencias RAG -->
<section>
  <h2>Preferencias RAG</h2>
  <form method="POST" action="?/update_preferences" use:enhance>
    <label>Colección default
      <select name="default_collection">
        <option value="">— Ninguna —</option>
        {#each data.collections as col}
          <option value={col} selected={data.preferences.default_collection === col}>{col}</option>
        {/each}
      </select>
    </label>
    <label>Modo de consulta
      <select name="default_query_mode">
        <option value="standard" selected={data.preferences.default_query_mode === 'standard'}>Estándar</option>
        <option value="crossdoc" selected={data.preferences.default_query_mode === 'crossdoc'}>Cross-doc</option>
      </select>
    </label>
    <label>Top K vectorial
      <input type="number" name="vdb_top_k" value={data.preferences.vdb_top_k} min="1" max="50" />
    </label>
    <label>Top K reranker
      <input type="number" name="reranker_top_k" value={data.preferences.reranker_top_k} min="1" max="20" />
    </label>
    <label>Máx. sub-queries
      <input type="number" name="max_sub_queries" value={data.preferences.max_sub_queries} min="1" max="10" />
    </label>
    <label><input type="checkbox" name="follow_up_retries" checked={data.preferences.follow_up_retries} /> Reintentar sub-queries fallidas</label>
    <label><input type="checkbox" name="show_decomposition" checked={data.preferences.show_decomposition} /> Mostrar descomposición (debug)</label>
    {#if form?.success && form.section === 'preferences'}<p class="text-green-500">Guardado</p>{/if}
    <button type="submit">Guardar preferencias</button>
  </form>
</section>

<!-- Sección 4: Notificaciones -->
<section>
  <h2>Notificaciones</h2>
  <form method="POST" action="?/update_notifications" use:enhance>
    <label><input type="checkbox" name="notify_ingestion_done" checked={data.preferences.notify_ingestion_done} /> Notificarme cuando termine una ingesta</label>
    <label><input type="checkbox" name="notify_system_alerts" checked={data.preferences.notify_system_alerts} /> Notificarme sobre alertas del sistema</label>
    {#if form?.success && form.section === 'notifications'}<p class="text-green-500">Guardado</p>{/if}
    <button type="submit">Guardar</button>
  </form>
</section>

<!-- Sección 5: API Key + Sesión (sin cambios) -->
```

**Nota importante:** La UI debe respetar el design system del proyecto. Antes de reescribir, leer cómo están estilizadas las secciones existentes y usar las mismas clases CSS variables (`var(--bg-surface)`, `var(--border)`, `var(--accent)`, etc.).

**Step 3: Correr tests frontend**

```bash
npm test -- --run 2>&1 | tail -4
```

Expected: todos pasan

**Step 4: Commit**

```bash
git add src/routes/\(app\)/settings/+page.svelte
git commit -m "feat(fase8): settings UI — perfil con avatar, contraseña, RAG prefs, notificaciones"
```

---

### Task 8: Tests BFF — settings server actions

**Files:**
- Create: `services/sda-frontend/src/routes/(app)/settings/settings.test.ts`

**Step 1: Crear tests**

```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('settings server actions', () => {
    beforeEach(() => {
        vi.resetModules();
        vi.stubEnv('GATEWAY_URL', 'http://gateway:9000');
        vi.stubEnv('SYSTEM_API_KEY', 'test-key');
    });
    afterEach(() => {
        vi.unstubAllEnvs();
        vi.unstubAllGlobals();
    });

    const mockUser = { id: 7, name: 'Test', email: 't@t.com', role: 'user', area_id: 1 };

    it('update_profile retorna success', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true, json: async () => ({ ok: true })
        }));
        const { actions } = await import('./+page.server.js');
        const fd = new FormData();
        fd.set('name', 'Nuevo'); fd.set('avatar_color', '#ff0000'); fd.set('ui_language', 'es');
        const result = await actions.update_profile({
            request: { formData: async () => fd },
            locals: { user: mockUser },
        } as any);
        expect(result).toMatchObject({ success: true, section: 'profile' });
    });

    it('update_password retorna error si contraseñas no coinciden', async () => {
        const { actions } = await import('./+page.server.js');
        const fd = new FormData();
        fd.set('current_password', 'vieja'); fd.set('new_password', 'nueva1234');
        fd.set('confirm_password', 'diferente');
        const result = await actions.update_password({
            request: { formData: async () => fd },
            locals: { user: mockUser },
        } as any);
        expect(result).toMatchObject({ status: 400, data: { field: 'confirm_password' } });
    });

    it('update_preferences guarda prefs correctamente', async () => {
        vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
            ok: true, json: async () => ({ ok: true })
        }));
        const { actions } = await import('./+page.server.js');
        const fd = new FormData();
        fd.set('vdb_top_k', '15'); fd.set('default_query_mode', 'crossdoc');
        fd.set('follow_up_retries', 'on');
        const result = await actions.update_preferences({
            request: { formData: async () => fd },
            locals: { user: mockUser },
        } as any);
        expect(result).toMatchObject({ success: true, section: 'preferences' });
    });
});
```

**Step 2: Correr tests**

```bash
npm test -- --run 2>&1 | tail -4
```

Expected: 356+ tests PASS

**Step 3: Commit**

```bash
git add src/routes/\(app\)/settings/settings.test.ts
git commit -m "test(fase8): BFF settings server actions"
```

---

### Task 9: GATE 4 — Simplify

```bash
# Desde la raíz del worktree
cd /Users/enzo/rag-saldivia/.worktrees/fase8-settings-pro
uv run pytest saldivia/tests/ -q --tb=short
cd services/sda-frontend && npm test -- --run
```

Invocar skill `/simplify` para revisar todos los cambios.

---

### Task 10: GATE 5 — Review

Invocar subagentes:
- `gateway-reviewer` para `saldivia/gateway.py` + `saldivia/auth/database.py`
- `frontend-reviewer` para los componentes Svelte y rutas BFF

---

### Task 11: GATE 6 — Docs + PR

```bash
# Doc writer
# Luego:
cd /Users/enzo/rag-saldivia/.worktrees/fase8-settings-pro
git push -u origin fase8/settings-pro
gh pr create --title "feat(fase8): Settings Pro — perfil, prefs RAG, contraseña, notificaciones"
```
