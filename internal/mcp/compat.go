package mcp

import (
	"context"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/permission"
)

func ConnectFromConfig(ctx context.Context, r *Registry, cfg *config.Config) error {
	for name, serverCfg := range config.GetMCPServers(cfg) {
		if err := r.Connect(ctx, name, serverCfg.Command, serverCfg.Args...); err != nil {
			return err
		}
	}
	return nil
}

func BuildPermsFromConfig(cfg *config.Config, store permission.Store, sessionID string) *permission.Manager {
	m := permission.NewManager(sessionID, store)
	dc := config.GetDefaultPermissionDecision(cfg)
	dec, _ := permission.ParseDecision(dc)
	level, _ := permission.ParseLevel(config.GetDefaultPermissionLevel(cfg))
	if level == 0 && dec != permission.DecisionDeny {
		level = permission.LevelOnce
	}
	rules := config.GetPermissionRules(cfg)
	permRules := make([]permission.Rule, len(rules))
	for i, r := range rules {
		d, _ := permission.ParseDecision(r.Decision)
		l, _ := permission.ParseLevel(r.Level)
		if l == 0 && d != permission.DecisionDeny {
			l = permission.LevelOnce
		}
		permRules[i] = permission.Rule{Scope: r.Scope, Name: r.Name, Decision: d, Level: l}
	}
	if len(permRules) == 0 {
		permRules = []permission.Rule{{Scope: "*", Name: "*", Decision: dec, Level: level}}
	}
	m.SetRules(permRules)
	return m
}
