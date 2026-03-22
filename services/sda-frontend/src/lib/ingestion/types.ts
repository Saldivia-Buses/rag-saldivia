export type Tier = 'tiny' | 'small' | 'medium' | 'large';

export interface TierConfig {
    label: string;
    color: 'green' | 'blue' | 'amber' | 'red';
    maxPages: number;
    pollInterval: number;       // segundos entre polls
    deadlockThreshold: number;  // segundos sin progreso → stalled
    expectedMaxDuration: number; // segundos máximos esperados
    secondsPerPage: number;     // para estimación de ETA por páginas
}

export const TIER_CONFIG = {
    tiny: {
        label: 'Pequeño',
        color: 'green',
        maxPages: 20,
        pollInterval: 2,
        deadlockThreshold: 30,
        expectedMaxDuration: 45,
        secondsPerPage: 1.5,
    },
    small: {
        label: 'Mediano',
        color: 'blue',
        maxPages: 80,
        pollInterval: 3,
        deadlockThreshold: 60,
        expectedMaxDuration: 120,
        secondsPerPage: 2.0,
    },
    medium: {
        label: 'Grande',
        color: 'amber',
        maxPages: 250,
        pollInterval: 5,
        deadlockThreshold: 90,
        expectedMaxDuration: 480,
        secondsPerPage: 2.5,
    },
    large: {
        label: 'Muy grande',
        color: 'red',
        maxPages: Infinity,
        pollInterval: 10,
        deadlockThreshold: 120,
        expectedMaxDuration: 1800,
        secondsPerPage: 3.5,
    },
} as const satisfies Record<Tier, TierConfig>;

/**
 * Clasifica tier por número de páginas (para PDFs).
 */
export function classifyTier(pages: number): Tier {
    if (pages <= 20)  return 'tiny';
    if (pages <= 80)  return 'small';
    if (pages <= 250) return 'medium';
    return 'large';
}

/**
 * Clasifica tier por tamaño de archivo en bytes (para non-PDFs).
 */
export function classifyTierBySize(bytes: number): Tier {
    if (bytes < 100_000)   return 'tiny';
    if (bytes < 500_000)   return 'small';
    if (bytes < 5_000_000) return 'medium';
    return 'large';
}

/**
 * Estima segundos restantes para completar la ingesta.
 * @param tier - Tier del job
 * @param progress - Progreso actual (0-100)
 * @param elapsedSeconds - Segundos transcurridos desde el inicio
 */
export function estimateETA(tier: Tier, progress: number, elapsedSeconds: number): number {
    if (progress >= 100) return 0;
    if (progress === 0) return TIER_CONFIG[tier].expectedMaxDuration;

    const totalEstimated = (elapsedSeconds / progress) * 100;
    const remainingByElapsed = Math.max(0, totalEstimated - elapsedSeconds);

    // Cap: nunca estimar más que lo que esperaría el tier para el progreso restante
    const remainingByTier = TIER_CONFIG[tier].expectedMaxDuration * (1 - progress / 100);
    const remaining = Math.min(remainingByElapsed, remainingByTier);

    return Math.round(remaining);
}
