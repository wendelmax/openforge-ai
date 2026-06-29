package pm

// InstallGuide contains platform-specific installation instructions for a provider.
type InstallGuide struct {
	Provider    ProviderType      `json:"provider"`
	Name        string            `json:"name"`
	Website     string            `json:"website"`
	Description string            `json:"description"`
	Platforms   []PlatformInstall `json:"platforms"`
	PostInstall []string          `json:"post_install,omitempty"`
	VerifyCmd   string            `json:"verify_cmd,omitempty"`
}

// PlatformInstall has installation commands for one OS.
type PlatformInstall struct {
	OS      string   `json:"os"`
	Install []string `json:"install"`
	Verify  []string `json:"verify"`
	Notes   string   `json:"notes,omitempty"`
}

// InstallGuides maps provider types to their installation instructions.
var InstallGuides = map[ProviderType]*InstallGuide{
	ProviderOllama: {
		Provider:    ProviderOllama,
		Name:        "Ollama",
		Website:     "https://ollama.com",
		Description: "Run Llama, Mistral, Qwen, DeepSeek, and hundreds of models locally. GGUF-based, GPU acceleration.",
		Platforms: []PlatformInstall{
			{OS: "linux", Install: []string{"curl -fsSL https://ollama.com/install.sh | sh"}, Verify: []string{"ollama --version", "curl -s http://localhost:11434/api/tags"}},
			{OS: "darwin", Install: []string{"brew install ollama"}, Verify: []string{"ollama --version"}},
			{OS: "windows", Install: []string{"winget install Ollama.Ollama"}, Verify: []string{"ollama --version"}, Notes: "Or download from https://ollama.com/download/windows"},
		},
		PostInstall: []string{"Start: ollama serve", "Pull model: ollama pull llama3.2:3b"},
		VerifyCmd:   "curl -s http://localhost:11434/api/tags",
	},
	ProviderLlamaCpp: {
		Provider:    ProviderLlamaCpp,
		Name:        "llama.cpp",
		Website:     "https://github.com/ggerganov/llama.cpp",
		Description: "Lightweight C++ inference for GGUF models. CPU-first with GPU acceleration.",
		Platforms: []PlatformInstall{
			{OS: "linux", Install: []string{"brew install llama.cpp"}, Verify: []string{"llama-server --version 2>&1 || echo 'built from source'"}},
			{OS: "darwin", Install: []string{"brew install llama.cpp"}, Verify: []string{"llama-server --version 2>&1 || echo 'built from source'"}},
			{OS: "windows", Install: []string{"# Download from https://github.com/ggerganov/llama.cpp/releases"}, Verify: []string{"llama-server.exe --version 2>&1 || echo 'check build/bin/Release/'"}, Notes: "Pre-built binaries on GitHub Releases"},
		},
		PostInstall: []string{"Start: llama-server -m model.gguf --port 8080"},
		VerifyCmd:   "curl -s http://localhost:8080/v1/models",
	},
	ProviderVLLM: {
		Provider:    ProviderVLLM,
		Name:        "vLLM",
		Website:     "https://github.com/vllm-project/vllm",
		Description: "High-performance LLM serving with PagedAttention. Best for multi-GPU.",
		Platforms: []PlatformInstall{
			{OS: "linux", Install: []string{"pip install vllm"}, Verify: []string{"python -c 'import vllm; print(vllm.__version__)'"}, Notes: "Requires NVIDIA GPU + CUDA 11.8+"},
			{OS: "darwin", Install: []string{"pip install vllm"}, Verify: []string{"python -c 'import vllm; print(vllm.__version__)'"}, Notes: "Apple Silicon experimental"},
			{OS: "windows", Install: []string{"pip install vllm"}, Verify: []string{"python -c 'import vllm; print(vllm.__version__)'"}, Notes: "Requires CUDA toolkit"},
		},
		PostInstall: []string{"Start: vllm serve meta-llama/Llama-3.2-3B-Instruct --port 8000"},
		VerifyCmd:   "curl -s http://localhost:8000/v1/models",
	},
	ProviderLMStudio: {
		Provider:    ProviderLMStudio,
		Name:        "LM Studio",
		Website:     "https://lmstudio.ai",
		Description: "User-friendly desktop app for running local models. Built-in model browser.",
		Platforms: []PlatformInstall{
			{OS: "linux", Install: []string{"# Download .AppImage from https://lmstudio.ai"}, Verify: []string{"curl -s http://localhost:1234/v1/models"}},
			{OS: "darwin", Install: []string{"brew install --cask lm-studio"}, Verify: []string{"curl -s http://localhost:1234/v1/models"}},
			{OS: "windows", Install: []string{"winget install LMStudio.LMStudio"}, Verify: []string{"curl -s http://localhost:1234/v1/models"}},
		},
		PostInstall: []string{"1. Open LM Studio", "2. Download model", "3. Start Server from Local Server tab"},
		VerifyCmd:   "curl -s http://localhost:1234/v1/models",
	},
	ProviderOpenVINO: {
		Provider:    ProviderOpenVINO,
		Name:        "OpenVINO",
		Website:     "https://docs.openvino.ai",
		Description: "Intel's native inference runtime. Best for CPU, Intel GPU, and Intel NPU.",
		Platforms: []PlatformInstall{
			{OS: "linux", Install: []string{"pip install openvino openvino-genai"}, Verify: []string{"python -c 'import openvino; print(openvino.__version__)'"}, Notes: "Intel GPU: sudo apt install intel-opencl-icd"},
			{OS: "windows", Install: []string{"pip install openvino openvino-genai"}, Verify: []string{"python -c \"import openvino; print(openvino.__version__)\""}, Notes: "NPU drivers: https://intel.com/npu-driver"},
			{OS: "darwin", Install: []string{"pip install openvino openvino-genai"}, Verify: []string{"python -c 'import openvino; print(openvino.__version__)'"}, Notes: "Intel HW only"},
		},
		PostInstall: []string{"OpenVINO is built into OpenForge — no server process needed", "Convert models: optimum-cli export openvino --model <id> <dir>"},
		VerifyCmd:   "python -c 'import openvino; print(openvino.__version__)'",
	},
}
