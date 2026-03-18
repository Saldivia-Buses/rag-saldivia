import { describe, it, expect } from 'vitest';

describe('auth', () => {
    it('returns null for missing cookie', async () => {
        // Set a dummy JWT_SECRET to avoid empty secret error
        process.env.JWT_SECRET = 'test-secret-for-testing';
        // Dynamic import to pick up env
        const { verifySession } = await import('./auth.js');
        const mockCookies = { get: (_name: string) => undefined } as any;
        const result = await verifySession(mockCookies);
        expect(result).toBeNull();
    });
});
