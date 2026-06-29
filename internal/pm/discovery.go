package pm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DiscoveryResult summarizes auto-detection for a single provider.
type DiscoveryResult struct {
	Provider    ProviderType    `json:"provider"`
	Installed   bool            `json:"installed"`
	Running     bool            `json:"running"`
	Health      *ProviderHealth `json:"health,omitempty"`
	Endpoint    string          `json:"endpoint,omitempty"`
	InstallHint string          `json:"install_hint,omitempty"`
}

// Discover scans the local system for installed and running inference providers.
func Discover(ctx context.Context) ([]DiscoveryResult, error) {
	results := make([]DiscoveryResult, 0, 5)
	results = append(results, detectOpenVINO())
	results = append(results, detectOllama(ctx))
	results = append(results, detectLlamaCpp(ctx))
	results = append(results, detectVLLM(ctx))
	results = append(results, detectLMStudio(ctx))
	return results, nil
}

func detectOpenVINO() DiscoveryResult {
	r := DiscoveryResult{Provider: ProviderOpenVINO}
	libName := openvinoLibName()
	if libName == "" {
		r.InstallHint = installHint(ProviderOpenVINO)
		return r
	}
	for _, dir := range openvinoSearchPaths() {
		if fileExists(filepath.Join(dir, libName)) {
			r.Installed = true
			r.Running = true
			r.Health = &ProviderHealth{Status: StatusAvailable}
			return r
		}
	}
	for _, env := range []string{"INTEL_OPENVINO_DIR", "OPENVINO_HOME"} {
		if v := os.Getenv(env); v != "" {
			if fileExists(filepath.Join(v, libName)) {
				r.Installed = true
				r.Running = true
				r.Health = &ProviderHealth{Status: StatusAvailable}
				return r
			}
			r.Installed = true
			r.InstallHint = "OpenVINO environment set but shared library not found. Reinstall OpenVINO."
			return r
		}
	}
	r.InstallHint = installHint(ProviderOpenVINO)
	return r
}

func openvinoLibName() string {
	switch runtime.GOOS {
	case "windows":
		return "openvino_c.dll"
	case "darwin":
		return "libopenvino_c.dylib"
	default:
		return "libopenvino_c.so"
	}
}

func openvinoSearchPaths() []string {
	switch runtime.GOOS {
	case "windows":
		var out []string
		for _, d := range []string{"C:", "D:"} {
			out = append(out,
				filepath.Join(d, "Program Files (x86)", "Intel", "openvino_2025", "runtime", "bin", "intel64", "Release"),
				filepath.Join(d, "Program Files", "Intel", "openvino_2025", "runtime", "bin", "intel64", "Release"),
				filepath.Join(d, "Program Files (x86)", "Intel", "openvino_2024", "runtime", "bin", "intel64", "Release"),
			)
		}
		for _, p := range filepath.SplitList(os.Getenv("PATH")) {
			out = append(out, p)
		}
		return out
	case "darwin":
		return []string{"/opt/intel/openvino_2025/runtime/lib/intel64", "/usr/local/lib", "/usr/lib"}
	default:
		return []string{
			"/opt/intel/openvino_2025/runtime/lib/intel64",
			"/opt/intel/openvino_2024/runtime/lib/intel64",
			"/usr/lib", "/usr/local/lib", "/usr/lib/x86_64-linux-gnu",
		}
	}
}

func detectOllama(ctx context.Context) DiscoveryResult {
	r := DiscoveryResult{Provider: ProviderOllama}
	r.Installed = binaryOnPath("ollama")
	health, err := httpHealthCheck(ctx, "http://localhost:11434/api/tags", 2*time.Second, parseOllamaTags)
	if err != nil {
		r.InstallHint = installHint(ProviderOllama)
		return r
	}
	r.Running = true
	r.Health = health
	r.Endpoint = "http://localhost:11434"
	return r
}

func detectLlamaCpp(ctx context.Context) DiscoveryResult {
	r := DiscoveryResult{Provider: ProviderLlamaCpp}
	r.Installed = binaryOnPath("llama-server") || binaryOnPath("llama-server.exe")
	for _, port := range []int{8080, 8081, 8082} {
		health, err := tryHTTPEndpoint(ctx, port)
		if err == nil {
			r.Running = true
			r.Health = health
			r.Endpoint = fmt.Sprintf("http://localhost:%d", port)
			return r
		}
	}
	r.InstallHint = installHint(ProviderLlamaCpp)
	return r
}

func detectVLLM(ctx context.Context) DiscoveryResult {
	r := DiscoveryResult{Provider: ProviderVLLM}
	hasPy := binaryOnPath("python") || binaryOnPath("python3")
	r.Installed = binaryOnPath("vllm") || (hasPy && pipHasVLLM())
	for _, port := range []int{8000, 8001} {
		health, err := tryHTTPEndpoint(ctx, port)
		if err == nil {
			r.Running = true
			r.Health = health
			r.Endpoint = fmt.Sprintf("http://localhost:%d", port)
			return r
		}
	}
	r.InstallHint = installHint(ProviderVLLM)
	return r
}

func detectLMStudio(ctx context.Context) DiscoveryResult {
	r := DiscoveryResult{Provider: ProviderLMStudio}
	r.Installed = binaryOnPath("lm-studio") || binaryOnPath("LM Studio.exe") || dirExists(lmStudioDir())
	health, err := httpHealthCheck(ctx, "http://localhost:1234/v1/models", 2*time.Second, parseOpenAIModels)
	if err != nil {
		r.InstallHint = installHint(ProviderLMStudio)
		return r
	}
	r.Running = true
	r.Health = health
	r.Endpoint = "http://localhost:1234"
	return r
}

func lmStudioDir() string {
	switch runtime.GOOS {
	case "windows":
		if d := os.Getenv("LOCALAPPDATA"); d != "" {
			return filepath.Join(d, "LM-Studio")
		}
	case "darwin":
		return "/Applications/LM Studio.app"
	default:
		if home, _ := os.UserHomeDir(); home != "" {
			return filepath.Join(home, ".lmstudio")
		}
	}
	return ""
}

func tryHTTPEndpoint(ctx context.Context, port int) (*ProviderHealth, error) {
	url := fmt.Sprintf("http://localhost:%d/health", port)
	health, err := httpHealthCheck(ctx, url, 2*time.Second, nil)
	if err == nil {
		return health, nil
	}
	url = fmt.Sprintf("http://localhost:%d/v1/models", port)
	return httpHealthCheck(ctx, url, 2*time.Second, parseOpenAIModels)
}

func httpHealthCheck(ctx context.Context, url string, timeout time.Duration, parse func([]byte) int) (*ProviderHealth, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	latency := time.Since(start)
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	h := &ProviderHealth{Status: StatusAvailable, Latency: latency}
	if parse != nil {
		buf := make([]byte, 16384)
		n, _ := resp.Body.Read(buf)
		if n > 0 {
			h.Models = parse(buf[:n])
		}
	}
	return h, nil
}

func parseOllamaTags(body []byte) int {
	var v struct {
		Models []struct{ Name string `json:"name"` } `json:"models"`
	}
	if json.Unmarshal(body, &v) == nil {
		return len(v.Models)
	}
	return 0
}

func parseOpenAIModels(body []byte) int {
	var v struct {
		Data []struct{ ID string `json:"id"` } `json:"data"`
	}
	if json.Unmarshal(body, &v) == nil {
		return len(v.Data)
	}
	return 0
}

func binaryOnPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func pipHasVLLM() bool {
	for _, pattern := range pythonSitePackages() {
		matches, err := filepath.Glob(filepath.Join(pattern, "vllm"))
		if err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

func pythonSitePackages() []string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		localAppData := os.Getenv("LOCALAPPDATA")
		return []string{
			filepath.Join(appData, "Python", "Python312", "Lib", "site-packages"),
			filepath.Join(localAppData, "Programs", "Python", "Python312", "Lib", "site-packages"),
		}
	default:
		home, _ := os.UserHomeDir()
		return []string{
			"/usr/lib/python3.12/site-packages",
			"/usr/local/lib/python3.12/site-packages",
			filepath.Join(home, ".local", "lib", "python3.12", "site-packages"),
		}
	}
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func installHint(pt ProviderType) string {
	guide, ok := InstallGuides[pt]
	if !ok {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(guide.Description)
	sb.WriteString(" | ")
	sb.WriteString(guide.Website)
	for _, p := range guide.Platforms {
		if strings.EqualFold(p.OS, runtime.GOOS) {
			sb.WriteString(" | Install: ")
			sb.WriteString(strings.Join(p.Install, "; "))
			break
		}
	}
	return sb.String()
}
