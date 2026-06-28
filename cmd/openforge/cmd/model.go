package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/modelzoo"
	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:     "model",
	Aliases: []string{"models", "m"},
	Short:   "Manage AI models",
	Long:    `List, load, unload, and download models from the OpenVINO model zoo.`,
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List downloaded models",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(cfg.Models.Path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("no models downloaded yet")
				return nil
			}
			return fmt.Errorf("cannot read %q: %w", cfg.Models.Path, err)
		}

		found := false
		for _, e := range entries {
			if e.IsDir() {
				info := modelzoo.ByID(e.Name())
				if info != nil {
					fmt.Printf("  %-25s %s\n", e.Name(), info.Description)
				} else {
					fmt.Printf("  %-25s (unknown model)\n", e.Name())
				}
				found = true
			}
		}
		if !found {
			fmt.Println("no models downloaded yet")
		}
		return nil
	},
}

var modelLoadCmd = &cobra.Command{
	Use:   "load <model-id>",
	Short: "Load a model into memory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		modelID := args[0]
		device, _ := cmd.Flags().GetString("device")

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if device == "" {
			device = cfg.Models.Device
		}

		provider := openvino.NewProvider(cfg.Models.Path)
		if err := provider.Initialize(ctx); err != nil {
			return err
		}
		defer provider.Shutdown(ctx)

		modelPath := filepath.Join(cfg.Models.Path, modelID)
		if err := provider.Runtime().LoadModel(ctx, modelID, modelPath, device); err != nil {
			return fmt.Errorf("load failed: %w", err)
		}
		fmt.Printf("loaded: %s (device: %s)\n", modelID, device)
		return nil
	},
}

var modelUnloadCmd = &cobra.Command{
	Use:   "unload <model-id>",
	Short: "Unload a model from memory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		modelID := args[0]

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		provider := openvino.NewProvider(cfg.Models.Path)
		if err := provider.Initialize(ctx); err != nil {
			return err
		}
		defer provider.Shutdown(ctx)

		if err := provider.Runtime().UnloadModel(ctx, modelID); err != nil {
			return fmt.Errorf("unload failed: %w", err)
		}
		fmt.Printf("unloaded: %s\n", modelID)
		return nil
	},
}

var modelPullCmd = &cobra.Command{
	Use:   "pull <model-id>",
	Short: "Download a model from HuggingFace",
	Long:  `Download an OpenVINO-format model from HuggingFace. Use 'openforge model list-remote' to see available models.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		name := args[0]

		model := modelzoo.ByID(name)
		if model == nil {
			return fmt.Errorf("unknown model %q; use 'openforge model list-remote' to see available models", name)
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		fmt.Printf("Downloading %s (%s) from %s\n", model.Name, model.Precision, model.HuggingFace)
		fmt.Printf("  Size: ~%d MB\n", model.SizeMB)

		if err := modelzoo.DownloadModel(ctx, model, cfg.Models.Path); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		fmt.Printf("done: %s saved to %s\n", model.ID, cfg.Models.Path)
		return nil
	},
}

var modelListRemoteCmd = &cobra.Command{
	Use:   "list-remote",
	Short: "List available models in the OpenVINO zoo",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintf(w, "ID\tTYPE\tPRECISION\tSIZE\tDESCRIPTION\n")
		fmt.Fprintf(w, "--\t----\t---------\t----\t-----------\n")

		for _, m := range modelzoo.Catalog() {
			size := fmt.Sprintf("%d MB", m.SizeMB)
			mark := ""
			if m.DefaultModel {
				mark = "  [default]"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s%s\n", m.ID, m.Type, m.Precision, size, m.Description, mark)
		}

		w.Flush()
		fmt.Println()
		fmt.Println("Use 'openforge model pull <id>' to download a model.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(modelCmd)
	modelCmd.AddCommand(modelListCmd)
	modelCmd.AddCommand(modelLoadCmd)
	modelCmd.AddCommand(modelUnloadCmd)
	modelCmd.AddCommand(modelPullCmd)
	modelCmd.AddCommand(modelListRemoteCmd)

	modelLoadCmd.Flags().StringP("device", "d", "", "target device")
}
