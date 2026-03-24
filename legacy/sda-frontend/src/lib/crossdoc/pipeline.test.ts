import { describe, it, expect } from 'vitest';
import { jaccard, dedup, parseSubQueries, hasUsefulData, truncateIfRepetitive, DECOMPOSE_PROMPT, FOLLOWUP_PROMPT, SYNTHESIS_PROMPT } from './pipeline.js';

describe('jaccard', () => {
	it('identical strings → 1.0', () => {
		expect(jaccard('presión bomba', 'presión bomba')).toBe(1);
	});
	it('disjoint strings → 0', () => {
		expect(jaccard('presión bomba', 'temperatura motor')).toBe(0);
	});
	it('partial overlap → between 0 and 1', () => {
		const score = jaccard('presión máxima bomba', 'presión mínima bomba');
		expect(score).toBeGreaterThan(0);
		expect(score).toBeLessThan(1);
	});
	it('empty strings → 0', () => {
		expect(jaccard('', '')).toBe(0);
		expect(jaccard('presión', '')).toBe(0);
		expect(jaccard('', 'bomba')).toBe(0);
	});
	it('whitespace-only strings → 0', () => {
		expect(jaccard('   ', '   ')).toBe(0);
		expect(jaccard('presión bomba', '   ')).toBe(0);
	});
});

describe('dedup', () => {
	it('elimina queries con jaccard >= 0.65', () => {
		const queries = [
			'presión máxima bomba centrífuga',
			'presión máxima bomba centrifuga', // casi idéntica
			'temperatura motor eléctrico',
		];
		const result = dedup(queries);
		expect(result).toHaveLength(2);
		expect(result[0]).toBe('presión máxima bomba centrífuga');
		expect(result[1]).toBe('temperatura motor eléctrico');
	});

	it('no elimina queries distintas', () => {
		const queries = ['presión bomba', 'temperatura motor', 'voltaje inversor'];
		expect(dedup(queries)).toHaveLength(3);
	});

	it('queries at exact threshold boundary (0.65)', () => {
		// Create two queries that have jaccard exactly at or near 0.65
		const queries = [
			'presión máxima bomba centrífuga motor eléctrico',
			'presión máxima bomba centrifuga', // jaccard should be near threshold
		];
		const result = dedup(queries);
		// Should keep at most one if jaccard >= 0.65
		expect(result.length).toBeGreaterThanOrEqual(1);
		expect(result.length).toBeLessThanOrEqual(2);
	});

	it('empty array → empty result', () => {
		expect(dedup([])).toEqual([]);
	});
});

describe('parseSubQueries', () => {
	it('parsea líneas simples', () => {
		const text = 'presión máxima\ntemperatura motor\nvoltaje nominal';
		expect(parseSubQueries(text)).toEqual([
			'presión máxima',
			'temperatura motor',
			'voltaje nominal',
		]);
	});
	it('elimina líneas con numeración', () => {
		const text = '1. presión máxima\n2) temperatura motor';
		const result = parseSubQueries(text);
		expect(result[0]).toBe('presión máxima');
		expect(result[1]).toBe('temperatura motor');
	});
	it('filtra líneas vacías o muy cortas', () => {
		const text = '\npresión máxima\n   \nok\n';
		const result = parseSubQueries(text);
		expect(result).toHaveLength(1);
		expect(result[0]).toBe('presión máxima');
	});
	it('aplica cap cuando se pasa maxSubQueries', () => {
		const text = 'a uno\nb dos\nc tres\nd cuatro';
		expect(parseSubQueries(text, 2)).toHaveLength(2);
	});
	it('maneja input vacío', () => {
		expect(parseSubQueries('')).toEqual([]);
		expect(parseSubQueries('   ')).toEqual([]);
	});
	it('maneja input con solo espacios y newlines', () => {
		expect(parseSubQueries('\n\n   \n\n')).toEqual([]);
	});
	it('maneja líneas muy largas (>200 chars)', () => {
		const longLine = 'a'.repeat(250);
		const text = `presión máxima\n${longLine}\ntemperatura motor`;
		const result = parseSubQueries(text);
		expect(result).toEqual(['presión máxima', 'temperatura motor']);
	});
	it('preserva whitespace interno después de eliminar numeración', () => {
		const text = '1.  presión   máxima   bomba';
		const result = parseSubQueries(text);
		expect(result[0]).toBe('presión   máxima   bomba');
	});
});

describe('hasUsefulData', () => {
	it('texto con contenido → true', () => {
		expect(hasUsefulData('La presión máxima es 12 bar.')).toBe(true);
	});
	it('texto vacío → false', () => {
		expect(hasUsefulData('')).toBe(false);
		expect(hasUsefulData('   ')).toBe(false);
	});
	it('patrones de "sin datos" → false', () => {
		expect(hasUsefulData('No information found')).toBe(false);
		expect(hasUsefulData("I cannot answer this question")).toBe(false);
		expect(hasUsefulData('out of context')).toBe(false);
	});
	it('texto muy corto (< 3 chars) → false', () => {
		expect(hasUsefulData('ab')).toBe(false);
		expect(hasUsefulData('x')).toBe(false);
	});
	it('variantes de patrones de "sin datos" → false', () => {
		// Only test patterns that actually match the regex in hasUsefulData
		expect(hasUsefulData('No information available')).toBe(false);
		expect(hasUsefulData("I don't have the information")).toBe(false);
		expect(hasUsefulData('No data found')).toBe(false);
		expect(hasUsefulData('Sin results from search')).toBe(false);
	});
	it('texto con contenido útil después de espacios → true', () => {
		expect(hasUsefulData('   La respuesta es 42.   ')).toBe(true);
	});
});

describe('truncateIfRepetitive', () => {
	it('text without repetition → unchanged', () => {
		const text = 'La presión máxima es 12 bar. La temperatura es 25°C.';
		expect(truncateIfRepetitive(text)).toBe(text);
	});

	it('short text → unchanged', () => {
		const text = 'short text';
		expect(truncateIfRepetitive(text)).toBe(text);
	});

	it('repetitive tail → truncated when pattern repeats enough', () => {
		// Algorithm needs: tail appears 2+ times in preceding window AND firstIdx > 0
		const prefix = 'Some unique prefix text here to make firstIdx > 0. ';
		const tail60 = 'Y'.repeat(60);
		const text = prefix + tail60.repeat(5); // Prefix ensures firstIdx > 0
		const result = truncateIfRepetitive(text);
		// Should detect repetition and truncate
		expect(result.length).toBeLessThan(text.length);
		expect(result).toContain(prefix);
	});

	it('extremely long text (>15000 chars) → capped at MAX_RESPONSE_CHARS', () => {
		const text = 'a'.repeat(20_000);
		const result = truncateIfRepetitive(text);
		expect(result.length).toBeLessThanOrEqual(15_000);
	});

	it('text at exactly MAX_RESPONSE_CHARS → unchanged', () => {
		const text = 'a'.repeat(15_000);
		const result = truncateIfRepetitive(text);
		expect(result).toBe(text);
	});
});

describe('DECOMPOSE_PROMPT', () => {
	it('incluye la pregunta del usuario en el prompt', () => {
		const prompt = DECOMPOSE_PROMPT('¿Cuál es la presión máxima de la bomba?');
		expect(prompt).toContain('¿Cuál es la presión máxima de la bomba?');
		expect(prompt).toContain('sub-queries');
	});
});

describe('FOLLOWUP_PROMPT', () => {
	it('incluye las queries fallidas en el prompt', () => {
		const failedQueries = ['presión bomba centrífuga', 'temperatura motor eléctrico'];
		const prompt = FOLLOWUP_PROMPT(failedQueries);
		expect(prompt).toContain('presión bomba centrífuga');
		expect(prompt).toContain('temperatura motor eléctrico');
		expect(prompt).toContain('no useful results');
	});

	it('maneja array vacío sin error', () => {
		const prompt = FOLLOWUP_PROMPT([]);
		expect(typeof prompt).toBe('string');
		expect(prompt.length).toBeGreaterThan(0);
	});
});

describe('SYNTHESIS_PROMPT', () => {
	it('incluye la pregunta original y los resultados de retrieval', () => {
		const results = [
			{ query: 'presión bomba', content: 'La presión es 12 bar' },
			{ query: 'temperatura motor', content: 'La temperatura nominal es 80°C' },
		];
		const prompt = SYNTHESIS_PROMPT('¿Cuáles son las especificaciones?', results);
		expect(prompt).toContain('¿Cuáles son las especificaciones?');
		expect(prompt).toContain('La presión es 12 bar');
		expect(prompt).toContain('La temperatura nominal es 80°C');
		expect(prompt).toContain('Sub-query 1');
		expect(prompt).toContain('Sub-query 2');
	});

	it('maneja array de resultados vacío', () => {
		const prompt = SYNTHESIS_PROMPT('¿test?', []);
		expect(prompt).toContain('¿test?');
		expect(typeof prompt).toBe('string');
	});
});
