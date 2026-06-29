package server

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/engine"
	"github.com/openforge-ai/openforge/runtime"
)

//go:embed openapi.json
var openAPISpec []byte

// Server represents the HTTP API server for OpenForge.
type Server struct {
	engine   *engine.Engine
	config   *config.Config
	gin      *gin.Engine
	http     *http.Server
	startedAt time.Time
}

// New creates a new Server with the given engine and config.
func New(eng *engine.Engine, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		engine:    eng,
		config:    cfg,
		gin:       gin.New(),
		startedAt: time.Now(),
	}

	s.gin.Use(s.requestIDMiddleware())
	s.gin.Use(s.loggingMiddleware())
	s.gin.Use(gin.Recovery())

	s.gin.GET("/health", s.handleHealth)
	s.gin.GET("/v1/health", s.handleHealth)
	s.gin.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", openAPISpec)
	})

	v1 := s.gin.Group("/v1")
	{
		v1.POST("/chat", s.handleChat)
		v1.POST("/completion", s.handleCompletion)
		v1.POST("/embeddings", s.handleEmbeddings)
		v1.POST("/rerank", s.handleRerank)
		v1.POST("/model/load", s.handleModelLoad)
		v1.POST("/model/unload", s.handleModelUnload)
		v1.GET("/models", s.handleListModels)
		v1.GET("/devices", s.handleListDevices)
		v1.POST("/benchmark", s.handleBenchmark)
	}

	return s
}

// Start begins listening on addr and serving HTTP requests.
func (s *Server) Start(addr string) error {
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.gin,
	}
	return s.http.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server with a timeout context.
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.http.Shutdown(shutdownCtx)
}

func (s *Server) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set("request_id", rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		rid, _ := c.Get("request_id")
		slog.Debug("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
			"request_id", rid,
		)
	}
}

func (s *Server) requestID(c *gin.Context) string {
	rid, _ := c.Get("request_id")
	return rid.(string)
}

func (s *Server) respondError(c *gin.Context, status int, code, message string) {
	rid := s.requestID(c)
	slog.Warn("api error", "code", code, "message", message, "request_id", rid)
	c.JSON(status, runtime.ErrorResponse{
		Error: runtime.APIError{
			Code:      code,
			Message:   message,
			RequestID: rid,
		},
	})
}

func (s *Server) bindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		s.respondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return false
	}
	return true
}

func (s *Server) handleHealth(c *gin.Context) {
	models, _ := s.engine.Runtime().ListModels(c.Request.Context())
	loaded := 0
	for _, m := range models {
		if m.Loaded {
			loaded++
		}
	}
	devices, _ := s.engine.Runtime().ListDevices(c.Request.Context())
	activeDevice := "unknown"
	for _, d := range devices {
		if d.Available {
			activeDevice = d.ID
			break
		}
	}
	c.JSON(http.StatusOK, runtime.HealthResponse{
		Status:       "ok",
		Version:      "0.1.0-dev",
		Uptime:       time.Since(s.startedAt).Round(time.Second).String(),
		ModelsLoaded: loaded,
		ActiveDevice: activeDevice,
	})
}

func (s *Server) handleChat(c *gin.Context) {
	var req runtime.ChatRequest
	if !s.bindJSON(c, &req) {
		return
	}
	if req.Model == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}
	if len(req.Messages) == 0 {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "messages is required")
		return
	}

	params := runtime.InferenceParams{
		Device:      s.engine.Runtime().DeviceForWorkload("chat", req.Device),
		MaxTokens:   req.MaxTokens,
		Temperature: 0.7,
		TopP:        0.9,
	}
	if req.Temperature != nil {
		params.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		params.TopP = *req.TopP
	}
	if req.TopK != nil {
		params.TopK = *req.TopK
	}

	var prompt string
	for _, msg := range req.Messages {
		prompt += msg.Role + ": " + msg.Content + "\n"
	}

	if req.Stream {
		s.handleChatStream(c, req.Model, prompt, params)
		return
	}

	start := time.Now()
	result, err := s.engine.Runtime().Infer(c.Request.Context(), req.Model, prompt, params)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "inference_error", err.Error())
		return
	}
	elapsed := time.Since(start)

	resp := runtime.ChatResponse{
		ID:      "chatcmpl-" + uuid.New().String(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []runtime.Choice{
			{
				Index: 0,
				Message: &runtime.Message{
					Role:    "assistant",
					Content: result.Text,
				},
				FinishReason: strPtr("stop"),
			},
		},
		Usage: NewUsage(len(prompt), len(result.Tokens)),
		Timing: &runtime.Timing{
			Total: elapsed.Round(time.Millisecond).String(),
		},
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleChatStream(c *gin.Context, modelID, prompt string, params runtime.InferenceParams) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ch, err := s.engine.Runtime().InferStream(c.Request.Context(), modelID, prompt, params)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "inference_error", err.Error())
		return
	}

	flusher, flushOk := c.Writer.(http.Flusher)
	write := func(format string, args ...interface{}) {
		fmt.Fprintf(c.Writer, format, args...)
		if flushOk {
			flusher.Flush()
		}
	}

	for {
		select {
		case token, ok := <-ch:
			if !ok {
				write("data: [DONE]\n\n")
				return
			}
			payload := map[string]interface{}{
				"id":      "chatcmpl-" + uuid.New().String(),
				"object":  "chat.completion.chunk",
				"created": time.Now().Unix(),
				"model":   modelID,
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"delta": map[string]string{"content": token},
						"finish_reason": nil,
					},
				},
			}
			data, _ := json.Marshal(payload)
			write("data: %s\n\n", data)
		case <-c.Request.Context().Done():
			write("data: [DONE]\n\n")
			return
		}
	}
}

func (s *Server) handleCompletion(c *gin.Context) {
	var req runtime.CompletionRequest
	if !s.bindJSON(c, &req) {
		return
	}
	if req.Model == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}
	if req.Prompt == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "prompt is required")
		return
	}

	params := runtime.InferenceParams{
		Device:      s.engine.Runtime().DeviceForWorkload("completion", req.Device),
		MaxTokens:   req.MaxTokens,
		Temperature: 0.7,
		TopP:        0.9,
	}
	if req.Temperature != nil {
		params.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		params.TopP = *req.TopP
	}
	if req.TopK != nil {
		params.TopK = *req.TopK
	}

	start := time.Now()
	result, err := s.engine.Runtime().Infer(c.Request.Context(), req.Model, req.Prompt, params)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "inference_error", err.Error())
		return
	}
	elapsed := time.Since(start)

	resp := runtime.CompletionResponse{
		ID:      "cmpl-" + uuid.New().String(),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []runtime.CompletionChoice{
			{
				Index: 0,
				Text:  result.Text,
				FinishReason: strPtr("stop"),
			},
		},
		Usage: NewUsage(len(req.Prompt), len(result.Tokens)),
		Timing: &runtime.Timing{
			Total: elapsed.Round(time.Millisecond).String(),
		},
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleEmbeddings(c *gin.Context) {
	var req runtime.EmbeddingRequest
	if !s.bindJSON(c, &req) {
		return
	}
	if req.Model == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}
	if len(req.Input) == 0 {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "input is required")
		return
	}

	device := s.engine.Runtime().DeviceForWorkload("embedding", req.Device)
	result, err := s.engine.Runtime().Embed(c.Request.Context(), req.Model, req.Input, device)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "inference_error", err.Error())
		return
	}

	data := make([]runtime.EmbeddingData, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		data[i] = runtime.EmbeddingData{
			Object:    "embedding",
			Index:     i,
			Embedding: emb,
		}
	}

	resp := runtime.EmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  req.Model,
		Usage:  NewUsage(len(req.Input), 0),
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleRerank(c *gin.Context) {
	var req runtime.RerankRequest
	if !s.bindJSON(c, &req) {
		return
	}
	if req.Model == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}
	if req.Query == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "query is required")
		return
	}
	if len(req.Documents) == 0 {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "documents is required")
		return
	}

	ctx := c.Request.Context()
	allInputs := append([]string{req.Query}, req.Documents...)

	device := s.engine.Runtime().DeviceForWorkload("rerank", req.Device)
	embResult, err := s.engine.Runtime().Embed(ctx, req.Model, allInputs, device)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "rerank_error", err.Error())
		return
	}
	if len(embResult.Embeddings) < 2 {
		s.respondError(c, http.StatusInternalServerError, "rerank_error", "failed to generate embeddings")
		return
	}

	queryEmb := embResult.Embeddings[0]
	type scoredDoc struct {
		index int
		score float32
		doc   string
	}
	scored := make([]scoredDoc, len(req.Documents))
	for i, docEmb := range embResult.Embeddings[1:] {
		scored[i] = scoredDoc{
			index: i,
			score: cosineSimilarity(queryEmb, docEmb),
			doc:   req.Documents[i],
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	topN := req.TopN
	if topN <= 0 || topN > len(scored) {
		topN = len(scored)
	}

	results := make([]runtime.RerankResult, topN)
	for i := 0; i < topN; i++ {
		results[i] = runtime.RerankResult{
			Index:    scored[i].index,
			Score:    scored[i].score,
			Document: scored[i].doc,
		}
	}

	c.JSON(http.StatusOK, runtime.RerankResponse{
		Object:  "rerank",
		Model:   req.Model,
		Results: results,
	})
}

func (s *Server) handleModelLoad(c *gin.Context) {
	var req runtime.ModelLoadRequest
	if !s.bindJSON(c, &req) {
		return
	}
	if req.ModelID == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "model_id is required")
		return
	}
	device := req.Device
	if device == "" {
		device = s.config.Models.Device
	}

	start := time.Now()
	err := s.engine.Runtime().LoadModel(c.Request.Context(), req.ModelID, req.ModelID, device)
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "load_error", err.Error())
		return
	}
	elapsed := time.Since(start)

	c.JSON(http.StatusOK, runtime.ModelStatus{
		Object: "model",
		ID:     req.ModelID,
		Status: "loaded",
		Device: device,
		Timing: &runtime.Timing{
			LoadDuration: elapsed.Round(time.Millisecond).String(),
		},
	})
}

func (s *Server) handleModelUnload(c *gin.Context) {
	var req runtime.ModelUnloadRequest
	if !s.bindJSON(c, &req) {
		return
	}
	if req.ModelID == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "model_id is required")
		return
	}

	if err := s.engine.Runtime().UnloadModel(c.Request.Context(), req.ModelID); err != nil {
		s.respondError(c, http.StatusInternalServerError, "unload_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, runtime.ModelStatus{
		Object: "model",
		ID:     req.ModelID,
		Status: "unloaded",
	})
}

func (s *Server) handleListModels(c *gin.Context) {
	models, err := s.engine.Runtime().ListModels(c.Request.Context())
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "list_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, runtime.ModelListResponse{
		Object: "list",
		Data:   models,
	})
}

func (s *Server) handleListDevices(c *gin.Context) {
	devices, err := s.engine.Runtime().ListDevices(c.Request.Context())
	if err != nil {
		s.respondError(c, http.StatusInternalServerError, "device_error", err.Error())
		return
	}
	c.JSON(http.StatusOK, runtime.DeviceListResponse{
		Object: "list",
		Data:   devices,
	})
}

func (s *Server) handleBenchmark(c *gin.Context) {
	var req runtime.BenchmarkRequest
	if !s.bindJSON(c, &req) {
		return
	}
	if req.Model == "" {
		s.respondError(c, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}

	ctx := c.Request.Context()
	device := req.Device
	if device == "" {
		device = s.config.Models.Device
	}

	iterations := req.Iterations
	if iterations <= 0 {
		iterations = 10
	}

	models, _ := s.engine.Runtime().ListModels(ctx)
	loaded := false
	for _, m := range models {
		if m.ID == req.Model && m.Loaded {
			loaded = true
			break
		}
	}
	if !loaded {
		if err := s.engine.Runtime().LoadModel(ctx, req.Model, req.Model, device); err != nil {
			s.respondError(c, http.StatusInternalServerError, "load_error", err.Error())
			return
		}
		defer s.engine.Runtime().UnloadModel(ctx, req.Model)
	}

	prompt := "Benchmarking prompt: write a short paragraph about artificial intelligence."
	params := runtime.InferenceParams{
		MaxTokens:   50,
		Temperature: 0.0,
	}

	latencies := make([]time.Duration, 0, iterations)
	var totalTokens int
	var ttft time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		result, err := s.engine.Runtime().Infer(ctx, req.Model, prompt, params)
		elapsed := time.Since(start)
		if err != nil {
			s.respondError(c, http.StatusInternalServerError, "benchmark_error", err.Error())
			return
		}
		if i == 0 {
			ttft = elapsed
		}
		latencies = append(latencies, elapsed)
		totalTokens += len(result.Tokens)
	}

	totalDuration := time.Duration(0)
	for _, d := range latencies {
		totalDuration += d
	}
	var tokensPerSec float64
	if totalTokens > 0 && totalDuration > 0 {
		tokensPerSec = float64(totalTokens) / totalDuration.Seconds()
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	p50 := latencies[int(float64(iterations)*0.5)].Seconds() * 1000
	p95 := latencies[int(float64(iterations)*0.95)].Seconds() * 1000
	p99 := latencies[int(float64(iterations)*0.99)].Seconds() * 1000

	c.JSON(http.StatusOK, runtime.BenchmarkResponse{
		Model:           req.Model,
		Device:          device,
		TokensPerSecond: tokensPerSec,
		TTFTMs:          ttft.Seconds() * 1000,
		LatencyP50Ms:    p50,
		LatencyP95Ms:    p95,
		LatencyP99Ms:    p99,
		MemoryMB:        0,
	})
}

func (s *Server) renderJSON(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func strPtr(s string) *string {
	return &s
}

// NewUsage creates a Usage struct with prompt, completion, and total token counts.
// Deprecated: use runtime.Usage directly instead.
func NewUsage(promptTokens, completionTokens int) *runtime.Usage {
	return &runtime.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return float32(dot / denom)
}
