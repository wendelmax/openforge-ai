// Package modelzoo maintains the catalog of available models for the runtime.
package modelzoo

// ModelType categorizes a model's intended use case.
type ModelType string

const (
	// TypeChat models are designed for general chat and instruction following.
	TypeChat ModelType = "chat"
	// TypeEmbed models generate vector embeddings from text.
	TypeEmbed ModelType = "embedding"
	// TypeReranker models are cross-encoders for reranking documents.
	TypeReranker ModelType = "reranker"
	// TypeCode models are specialized for code completion and generation.
	TypeCode ModelType = "code"
	// TypeVision models handle image inputs alongside text.
	TypeVision ModelType = "vision"
)

// Model describes a downloadable AI model and its metadata.
type Model struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	HuggingFace  string            `json:"huggingface"`
	Type         ModelType         `json:"type"`
	Precision    string            `json:"precision"`
	Description  string            `json:"description"`
	SizeMB       int64             `json:"size_mb"`
	RamMinMB     int64             `json:"ram_min_mb"`
	ContextLen   int               `json:"context_len"`
	DefaultModel bool              `json:"default_model"`
	Files        []string          `json:"files"`
	Aliases      []string          `json:"aliases,omitempty"`
}

// Catalog returns the complete list of available models.
func Catalog() []Model {
	return []Model{
		{
			ID:           "phi-3-mini",
			Name:         "Phi-3 Mini 3.8B",
			HuggingFace:  "OpenVINO/phi-3-mini-int4-ov",
			Type:         TypeChat,
			Precision:    "INT4",
			Description:  "Microsoft Phi-3 Mini 3.8B — optimized for chat, high quality, low resource",
			SizeMB:       2400,
			RamMinMB:     4096,
			ContextLen:   4096,
			DefaultModel: true,
		},
		{
			ID:           "phi-3-mini-int4",
			Name:         "Phi-3 Mini 3.8B INT4",
			HuggingFace:  "OpenVINO/phi-3-mini-int4-ov",
			Type:         TypeChat,
			Precision:    "INT4",
			Description:  "Phi-3 Mini 3.8B INT4 quantized — chat",
			SizeMB:       2400,
			RamMinMB:     4096,
			ContextLen:   4096,
			DefaultModel: false,
			Aliases:      []string{"phi-3-mini-int4-ov"},
		},
		{
			ID:           "phi-3-mini-fp16",
			Name:         "Phi-3 Mini 3.8B FP16",
			HuggingFace:  "OpenVINO/phi-3-mini-fp16-ov",
			Type:         TypeChat,
			Precision:    "FP16",
			Description:  "Phi-3 Mini 3.8B FP16 — higher quality, more RAM",
			SizeMB:       7200,
			RamMinMB:     12288,
			ContextLen:   4096,
			DefaultModel: false,
		},
		{
			ID:           "llama-3.2-3b",
			Name:         "Llama 3.2 3B",
			HuggingFace:  "OpenVINO/llama-3.2-3b-int4-ov",
			Type:         TypeChat,
			Precision:    "INT4",
			Description:  "Meta Llama 3.2 3B INT4 — chat, instruction following",
			SizeMB:       1800,
			RamMinMB:     4096,
			ContextLen:   8192,
			DefaultModel: false,
		},
		{
			ID:           "llama-3.2-3b-int4",
			Name:         "Llama 3.2 3B INT4",
			HuggingFace:  "OpenVINO/llama-3.2-3b-int4-ov",
			Type:         TypeChat,
			Precision:    "INT4",
			Description:  "Meta Llama 3.2 3B INT4 — chat, instruction following",
			SizeMB:       1800,
			RamMinMB:     4096,
			ContextLen:   8192,
			DefaultModel: false,
			Aliases:      []string{"llama-3.2-3b-int4-ov"},
		},
		{
			ID:           "qwen2-0.5b",
			Name:         "Qwen2 0.5B",
			HuggingFace:  "OpenVINO/qwen2-0.5b-int4-ov",
			Type:         TypeChat,
			Precision:    "INT4",
			Description:  "Qwen2 0.5B INT4 — lightweight chat, runs on CPU",
			SizeMB:       350,
			RamMinMB:     1024,
			ContextLen:   32768,
			DefaultModel: false,
		},
		{
			ID:           "qwen2-0.5b-int4",
			Name:         "Qwen2 0.5B INT4",
			HuggingFace:  "OpenVINO/qwen2-0.5b-int4-ov",
			Type:         TypeChat,
			Precision:    "INT4",
			Description:  "Qwen2 0.5B INT4 — lightweight chat",
			SizeMB:       350,
			RamMinMB:     1024,
			ContextLen:   32768,
			DefaultModel: false,
			Aliases:      []string{"qwen2-0.5b-int4-ov"},
		},
		{
			ID:           "codegemma-2b",
			Name:         "CodeGemma 2B",
			HuggingFace:  "OpenVINO/codegemma-2b-int4-ov",
			Type:         TypeCode,
			Precision:    "INT4",
			Description:  "Google CodeGemma 2B INT4 — code completion and generation",
			SizeMB:       1300,
			RamMinMB:     3072,
			ContextLen:   8192,
			DefaultModel: false,
		},
		{
			ID:           "codegemma-2b-int4",
			Name:         "CodeGemma 2B INT4",
			HuggingFace:  "OpenVINO/codegemma-2b-int4-ov",
			Type:         TypeCode,
			Precision:    "INT4",
			Description:  "Google CodeGemma 2B INT4 — code completion",
			SizeMB:       1300,
			RamMinMB:     3072,
			ContextLen:   8192,
			DefaultModel: false,
			Aliases:      []string{"codegemma-2b-int4-ov"},
		},
		{
			ID:           "tinyllama-1.1b",
			Name:         "TinyLlama 1.1B",
			HuggingFace:  "OpenVINO/tinyllama-1.1b-int4-ov",
			Type:         TypeChat,
			Precision:    "INT4",
			Description:  "TinyLlama 1.1B INT4 — ultra lightweight, runs anywhere",
			SizeMB:       650,
			RamMinMB:     2048,
			ContextLen:   2048,
			DefaultModel: false,
		},
		{
			ID:           "tinyllama-1.1b-int4",
			Name:         "TinyLlama 1.1B INT4",
			HuggingFace:  "OpenVINO/tinyllama-1.1b-int4-ov",
			Type:         TypeChat,
			Precision:    "INT4",
			Description:  "TinyLlama 1.1B INT4 — ultra lightweight",
			SizeMB:       650,
			RamMinMB:     2048,
			ContextLen:   2048,
			DefaultModel: false,
			Aliases:      []string{"tinyllama-1.1b-int4-ov"},
		},
		{
			ID:           "bge-small-en-v1.5",
			Name:         "BGE Small EN v1.5",
			HuggingFace:  "OpenVINO/bge-small-en-v1.5-ov",
			Type:         TypeEmbed,
			Precision:    "FP16",
			Description:  "BAAI BGE Small English v1.5 — embeddings, 384 dim",
			SizeMB:       130,
			RamMinMB:     512,
			ContextLen:   512,
			DefaultModel: false,
		},
		{
			ID:           "bge-base-en-v1.5",
			Name:         "BGE Base EN v1.5",
			HuggingFace:  "OpenVINO/bge-base-en-v1.5-ov",
			Type:         TypeEmbed,
			Precision:    "FP16",
			Description:  "BAAI BGE Base English v1.5 — embeddings, 768 dim",
			SizeMB:       430,
			RamMinMB:     1024,
			ContextLen:   512,
			DefaultModel: false,
		},
		{
			ID:           "bge-reranker-v2-m3",
			Name:         "BGE Reranker v2 M3",
			HuggingFace:  "OpenVINO/bge-reranker-v2-m3-ov",
			Type:         TypeReranker,
			Precision:    "FP16",
			Description:  "BAAI BGE Reranker v2 M3 — cross-encoder reranking",
			SizeMB:       2200,
			RamMinMB:     4096,
			ContextLen:   8192,
			DefaultModel: false,
		},
	}
}
