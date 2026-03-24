import { describe, it, expect } from 'vitest';
import { computeSHA256 } from './hash.js';

describe('computeSHA256', () => {
    it('produces a 64-char hex string', async () => {
        const file = new File(['hello world'], 'test.txt', { type: 'text/plain' });
        const hash = await computeSHA256(file);
        expect(hash).toHaveLength(64);
        expect(hash).toMatch(/^[0-9a-f]+$/);
    });

    it('same content produces same hash', async () => {
        const file1 = new File(['abc'], 'a.txt');
        const file2 = new File(['abc'], 'b.txt');
        expect(await computeSHA256(file1)).toBe(await computeSHA256(file2));
    });

    it('different content produces different hash', async () => {
        const file1 = new File(['abc'], 'a.txt');
        const file2 = new File(['xyz'], 'b.txt');
        expect(await computeSHA256(file1)).not.toBe(await computeSHA256(file2));
    });
});
