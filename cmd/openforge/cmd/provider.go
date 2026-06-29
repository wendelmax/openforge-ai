package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/openforge-ai/openforge/internal/pm"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage local inference providers (OpenVINO, Ollama, llama.cpp, vLLM, LM Studio)",
	Long: `Auto-detect, install, and manage local LLM inference runtimes.

Commands:
  openforge provider list      List all detected providers with status
  openforge provider detect    Scan hardware and detect installed runtimes
  openforge provider install   Install a provider (interactive)
  openforge provider guide     Show platform-specific installation guide
  openforge provider start     Start a provider runtime
  openforge provider stop      Stop a provider runtime
  openforge provider info      Show detailed info about a provider`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProviderList()
	},
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all detected providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProviderList()
	},
}

var providerDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Scan hardware and detect installed runtimes",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProviderDetect()
	},
}

var providerInstallCmd = &cobra.Command{
	Use:   "install [provider]",
	Short: "Install a provider runtime",
	Long: `Install a local inference runtime.

Examples:
  openforge provider install ollama
  openforge provider install --auto`,
	RunE: func(cmd *cobra.Command, args []string) error {
		auto, _ := cmd.Flags().GetBool("auto")
		return runProviderInstall(args, auto)
	},
}

var providerGuideCmd = &cobra.Command{
	Use:   "guide [provider]",
	Short: "Show installation guide for a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProviderGuide(args)
	},
}

var providerInfoCmd = &cobra.Command{
	Use:   "info [provider]",
	Short: "Show detailed info about a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProviderInfo(args)
	},
}

var providerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check health of all providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProviderStatus()
	},
}

var providerStartCmd = &cobra.Command{
	Use:   "start [provider]",
	Short: "Start a provider runtime",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented: use the provider's native command to start it")
	},
}

var providerStopCmd = &cobra.Command{
	Use:   "stop [provider]",
	Short: "Stop a provider runtime",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented: use the provider's native command to stop it")
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize OpenForge configuration with auto-detection",
	Long: `Detect hardware and generate an optimized config.yaml.

Examples:
  openforge config init --auto     # Full auto: detect + recommend
  openforge config init --provider ollama`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigInit()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage OpenForge configuration",
}

func init() {
	rootCmd.AddCommand(providerCmd)
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerDetectCmd)
	providerCmd.AddCommand(providerInstallCmd)
	providerCmd.AddCommand(providerGuideCmd)
	providerCmd.AddCommand(providerInfoCmd)
	providerCmd.AddCommand(providerStatusCmd)
	providerCmd.AddCommand(providerStartCmd)
	providerCmd.AddCommand(providerStopCmd)

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)

	providerInstallCmd.Flags().Bool("auto", false, "auto-select best provider for hardware")
	providerInstallCmd.Flags().Bool("yes", false, "skip confirmation prompts")
	configInitCmd.Flags().Bool("auto", false, "automatic detection without prompts")
	configInitCmd.Flags().String("provider", "", "force a specific provider (ollama, openvino, etc.)")
}

// --- command implementations ---

func runProviderList() error {
	results, err := pm.Discover(context.Background())
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROVIDER\tINSTALLED\tRUNNING\tENDPOINT\tMODELS")
	for _, r := range results {
		instStr := "no"
		if r.Installed {
			instStr = "yes"
		}
		runStr := "no"
		if r.Running {
			runStr = "yes"
		}
		endpoint := r.Endpoint
		if endpoint == "" {
			endpoint = "-"
		}
		models := "-"
		if r.Health != nil && r.Health.Models > 0 {
			models = fmt.Sprintf("%d", r.Health.Models)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Provider, instStr, runStr, endpoint, models)
	}
	w.Flush()

	fmt.Println()
	fmt.Println("Run 'openforge provider install <name>' to install a runtime.")
	fmt.Println("Run 'openforge provider guide <name>' for platform-specific instructions.")
	fmt.Println("Run 'openforge provider detect' for detailed hardware scan.")
	return nil
}

func runProviderDetect() error {
	fmt.Print("Scanning hardware and runtimes...\n")

	results, err := pm.Discover(context.Background())
	if err != nil {
		return err
	}

	fmt.Printf("OS:      %s\n", runtime.GOOS)
	fmt.Printf("Arch:    %s\n", runtime.GOARCH)
	fmt.Println()

	for _, r := range results {
		icon := "❌"
		if r.Running {
			icon = "✅"
		} else if r.Installed {
			icon = "🟡"
		}

		fmt.Printf("%s  %s (%s)\n", icon, r.Provider, statusString(r))
		if r.Endpoint != "" {
			fmt.Printf("   Endpoint:  %s\n", r.Endpoint)
		}
		if r.Health != nil {
			fmt.Printf("   Latency:   %v\n", r.Health.Latency.Round(time.Millisecond))
			if r.Health.Error != "" {
				fmt.Printf("   Error:     %s\n", r.Health.Error)
			}
		}
		if !r.Installed && r.InstallHint != "" {
			fmt.Printf("   Hint:      %s\n", r.InstallHint)
		}
		fmt.Println()
	}

	if !hasRunning(results) {
		fmt.Println("⚠  No local inference runtime detected.")
		fmt.Println()
		fmt.Println("Quick start:")
		fmt.Println("  openforge provider install ollama     # Easiest: installs Ollama")
		fmt.Println("  openforge provider install openvino   # Intel NPU/GPU (requires Intel HW)")
		fmt.Println("  openforge provider guide              # Full platform-specific guides")
	}
	return nil
}

func runProviderInstall(args []string, auto bool) error {
	// Handle --auto: find best provider for current HW
	if auto || (len(args) == 0 && !auto) {
		return runProviderGuide(args)
	}

	target := args[0]
	fmt.Printf("Installation guide for: %s\n\n", target)

	guide, ok := pm.InstallGuides[pm.ProviderType(target)]
	if !ok {
		return fmt.Errorf("unknown provider: %s. Available: openvino, ollama, llamacpp, vllm, lmstudio", target)
	}

	fmt.Printf("  %s\n", guide.Description)
	fmt.Printf("  Website: %s\n\n", guide.Website)

	var found bool
	for _, p := range guide.Platforms {
		if p.OS == runtime.GOOS {
			found = true
			fmt.Printf("  === %s ===\n\n", p.OS)
			fmt.Println("  Install:")
			for _, cmd := range p.Install {
				fmt.Printf("    $ %s\n", cmd)
			}
			fmt.Println()
			if len(p.Verify) > 0 {
				fmt.Println("  Verify:")
				for _, cmd := range p.Verify {
					fmt.Printf("    $ %s\n", cmd)
				}
			}
			if p.Notes != "" {
				fmt.Printf("\n  Note: %s\n", p.Notes)
			}
			fmt.Println()
		}
	}

	if !found {
		fmt.Println("  No platform-specific instructions for this OS.")
		fmt.Println("  See the website for manual installation.")
	}

	if len(guide.PostInstall) > 0 {
		fmt.Println("  After installation:")
		for _, step := range guide.PostInstall {
			fmt.Printf("    %s\n", step)
		}
		fmt.Println()
	}

	fmt.Printf("  Verify: %s\n", guide.VerifyCmd)
	return nil
}

func runProviderGuide(args []string) error {
	if len(args) > 0 {
		return runProviderInstall(args, false)
	}

	system := identifySystem()

	fmt.Print("=== OpenForge Provider Installation Guide ===\n\n")
	fmt.Printf("Detected: %s\n\n", system)

	fmt.Println("--- Ollama (Recommended for most users) ---")
	showShortGuide(pm.ProviderOllama)

	fmt.Println("\n--- OpenVINO (Best for Intel NPU/GPU) ---")
	showShortGuide(pm.ProviderOpenVINO)

	fmt.Println("\n--- vLLM (Best for NVIDIA GPU setups) ---")
	showShortGuide(pm.ProviderVLLM)

	fmt.Println("\n--- LM Studio (GUI app, easy to use) ---")
	showShortGuide(pm.ProviderLMStudio)

	fmt.Println("\n--- llama.cpp (Minimal, CPU-first) ---")
	showShortGuide(pm.ProviderLlamaCpp)

	fmt.Println("\nRun 'openforge provider install <name>' for detailed instructions.")
	fmt.Println("Run 'openforge provider detect' to scan your system.")

	return nil
}

func runProviderInfo(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("specify a provider name: openforge provider info ollama")
	}
	target := args[0]

	guide, ok := pm.InstallGuides[pm.ProviderType(target)]
	if !ok {
		return fmt.Errorf("unknown provider: %s", target)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(guide)
}

func runProviderStatus() error {
	results, err := pm.Discover(context.Background())
	if err != nil {
		return err
	}

	for _, r := range results {
		fmt.Printf("%-12s  installed=%v  running=%v", r.Provider, r.Installed, r.Running)
		if r.Health != nil {
			fmt.Printf("  status=%s", r.Health.Status)
			if r.Health.Latency > 0 {
				fmt.Printf("  latency=%v", r.Health.Latency.Round(time.Millisecond))
			}
		}
		fmt.Println()
	}
	return nil
}

func runConfigInit() error {
	fmt.Print("OpenForge configuration initialization\n\n")

	results, err := pm.Discover(context.Background())
	if err != nil {
		return err
	}

	var running []pm.DiscoveryResult
	for _, r := range results {
		if r.Running {
			running = append(running, r)
		}
	}

	if len(running) == 0 {
		fmt.Println("No local inference runtime found.")
		fmt.Println()
		fmt.Println("Run 'openforge provider install ollama' to get started,")
		fmt.Println("or 'openforge provider guide' to see all options.")
		return nil
	}

	fmt.Println("Available runtimes:")
	for _, r := range running {
		fmt.Printf("  ✅ %s — %s\n", r.Provider, r.Endpoint)
	}
	fmt.Println()

	var buf []byte
	// Write a basic config YAML
	buf = append(buf, "# OpenForge configuration — auto-generated\n# Edit as needed\n\n"...)
	buf = append(buf, "providers:\n"...)
	buf = append(buf, "  chain: [openvino, ollama, llamacpp, vllm, lmstudio]\n\n"...)
	buf = append(buf, "  workloads:\n"...)
	buf = append(buf, "    chat: auto\n"...)
	buf = append(buf, "    embed: auto\n"...)
	buf = append(buf, "    rerank: auto\n"...)
	buf = append(buf, "    code: auto\n\n"...)

	for _, r := range running {
		buf = append(buf, fmt.Sprintf("  %s:\n", r.Provider)...)
		buf = append(buf, "    enabled: true\n"...)
		if r.Endpoint != "" {
			buf = append(buf, fmt.Sprintf("    endpoint: %s\n", r.Endpoint)...)
		}
		buf = append(buf, "\n"...)
	}

	buf = append(buf, "server:\n"...)
	buf = append(buf, "  host: 127.0.0.1\n"...)
	buf = append(buf, "  port: 9090\n"...)
	buf = append(buf, "  timeout: 30\n"...)
	buf = append(buf, "\n"...)
	buf = append(buf, "logging:\n"...)
	buf = append(buf, "  level: info\n"...)
	buf = append(buf, "  format: text\n"...)

	path := "openforge.yaml"
	if err := os.WriteFile(path, buf, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("✅ Configuration written to %s\n\n", path)
	fmt.Println("You can now run:")
	fmt.Println("  openforge              # Launch interactive TUI")
	fmt.Println("  openforge serve        # Start HTTP API server")

	return nil
}

// --- helpers ---

func statusString(r pm.DiscoveryResult) string {
	if r.Running {
		return "running"
	}
	if r.Installed {
		return "installed (not running)"
	}
	return "not installed"
}

func hasRunning(results []pm.DiscoveryResult) bool {
	for _, r := range results {
		if r.Running {
			return true
		}
	}
	return false
}

func identifySystem() string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("Windows (native) (%s)", runtime.GOARCH)
	case "darwin":
		return fmt.Sprintf("macOS %s", runtime.GOARCH)
	default:
		return fmt.Sprintf("Linux (%s)", runtime.GOARCH)
	}
}

func showShortGuide(pt pm.ProviderType) {
	guide, ok := pm.InstallGuides[pt]
	if !ok {
		return
	}
	for _, p := range guide.Platforms {
		if p.OS == runtime.GOOS {
			for _, cmd := range p.Install {
				fmt.Printf("  $ %s\n", cmd)
			}
			return
		}
	}
	fmt.Println("  No instructions for this OS — see website.")
}

var _ = io.Discard // skip unused import
