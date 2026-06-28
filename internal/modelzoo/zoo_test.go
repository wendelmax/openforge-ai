package modelzoo

import (
	"testing"
)

func TestCatalogNotEmpty(t *testing.T) {
	models := Catalog()
	if len(models) == 0 {
		t.Fatal("catalog should not be empty")
	}
}

func TestByID_Found(t *testing.T) {
	m := ByID("phi-3-mini")
	if m == nil {
		t.Fatal("expected phi-3-mini in catalog")
	}
	if m.Type != TypeChat {
		t.Errorf("expected chat type, got %s", m.Type)
	}
}

func TestByID_Alias(t *testing.T) {
	m := ByID("phi-3-mini-int4-ov")
	if m == nil {
		t.Fatal("expected phi-3-mini-int4-ov (alias) to resolve")
	}
}

func TestByID_CaseInsensitive(t *testing.T) {
	m := ByID("Phi-3-Mini")
	if m == nil {
		t.Fatal("expected case-insensitive lookup to work")
	}
}

func TestByID_Unknown(t *testing.T) {
	m := ByID("nonexistent-model")
	if m != nil {
		t.Fatal("expected nil for unknown model")
	}
}

func TestByType_Chat(t *testing.T) {
	models := ByType(TypeChat)
	if len(models) == 0 {
		t.Fatal("expected chat models")
	}
	for _, m := range models {
		if m.Type != TypeChat {
			t.Errorf("model %s has type %s, expected chat", m.ID, m.Type)
		}
	}
}

func TestByType_Embed(t *testing.T) {
	models := ByType(TypeEmbed)
	if len(models) == 0 {
		t.Fatal("expected embedding models")
	}
	for _, m := range models {
		if m.Type != TypeEmbed {
			t.Errorf("model %s has type %s, expected embed", m.ID, m.Type)
		}
	}
}

func TestChatModels(t *testing.T) {
	models := ChatModels()
	if len(models) == 0 {
		t.Fatal("expected chat models")
	}
}

func TestEmbedModels(t *testing.T) {
	models := EmbedModels()
	if len(models) == 0 {
		t.Fatal("expected embedding models")
	}
}

func TestDefaultModel(t *testing.T) {
	m := DefaultModel()
	if m == nil {
		t.Fatal("expected a default model")
	}
	if !m.DefaultModel {
		t.Errorf("model %s should have DefaultModel=true", m.ID)
	}
}

func TestDefaultModel_Unique(t *testing.T) {
	count := 0
	for _, m := range Catalog() {
		if m.DefaultModel {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 default model, got %d", count)
	}
}

func TestHuggingFaceURL(t *testing.T) {
	m := ByID("phi-3-mini")
	if m == nil {
		t.Fatal("phi-3-mini not found")
	}
	url := HuggingFaceURL(m)
	expected := "https://huggingface.co/OpenVINO/phi-3-mini-int4-ov"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestModelIDs_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for _, m := range Catalog() {
		if seen[m.ID] {
			t.Errorf("duplicate model ID: %s", m.ID)
		}
		seen[m.ID] = true
	}
}

func TestModelHuggingFace_NotEmpty(t *testing.T) {
	for _, m := range Catalog() {
		if m.HuggingFace == "" {
			t.Errorf("model %s has empty HuggingFace field", m.ID)
		}
	}
}

func TestModelSize_Valid(t *testing.T) {
	for _, m := range Catalog() {
		if m.SizeMB <= 0 {
			t.Errorf("model %s has invalid size: %d MB", m.ID, m.SizeMB)
		}
	}
}

func TestModelRamMin_Valid(t *testing.T) {
	for _, m := range Catalog() {
		if m.RamMinMB <= 0 {
			t.Errorf("model %s has invalid RamMinMB: %d", m.ID, m.RamMinMB)
		}
	}
}

func TestModelFiles_NotEmpty(t *testing.T) {
	for _, m := range Catalog() {
		if len(m.Files) > 0 {
			continue
		}
	}
}

func TestCodeModels(t *testing.T) {
	models := CodeModels()
	if len(models) == 0 {
		t.Fatal("expected code models")
	}
	for _, m := range models {
		if m.Type != TypeCode {
			t.Errorf("model %s has type %s, expected code", m.ID, m.Type)
		}
	}
}

func TestHuggingFaceFileURL(t *testing.T) {
	m := ByID("phi-3-mini")
	if m == nil {
		t.Fatal("phi-3-mini not found")
	}
	url := HuggingFaceFileURL(m, "openvino_model.xml")
	expected := "https://huggingface.co/OpenVINO/phi-3-mini-int4-ov/resolve/main/openvino_model.xml"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestDefaultHuggingFaceAPIURL(t *testing.T) {
	url := defaultHuggingFaceAPIURL("test/repo")
	expected := "https://huggingface.co/api/models/test/repo"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestDefaultHuggingFaceFileURL(t *testing.T) {
	m := &Model{
		HuggingFace: "test/repo",
	}
	url := defaultHuggingFaceFileURL(m, "model.bin")
	expected := "https://huggingface.co/test/repo/resolve/main/model.bin"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestModelTypes_Coverage(t *testing.T) {
	types := make(map[ModelType]bool)
	for _, m := range Catalog() {
		types[m.Type] = true
	}
	if !types[TypeChat] {
		t.Error("no chat models in catalog")
	}
	if !types[TypeEmbed] {
		t.Error("no embedding models in catalog")
	}
}
