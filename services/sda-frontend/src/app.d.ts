// src/app.d.ts
import type { SessionUser } from '$lib/server/gateway';
declare global {
    namespace App {
        interface Locals {
            user: SessionUser | null;
        }
    }
}
export {};
