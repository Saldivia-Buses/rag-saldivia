/**
 * crypto.ts — Shared AES-256-GCM encryption for secrets stored in DB.
 *
 * Used by: external-sources.ts (connector credentials), sso.ts (client secrets)
 * Key source: SYSTEM_API_KEY env var (base64-encoded 32-byte key)
 */

import { createCipheriv, createDecipheriv, randomBytes } from "crypto"

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

/**
 * Encrypt plaintext with AES-256-GCM using SYSTEM_API_KEY.
 * Returns base64(iv + tag + ciphertext). Falls back to plaintext if no key (dev mode).
 */
export function encryptSecret(plaintext: string): string {
  const key = getEncryptionKey()
  if (!key) return plaintext
  const iv = randomBytes(IV_LEN)
  const cipher = createCipheriv(ALGO, key, iv)
  const encrypted = Buffer.concat([cipher.update(plaintext, "utf8"), cipher.final()])
  const tag = cipher.getAuthTag()
  return Buffer.concat([iv, tag, encrypted]).toString("base64")
}

/**
 * Decrypt a secret encrypted with encryptSecret().
 * Supports lazy migration: if the stored value is plaintext (not valid ciphertext),
 * it returns the value as-is.
 */
export function decryptSecret(stored: string): string {
  const key = getEncryptionKey()
  if (!key) return stored
  const buf = Buffer.from(stored, "base64")
  if (buf.length < IV_LEN + TAG_LEN + 1) return stored
  const iv = buf.subarray(0, IV_LEN)
  const tag = buf.subarray(IV_LEN, IV_LEN + TAG_LEN)
  const ciphertext = buf.subarray(IV_LEN + TAG_LEN)
  try {
    const decipher = createDecipheriv(ALGO, key, iv)
    decipher.setAuthTag(tag)
    return Buffer.concat([decipher.update(ciphertext), decipher.final()]).toString("utf8")
  } catch {
    return stored
  }
}
