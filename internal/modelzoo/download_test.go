package modelzoo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadModel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/models/OpenVINO/test-model-ov" {
			resp := map[string]interface{}{
				"siblings": []map[string]interface{}{
					{"rfilename": "openvino_model.xml", "size": 100},
					{"rfilename": "openvino_model.bin", "size": 1000},
					{"rfilename": "config.json", "size": 50},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock content"))
	}))
	defer ts.Close()

	origURL := huggingFaceFileURL
	huggingFaceFileURL = func(m *Model, filename string) string {
		return ts.URL + "/resolve/main/" + filename
	}
	origAPI := huggingFaceAPIURL
	huggingFaceAPIURL = func(repoID string) string {
		return ts.URL + "/api/models/" + repoID
	}
	defer func() {
		huggingFaceFileURL = origURL
		huggingFaceAPIURL = origAPI
	}()

	model := &Model{
		ID:          "test-model",
		HuggingFace: "OpenVINO/test-model-ov",
		Precision:   "INT4",
		SizeMB:      1,
	}

	destDir := t.TempDir()
	err := DownloadModel(context.Background(), model, destDir)
	require.NoError(t, err)

	assert.DirExists(t, filepath.Join(destDir, "test-model"))
	assert.FileExists(t, filepath.Join(destDir, "test-model", "openvino_model.xml"))
	assert.FileExists(t, filepath.Join(destDir, "test-model", "config.json"))

	manifestPath := filepath.Join(destDir, "test-model", ".openforge.json")
	assert.FileExists(t, manifestPath)

	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	var m manifest
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)
	assert.Equal(t, "test-model", m.ModelID)
	assert.Equal(t, "OpenVINO/test-model-ov", m.HuggingFace)
	assert.Len(t, m.Files, 3)
}

func TestDownloadModelAPIFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	origAPI := huggingFaceAPIURL
	huggingFaceAPIURL = func(repoID string) string {
		return ts.URL + "/api/models/" + repoID
	}
	defer func() { huggingFaceAPIURL = origAPI }()

	model := &Model{
		ID:          "test-model",
		HuggingFace: "OpenVINO/nonexistent",
	}

	err := DownloadModel(context.Background(), model, t.TempDir())
	assert.Error(t, err)
}

func TestListRepoFiles(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"siblings": []map[string]interface{}{
				{"rfilename": "openvino_model.xml", "size": 100},
				{"rfilename": "openvino_model.bin", "size": 1000},
				{"rfilename": "config.json", "size": 50},
				{"rfilename": ".gitattributes", "size": 10},
				{"rfilename": "README.md", "size": 200},
				{"rfilename": "safe_assets/checkpoint", "size": 500},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	origAPI := huggingFaceAPIURL
	huggingFaceAPIURL = func(repoID string) string {
		return ts.URL + "/api/models/" + repoID
	}
	defer func() { huggingFaceAPIURL = origAPI }()

	files, err := listRepoFiles(context.Background(), "test-repo")
	require.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Contains(t, files, "openvino_model.xml")
	assert.Contains(t, files, "openvino_model.bin")
	assert.Contains(t, files, "config.json")
	assert.NotContains(t, files, ".gitattributes")
	assert.NotContains(t, files, "README.md")
	assert.NotContains(t, files, "safe_assets/checkpoint")
}

func TestListRepoFilesAPIFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	origAPI := huggingFaceAPIURL
	huggingFaceAPIURL = func(repoID string) string {
		return ts.URL + "/api/models/" + repoID
	}
	defer func() { huggingFaceAPIURL = origAPI }()

	_, err := listRepoFiles(context.Background(), "test-repo")
	assert.Error(t, err)
}

func TestWriteManifest(t *testing.T) {
	dir := t.TempDir()

	model := &Model{
		ID:          "test-model",
		HuggingFace: "OpenVINO/test-model-ov",
	}
	files := []string{"openvino_model.xml", "openvino_model.bin"}

	err := writeManifest(dir, model, files)
	require.NoError(t, err)

	manifestPath := filepath.Join(dir, ".openforge.json")
	assert.FileExists(t, manifestPath)

	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var m manifest
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)
	assert.Equal(t, ManifestVersion, m.Version)
	assert.Equal(t, "test-model", m.ModelID)
	assert.Equal(t, "OpenVINO/test-model-ov", m.HuggingFace)
	assert.Len(t, m.Files, 2)
}

func TestListRepoFilesNetworkError(t *testing.T) {
	origAPI := huggingFaceAPIURL
	huggingFaceAPIURL = func(repoID string) string {
		return "http://127.0.0.1:1/api/models/" + repoID
	}
	defer func() { huggingFaceAPIURL = origAPI }()

	_, err := listRepoFiles(context.Background(), "test-repo")
	assert.Error(t, err)
}

func TestDownloadFileHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	origURL := huggingFaceFileURL
	huggingFaceFileURL = func(m *Model, filename string) string {
		return ts.URL + "/" + filename
	}
	defer func() { huggingFaceFileURL = origURL }()

	model := &Model{
		ID:          "test-model",
		HuggingFace: "OpenVINO/test-model-ov",
	}

	err := downloadFile(context.Background(), model, "model.bin", t.TempDir())
	assert.Error(t, err)
}

func TestDownloadFileNetworkError(t *testing.T) {
	origURL := huggingFaceFileURL
	huggingFaceFileURL = func(m *Model, filename string) string {
		return "http://127.0.0.1:1/nonexistent"
	}
	defer func() { huggingFaceFileURL = origURL }()

	model := &Model{
		ID:          "test-model",
		HuggingFace: "OpenVINO/test-model-ov",
	}

	err := downloadFile(context.Background(), model, "model.bin", t.TempDir())
	assert.Error(t, err)
}

func TestDownloadModelListFilesError(t *testing.T) {
	// API server returns error
	err := DownloadModel(context.Background(), &Model{
		ID:          "test-model",
		HuggingFace: "nonexistent",
	}, t.TempDir())
	assert.Error(t, err)
}

func TestDownloadFileCreateError(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "readonly")
	require.NoError(t, os.MkdirAll(subdir, 0755))
	require.NoError(t, os.Chmod(subdir, 0444))
	t.Cleanup(func() { os.Chmod(subdir, 0755) })

	origURL := huggingFaceFileURL
	huggingFaceFileURL = func(m *Model, filename string) string {
		return "http://127.0.0.1:1/" + filename
	}
	defer func() { huggingFaceFileURL = origURL }()

	model := &Model{ID: "test", HuggingFace: "test/repo"}
	err := downloadFile(context.Background(), model, "file.bin", subdir)
	assert.Error(t, err)
}

func TestWriteManifestError(t *testing.T) {
	model := &Model{
		ID:          "test-model",
		HuggingFace: "OpenVINO/test-model-ov",
	}
	files := []string{"model.xml"}

	dir := t.TempDir()
	modelDir := filepath.Join(dir, "test-model")
	os.MkdirAll(modelDir, 0755)
	// Make the model dir read-only so writeManifest fails
	os.Chmod(modelDir, 0444)
	defer os.Chmod(modelDir, 0755)

	err := writeManifest(modelDir, model, files)
	assert.Error(t, err)
}

func TestListRepoFilesBadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	origAPI := huggingFaceAPIURL
	huggingFaceAPIURL = func(repoID string) string {
		return ts.URL + "/api/models/" + repoID
	}
	defer func() { huggingFaceAPIURL = origAPI }()

	_, err := listRepoFiles(context.Background(), "test-repo")
	assert.Error(t, err)
}

func TestDownloadModelMkdirAllError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"siblings": []map[string]interface{}{
				{"rfilename": "model.xml", "size": 100},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	origAPI := huggingFaceAPIURL
	huggingFaceAPIURL = func(repoID string) string {
		return ts.URL + "/api/models/" + repoID
	}
	defer func() { huggingFaceAPIURL = origAPI }()

	destDir := t.TempDir()
	os.Chmod(destDir, 0555)
	t.Cleanup(func() { os.Chmod(destDir, 0755) })

	model := &Model{ID: "test-model", HuggingFace: "test/repo"}
	err := DownloadModel(context.Background(), model, destDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating")
}

func TestDownloadModelDownloadFileError(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"siblings": []map[string]interface{}{
				{"rfilename": "model.xml", "size": 100},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer fileServer.Close()

	origAPI := huggingFaceAPIURL
	origURL := huggingFaceFileURL
	huggingFaceAPIURL = func(repoID string) string {
		return apiServer.URL + "/api/models/" + repoID
	}
	huggingFaceFileURL = func(m *Model, filename string) string {
		return fileServer.URL + "/" + filename
	}
	defer func() {
		huggingFaceAPIURL = origAPI
		huggingFaceFileURL = origURL
	}()

	model := &Model{ID: "test-model", HuggingFace: "test/repo"}
	err := DownloadModel(context.Background(), model, t.TempDir())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "downloading")
}

func TestDownloadFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mock model data"))
	}))
	defer ts.Close()

	origURL := huggingFaceFileURL
	huggingFaceFileURL = func(m *Model, filename string) string {
		return ts.URL + "/" + filename
	}
	defer func() { huggingFaceFileURL = origURL }()

	dir := t.TempDir()
	model := &Model{
		ID:          "test-model",
		HuggingFace: "OpenVINO/test-model-ov",
	}

	err := downloadFile(context.Background(), model, "model.bin", dir)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "model.bin"))
	data, _ := os.ReadFile(filepath.Join(dir, "model.bin"))
	assert.Equal(t, "mock model data", string(data))
}
