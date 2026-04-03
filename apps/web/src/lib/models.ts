/**
 * Available LLM models for the chat interface.
 * These map to OpenRouter model IDs.
 */

export interface LLMModel {
  id: string;
  name: string;
  provider: string;
  providerLogo: string;
  description: string;
  contextWindow: string;
  speed: "fast" | "medium" | "slow";
}

export const MODELS: LLMModel[] = [
  // Claude
  {
    id: "anthropic/claude-sonnet-4-6",
    name: "Claude Sonnet 4.6",
    provider: "Claude",
    providerLogo: "anthropic",
    description: "Excelente balance velocidad/calidad",
    contextWindow: "200K",
    speed: "fast",
  },
  {
    id: "anthropic/claude-opus-4-6",
    name: "Claude Opus 4.6",
    provider: "Claude",
    providerLogo: "anthropic",
    description: "Maximo razonamiento y calidad",
    contextWindow: "1M",
    speed: "slow",
  },
  {
    id: "anthropic/claude-haiku-4-5",
    name: "Claude Haiku 4.5",
    provider: "Claude",
    providerLogo: "anthropic",
    description: "Ultra rapido para tareas simples",
    contextWindow: "200K",
    speed: "fast",
  },
  // Google
  {
    id: "google/gemini-2.5-pro-preview",
    name: "Gemini 2.5 Pro",
    provider: "Google",
    providerLogo: "google",
    description: "Largo contexto, multimodal",
    contextWindow: "1M",
    speed: "medium",
  },
  {
    id: "google/gemini-2.5-flash-preview",
    name: "Gemini 2.5 Flash",
    provider: "Google",
    providerLogo: "google",
    description: "Ultra rapido de Google",
    contextWindow: "1M",
    speed: "fast",
  },
  // DeepSeek
  {
    id: "deepseek/deepseek-r1",
    name: "DeepSeek R1",
    provider: "DeepSeek",
    providerLogo: "deepseek",
    description: "Razonamiento avanzado open source",
    contextWindow: "164K",
    speed: "slow",
  },
  // Meta
  {
    id: "meta-llama/llama-4-maverick",
    name: "Llama 4 Maverick",
    provider: "Meta",
    providerLogo: "llama",
    description: "Open source de Meta",
    contextWindow: "1M",
    speed: "medium",
  },
  // Qwen
  {
    id: "qwen/qwen3-235b-a22b",
    name: "Qwen 3 235B",
    provider: "Alibaba",
    providerLogo: "alibaba",
    description: "MoE abierto de alto rendimiento",
    contextWindow: "128K",
    speed: "medium",
  },
  // Mistral
  {
    id: "mistralai/mistral-large-2411",
    name: "Mistral Large",
    provider: "Mistral",
    providerLogo: "mistral",
    description: "Flagship europeo multilingue",
    contextWindow: "128K",
    speed: "medium",
  },
];

export const DEFAULT_MODEL = MODELS[0]; // Claude Sonnet 4.6
