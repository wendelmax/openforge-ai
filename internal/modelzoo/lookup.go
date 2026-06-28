package modelzoo

import "strings"

// ByID looks up a model by its ID or alias, returning nil if not found.
func ByID(id string) *Model {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, m := range Catalog() {
		if strings.EqualFold(m.ID, id) {
			return &m
		}
		for _, a := range m.Aliases {
			if strings.EqualFold(a, id) {
				return &m
			}
		}
	}
	return nil
}

// ByType returns all models matching the given type.
func ByType(t ModelType) []Model {
	var result []Model
	for _, m := range Catalog() {
		if m.Type == t {
			result = append(result, m)
		}
	}
	return result
}

// ChatModels returns all chat-type models.
func ChatModels() []Model { return ByType(TypeChat) }
// EmbedModels returns all embedding-type models.
func EmbedModels() []Model { return ByType(TypeEmbed) }
// CodeModels returns all code-type models.
func CodeModels() []Model { return ByType(TypeCode) }

// DefaultModel returns the first model marked as the default, or nil.
func DefaultModel() *Model {
	for _, m := range Catalog() {
		if m.DefaultModel {
			return &m
		}
	}
	return nil
}

var huggingFaceAPIURL = defaultHuggingFaceAPIURL

func defaultHuggingFaceAPIURL(repoID string) string {
	return "https://huggingface.co/api/models/" + repoID
}

var huggingFaceFileURL = defaultHuggingFaceFileURL

func defaultHuggingFaceFileURL(m *Model, filename string) string {
	return "https://huggingface.co/" + m.HuggingFace + "/resolve/main/" + filename
}

// HuggingFaceURL returns the Hugging Face model page URL for the given model.
func HuggingFaceURL(m *Model) string {
	return "https://huggingface.co/" + m.HuggingFace
}

// HuggingFaceFileURL returns the download URL for a specific model file.
func HuggingFaceFileURL(m *Model, filename string) string {
	return huggingFaceFileURL(m, filename)
}
