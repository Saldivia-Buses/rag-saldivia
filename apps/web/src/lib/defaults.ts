/**
 * App-wide defaults — temporary until proper user-level config (Plan 24).
 * Uses NEXT_PUBLIC_ prefix so these are available in both server and client code.
 */

export const DEFAULT_COLLECTION =
  process.env["NEXT_PUBLIC_DEFAULT_COLLECTION"] ?? "default"
