package tool

import (
	"fmt"
	"sort"
	"strings"
)

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) List() []Tool {
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func (r *Registry) Descriptions() string {
	var b strings.Builder
	for _, t := range r.List() {
		b.WriteString(fmt.Sprintf("- %s: %s\n", t.Name(), t.Description()))
	}
	return b.String()
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(NewBashTool())
	r.Register(NewViewTool())
	r.Register(NewWriteTool())
	r.Register(NewEditTool())
	r.Register(NewGrepTool())
	r.Register(NewGlobTool())
	r.Register(NewLsTool())
	r.Register(NewTodosTool())
	r.Register(NewFetchTool(false))
	return r
}
