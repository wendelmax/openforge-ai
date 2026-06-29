package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pm "github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/runtime"
)

// OpenAICompatProvider wraps any OpenAI-compatible HTTP API endpoint.
// Works with vLLM, llama.cpp server, LM Studio, and cloud providers.
type OpenAICompatProvider struct {
	baseURL string
	client  *http.Client
	info    pm.ProviderInfo
	apiKey  string
}

func NewOpenAICompatProvider(baseURL string, info pm.ProviderInfo, apiKey string) *OpenAICompatProvider {
	return &OpenAICompatProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{MaxIdleConns: 4, IdleConnTimeout: 30 * time.Second},
		},
		info:   info,
		apiKey: apiKey,
	}
}

func (p *OpenAICompatProvider) Info() pm.ProviderInfo { return p.info }

func (p *OpenAICompatProvider) Status(ctx context.Context) (*pm.ProviderHealth, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/v1/models", nil)
	if err != nil {
		return pm.HealthError(err), nil
	}
	p.setAuth(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return pm.HealthUnavailable(err.Error()), nil
	}
	defer resp.Body.Close()
	latency := time.Since(start)
	if resp.StatusCode != http.StatusOK {
		h := pm.HealthError(fmt.Errorf("HTTP %d", resp.StatusCode))
		h.Latency = latency
		return h, nil
	}
	h := pm.HealthAvailable()
	h.Latency = latency
	return h, nil
}

// --- JSON helpers ---

type oaiMsg struct{ Role, Content string }

type oaiChatReq struct {
	Model       string   `json:"model"`
	Messages    []oaiMsg `json:"messages"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float32  `json:"temperature,omitempty"`
	Stream      bool     `json:"stream"`
}

type oaiChatResp struct {
	Choices []struct{ Message struct{ Content string } `json:"message"` } `json:"choices"`
	Usage   *oaiUsage `json:"usage"`
	Model   string    `json:"model"`
}

type oaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type oaiStreamChunk struct {
	Choices []struct {
		Delta        struct{ Content string } `json:"delta"`
		FinishReason *string                  `json:"finish_reason"`
	} `json:"choices"`
	Model string `json:"model"`
}

type oaiEmbedReq struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type oaiEmbedResp struct {
	Data  []struct{ Embedding []float32 `json:"embedding"` } `json:"data"`
	Model string  `json:"model"`
	Usage *oaiUsage `json:"usage"`
}

type oaiModelList struct {
	Data []struct{ ID string `json:"id"` } `json:"data"`
}

// --- Chat ---

func (p *OpenAICompatProvider) Chat(ctx context.Context, req *pm.ChatRequest) (*pm.ChatResponse, error) {
	msgs := make([]oaiMsg, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = oaiMsg{Role: m.Role, Content: m.Content}
	}
	body := oaiChatReq{Model: req.Model, Messages: msgs, MaxTokens: req.MaxTokens, Temperature: req.Temperature}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/chat/completions", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	p.setAuth(httpReq)
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai chat: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai chat: HTTP %d: %s", resp.StatusCode, string(b))
	}
	var oResp oaiChatResp
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return nil, fmt.Errorf("openai chat decode: %w", err)
	}
	content := ""
	if len(oResp.Choices) > 0 {
		content = oResp.Choices[0].Message.Content
	}
	cr := &pm.ChatResponse{Model: oResp.Model, Content: content, Provider: p.info.Type}
	if oResp.Usage != nil {
		cr.Usage = &runtime.Usage{
			PromptTokens: oResp.Usage.PromptTokens, CompletionTokens: oResp.Usage.CompletionTokens,
			TotalTokens: oResp.Usage.TotalTokens,
		}
	}
	return cr, nil
}

// --- ChatStream ---

func (p *OpenAICompatProvider) ChatStream(ctx context.Context, req *pm.ChatRequest) (<-chan pm.Token, error) {
	msgs := make([]oaiMsg, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = oaiMsg{Role: m.Role, Content: m.Content}
	}
	body := oaiChatReq{Model: req.Model, Messages: msgs, MaxTokens: req.MaxTokens, Temperature: req.Temperature, Stream: true}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/chat/completions", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	p.setAuth(httpReq)
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai stream: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai stream: HTTP %d: %s", resp.StatusCode, string(b))
	}
	ch := make(chan pm.Token, 64)
	go func() {
		defer resp.Body.Close()
		defer close(ch)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}
			var chunk oaiStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				ch <- pm.Token{Content: fmt.Sprintf("\n[stream error: %v]", err), Done: true}
				return
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			choice := chunk.Choices[0]
			done := choice.FinishReason != nil && *choice.FinishReason != ""
			ch <- pm.Token{Content: choice.Delta.Content, Model: chunk.Model, Done: done}
			if done {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- pm.Token{Content: fmt.Sprintf("\n[stream error: %v]", err), Done: true}
		}
	}()
	return ch, nil
}

// --- Embed ---

func (p *OpenAICompatProvider) Embed(ctx context.Context, req *pm.EmbedRequest) (*pm.EmbedResponse, error) {
	body := oaiEmbedReq{Model: req.Model, Input: req.Inputs}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/embeddings", bytes.NewReader(payload))
	httpReq.Header.Set("Content-Type", "application/json")
	p.setAuth(httpReq)
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai embed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai embed: HTTP %d: %s", resp.StatusCode, string(b))
	}
	var oResp oaiEmbedResp
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return nil, fmt.Errorf("openai embed decode: %w", err)
	}
	embeddings := make([][]float32, len(oResp.Data))
	for i, d := range oResp.Data {
		embeddings[i] = d.Embedding
	}
	er := &pm.EmbedResponse{Model: oResp.Model, Embeddings: embeddings, Provider: p.info.Type}
	if oResp.Usage != nil {
		er.Usage = &runtime.Usage{
			PromptTokens: oResp.Usage.PromptTokens, CompletionTokens: oResp.Usage.CompletionTokens,
			TotalTokens: oResp.Usage.TotalTokens,
		}
	}
	return er, nil
}

// --- ListModels ---

func (p *OpenAICompatProvider) ListModels(ctx context.Context) ([]pm.Model, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/v1/models", nil)
	p.setAuth(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai list models: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai list models: HTTP %d: %s", resp.StatusCode, string(b))
	}
	var list oaiModelList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("openai list models decode: %w", err)
	}
	models := make([]pm.Model, len(list.Data))
	for i, m := range list.Data {
		models[i] = pm.Model{ID: m.ID, Name: m.ID, Provider: p.info.Type}
	}
	return models, nil
}

// --- No-ops ---

func (p *OpenAICompatProvider) LoadModel(ctx context.Context, _ string) error   { return nil }
func (p *OpenAICompatProvider) UnloadModel(ctx context.Context, _ string) error { return nil }
func (p *OpenAICompatProvider) Start(ctx context.Context) error                  { return nil }
func (p *OpenAICompatProvider) Stop(ctx context.Context) error                   { return nil }

func (p *OpenAICompatProvider) setAuth(req *http.Request) {
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
}
