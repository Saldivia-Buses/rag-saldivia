import { describe, it, expect } from 'vitest';
import { generateFollowUps } from './followups.js';

describe('generateFollowUps', () => {
    it('retorna 2-3 sugerencias para contenido con varias oraciones', () => {
        const content = 'El sistema de autenticación utiliza JWT. Los tokens expiran en 8 horas. Cada usuario tiene un rol asignado. El admin puede gestionar usuarios.';
        const result = generateFollowUps(content, '');
        expect(result.length).toBeGreaterThanOrEqual(2);
        expect(result.length).toBeLessThanOrEqual(3);
    });

    it('retorna array vacío para contenido muy corto (menos de 2 oraciones largas)', () => {
        const result = generateFollowUps('Ok.', '');
        expect(result).toHaveLength(0);
    });

    it('cada sugerencia termina en ?', () => {
        const content = 'El sistema de autenticación utiliza JWT. Los tokens expiran en 8 horas. Cada usuario tiene un rol asignado.';
        const result = generateFollowUps(content, '');
        result.forEach(s => expect(s).toMatch(/\?$/));
    });

    it('no retorna más de 3 sugerencias aunque haya muchas oraciones', () => {
        const content = 'A. B tiene algo importante. C tiene algo importante. D tiene algo importante. E tiene algo importante. F tiene algo importante.';
        const result = generateFollowUps(content, '');
        expect(result.length).toBeLessThanOrEqual(3);
    });
});
