package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/provider/openvino"
	"github.com/openforge-ai/openforge/internal/skill"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage and run AI skills",
	Long:  `List, run, create, and validate reusable AI pipeline skills.`,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "skills"
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("no skills directory found at %q", dir)
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
				name := strings.TrimSuffix(e.Name(), ".yaml")
				name = strings.TrimSuffix(name, ".yml")
				fmt.Println(name)
			}
		}
		return nil
	},
}

var skillRunCmd = &cobra.Command{
	Use:   "run <skill-name>",
	Short: "Execute a skill pipeline",
	Long:  "Runs a YAML skill pipeline using the configured runtime.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		params, _ := cmd.Flags().GetStringToString("param")

		dir := "skills"
		yamlPath := filepath.Join(dir, name+".yaml")
		if _, err := os.Stat(yamlPath); err != nil {
			yamlPath = filepath.Join(dir, name+".yml")
			if _, err := os.Stat(yamlPath); err != nil {
				return fmt.Errorf("skill %q not found in skills/", name)
			}
		}

		data, err := os.ReadFile(yamlPath)
		if err != nil {
			return fmt.Errorf("read skill: %w", err)
		}

		var s skill.Skill
		if err := yaml.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("parse skill: %w", err)
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		ctx := context.Background()
		provider := openvino.NewProvider(cfg.Models.Path)
		if err := provider.Initialize(ctx); err != nil {
			return fmt.Errorf("initialize runtime: %w", err)
		}
		defer provider.Shutdown(ctx)

		inputs := make(map[string]interface{})
		for k, v := range params {
			inputs[k] = v
		}

		executor := skill.NewExecutor(provider.Runtime())
		outputs, err := executor.Execute(ctx, s, inputs)
		if err != nil {
			return fmt.Errorf("skill failed: %w", err)
		}

		for stepName, result := range outputs {
			switch v := result.(type) {
			case string:
				fmt.Printf("[%s] %s\n", stepName, v)
			case []float32:
				fmt.Printf("[%s] embedding (len=%d, first=%.4f)\n", stepName, len(v), v[0])
			default:
				fmt.Printf("[%s] %v\n", stepName, v)
			}
		}
		return nil
	},
}

var skillCreateCmd = &cobra.Command{
	Use:   "create <skill-name>",
	Short: "Scaffold a new skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path := filepath.Join("skills", name+".yaml")

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("skill %q already exists", path)
		}

		tmpl := fmt.Sprintf(`name: %s
description: "TODO: describe what this skill does"
version: 1.0.0

inputs:
  input:
    type: string
    description: Input text
    required: true

steps:
  - id: prompt
    type: prompt
    model: llama-3.2-3b
    user: "{{.inputs.input}}"
    output: result

  - id: output
    type: format
    template: "{{.steps.prompt.output}}"
    output: final
`, name)

		if err := os.MkdirAll("skills", 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(tmpl), 0644); err != nil {
			return err
		}
		fmt.Printf("created: %s\n", path)
		return nil
	},
}

var skillValidateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a skill YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		var s skill.Skill
		if err := yaml.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}

		if s.Name == "" {
			return fmt.Errorf("skill name is required")
		}
		if len(s.Steps) == 0 {
			return fmt.Errorf("at least one step is required")
		}
		for i, step := range s.Steps {
			if step.Name == "" {
				return fmt.Errorf("step %d: name is required", i+1)
			}
			if step.Output == "" {
				return fmt.Errorf("step %d: output is required", i+1)
			}
			validTypes := map[skill.StepType]bool{
				skill.StepPrompt: true, skill.StepEmbed: true,
				skill.StepRerank: true, skill.StepFormat: true, skill.StepCond: true,
			}
			if !validTypes[step.Type] {
				return fmt.Errorf("step %d: invalid type %q", i+1, step.Type)
			}
		}

		fmt.Printf("✅ skill %q is valid (%d steps)\n", s.Name, len(s.Steps))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillRunCmd)
	skillCmd.AddCommand(skillCreateCmd)
	skillCmd.AddCommand(skillValidateCmd)

	skillRunCmd.Flags().StringToStringP("param", "p", nil, "skill parameters (key=value)")
}
