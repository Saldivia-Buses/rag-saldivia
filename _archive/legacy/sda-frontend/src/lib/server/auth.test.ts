import { describe, it, expect, beforeEach, vi } from 'vitest';
import { SignJWT } from 'jose';

describe('auth', () => {
    beforeEach(() => {
        process.env.JWT_SECRET = 'test-secret-for-testing-purposes-only';
    });

    it('returns null for missing cookie', async () => {
        const { verifySession } = await import('./auth.js');
        const mockCookies = { get: (_name: string) => undefined } as any;
        const result = await verifySession(mockCookies);
        expect(result).toBeNull();
    });

    it('returns null for expired JWT token', async () => {
        const { verifySession } = await import('./auth.js');
        const secret = new TextEncoder().encode(process.env.JWT_SECRET);

        // Create expired token (exp in the past)
        const expiredToken = await new SignJWT({
            user_id: 1,
            email: 'test@example.com',
            name: 'Test User',
            role: 'user',
            area_id: 1,
        })
            .setProtectedHeader({ alg: 'HS256' })
            .setExpirationTime(0) // Unix epoch = clearly expired
            .sign(secret);

        const mockCookies = { get: (_name: string) => expiredToken } as any;
        const result = await verifySession(mockCookies);
        expect(result).toBeNull();
    });

    it('returns null for non-JWT string in cookie', async () => {
        const { verifySession } = await import('./auth.js');
        const mockCookies = { get: (_name: string) => 'not-a-jwt-token' } as any;
        const result = await verifySession(mockCookies);
        expect(result).toBeNull();
    });

    it('returns null for malformed JWT', async () => {
        const { verifySession } = await import('./auth.js');
        const mockCookies = { get: (_name: string) => 'header.payload.invalid' } as any;
        const result = await verifySession(mockCookies);
        expect(result).toBeNull();
    });

    it('returns user data for valid JWT token', async () => {
        const { verifySession } = await import('./auth.js');
        const secret = new TextEncoder().encode(process.env.JWT_SECRET);

        // Create valid token
        const validToken = await new SignJWT({
            user_id: 42,
            email: 'valid@example.com',
            name: 'Valid User',
            role: 'admin',
            area_id: 5,
        })
            .setProtectedHeader({ alg: 'HS256' })
            .setExpirationTime('2h')
            .sign(secret);

        const mockCookies = { get: (_name: string) => validToken } as any;
        const result = await verifySession(mockCookies);

        expect(result).not.toBeNull();
        expect(result?.id).toBe(42);
        expect(result?.email).toBe('valid@example.com');
        expect(result?.name).toBe('Valid User');
        expect(result?.role).toBe('admin');
        expect(result?.area_id).toBe(5);
    });

    it('setSessionCookie configures cookie with correct attributes', async () => {
        const { setSessionCookie } = await import('./auth.js');
        const mockSet = vi.fn();
        const mockCookies = { set: mockSet } as any;

        setSessionCookie(mockCookies, 'test-token-value');

        expect(mockSet).toHaveBeenCalledWith('sda_session', 'test-token-value', {
            path: '/',
            httpOnly: true,
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'strict',
            maxAge: 60 * 60 * 8, // 8 hours
        });
    });

    it('clearSessionCookie removes cookie', async () => {
        const { clearSessionCookie } = await import('./auth.js');
        const mockDelete = vi.fn();
        const mockCookies = { delete: mockDelete } as any;

        clearSessionCookie(mockCookies);

        expect(mockDelete).toHaveBeenCalledWith('sda_session', { path: '/' });
    });

    it('handles JWT missing required claim gracefully', async () => {
        const { verifySession } = await import('./auth.js');
        const secret = new TextEncoder().encode(process.env.JWT_SECRET);

        // Token missing 'email' claim
        const incompleteToken = await new SignJWT({
            user_id: 1,
            // email missing
            name: 'Test User',
            role: 'user',
        })
            .setProtectedHeader({ alg: 'HS256' })
            .setExpirationTime('2h')
            .sign(secret);

        const mockCookies = { get: (_name: string) => incompleteToken } as any;
        const result = await verifySession(mockCookies);

        // Should still return user object, but with undefined email
        expect(result).not.toBeNull();
        expect(result?.email).toBeUndefined();
    });

    it('retorna null cuando JWT_SECRET no está configurado', async () => {
        const savedSecret = process.env.JWT_SECRET;
        delete process.env.JWT_SECRET;

        const { verifySession } = await import('./auth.js');
        // Create a token with a known secret (not matching empty env)
        const secret = new TextEncoder().encode('some-secret');
        const token = await new SignJWT({ user_id: 1, name: 'Test', role: 'user', area_id: 1 })
            .setProtectedHeader({ alg: 'HS256' })
            .setExpirationTime('2h')
            .sign(secret);

        const mockCookies = { get: (_name: string) => token } as any;
        const result = await verifySession(mockCookies);

        // getJwtSecret() throws when JWT_SECRET missing → caught → null
        expect(result).toBeNull();

        process.env.JWT_SECRET = savedSecret;
    });

    it('name usa string vacío cuando el claim falta en el JWT', async () => {
        const { verifySession } = await import('./auth.js');
        const secret = new TextEncoder().encode(process.env.JWT_SECRET);

        // Token without 'name' claim
        const tokenSinName = await new SignJWT({
            user_id: 99,
            email: 'noname@example.com',
            role: 'user',
            area_id: 3,
            // name is intentionally missing
        })
            .setProtectedHeader({ alg: 'HS256' })
            .setExpirationTime('2h')
            .sign(secret);

        const mockCookies = { get: (_name: string) => tokenSinName } as any;
        const result = await verifySession(mockCookies);

        expect(result).not.toBeNull();
        expect(result?.name).toBe('');  // ?? '' fallback
    });
});
