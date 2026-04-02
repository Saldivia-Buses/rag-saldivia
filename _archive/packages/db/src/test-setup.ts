/**
 * Preload para tests de @rag-saldivia/db.
 * Mockea ioredis con ioredis-mock para que los tests unitarios
 * no requieran Redis corriendo.
 */

import { mock } from "bun:test"
import IORedisMock from "ioredis-mock"

// Asegurar que REDIS_URL esté configurado (para que getRedisClient no lance)
process.env["REDIS_URL"] ??= "redis://localhost:6379"

// Reemplazar ioredis con el mock en el entorno de tests
mock.module("ioredis", () => ({ default: IORedisMock }))
