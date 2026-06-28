//go:build cgo

package openvino

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/openforge-ai/openforge/runtime"
)

func (r *OpenVINORuntime) Benchmark(ctx context.Context, modelID string, iterations int, prompt string, maxTokens int) (runtime.BenchmarkResults, error) {
	r.mu.RLock()
	lm, ok := r.models[modelID]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("model %q not loaded", modelID)
	}

	results := make(runtime.BenchmarkResults)
	for device, h := range lm.compiledByDevice {
		_ = h
		slog.Info("benchmarking device", "device", device, "iterations", iterations)

		inputIDs := []int64{1, 2, 3, 4, 5}
		var totalDuration time.Duration
		successfulIterations := 0

		for i := 0; i < iterations; i++ {
			params := runtime.InferenceParams{
				Device:      device,
				MaxTokens:   maxTokens,
				Temperature: 0.1,
			}
			start := time.Now()
			_, err := r.generate(ctx, lm, inputIDs, params)
			if err != nil {
				slog.Warn("benchmark iteration failed", "device", device, "error", err)
				continue
			}
			totalDuration += time.Since(start)
			successfulIterations++
		}

		if successfulIterations == 0 {
			slog.Warn("no successful benchmark iterations", "device", device)
			continue
		}

		avgDuration := totalDuration / time.Duration(successfulIterations)
		tokensPerSec := float64(maxTokens) / avgDuration.Seconds()

		results[device] = runtime.DeviceBenchmark{
			ChatTokensPerSec: tokensPerSec,
			EmbedLatencyMS:   0,
		}
	}

	return results, nil
}
