/**
 * Available LLM models for the chat interface.
 * These map to OpenRouter model IDs when using the OpenRouter backend.
 * When using the NVIDIA Blueprint, the model is ignored (Blueprint has its own).
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
  {
    id: "anthropic/claude-sonnet-4",
    name: "Claude Sonnet 4",
    provider: "Anthropic",
    providerLogo: "anthropic",
    description: "Excelente balance velocidad/calidad",
    contextWindow: "200K",
    speed: "fast",
  },
  {
    id: "anthropic/claude-opus-4",
    name: "Claude Opus 4",
    provider: "Anthropic",
    providerLogo: "anthropic",
    description: "Maximo razonamiento y calidad",
    contextWindow: "200K",
    speed: "slow",
  },
  {
    id: "anthropic/claude-haiku-4",
    name: "Claude Haiku 4",
    provider: "Anthropic",
    providerLogo: "anthropic",
    description: "Ultra rapido para tareas simples",
    contextWindow: "200K",
    speed: "fast",
  },
  {
    id: "openai/gpt-4.1",
    name: "GPT-4.1",
    provider: "OpenAI",
    providerLogo: "openai",
    description: "Modelo flagship de OpenAI",
    contextWindow: "1M",
    speed: "medium",
  },
  {
    id: "openai/gpt-4.1-mini",
    name: "GPT-4.1 Mini",
    provider: "OpenAI",
    providerLogo: "openai",
    description: "Rapido y economico",
    contextWindow: "1M",
    speed: "fast",
  },
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
  {
    id: "deepseek/deepseek-r1",
    name: "DeepSeek R1",
    provider: "DeepSeek",
    providerLogo: "deepseek",
    description: "Razonamiento avanzado open source",
    contextWindow: "64K",
    speed: "slow",
  },
  {
    id: "meta-llama/llama-4-maverick",
    name: "Llama 4 Maverick",
    provider: "Meta",
    providerLogo: "llama",
    description: "Open source de Meta",
    contextWindow: "1M",
    speed: "medium",
  },
];

export const DEFAULT_MODEL = MODELS[0]; // Claude Sonnet 4
