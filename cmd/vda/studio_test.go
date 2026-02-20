package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUntarGzAndFindGeneratedAppDir(t *testing.T) {
	payload := mustTarGz(map[string]string{
		"bundle-root/apps/acme/clients/web/index.html":          "<h1>web</h1>",
		"bundle-root/apps/acme/clients/mobile/index.html":       "<h1>mobile</h1>",
		"bundle-root/apps/acme/services/api/cmd/server/main.go": "package main",
	})
	dest := t.TempDir()
	root, err := untarGz(payload, dest)
	if err != nil {
		t.Fatalf("untarGz returned error: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(root), "/bundle-root") {
		t.Fatalf("expected extracted root to end with /bundle-root, got %q", root)
	}
	appDir, err := findGeneratedAppDir(root)
	if err != nil {
		t.Fatalf("findGeneratedAppDir returned error: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(appDir), "/apps/acme") {
		t.Fatalf("expected app dir to end with /apps/acme, got %q", appDir)
	}
}

func TestUntarGzBlocksTraversal(t *testing.T) {
	payload := mustTarGz(map[string]string{
		"bundle-root/../../escape.txt": "bad",
	})
	dest := t.TempDir()
	_, err := untarGz(payload, dest)
	if err == nil {
		t.Fatalf("expected traversal error")
	}
	if !strings.Contains(err.Error(), "path traversal") {
		t.Fatalf("expected path traversal error, got %v", err)
	}
}

func mustTarGz(files map[string]string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		body := []byte(content)
		h := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(h); err != nil {
			panic(err)
		}
		if _, err := tw.Write(body); err != nil {
			panic(err)
		}
	}
	if err := tw.Close(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestStartStaticServer(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	srv, errs, err := startStaticServer(dir, 0)
	if err != nil {
		t.Fatalf("startStaticServer returned error: %v", err)
	}
	if srv != nil {
		_ = srv.Shutdown(context.Background())
	}
	if errs != nil {
		select {
		case <-errs:
		default:
		}
	}
}
