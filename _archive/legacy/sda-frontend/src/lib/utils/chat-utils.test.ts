import { describe, it, expect } from 'vitest';
import { filterSessions } from './chat-utils.js';

describe('filterSessions', () => {
    const sessions = [
        { id: '1', title: 'Manual Aries 365', updated_at: '2026-03-18T10:00:00' },
        { id: '2', title: 'Normativas homologación', updated_at: '2026-03-17T10:00:00' },
        { id: '3', title: 'Motor ZF especificaciones', updated_at: '2026-03-16T10:00:00' },
    ];

    it('retorna todas las sesiones cuando query es vacío', () => {
        expect(filterSessions(sessions, '')).toHaveLength(3);
    });

    it('retorna todas las sesiones cuando query es solo espacios', () => {
        expect(filterSessions(sessions, '   ')).toHaveLength(3);
    });

    it('filtra por título case-insensitive', () => {
        const result = filterSessions(sessions, 'aries');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('1');
    });

    it('filtra con mayúsculas en el query', () => {
        const result = filterSessions(sessions, 'MOTOR');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('3');
    });

    it('retorna array vacío si no hay coincidencias', () => {
        expect(filterSessions(sessions, 'inexistente xyz')).toHaveLength(0);
    });

    it('encuentra coincidencias parciales', () => {
        const result = filterSessions(sessions, 'homol');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('2');
    });
});
