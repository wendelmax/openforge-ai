package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/openforge-ai/openforge/runtime"
	"github.com/spf13/cobra"
)

var benchmarkFlags struct {
	model      string
	device     string
	iterations int
	all        bool
	output     string
}

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run model benchmarks",
	Long:  `Benchmark model performance across devices and configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		provider := openvino.NewProvider(cfg.Models.Path)
		if err := provider.Initialize(ctx); err != nil {
			return err
		}
		defer provider.Shutdown(ctx)

		rt := provider.Runtime()
		device := benchmarkFlags.device
		if device == "" {
			device = cfg.Models.Device
		}

		modelID := benchmarkFlags.model
		if modelID == "" {
			return fmt.Errorf("model is required; use --model")
		}

		iter := benchmarkFlags.iterations
		if iter <= 0 {
			iter = 50
		}

		fmt.Printf("benchmarking %s on %s (%d iterations)...\n", modelID, device, iter)

		if err := rt.LoadModel(ctx, modelID, modelID, device); err != nil {
			return fmt.Errorf("load model: %w", err)
		}
		defer rt.UnloadModel(ctx, modelID)

		time.Sleep(500 * time.Millisecond)

		var latencies []float64
		prompt := "Write a short poem about programming."

		for i := 0; i < iter; i++ {
			start := time.Now()
			_, err := rt.Infer(ctx, modelID, prompt, runtime.InferenceParams{
				MaxTokens: 128,
				Temperature: 0.7,
			})
			if err != nil {
				return fmt.Errorf("inference %d: %w", i, err)
			}
			elapsed := time.Since(start)
			latencies = append(latencies, elapsed.Seconds()*1000)
		}

		result := computeStats(latencies)
		result.Model = modelID
		result.Device = device

		fmt.Printf("\nresults for %s on %s:\n", modelID, device)
		fmt.Printf("  tokens/sec:      %.1f\n", result.TokensPerSecond)
		fmt.Printf("  avg latency:     %.1f ms\n", result.LatencyP50Ms)
		fmt.Printf("  p95 latency:     %.1f ms\n", result.LatencyP95Ms)
		fmt.Printf("  p99 latency:     %.1f ms\n", result.LatencyP99Ms)

		if benchmarkFlags.output != "" {
			data, _ := json.MarshalIndent(result, "", "  ")
			os.WriteFile(benchmarkFlags.output, data, 0644)
			fmt.Printf("results written to %s\n", benchmarkFlags.output)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(benchmarkCmd)
	benchmarkCmd.Flags().StringVarP(&benchmarkFlags.model, "model", "m", "", "model to benchmark")
	benchmarkCmd.Flags().StringVarP(&benchmarkFlags.device, "device", "d", "", "target device")
	benchmarkCmd.Flags().IntVarP(&benchmarkFlags.iterations, "iterations", "n", 50, "number of iterations")
	benchmarkCmd.Flags().BoolVar(&benchmarkFlags.all, "all", false, "benchmark all models")
	benchmarkCmd.Flags().StringVarP(&benchmarkFlags.output, "output", "o", "", "output file (JSON)")
}

func computeStats(latencies []float64) runtime.BenchmarkResponse {
	n := len(latencies)
	if n == 0 {
		return runtime.BenchmarkResponse{}
	}

	var sum float64
	for _, l := range latencies {
		sum += l
	}
	avg := sum / float64(n)

	sorted := make([]float64, n)
	copy(sorted, latencies)
	sortFloats(sorted)

	p50 := sorted[n/2]
	p95 := sorted[int(float64(n)*0.95)]
	p99 := sorted[int(float64(n)*0.99)]

	tps := 1000.0 / (avg / 128.0)

	return runtime.BenchmarkResponse{
		TokensPerSecond: tps,
		TTFTMs:          sorted[0],
		LatencyP50Ms:    p50,
		LatencyP95Ms:    p95,
		LatencyP99Ms:    p99,
	}
}

func sortFloats(a []float64) {
	n := len(a)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if a[j] < a[i] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}
