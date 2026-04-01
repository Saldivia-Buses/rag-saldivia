import { eq, and, lte } from "drizzle-orm"
import { getDb } from "../connection"
import { externalSources } from "../schema"
import type { NewExternalSource } from "../schema"
import { randomUUID, createCipheriv, createDecipheriv, randomBytes } from "crypto"

// --- Credential encryption helpers ---

const ALGO = "aes-256-gcm" as const
const IV_LEN = 12
const TAG_LEN = 16

function getEncryptionKey(): Buffer | null {
  const raw = process.env["SYSTEM_API_KEY"]
  if (!raw) return null
  const buf = Buffer.from(raw, "base64")
  if (buf.length !== 32) return null
  return buf
}

export function encryptCredentials(plaintext: string): string {
  const key = getEncryptionKey()
  if (!key) return plaintext // no key = store as-is (dev mode)
  const iv = randomBytes(IV_LEN)
  const cipher = createCipheriv(ALGO, key, iv)
  const encrypted = Buffer.concat([cipher.update(plaintext, "utf8"), cipher.final()])
  const tag = cipher.getAuthTag()
  // Format: base64(iv + tag + ciphertext)
  return Buffer.concat([iv, tag, encrypted]).toString("base64")
}

export function decryptCredentials(stored: string): string {
  const key = getEncryptionKey()
  if (!key) return stored
  // Try to decode as base64 — if it fails, it's plaintext (lazy migration)
  let buf: Buffer
  try {
    buf = Buffer.from(stored, "base64")
  } catch {
    return stored
  }
  // Minimum length: IV(12) + TAG(16) + at least 1 byte
  if (buf.length < IV_LEN + TAG_LEN + 1) return stored
  const iv = buf.subarray(0, IV_LEN)
  const tag = buf.subarray(IV_LEN, IV_LEN + TAG_LEN)
  const ciphertext = buf.subarray(IV_LEN + TAG_LEN)
  try {
    const decipher = createDecipheriv(ALGO, key, iv)
    decipher.setAuthTag(tag)
    return Buffer.concat([decipher.update(ciphertext), decipher.final()]).toString("utf8")
  } catch {
    // Decryption failed — likely plaintext stored before encryption was enabled
    return stored
  }
}

// --- Query functions ---

export async function createExternalSource(data: Omit<NewExternalSource, "id" | "createdAt" | "lastSync" | "active">) {
  const db = getDb()
  const values = {
    id: randomUUID(),
    ...data,
    credentials: data.credentials ? encryptCredentials(data.credentials) : "{}",
    active: true,
    createdAt: Date.now(),
  }
  const [row] = await db.insert(externalSources).values(values).returning()
  return { ...row!, credentials: data.credentials ?? "{}" }
}

export async function listExternalSources(userId: number) {
  const db = getDb()
  const rows = await db.select().from(externalSources).where(eq(externalSources.userId, userId))
  return rows.map((r) => ({ ...r, credentials: decryptCredentials(r.credentials) }))
}

export async function listActiveSourcesToSync() {
  const db = getDb()
  const now = Date.now()
  const allActive = await db.select().from(externalSources).where(eq(externalSources.active, true))
  return allActive
    .filter((s) => {
      const lastSync = s.lastSync ?? 0
      const intervalMs = s.schedule === "hourly" ? 3600_000 : s.schedule === "weekly" ? 7 * 86400_000 : 86400_000
      return now - lastSync >= intervalMs
    })
    .map((r) => ({ ...r, credentials: decryptCredentials(r.credentials) }))
}

export async function updateSourceLastSync(id: string) {
  const db = getDb()
  await db.update(externalSources).set({ lastSync: Date.now() }).where(eq(externalSources.id, id))
}

export async function deleteExternalSource(id: string, userId: number) {
  const db = getDb()
  await db.delete(externalSources).where(and(eq(externalSources.id, id), eq(externalSources.userId, userId)))
}
