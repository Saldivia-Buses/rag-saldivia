#!/usr/bin/env bun
/**
 * Seed de datos de desarrollo.
 * Crea un usuario admin por defecto y un área de ejemplo.
 *
 * Uso: bun run db:seed
 *      bun packages/db/src/seed.ts
 *
 * IMPORTANTE: Solo para desarrollo. No correr en producción.
 */

import { getDb } from "./connection"
import {
  areas, users, userAreas, areaCollections, promptTemplates,
  roles, permissions, rolePermissions, userRoleAssignments,
} from "./schema"
import { hashSync } from "bcrypt-ts"
import { eq, sql } from "drizzle-orm"

const db = getDb()

async function seed() {
  console.warn("Seeding base de datos de desarrollo...")

  const now = Date.now()

  // Schema migrations
  await db.run(sql`ALTER TABLE users ADD COLUMN last_seen INTEGER`).catch(() => {})

  // ── Área por defecto ───────────────────────────────────────────────────
  const [defaultArea] = await db
    .insert(areas)
    .values({
      name: "General",
      description: "Área general de acceso",
      createdAt: now,
    })
    .onConflictDoNothing()
    .returning()

  const areaId = defaultArea?.id

  if (!areaId) {
    console.warn("  El área 'General' ya existe — skipping")
  } else {
    console.warn(`  Área creada: General (id=${areaId})`)
  }

  // Obtener el área existente si ya existía
  const existingArea = areaId
    ? { id: areaId }
    : await db.query.areas.findFirst({ where: (a, { eq }) => eq(a.name, "General") })

  if (!existingArea) throw new Error("No se pudo crear o encontrar el área General")

  // ── Colección de ejemplo ───────────────────────────────────────────────
  await db
    .insert(areaCollections)
    .values({
      areaId: existingArea.id,
      collectionName: "tecpia",
      permission: "write",
    })
    .onConflictDoNothing()

  console.warn("  Colección 'tecpia' asignada al área General")

  // ── Usuario admin ──────────────────────────────────────────────────────
  const adminPasswordHash = hashSync("changeme", 10)
  const adminApiKeyHash = Bun.hash("rsk_dev_admin_api_key_changeme").toString()

  const [adminUser] = await db
    .insert(users)
    .values({
      email: "admin@localhost",
      name: "Admin (dev)",
      role: "admin",
      apiKeyHash: adminApiKeyHash,
      passwordHash: adminPasswordHash,
      preferences: {},
      active: true,
      createdAt: now,
    })
    .onConflictDoNothing()
    .returning()

  if (!adminUser) {
    console.warn("  Usuario admin@localhost ya existe — skipping")
  } else {
    console.warn(`  Usuario admin creado: admin@localhost (id=${adminUser.id})`)
    console.warn("    Email:    admin@localhost")
    console.warn("    Password: changeme")
    console.warn("    IMPORTANTE: cambiar antes de usar en producción")

    // Asignar admin al área general
    await db
      .insert(userAreas)
      .values({ userId: adminUser.id, areaId: existingArea.id })
      .onConflictDoNothing()
  }

  // ── Usuario de prueba ──────────────────────────────────────────────────
  const userPasswordHash = hashSync("test1234", 10)
  const userApiKeyHash = Bun.hash("rsk_dev_user_api_key_changeme").toString()

  const [testUser] = await db
    .insert(users)
    .values({
      email: "user@localhost",
      name: "Usuario de prueba",
      role: "user",
      apiKeyHash: userApiKeyHash,
      passwordHash: userPasswordHash,
      preferences: {},
      active: true,
      createdAt: now,
    })
    .onConflictDoNothing()
    .returning()

  if (!testUser) {
    console.warn("  Usuario user@localhost ya existe — skipping")
  } else {
    console.warn(`  Usuario de prueba creado: user@localhost (id=${testUser.id})`)

    await db
      .insert(userAreas)
      .values({ userId: testUser.id, areaId: existingArea.id })
      .onConflictDoNothing()
  }

  // ── Prompt Templates por defecto ─────────────────────────────────────────
  const adminId = adminUser?.id ?? (await db.query.users.findFirst({ where: (u, { eq }) => eq(u.email, "admin@localhost") }))?.id ?? 1

  const defaultTemplates = [
    { title: "Buscar documentos", prompt: "Buscá información sobre ", focusMode: "detallado" },
    { title: "Resumir contenido", prompt: "Hacé un resumen ejecutivo de ", focusMode: "ejecutivo" },
    { title: "Analizar datos", prompt: "Analizá los datos sobre ", focusMode: "detallado" },
    { title: "Comparar alternativas", prompt: "Compará las diferencias entre ", focusMode: "comparativo" },
    { title: "Explicación técnica", prompt: "Explicá técnicamente cómo funciona ", focusMode: "tecnico" },
    { title: "Preguntas frecuentes", prompt: "¿Cuáles son las preguntas más frecuentes sobre ", focusMode: "detallado" },
  ]

  for (const t of defaultTemplates) {
    await db
      .insert(promptTemplates)
      .values({ ...t, createdBy: adminId, active: true, createdAt: now })
      .onConflictDoNothing()
  }

  console.warn(`  ${defaultTemplates.length} prompt templates creados`)

  // ── RBAC: Crear tablas ────────────────────────────────────────────────
  await db.run(sql`CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    level INTEGER NOT NULL DEFAULT 0,
    color TEXT NOT NULL DEFAULT '#6e6c69',
    icon TEXT NOT NULL DEFAULT 'user',
    is_system INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL
  )`)

  await db.run(sql`CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT ''
  )`)

  await db.run(sql`CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
  )`)

  await db.run(sql`CREATE TABLE IF NOT EXISTS user_role_assignments (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id)
  )`)


  console.warn("  Tablas RBAC creadas")

  // ── RBAC: Seed roles por defecto ────────────────────────────────────────
  const defaultRoles = [
    { name: "Admin", description: "Acceso total al sistema", level: 100, color: "#2563eb", icon: "shield", isSystem: true },
    { name: "Manager", description: "Gestión de colecciones y usuarios", level: 50, color: "#d97706", icon: "user-cog", isSystem: true },
    { name: "Usuario", description: "Acceso básico al chat", level: 10, color: "#6e6c69", icon: "user", isSystem: true },
  ]

  for (const r of defaultRoles) {
    await db
      .insert(roles)
      .values({ ...r, createdAt: now })
      .onConflictDoNothing()
  }

  console.warn(`  ${defaultRoles.length} roles por defecto creados`)

  // ── RBAC: Seed permisos ─────────────────────────────────────────────────
  const defaultPermissions = [
    // Administración
    { key: "admin.access", label: "Acceder al panel admin", category: "Administración", description: "Ver el panel de administración" },
    { key: "admin.dashboard", label: "Ver dashboard", category: "Administración", description: "Ver estadísticas y métricas del sistema" },
    { key: "admin.audit_log", label: "Ver auditoría", category: "Administración", description: "Ver el registro de auditoría del sistema" },
    // Usuarios
    { key: "users.view", label: "Ver usuarios", category: "Usuarios", description: "Ver la lista de usuarios" },
    { key: "users.manage", label: "Gestionar usuarios", category: "Usuarios", description: "Crear, editar y eliminar usuarios" },
    { key: "users.reset_password", label: "Resetear contraseñas", category: "Usuarios", description: "Resetear la contraseña de otros usuarios" },
    { key: "users.view_online", label: "Ver usuarios online", category: "Usuarios", description: "Ver el estado de conexión de los usuarios" },
    // Roles
    { key: "roles.manage", label: "Gestionar roles", category: "Roles", description: "Crear y editar roles y permisos" },
    { key: "roles.view", label: "Ver roles", category: "Roles", description: "Ver los roles del sistema" },
    // Chat
    { key: "chat.use", label: "Usar el chat", category: "Chat", description: "Enviar mensajes y usar el RAG" },
    { key: "chat.manage_all", label: "Gestionar todos los chats", category: "Chat", description: "Ver y eliminar chats de otros usuarios" },
    { key: "chat.export", label: "Exportar chats", category: "Chat", description: "Exportar conversaciones a PDF o texto" },
    { key: "chat.share", label: "Compartir chats", category: "Chat", description: "Crear enlaces para compartir conversaciones" },
    // Colecciones
    { key: "collections.read", label: "Ver colecciones", category: "Colecciones", description: "Ver la lista de colecciones" },
    { key: "collections.manage", label: "Gestionar colecciones", category: "Colecciones", description: "Crear y eliminar colecciones" },
    // Documentos
    { key: "upload.use", label: "Subir documentos", category: "Documentos", description: "Subir archivos al RAG" },
    { key: "upload.bulk", label: "Carga masiva", category: "Documentos", description: "Subir múltiples archivos a la vez" },
    { key: "upload.delete", label: "Eliminar documentos", category: "Documentos", description: "Eliminar documentos del RAG" },
    // Templates
    { key: "templates.view", label: "Ver templates", category: "Templates", description: "Ver templates de prompt disponibles" },
    { key: "templates.manage", label: "Gestionar templates", category: "Templates", description: "Crear y eliminar templates de prompt" },
    // Sistema
    { key: "settings.view", label: "Ver configuración", category: "Sistema", description: "Ver la configuración del sistema" },
    { key: "settings.manage", label: "Modificar configuración", category: "Sistema", description: "Cambiar configuración global del sistema" },
    { key: "settings.webhooks", label: "Gestionar webhooks", category: "Sistema", description: "Configurar integraciones via webhooks" },
    { key: "settings.api_keys", label: "Gestionar API keys", category: "Sistema", description: "Crear y revocar API keys del sistema" },
  ]

  for (const p of defaultPermissions) {
    await db
      .insert(permissions)
      .values(p)
      .onConflictDoNothing()
  }

  console.warn(`  ${defaultPermissions.length} permisos creados`)

  // ── RBAC: Role-permission matrix ────────────────────────────────────────
  const allRoles = await db.select().from(roles)
  const allPerms = await db.select().from(permissions)

  const roleMap = Object.fromEntries(allRoles.map((r) => [r.name, r.id]))
  const permMap = Object.fromEntries(allPerms.map((p) => [p.key, p.id]))

  // Admin: all permissions
  // Manager: broad access minus admin-only features
  // Usuario: basic chat and read access
  const matrix: Record<string, string[]> = {
    Admin: defaultPermissions.map((p) => p.key),
    Manager: [
      "admin.access", "admin.dashboard", "users.view", "users.view_online", "roles.view",
      "chat.use", "chat.manage_all", "chat.export", "chat.share",
      "collections.read", "collections.manage",
      "upload.use", "upload.bulk",
      "templates.view", "templates.manage",
      "settings.view",
    ],
    Usuario: [
      "chat.use", "chat.share",
      "collections.read",
      "templates.view",
      "settings.view",
    ],
  }

  for (const [roleName, permKeys] of Object.entries(matrix)) {
    const rId = roleMap[roleName]
    if (!rId) continue
    for (const key of permKeys) {
      const pId = permMap[key]
      if (!pId) continue
      await db
        .insert(rolePermissions)
        .values({ roleId: rId, permissionId: pId })
        .onConflictDoNothing()
    }
  }

  console.warn("  Matrix role-permissions poblada")

  // ── RBAC: Migrar usuarios existentes ────────────────────────────────────
  const allUsers = await db.select({ id: users.id, role: users.role }).from(users)
  const legacyRoleMap: Record<string, string> = {
    admin: "Admin",
    area_manager: "Manager",
    user: "Usuario",
  }

  for (const u of allUsers) {
    const rbacRoleName = legacyRoleMap[u.role] ?? "Usuario"
    const rId = roleMap[rbacRoleName]
    if (!rId) continue
    await db
      .insert(userRoleAssignments)
      .values({ userId: u.id, roleId: rId, assignedAt: now })
      .onConflictDoNothing()
  }

  console.warn(`  ${allUsers.length} usuarios migrados al RBAC`)

  console.warn("\nSeed completado.")
  console.warn("Credenciales de desarrollo:")
  console.warn("  admin@localhost / changeme  (admin)")
  console.warn("  user@localhost / test1234   (user)")
}

seed().catch((err) => {
  console.error("Seed falló:", err)
  process.exit(1)
})
