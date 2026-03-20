// src/lib/server/gateway.test.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
    GatewayError,
    gatewayLogin,
    gatewayGetMe,
    gatewayRefreshKey,
    gatewayListCollections,
    gatewayCollectionStats,
    gatewayCreateCollection,
    gatewayDeleteCollection,
    gatewayListSessions,
    gatewayCreateSession,
    gatewayGetSession,
    gatewayDeleteSession,
    gatewayListUsers,
    gatewayCreateUser,
    gatewayUpdateUser,
    gatewayDeleteUser,
    gatewayListAreas,
    gatewayGetAreaCollections,
    gatewayGrantCollection,
    gatewayRevokeCollection,
    gatewayGetAudit,
    gatewayGenerateText,
    gatewayGenerateStream,
} from './gateway';

// Mock environment variables
vi.stubEnv('GATEWAY_URL', 'http://localhost:9000');
vi.stubEnv('SYSTEM_API_KEY', 'test-system-key');

const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

describe('gateway.ts', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    describe('GatewayError', () => {
        it('should create error with status and detail', () => {
            const err = new GatewayError(404, 'Not found');
            expect(err.name).toBe('GatewayError');
            expect(err.status).toBe(404);
            expect(err.detail).toBe('Not found');
            expect(err.message).toBe('Gateway error (404): Not found');
        });
    });

    describe('Auth functions', () => {
        it('gatewayLogin should POST credentials and return token + user', async () => {
            const mockResponse = {
                token: 'jwt-token-123',
                user: { id: 1, email: 'test@example.com', name: 'Test User', role: 'user', area_id: 1 },
            };
            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            });

            const result = await gatewayLogin('test@example.com', 'password123');

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/auth/session',
                expect.objectContaining({
                    method: 'POST',
                    headers: expect.objectContaining({
                        Authorization: 'Bearer test-system-key',
                        'Content-Type': 'application/json',
                    }),
                    body: JSON.stringify({ email: 'test@example.com', password: 'password123' }),
                })
            );
            expect(result).toEqual(mockResponse);
        });

        it('gatewayGetMe should fetch user by ID', async () => {
            const mockUser = { id: 42, email: 'alice@example.com', name: 'Alice', role: 'admin', area_id: 2 };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockUser });

            const result = await gatewayGetMe(42);

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/auth/me?user_id=42',
                expect.objectContaining({
                    headers: expect.objectContaining({
                        Authorization: 'Bearer test-system-key',
                    }),
                })
            );
            expect(result).toEqual(mockUser);
        });

        it('gatewayRefreshKey should POST and return new API key', async () => {
            const mockResponse = { api_key: 'new-api-key-xyz' };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockResponse });

            const result = await gatewayRefreshKey();

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/auth/refresh-key',
                expect.objectContaining({ method: 'POST' })
            );
            expect(result).toEqual(mockResponse);
        });
    });

    describe('Collections functions', () => {
        it('gatewayListCollections should fetch collection list', async () => {
            const mockResponse = { collections: ['coll-a', 'coll-b'] };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockResponse });

            const result = await gatewayListCollections();

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/v1/collections', expect.any(Object));
            expect(result).toEqual(mockResponse);
        });

        it('gatewayCollectionStats should fetch stats for a collection', async () => {
            const mockStats = { collection: 'test-coll', entity_count: 100, document_count: 10 };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockStats });

            const result = await gatewayCollectionStats('test-coll');

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/v1/collections/test-coll/stats', expect.any(Object));
            expect(result).toEqual(mockStats);
        });

        it('gatewayCreateCollection should POST new collection', async () => {
            const mockResponse = { name: 'new-coll' };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockResponse });

            const result = await gatewayCreateCollection('new-coll', 'custom-schema');

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/v1/collections',
                expect.objectContaining({
                    method: 'POST',
                    body: JSON.stringify({ name: 'new-coll', schema: 'custom-schema' }),
                })
            );
            expect(result).toEqual(mockResponse);
        });

        it('gatewayCreateCollection should use default schema when not provided', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ name: 'default-coll' }) });

            await gatewayCreateCollection('default-coll');

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/v1/collections',
                expect.objectContaining({
                    body: JSON.stringify({ name: 'default-coll', schema: 'default' }),
                })
            );
        });

        it('gatewayDeleteCollection should DELETE collection', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true }) });

            const result = await gatewayDeleteCollection('old-coll');

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/v1/collections/old-coll',
                expect.objectContaining({ method: 'DELETE' })
            );
            expect(result).toEqual({ ok: true });
        });
    });

    describe('Chat sessions functions', () => {
        it('gatewayListSessions should fetch sessions for user', async () => {
            const mockSessions = {
                sessions: [
                    { id: 's1', title: 'Session 1', collection: 'coll-a', crossdoc: false, updated_at: '2026-03-19T10:00:00Z' },
                ],
            };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockSessions });

            const result = await gatewayListSessions(5);

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/chat/sessions?user_id=5', expect.any(Object));
            expect(result).toEqual(mockSessions);
        });

        it('gatewayCreateSession should POST new session', async () => {
            const mockResponse = { id: 's-new', title: 'New Session', collection: 'test-coll' };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockResponse });

            const result = await gatewayCreateSession(8, 'test-coll', true);

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/chat/sessions?user_id=8',
                expect.objectContaining({
                    method: 'POST',
                    body: JSON.stringify({ collection: 'test-coll', crossdoc: true }),
                })
            );
            expect(result).toEqual(mockResponse);
        });

        it('gatewayCreateSession should default crossdoc to false', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ id: 's2', title: 'Test', collection: 'c' }) });

            await gatewayCreateSession(1, 'c');

            expect(mockFetch).toHaveBeenCalledWith(
                expect.any(String),
                expect.objectContaining({
                    body: JSON.stringify({ collection: 'c', crossdoc: false }),
                })
            );
        });

        it('gatewayGetSession should fetch session detail', async () => {
            const mockSession = {
                id: 's1',
                title: 'Test',
                collection: 'coll',
                crossdoc: false,
                updated_at: '2026-03-19T10:00:00Z',
                messages: [{ role: 'user', content: 'Hello', timestamp: '2026-03-19T10:01:00Z' }],
            };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockSession });

            const result = await gatewayGetSession('s1', 3);

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/chat/sessions/s1?user_id=3', expect.any(Object));
            expect(result).toEqual(mockSession);
        });

        it('gatewayDeleteSession should DELETE session', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true }) });

            const result = await gatewayDeleteSession('s-old', 9);

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/chat/sessions/s-old?user_id=9',
                expect.objectContaining({ method: 'DELETE' })
            );
            expect(result).toEqual({ ok: true });
        });
    });

    describe('Admin functions', () => {
        it('gatewayListUsers should fetch users', async () => {
            const mockUsers = {
                users: [{ id: 1, email: 'admin@example.com', name: 'Admin', area_id: 1, role: 'admin', active: true, last_login: null }],
            };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockUsers });

            const result = await gatewayListUsers();

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/admin/users', expect.any(Object));
            expect(result).toEqual(mockUsers);
        });

        it('gatewayCreateUser should POST new user', async () => {
            const userData = { email: 'newuser@example.com', name: 'New User', area_id: 2, role: 'user', password: 'pass123' };
            const mockResponse = { id: 10, email: 'newuser@example.com', api_key: 'api-key-abc' };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockResponse });

            const result = await gatewayCreateUser(userData);

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/admin/users',
                expect.objectContaining({
                    method: 'POST',
                    body: JSON.stringify(userData),
                })
            );
            expect(result).toEqual(mockResponse);
        });

        it('gatewayUpdateUser should PUT user updates', async () => {
            const updates = { name: 'Updated Name', active: false };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true }) });

            const result = await gatewayUpdateUser(5, updates);

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/admin/users/5',
                expect.objectContaining({
                    method: 'PUT',
                    body: JSON.stringify(updates),
                })
            );
            expect(result).toEqual({ ok: true });
        });

        it('gatewayDeleteUser should DELETE user', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true }) });

            const result = await gatewayDeleteUser(7);

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/admin/users/7',
                expect.objectContaining({ method: 'DELETE' })
            );
            expect(result).toEqual({ ok: true });
        });

        it('gatewayListAreas should fetch areas', async () => {
            const mockAreas = { areas: [{ id: 1, name: 'Engineering', description: 'Dev team' }] };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockAreas });

            const result = await gatewayListAreas();

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/admin/areas', expect.any(Object));
            expect(result).toEqual(mockAreas);
        });

        it('gatewayGetAreaCollections should fetch area collections', async () => {
            const mockCollections = { collections: [{ name: 'coll-x', permission: 'read' }] };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockCollections });

            const result = await gatewayGetAreaCollections(3);

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/admin/areas/3/collections', expect.any(Object));
            expect(result).toEqual(mockCollections);
        });

        it('gatewayGrantCollection should POST grant with default permission', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true }) });

            const result = await gatewayGrantCollection(2, 'test-coll');

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/admin/areas/2/collections',
                expect.objectContaining({
                    method: 'POST',
                    body: JSON.stringify({ collection_name: 'test-coll', permission: 'read' }),
                })
            );
            expect(result).toEqual({ ok: true });
        });

        it('gatewayGrantCollection should POST grant with custom permission', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true }) });

            await gatewayGrantCollection(2, 'admin-coll', 'write');

            expect(mockFetch).toHaveBeenCalledWith(
                expect.any(String),
                expect.objectContaining({
                    body: JSON.stringify({ collection_name: 'admin-coll', permission: 'write' }),
                })
            );
        });

        it('gatewayRevokeCollection should DELETE grant', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ ok: true }) });

            const result = await gatewayRevokeCollection(4, 'revoked-coll');

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/admin/areas/4/collections/revoked-coll',
                expect.objectContaining({ method: 'DELETE' })
            );
            expect(result).toEqual({ ok: true });
        });

        it('gatewayGetAudit should fetch audit log with query params', async () => {
            const mockAudit = {
                entries: [
                    { id: 1, user_id: 5, action: 'query', collection: 'test', query_preview: 'Hello', ip_address: '127.0.0.1', timestamp: '2026-03-19T10:00:00Z' },
                ],
            };
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => mockAudit });

            const result = await gatewayGetAudit({ user_id: 5, action: 'query', limit: 10 });

            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/admin/audit?user_id=5&action=query&limit=10',
                expect.any(Object)
            );
            expect(result).toEqual(mockAudit);
        });

        it('gatewayGetAudit should handle empty params', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ entries: [] }) });

            await gatewayGetAudit({});

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/admin/audit', expect.any(Object));
        });

        it('gatewayGetAudit should filter out null/undefined params', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ entries: [] }) });

            await gatewayGetAudit({ user_id: 1, action: undefined, collection: null });

            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9000/admin/audit?user_id=1', expect.any(Object));
        });
    });

    describe('Generate functions', () => {
        it('gatewayGenerateText should stream and accumulate text', async () => {
            const mockBody = {
                getReader: () => ({
                    read: vi
                        .fn()
                        .mockResolvedValueOnce({
                            done: false,
                            value: new TextEncoder().encode('data: {"choices":[{"delta":{"content":"Hello"}}]}\n'),
                        })
                        .mockResolvedValueOnce({
                            done: false,
                            value: new TextEncoder().encode('data: {"choices":[{"delta":{"content":" world"}}]}\n'),
                        })
                        .mockResolvedValueOnce({
                            done: false,
                            value: new TextEncoder().encode('data: [DONE]\n'),
                        })
                        .mockResolvedValueOnce({ done: true, value: undefined }),
                    releaseLock: vi.fn(),
                }),
            };
            mockFetch.mockResolvedValueOnce({ ok: true, body: mockBody });

            const result = await gatewayGenerateText({
                messages: [{ role: 'user', content: 'Test' }],
                use_knowledge_base: true,
                collection_names: ['test-coll'],
            });

            expect(result).toBe('Hello world');
            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/v1/generate',
                expect.objectContaining({
                    method: 'POST',
                    body: JSON.stringify({
                        messages: [{ role: 'user', content: 'Test' }],
                        use_knowledge_base: true,
                        collection_names: ['test-coll'],
                    }),
                })
            );
        });

        it('gatewayGenerateText should handle empty body', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, body: null });

            const result = await gatewayGenerateText({ messages: [] });

            expect(result).toBe('');
        });

        it('gatewayGenerateText should skip malformed JSON lines', async () => {
            const mockBody = {
                getReader: () => ({
                    read: vi
                        .fn()
                        .mockResolvedValueOnce({
                            done: false,
                            value: new TextEncoder().encode('data: {"choices":[{"delta":{"content":"OK"}}]}\ndata: {malformed}\n'),
                        })
                        .mockResolvedValueOnce({ done: true, value: undefined }),
                    releaseLock: vi.fn(),
                }),
            };
            mockFetch.mockResolvedValueOnce({ ok: true, body: mockBody });

            const result = await gatewayGenerateText({ messages: [] });

            expect(result).toBe('OK');
        });

        it('gatewayGenerateText should handle abort signal', async () => {
            const controller = new AbortController();
            const mockBody = {
                getReader: () => ({
                    read: vi.fn().mockResolvedValue({ done: true, value: undefined }),
                    releaseLock: vi.fn(),
                }),
            };
            mockFetch.mockResolvedValueOnce({ ok: true, body: mockBody });

            await gatewayGenerateText({ messages: [] }, controller.signal);

            expect(mockFetch).toHaveBeenCalledWith(
                expect.any(String),
                expect.objectContaining({ signal: controller.signal })
            );
        });

        it('gatewayGenerateStream should return Response for streaming', async () => {
            const mockResponse = new Response('stream', { status: 200 });
            mockFetch.mockResolvedValueOnce(mockResponse);

            const result = await gatewayGenerateStream({ messages: [{ role: 'user', content: 'Test' }] });

            expect(result).toBe(mockResponse);
            expect(mockFetch).toHaveBeenCalledWith(
                'http://localhost:9000/v1/generate',
                expect.objectContaining({
                    method: 'POST',
                    body: JSON.stringify({ messages: [{ role: 'user', content: 'Test' }] }),
                })
            );
        });

        it('gatewayGenerateStream should throw on non-ok response', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 500,
                text: async () => 'Internal error',
            });

            await expect(gatewayGenerateStream({ messages: [] })).rejects.toMatchObject({
                message: 'Gateway error (500): Internal error',
                status: 500,
            });
        });

        it('gatewayGenerateText lanza GatewayError cuando la respuesta no es ok', async () => {
            // Cubre línea 178: if (!resp.ok) throw new GatewayError(resp.status, await resp.text())
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 429,
                text: async () => 'Rate limit exceeded',
            });

            await expect(gatewayGenerateText({ messages: [] })).rejects.toMatchObject({
                status: 429,
                message: 'Gateway error (429): Rate limit exceeded',
            });
        });

        it('gatewayGenerateText skips líneas SSE que no empiezan con "data:"', async () => {
            // Cubre línea 193: if (!line.startsWith('data: ')) continue;
            const mockBody = {
                getReader: () => ({
                    read: vi.fn()
                        .mockResolvedValueOnce({
                            done: false,
                            value: new TextEncoder().encode(
                                ': keep-alive\n' +                      // comentario SSE → debe skipear
                                'event: ping\n' +                       // evento SSE → debe skipear
                                'data: {"choices":[{"delta":{"content":"texto"}}]}\n'
                            ),
                        })
                        .mockResolvedValueOnce({ done: true, value: undefined }),
                    releaseLock: vi.fn(),
                }),
            };
            mockFetch.mockResolvedValueOnce({ ok: true, body: mockBody });

            const result = await gatewayGenerateText({ messages: [] });

            // Solo el chunk con 'data:' debe haberse acumulado
            expect(result).toBe('texto');
        });

        it('gatewayGenerateText acumula string vacío cuando delta.content es undefined', async () => {
            // Cubre línea 198: ?? '' cuando choices[0].delta.content es undefined o null
            const mockBody = {
                getReader: () => ({
                    read: vi.fn()
                        .mockResolvedValueOnce({
                            done: false,
                            value: new TextEncoder().encode(
                                'data: {"choices":[{"delta":{}}]}\n' +      // sin content → ?? ''
                                'data: {"choices":[]}\n' +                  // sin choices → ?? ''
                                'data: {"choices":[{"delta":{"content":"ok"}}]}\n'
                            ),
                        })
                        .mockResolvedValueOnce({ done: true, value: undefined }),
                    releaseLock: vi.fn(),
                }),
            };
            mockFetch.mockResolvedValueOnce({ ok: true, body: mockBody });

            const result = await gatewayGenerateText({ messages: [] });

            // Los primeros dos chunks aportan '' (el ?? ''), el tercero aporta 'ok'
            expect(result).toBe('ok');
        });

        it('usa URL localhost:9000 por defecto cuando GATEWAY_URL no está configurado', async () => {
            // Cubre la rama ?? 'http://localhost:9000' de la línea 3 del módulo
            vi.resetModules();
            vi.unstubAllEnvs();
            vi.stubEnv('SYSTEM_API_KEY', 'test-system-key');
            // No stubeamos GATEWAY_URL → el módulo debe usar el fallback

            const localMock = vi.fn().mockResolvedValue({
                ok: true,
                json: async () => ({ token: 'x', user: { id: 1, email: 'a@b.com', name: 'A', role: 'user', area_id: 1 } }),
            });
            vi.stubGlobal('fetch', localMock);

            const { gatewayLogin: loginFresh } = await import('./gateway.js');
            await loginFresh('a@b.com', 'pass');

            expect(localMock).toHaveBeenCalledWith(
                expect.stringContaining('http://localhost:9000'),
                expect.any(Object)
            );

            // Restaurar para tests siguientes
            vi.stubEnv('GATEWAY_URL', 'http://localhost:9000');
            vi.stubGlobal('fetch', mockFetch);
        });
    });

    describe('Error handling', () => {
        it('should throw GatewayError on 4xx responses', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 404,
                text: async () => 'Not found',
            });

            await expect(gatewayListCollections()).rejects.toMatchObject({
                message: 'Gateway error (404): Not found',
                status: 404,
            });
        });

        it('should throw GatewayError on 5xx responses', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 503,
                text: async () => 'Service unavailable',
            });

            await expect(gatewayGetMe(1)).rejects.toMatchObject({
                message: 'Gateway error (503): Service unavailable',
                status: 503,
            });
        });

        it('should throw GatewayError on timeout (AbortError)', async () => {
            // Simulate an AbortError (what the AbortController signals to fetch)
            const abortErr = new Error('The operation was aborted');
            abortErr.name = 'AbortError';
            mockFetch.mockRejectedValueOnce(abortErr);

            await expect(gatewayListCollections()).rejects.toMatchObject({
                status: 504,
                message: expect.stringMatching(/timeout/i),
            });
        });

        it('should throw GatewayError on network errors', async () => {
            mockFetch.mockRejectedValueOnce(new Error('Network failed'));

            await expect(gatewayListSessions(1)).rejects.toMatchObject({
                message: 'Gateway error (502): Gateway unreachable: Network failed',
                status: 502,
            });
        });

        it('should rethrow GatewayError as-is', async () => {
            const customError = new GatewayError(401, 'Unauthorized');
            mockFetch.mockRejectedValueOnce(customError);

            await expect(gatewayLogin('a', 'b')).rejects.toBe(customError);
        });
    });

    describe('Timeout handling', () => {
        it('should use default timeout if not specified', async () => {
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ collections: [] }) });

            await gatewayListCollections();

            // AbortController's signal should have been passed
            expect(mockFetch).toHaveBeenCalledWith(
                expect.any(String),
                expect.objectContaining({
                    signal: expect.any(AbortSignal),
                })
            );
        });

        it('should clear timeout after successful response', async () => {
            const clearTimeoutSpy = vi.spyOn(global, 'clearTimeout');
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({ collections: [] }) });

            await gatewayListCollections();

            expect(clearTimeoutSpy).toHaveBeenCalled();
        });

        it('should clear timeout even on error', async () => {
            const clearTimeoutSpy = vi.spyOn(global, 'clearTimeout');
            mockFetch.mockRejectedValueOnce(new Error('fail'));

            await expect(gatewayListCollections()).rejects.toThrow();

            expect(clearTimeoutSpy).toHaveBeenCalled();
        });
    });

    describe('Missing SYSTEM_API_KEY', () => {
        it('should throw if SYSTEM_API_KEY is not set', async () => {
            vi.stubEnv('SYSTEM_API_KEY', '');
            mockFetch.mockResolvedValueOnce({ ok: true, json: async () => ({}) });

            await expect(gatewayListCollections()).rejects.toThrow('SYSTEM_API_KEY environment variable is required');
        });
    });
});
