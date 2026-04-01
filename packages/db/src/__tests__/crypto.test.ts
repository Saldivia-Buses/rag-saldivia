import { describe, test, expect, beforeAll, afterAll } from "bun:test"
import { encryptSecret, decryptSecret } from "../crypto"

// Use a test key (32 bytes base64)
const TEST_KEY = Buffer.from("a".repeat(32)).toString("base64")

describe("crypto", () => {
  let originalKey: string | undefined

  beforeAll(() => {
    originalKey = process.env["SYSTEM_API_KEY"]
    process.env["SYSTEM_API_KEY"] = TEST_KEY
  })

  afterAll(() => {
    if (originalKey !== undefined) {
      process.env["SYSTEM_API_KEY"] = originalKey
    } else {
      delete process.env["SYSTEM_API_KEY"]
    }
  })

  test("encrypt/decrypt roundtrip", () => {
    const plaintext = '{"client_secret":"super-secret-123"}'
    const encrypted = encryptSecret(plaintext)
    expect(encrypted).not.toBe(plaintext)
    const decrypted = decryptSecret(encrypted)
    expect(decrypted).toBe(plaintext)
  })

  test("different plaintexts produce different ciphertexts", () => {
    const a = encryptSecret("secret-a")
    const b = encryptSecret("secret-b")
    expect(a).not.toBe(b)
  })

  test("same plaintext produces different ciphertexts (random IV)", () => {
    const a = encryptSecret("same-value")
    const b = encryptSecret("same-value")
    expect(a).not.toBe(b)
    // But both decrypt to the same value
    expect(decryptSecret(a)).toBe("same-value")
    expect(decryptSecret(b)).toBe("same-value")
  })

  test("tampered ciphertext returns stored value (graceful fallback)", () => {
    const encrypted = encryptSecret("test-secret")
    // Flip a byte in the middle
    const buf = Buffer.from(encrypted, "base64")
    buf[20] = buf[20]! ^ 0xff
    const tampered = buf.toString("base64")
    // Should not throw, returns the tampered base64 string as-is
    const result = decryptSecret(tampered)
    expect(result).toBe(tampered)
  })

  test("plaintext input returns as-is (lazy migration)", () => {
    const plaintext = "not-encrypted-value"
    const result = decryptSecret(plaintext)
    expect(result).toBe(plaintext)
  })

  test("short strings roundtrip", () => {
    const encrypted = encryptSecret("x")
    const decrypted = decryptSecret(encrypted)
    expect(decrypted).toBe("x")
  })

  test("unicode content roundtrips", () => {
    const plaintext = '{"name":"Señor González","key":"contraseña-única"}'
    const encrypted = encryptSecret(plaintext)
    const decrypted = decryptSecret(encrypted)
    expect(decrypted).toBe(plaintext)
  })
})

describe("crypto without key", () => {
  let originalKey: string | undefined

  beforeAll(() => {
    originalKey = process.env["SYSTEM_API_KEY"]
    delete process.env["SYSTEM_API_KEY"]
  })

  afterAll(() => {
    if (originalKey !== undefined) {
      process.env["SYSTEM_API_KEY"] = originalKey
    }
  })

  test("encrypt returns plaintext when no key", () => {
    const plaintext = "no-encryption"
    expect(encryptSecret(plaintext)).toBe(plaintext)
  })

  test("decrypt returns stored value when no key", () => {
    const stored = "some-stored-value"
    expect(decryptSecret(stored)).toBe(stored)
  })
})
