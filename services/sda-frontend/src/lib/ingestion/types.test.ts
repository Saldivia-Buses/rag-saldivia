import { describe, it, expect } from 'vitest';
import { classifyTier, classifyTierBySize, estimateETA, TIER_CONFIG } from './types.js';

describe('classifyTier', () => {
    it('clasifica correctamente por páginas', () => {
        expect(classifyTier(1)).toBe('tiny');
        expect(classifyTier(20)).toBe('tiny');
        expect(classifyTier(21)).toBe('small');
        expect(classifyTier(80)).toBe('small');
        expect(classifyTier(81)).toBe('medium');
        expect(classifyTier(250)).toBe('medium');
        expect(classifyTier(251)).toBe('large');
        expect(classifyTier(5000)).toBe('large');
    });
});

describe('classifyTierBySize', () => {
    it('clasifica por tamaño de archivo', () => {
        expect(classifyTierBySize(50_000)).toBe('tiny');
        expect(classifyTierBySize(99_999)).toBe('tiny');
        expect(classifyTierBySize(100_000)).toBe('small');
        expect(classifyTierBySize(499_999)).toBe('small');
        expect(classifyTierBySize(500_000)).toBe('medium');
        expect(classifyTierBySize(4_999_999)).toBe('medium');
        expect(classifyTierBySize(5_000_000)).toBe('large');
    });
});

describe('estimateETA', () => {
    it('devuelve expectedMaxDuration cuando progress es 0', () => {
        expect(estimateETA('medium', 0, 0)).toBe(TIER_CONFIG.medium.expectedMaxDuration);
    });

    it('devuelve 0 cuando progress es 100', () => {
        expect(estimateETA('tiny', 100, 30)).toBe(0);
    });

    it('calcula ETA correctamente con progreso real', () => {
        // 50% en 60s → total estimado = 120s → restante = 60s
        expect(estimateETA('small', 50, 60)).toBe(60);
    });

    it('nunca devuelve negativo', () => {
        expect(estimateETA('tiny', 99, 10000)).toBe(0);
    });
});

describe('TIER_CONFIG', () => {
    it('todos los tiers tienen los campos requeridos', () => {
        for (const tier of ['tiny', 'small', 'medium', 'large'] as const) {
            const config = TIER_CONFIG[tier];
            expect(config.pollInterval).toBeGreaterThan(0);
            expect(config.deadlockThreshold).toBeGreaterThan(0);
            expect(config.expectedMaxDuration).toBeGreaterThan(0);
            expect(['green', 'blue', 'amber', 'red']).toContain(config.color);
        }
    });
});
