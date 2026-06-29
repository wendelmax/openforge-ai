package skills

import (
	"os"
	"path/filepath"
	"strings"
)

type SkillFile struct {
	Name        string
	Description string
	Content     string
	Path        string
}

func DiscoverSkills(dirs ...string) ([]SkillFile, error) {
	var skills []SkillFile
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		expanded := expandHome(dir)
		_ = filepath.Walk(expanded, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.EqualFold(info.Name(), "skill.md") {
				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				content := string(data)
				name := filepath.Base(filepath.Dir(path))
				if name == "." || name == "" {
					name = info.Name()
				}
				desc := extractFrontmatter(content, "description")
				skills = append(skills, SkillFile{
					Name: name, Description: desc, Content: content, Path: path,
				})
			}
			return nil
		})
	}
	return skills, nil
}

func extractFrontmatter(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key+":") {
			return strings.TrimSpace(strings.TrimPrefix(line, key+":"))
		}
	}
	return ""
}

func InjectSkills(prompt string, skills []SkillFile) string {
	if len(skills) == 0 {
		return prompt
	}
	var sb strings.Builder
	sb.WriteString(prompt)
	sb.WriteString("\n## Loaded Skills\n\n")
	for _, s := range skills {
		sb.WriteString("### " + s.Name + "\n")
		if s.Description != "" {
			sb.WriteString(s.Description + "\n\n")
		}
		sb.WriteString(s.Content + "\n\n")
	}
	return sb.String()
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
