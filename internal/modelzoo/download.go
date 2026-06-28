package modelzoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type downloadProgress struct {
	mu       sync.Mutex
	bytes    int64
	total    int64
	started  time.Time
	filename string
}

func (p *downloadProgress) Write(b []byte) (int, error) {
	n := len(b)
	p.mu.Lock()
	p.bytes += int64(n)
	current := p.bytes
	total := p.total
	filename := p.filename
	p.mu.Unlock()

	if total > 0 {
		pct := float64(current) * 100 / float64(total)
		elapsed := time.Since(p.started)
		speed := float64(current) / elapsed.Seconds() / 1024 / 1024
		fmt.Fprintf(os.Stderr, "\r  %s: %.0f%% (%.1f MB/s)      ", filename, pct, speed)
	}
	return n, nil
}

// DownloadModel downloads all model files from Hugging Face to the destination directory.
func DownloadModel(ctx context.Context, model *Model, destDir string) error {
	files, err := listRepoFiles(ctx, model.HuggingFace)
	if err != nil {
		return fmt.Errorf("listing files for %s: %w", model.HuggingFace, err)
	}

	modelDir := filepath.Join(destDir, model.ID)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", modelDir, err)
	}

	if err := writeManifest(modelDir, model, files); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	for _, f := range files {
		if err := downloadFile(ctx, model, f, modelDir); err != nil {
			return fmt.Errorf("downloading %s: %w", f, err)
		}
	}

	return nil
}

func listRepoFiles(ctx context.Context, repoID string) ([]string, error) {
	apiURL := huggingFaceAPIURL(repoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d for %s", resp.StatusCode, repoID)
	}

	var result struct {
		Siblings []struct {
			Rfilename string `json:"rfilename"`
			Size      int64  `json:"size"`
		} `json:"siblings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	ovExt := map[string]bool{
		".xml": true, ".bin": true, ".json": true,
	}
	var files []string
	for _, s := range result.Siblings {
		ext := strings.ToLower(filepath.Ext(s.Rfilename))
		if ovExt[ext] && !strings.Contains(s.Rfilename, "safe_") {
			files = append(files, s.Rfilename)
		}
	}

	return files, nil
}

func downloadFile(ctx context.Context, model *Model, filename, destDir string) error {
	localPath := filepath.Join(destDir, filename)
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}

	url := huggingFaceFileURL(model, filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	progress := &downloadProgress{
		total:    resp.ContentLength,
		started:  time.Now(),
		filename: filename,
	}

	written, err := io.Copy(f, io.TeeReader(resp.Body, progress))
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\r  %s: 100%% (%.1f MB) done\n", filename, float64(written)/1024/1024)
	return nil
}

// ManifestVersion is the current version of the download manifest format.
var ManifestVersion = 1

type manifest struct {
	Version   int      `json:"version"`
	ModelID   string   `json:"model_id"`
	HuggingFace string `json:"huggingface"`
	Downloaded string  `json:"downloaded"`
	Files     []string `json:"files"`
}

func writeManifest(dir string, model *Model, files []string) error {
	m := manifest{
		Version:     ManifestVersion,
		ModelID:     model.ID,
		HuggingFace: model.HuggingFace,
		Downloaded:  time.Now().UTC().Format(time.RFC3339),
		Files:       files,
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".openforge.json"), data, 0644)
}
