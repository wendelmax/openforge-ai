package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverSkills_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	skills, err := DiscoverSkills(dir)
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestDiscoverSkills_FindsSkillMD(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	content := "description: a test skill\n\n# My Skill\n\nDo stuff."
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644))
	skills, err := DiscoverSkills(dir)
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "my-skill", skills[0].Name)
	assert.Equal(t, "a test skill", skills[0].Description)
}

func TestInjectSkills_NoSkills(t *testing.T) {
	prompt := "hello world"
	assert.Equal(t, prompt, InjectSkills(prompt, nil))
	assert.Equal(t, prompt, InjectSkills(prompt, []SkillFile{}))
}

func TestInjectSkills_WithSkills(t *testing.T) {
	result := InjectSkills("base", []SkillFile{
		{Name: "skill1", Description: "does things", Content: "content here"},
	})
	assert.Contains(t, result, "base")
	assert.Contains(t, result, "skill1")
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := expandHome("~/projects/test")
	assert.Equal(t, filepath.Join(home, "projects", "test"), result)
}
