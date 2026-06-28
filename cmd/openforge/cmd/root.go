package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	verbose     bool
	showVersion bool
)

var rootCmd = &cobra.Command{
	Use:   "openforge",
	Short: "AI Runtime for Developers — 100% Local, 100% OpenVINO",
	Long: `OpenForge is an open-source AI inference framework that runs 
exclusively on OpenVINO Runtime. Execute LLMs, embeddings, and 
reranking models entirely offline on Intel hardware.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			return printVersion()
		}
		return runTUI()
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "V", false, "print version")
}

func printVersion() error {
	fmt.Printf("openforge %s\n", Version)
	fmt.Printf("commit: %s\n", Commit)
	fmt.Printf("date:   %s\n", Date)
	return nil
}

var (
	Version = "0.1.0-dev"
	Commit  = "none"
	Date    = "unknown"
)
