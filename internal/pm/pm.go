package pm

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
)

// ProviderManager orchestrates auto-discovery, lifecycle, and routing
// across all local inference providers.
type ProviderManager struct {
	mu        sync.RWMutex
	providers map[ProviderType]Provider
	chain     []ProviderType
	workloads map[WorkloadType]ProviderType
	active    ProviderType
	log       *slog.Logger
}

// Config holds ProviderManager configuration.
type Config struct {
	// Chain is the ordered list of providers to try (e.g., [openvino, ollama]).
	Chain []ProviderType
	// Workloads maps workload types to a preferred provider type.
	Workloads map[WorkloadType]ProviderType
	// Active is the default active provider (empty = auto).
	Active ProviderType
}

// New creates a ProviderManager with the given providers and config.
func New(providers []Provider, cfg Config) *ProviderManager {
	pm := &ProviderManager{
		providers: make(map[ProviderType]Provider),
		chain:     cfg.Chain,
		workloads: cfg.Workloads,
		log:       slog.With("component", "pm"),
	}
	for _, p := range providers {
		pm.providers[p.Info().Type] = p
	}
	if pm.chain == nil {
		pm.chain = []ProviderType{ProviderOpenVINO, ProviderOllama, ProviderLlamaCpp, ProviderVLLM, ProviderLMStudio}
	}
	if pm.workloads == nil {
		pm.workloads = map[WorkloadType]ProviderType{}
	}
	return pm
}

// Detect scans all registered providers and returns their health status.
func (pm *ProviderManager) Detect(ctx context.Context) (map[ProviderType]*ProviderHealth, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	type result struct {
		pt     ProviderType
		health *ProviderHealth
	}

	ch := make(chan result, len(pm.providers))
	var wg sync.WaitGroup

	for pt, prov := range pm.providers {
		wg.Add(1)
		go func(t ProviderType, p Provider) {
			defer wg.Done()
			health, err := p.Status(ctx)
			if err != nil {
				health = HealthError(err)
			}
			ch <- result{t, health}
		}(pt, prov)
	}

	wg.Wait()
	close(ch)

	results := make(map[ProviderType]*ProviderHealth)
	for r := range ch {
		results[r.pt] = r.health
	}
	return results, nil
}

// AutoSelect chooses the best available provider for the given workload.
func (pm *ProviderManager) AutoSelect(ctx context.Context, workload WorkloadType) (Provider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 1. Check explicit workload mapping
	if preferred, ok := pm.workloads[workload]; ok {
		if prov, ok := pm.providers[preferred]; ok {
			health, err := prov.Status(ctx)
			if err == nil && health.Status == StatusAvailable {
				pm.active = preferred
				return prov, nil
			}
		}
	}

	// 2. Walk the chain in order
	for _, pt := range pm.chain {
		prov, ok := pm.providers[pt]
		if !ok {
			continue
		}
		health, err := prov.Status(ctx)
		if err != nil {
			pm.log.Debug("provider check failed", "type", pt, "error", err)
			continue
		}
		if health.Status == StatusAvailable {
			pm.active = pt
			return prov, nil
		}
	}

	return nil, fmt.Errorf("no provider available for workload %q", workload)
}

// ProviderByType returns a specific provider by type.
func (pm *ProviderManager) ProviderByType(pt ProviderType) (Provider, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, ok := pm.providers[pt]
	return p, ok
}

// ActiveProvider returns the currently active provider, triggering AutoSelect if none set.
func (pm *ProviderManager) ActiveProvider(ctx context.Context) (Provider, error) {
	pm.mu.RLock()
	active := pm.active
	pm.mu.RUnlock()

	if active == "" {
		return pm.AutoSelect(ctx, WorkloadChat)
	}

	pm.mu.RLock()
	prov, ok := pm.providers[active]
	pm.mu.RUnlock()
	if !ok {
		return pm.AutoSelect(ctx, WorkloadChat)
	}

	health, err := prov.Status(ctx)
	if err != nil || health.Status != StatusAvailable {
		return pm.AutoSelect(ctx, WorkloadChat)
	}

	return prov, nil
}

// ListProviders returns all registered providers sorted by chain order.
func (pm *ProviderManager) ListProviders() []Provider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	ordered := make([]Provider, 0, len(pm.providers))
	seen := make(map[ProviderType]bool)

	for _, pt := range pm.chain {
		if prov, ok := pm.providers[pt]; ok && !seen[pt] {
			ordered = append(ordered, prov)
			seen[pt] = true
		}
	}

	return ordered
}

// Installer returns the installation guide for a provider type.
func (pm *ProviderManager) Installer(pt ProviderType) (*InstallGuide, error) {
	prov, ok := pm.providers[pt]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", pt)
	}
	if !prov.Info().NeedsInstall {
		return nil, fmt.Errorf("%s is built-in, no installation needed", pt)
	}
	return InstallGuides[pt], nil
}

// SortByPriority orders providers by hardware priority for the given workload.
func SortByPriority(providers []Provider, workload WorkloadType) []Provider {
	sorted := make([]Provider, len(providers))
	copy(sorted, providers)

	sort.SliceStable(sorted, func(i, j int) bool {
		return priority(sorted[i].Info(), workload) > priority(sorted[j].Info(), workload)
	})
	return sorted
}

func priority(info ProviderInfo, workload WorkloadType) int {
	nativeScore := 0
	if info.Native {
		nativeScore = 100
	}

	supportsWorkload := 0
	for _, w := range info.SupportedWorkloads {
		if w == workload {
			supportsWorkload = 50
			break
		}
	}

	return nativeScore + supportsWorkload
}
