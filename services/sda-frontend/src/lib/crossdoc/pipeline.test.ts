import { describe, it, expect } from 'vitest';
import { jaccard, dedup, parseSubQueries, hasUsefulData } from './pipeline.js';

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
});
