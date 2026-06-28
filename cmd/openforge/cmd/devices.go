package cmd

import (
	"context"
	"fmt"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List available inference devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		provider := openvino.NewProvider(cfg.Models.Path)
		if err := provider.Initialize(ctx); err != nil {
			return fmt.Errorf("OpenVINO init failed: %w", err)
		}
		defer provider.Shutdown(ctx)

		devices, err := provider.Runtime().ListDevices(ctx)
		if err != nil {
			return fmt.Errorf("device enumeration failed: %w", err)
		}

		if len(devices) == 0 {
			fmt.Println("no devices detected")
			return nil
		}

		fmt.Printf("%-15s %-30s %-6s %s\n", "ID", "NAME", "TYPE", "STATUS")
		fmt.Println("-", 60)
		for _, d := range devices {
			status := "available"
			if !d.Available {
				status = "unavailable"
			}
			fmt.Printf("%-15s %-30s %-6s %s\n", d.ID, d.Name, d.Type, status)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(devicesCmd)
}
