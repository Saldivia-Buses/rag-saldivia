export interface SessionSummary {
    id: string;
    title: string;
    updated_at: string;
}

/**
 * Filtra sesiones por título (case-insensitive, substring match).
 * Retorna todas si query es vacío o solo espacios.
 */
export function filterSessions(sessions: SessionSummary[], query: string): SessionSummary[] {
    const q = query.trim().toLowerCase();
    if (!q) return sessions;
    return sessions.filter(s => s.title.toLowerCase().includes(q));
}
