const TEMPLATES = [
    (t: string) => `¿Podés ampliar sobre ${t}?`,
    (t: string) => `¿Cuáles son los riesgos de ${t}?`,
    (t: string) => `¿Hay más documentos sobre ${t}?`,
];

export function generateFollowUps(content: string, originalQuery: string): string[] {
    const sentences = content.split(/[.!]\s+/).map(s => s.trim()).filter(s => s.length > 20);
    if (sentences.length < 2) return [];

    const topics = sentences.slice(0, 4).map(s => {
        const words = s.split(/\s+/);
        return words.slice(-3).join(' ').replace(/[^a-záéíóúüñA-ZÁÉÍÓÚÜÑ\s]/g, '').trim();
    }).filter(t => t.length > 3);

    if (topics.length === 0) return [];

    return topics.slice(0, 3).map((topic, i) =>
        TEMPLATES[i % TEMPLATES.length](topic)
    ).filter(s => s.toLowerCase() !== originalQuery.toLowerCase());
}
