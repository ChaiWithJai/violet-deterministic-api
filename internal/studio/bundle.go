package studio

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

func (s *Service) BuildBundle(tenantID, jobID string) (string, []byte, bool, error) {
	job, ok := s.GetJob(tenantID, jobID)
	if !ok {
		return "", nil, false, nil
	}
	root := fmt.Sprintf("%s-%s", slugify(fallback(job.Confirmation.AppName, "generated-app")), job.JobID)
	if root == "" {
		root = job.JobID
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for _, artifact := range job.Files {
		rel := filepath.ToSlash(strings.TrimPrefix(strings.TrimSpace(artifact.Path), "/"))
		if rel == "" || rel == "." {
			continue
		}
		fullPath := filepath.ToSlash(filepath.Join(root, rel))
		body := []byte(artifact.Content)
		h := &tar.Header{
			Name: fullPath,
			Mode: 0o644,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(h); err != nil {
			return "", nil, false, err
		}
		if _, err := tw.Write(body); err != nil {
			return "", nil, false, err
		}
	}

	manifest, _ := json.MarshalIndent(job.ArtifactManifest, "", "  ")
	manifestPath := filepath.ToSlash(filepath.Join(root, "studio_artifact_manifest.json"))
	if err := tw.WriteHeader(&tar.Header{
		Name: manifestPath,
		Mode: 0o644,
		Size: int64(len(manifest)),
	}); err != nil {
		return "", nil, false, err
	}
	if _, err := tw.Write(manifest); err != nil {
		return "", nil, false, err
	}

	if err := tw.Close(); err != nil {
		return "", nil, false, err
	}
	if err := gz.Close(); err != nil {
		return "", nil, false, err
	}
	return root + ".tar.gz", buf.Bytes(), true, nil
}
