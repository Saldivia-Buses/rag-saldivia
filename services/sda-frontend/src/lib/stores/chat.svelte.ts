// Svelte 5 runes-based reactive store for chat state

export interface Message {
    role: 'user' | 'assistant';
    content: string;
    sources?: Source[];
    timestamp: string;
    crossdocResults?: import('$lib/crossdoc/types').SubResult[];
}

export interface Source {
    document: string;
    page?: number;
    excerpt: string;
}

export class ChatStore {
    messages = $state<Message[]>([]);
    sources = $state<Source[]>([]);
    streaming = $state(false);
    streamingContent = $state('');
    collection = $state('');
    crossdoc = $state(false);
    abortController = $state<AbortController | null>(null);

    addUserMessage(content: string) {
        this.messages.push({ role: 'user', content, timestamp: new Date().toISOString() });
    }

    startStream() {
        this.abortController = new AbortController();
        this.streaming = true;
        this.streamingContent = '';
        this.sources = [];
    }

    appendToken(token: string) {
        this.streamingContent += token;
    }

    setSources(sources: Source[]) {
        this.sources = sources;
    }

    stopStream() {
        this.abortController?.abort();
        this.finalizeStream();
    }

    finalizeStream(opts?: { crossdocResults?: import('$lib/crossdoc/types').SubResult[] }) {
        if (this.streamingContent) {
            this.messages.push({
                role: 'assistant',
                content: this.streamingContent,
                sources: [...this.sources],
                timestamp: new Date().toISOString(),
                crossdocResults: opts?.crossdocResults,
            });
        }
        this.streaming = false;
        this.streamingContent = '';
    }

    loadMessages(messages: Message[]) {
        this.messages = messages;
    }
}
