interface ExportSession {
    id: string;
    title: string;
    created_at: string;
}

interface ExportMessage {
    role: 'user' | 'assistant';
    content: string;
    timestamp: string;
}

export function buildMarkdown(session: ExportSession, messages: ExportMessage[]): string {
    const lines = [
        `# ${session.title}`,
        `_${new Date(session.created_at).toLocaleDateString('es-AR')}_`,
        '',
    ];
    for (const m of messages) {
        const label = m.role === 'user' ? '**Vos:**' : '**SDA:**';
        lines.push(`${label} ${m.content}`, '');
    }
    return lines.join('\n');
}

export function buildJSON(session: ExportSession, messages: ExportMessage[]): string {
    return JSON.stringify({ session, messages }, null, 2);
}

export function downloadFile(content: string, filename: string, mime: string): void {
    const blob = new Blob([content], { type: mime });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}
