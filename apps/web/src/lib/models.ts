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
}

export const MODELS: LLMModel[] = [
  // Claude
  {
    id: "anthropic/claude-sonnet-4-6",
    name: "Claude Sonnet 4.6",
    provider: "Anthropic",
    providerLogo: "anthropic",
    description: "Rapido, inteligente, ideal para todo",
    contextWindow: "200K",
  },
  {
    id: "anthropic/claude-opus-4-6",
    name: "Claude Opus 4.6",
    provider: "Anthropic",
    providerLogo: "anthropic",
    description: "Maximo razonamiento y precision",
    contextWindow: "1M",
  },
  {
    id: "anthropic/claude-haiku-4-5",
    name: "Claude Haiku 4.5",
    provider: "Anthropic",
    providerLogo: "anthropic",
    description: "Ultra rapido, ideal para tareas simples",
    contextWindow: "200K",
  },
  // Google
  {
    id: "google/gemini-2.5-pro",
    name: "Gemini 2.5 Pro",
    provider: "Google",
    providerLogo: "google",
    description: "Multimodal con contexto de 1M tokens",
    contextWindow: "1M",
  },
  {
    id: "google/gemini-2.5-flash",
    name: "Gemini 2.5 Flash",
    provider: "Google",
    providerLogo: "google",
    description: "Velocidad extrema de Google",
    contextWindow: "1M",
  },
  // DeepSeek
  {
    id: "deepseek/deepseek-r1-0528",
    name: "DeepSeek R1",
    provider: "DeepSeek",
    providerLogo: "deepseek",
    description: "Razonamiento avanzado open source",
    contextWindow: "164K",
  },
  // Meta
  {
    id: "meta-llama/llama-4-maverick",
    name: "Llama 4 Maverick",
    provider: "Meta",
    providerLogo: "llama",
    description: "MoE abierto de 400B parametros",
    contextWindow: "1M",
  },
  // Qwen
  {
    id: "qwen/qwen3-235b-a22b",
    name: "Qwen 3 235B",
    provider: "Qwen",
    providerLogo: "alibaba",
    description: "MoE de Alibaba, 22B activos",
    contextWindow: "128K",
  },
  // Mistral
  {
    id: "mistralai/mistral-medium-3",
    name: "Mistral Medium 3",
    provider: "Mistral",
    providerLogo: "mistral",
    description: "Flagship europeo multilingue",
    contextWindow: "128K",
  },
];

export const DEFAULT_MODEL = MODELS[0]; // Claude Sonnet 4.6
