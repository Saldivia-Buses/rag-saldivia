// Svelte 5 runes-based reactive store for chat state

export interface Message {
    role: 'user' | 'assistant';
    content: string;
    sources?: Source[];
    timestamp: string;
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

    addUserMessage(content: string) {
        this.messages.push({ role: 'user', content, timestamp: new Date().toISOString() });
    }

    startStream() {
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

    finalizeStream() {
        if (this.streamingContent) {
            this.messages.push({
                role: 'assistant',
                content: this.streamingContent,
                sources: [...this.sources],
                timestamp: new Date().toISOString()
            });
        }
        this.streaming = false;
        this.streamingContent = '';
    }

    loadMessages(messages: Message[]) {
        this.messages = messages;
    }
}
