import { describe, it, expect } from 'vitest';
import { isNearBottom } from './scroll.js';

describe('isNearBottom', () => {
    it('retorna true cuando está en el fondo exacto', () => {
        // scrollHeight - scrollTop - clientHeight = 0
        const el = { scrollHeight: 1000, scrollTop: 900, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(true);
    });

    it('retorna true cuando está dentro del threshold (por defecto 100px)', () => {
        // 1000 - 850 - 100 = 50 < 100
        const el = { scrollHeight: 1000, scrollTop: 850, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(true);
    });

    it('retorna false cuando está más arriba del threshold', () => {
        // 1000 - 700 - 100 = 200 > 100
        const el = { scrollHeight: 1000, scrollTop: 700, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(false);
    });

    it('retorna false cuando está al tope', () => {
        const el = { scrollHeight: 1000, scrollTop: 0, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(false);
    });

    it('respeta threshold customizado', () => {
        const el = { scrollHeight: 1000, scrollTop: 850, clientHeight: 100 };
        // distancia = 50
        expect(isNearBottom(el, 60)).toBe(true);   // 50 < 60
        expect(isNearBottom(el, 40)).toBe(false);  // 50 > 40
    });

    it('retorna true cuando el contenido es más corto que el viewport', () => {
        // scrollHeight <= clientHeight → siempre en el fondo
        const el = { scrollHeight: 80, scrollTop: 0, clientHeight: 100 };
        expect(isNearBottom(el)).toBe(true);
    });
});
