package temingo

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderReportsMistypedAssetPath is the regression test for the failure this
// feature exists to catch: a stylesheet reference pointing at nothing, which
// previously built cleanly and produced an unstyled page.
func TestRenderReportsMistypedAssetPath(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "style.css"), []byte(".a{color:red}"), 0o644); err != nil {
		t.Fatal(err)
	}
	page := `<html><head><link rel="stylesheet" href="/styl.css"></head><body><p>x</p></body></html>`
	if err := os.WriteFile(filepath.Join(src, "index.template.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	engine := DefaultEngine()
	engine.InputDir = src + string(filepath.Separator)
	engine.OutputDir = filepath.Join(dir, "output") + string(filepath.Separator)
	engine.TemingoignorePath = filepath.Join(dir, ".temingoignore")
	engine.Logger = slog.New(slog.NewTextHandler(&buf, nil))

	if err := engine.Render(); err != nil {
		t.Fatalf("Render() = %v, want nil - findings must not fail a non-strict build", err)
	}

	out := buf.String()
	if !strings.Contains(out, "missing-target") || !strings.Contains(out, "/styl.css") {
		t.Errorf("expected a missing-target finding for /styl.css, got:\n%s", out)
	}

	if _, err := os.Stat(filepath.Join(dir, "output", "index.html")); err != nil {
		t.Errorf("output was not written: %v", err)
	}
}

// TestRenderStrictFailsOnFinding proves the CI gate works.
func TestRenderStrictFailsOnFinding(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	page := `<html><body><a href="/nope/">x</a></body></html>`
	if err := os.WriteFile(filepath.Join(src, "index.template.html"), []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}

	engine := DefaultEngine()
	engine.InputDir = src + string(filepath.Separator)
	engine.OutputDir = filepath.Join(dir, "output") + string(filepath.Separator)
	engine.TemingoignorePath = filepath.Join(dir, ".temingoignore")
	engine.Logger = slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	engine.Strict = true

	if err := engine.Render(); err == nil {
		t.Errorf("Render() = nil, want an error under strict with a finding")
	}
}
