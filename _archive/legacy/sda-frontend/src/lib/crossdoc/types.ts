export interface CrossdocOptions {
	maxSubQueries: number; // 0 = ilimitado
	synthesisModel: string; // '' = usar LLM por defecto
	followUpRetries: boolean;
	showDecomposition: boolean;
	vdbTopK: number;
	rerankerTopK: number;
}

export interface SubResult {
	query: string;
	content: string;
	success: boolean;
}

export interface CrossdocProgress {
	phase: 'decomposing' | 'querying' | 'retrying' | 'synthesizing' | 'done' | 'error';
	subQueries: string[];
	completed: number;
	total: number;
	results: SubResult[];
	error?: string;
}

export const DEFAULT_CROSSDOC_OPTIONS: CrossdocOptions = {
	maxSubQueries: 4,
	synthesisModel: '',
	followUpRetries: true,
	showDecomposition: false,
	vdbTopK: 10,
	rerankerTopK: 5,
};
