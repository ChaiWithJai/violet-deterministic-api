package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

type launchedAPI struct {
	runtime string
	cmd     *exec.Cmd
	done    chan error
}

func handleStudio(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(2)
	}
	switch args[0] {
	case "launch":
		handleStudioLaunch(args[1:])
	default:
		usage()
		os.Exit(2)
	}
}

func handleStudioLaunch(args []string) {
	fs := flag.NewFlagSet("studio launch", flag.ExitOnError)
	baseURL := fs.String("base-url", getenv("VDA_BASE_URL", "http://localhost:4020"), "API base URL")
	token := fs.String("token", getenv("VDA_TOKEN", "dev-token"), "bearer token")
	jobID := fs.String("job-id", "", "studio job id")
	outDir := fs.String("out-dir", filepath.Join(".", "output", "launch"), "workspace root for extracted bundles")
	apiPort := fs.Int("api-port", 8090, "generated API port")
	webPort := fs.Int("web-port", 4173, "web preview port")
	mobilePort := fs.Int("mobile-port", 4174, "mobile preview port")
	_ = fs.Parse(args)

	if strings.TrimSpace(*jobID) == "" {
		must(fmt.Errorf("--job-id is required"))
	}
	if *apiPort <= 0 || *webPort <= 0 || *mobilePort <= 0 {
		must(fmt.Errorf("ports must be positive"))
	}

	fmt.Println("[studio] downloading bundle:", *jobID)
	bundlePayload, filename, err := downloadStudioBundle(strings.TrimRight(*baseURL, "/"), *token, *jobID)
	must(err)

	workspaceRoot, appDir, err := extractStudioBundle(bundlePayload, *outDir, *jobID, filename)
	must(err)

	apiServiceDir := filepath.Join(appDir, "services", "api")
	apiURL := fmt.Sprintf("http://localhost:%d", *apiPort)
	apiProc, err := startGeneratedAPI(apiServiceDir, *jobID, *apiPort)
	must(err)
	defer func() {
		_ = stopLaunchedAPI(apiProc)
	}()

	err = waitForEndpoint(apiURL+"/health", 20*time.Second)
	must(err)

	webDir := filepath.Join(appDir, "clients", "web")
	mobileDir := filepath.Join(appDir, "clients", "mobile")
	webServer, webErrs, err := startStaticServer(webDir, *webPort)
	must(err)
	mobileServer, mobileErrs, err := startStaticServer(mobileDir, *mobilePort)
	must(err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = webServer.Shutdown(ctx)
		_ = mobileServer.Shutdown(ctx)
	}()

	fmt.Println("")
	fmt.Println("Studio launch is live")
	fmt.Println("  Workspace:   ", workspaceRoot)
	fmt.Println("  App:         ", appDir)
	fmt.Println("  API runtime: ", apiProc.runtime)
	fmt.Println("  API health:  ", apiURL+"/health")
	fmt.Println("  API tools:   ", apiURL+"/v1/tools")
	fmt.Println("  Web preview: ", fmt.Sprintf("http://localhost:%d", *webPort))
	fmt.Println("  Mobile preview:", fmt.Sprintf("http://localhost:%d", *mobilePort))
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop all launched processes.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

	for {
		select {
		case <-sig:
			return
		case err := <-apiProc.done:
			must(fmt.Errorf("generated api runtime exited: %w", err))
		case err := <-webErrs:
			if err != nil {
				must(fmt.Errorf("web preview server failed: %w", err))
			}
		case err := <-mobileErrs:
			if err != nil {
				must(fmt.Errorf("mobile preview server failed: %w", err))
			}
		}
	}
}

func downloadStudioBundle(baseURL, token, jobID string) ([]byte, string, error) {
	endpoint := fmt.Sprintf("%s/v1/studio/jobs/%s/bundle", strings.TrimRight(baseURL, "/"), strings.TrimSpace(jobID))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	req.Header.Set("Accept", "application/gzip")
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("bundle request failed (%d): %s", resp.StatusCode, trimString(string(body), 280))
	}
	filename := "bundle.tar.gz"
	if cd := strings.TrimSpace(resp.Header.Get("Content-Disposition")); cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			if v, ok := params["filename"]; ok && strings.TrimSpace(v) != "" {
				filename = strings.TrimSpace(v)
			}
		}
	}
	return body, filename, nil
}

func extractStudioBundle(payload []byte, outRoot, jobID, filename string) (string, string, error) {
	stamp := time.Now().UTC().Format("20060102-150405")
	extractRoot := filepath.Join(outRoot, fmt.Sprintf("%s-%s", sanitizePathFragment(jobID), stamp))
	if err := os.MkdirAll(extractRoot, 0o755); err != nil {
		return "", "", err
	}

	rootDir, err := untarGz(payload, extractRoot)
	if err != nil {
		return "", "", fmt.Errorf("extract %s: %w", filename, err)
	}
	if rootDir == "" {
		rootDir = extractRoot
	}

	appDir, err := findGeneratedAppDir(rootDir)
	if err != nil {
		return "", "", err
	}
	return rootDir, appDir, nil
}

func untarGz(payload []byte, destination string) (string, error) {
	gr, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	destClean := filepath.Clean(destination)
	destPrefix := destClean + string(os.PathSeparator)
	rootComponent := ""

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}

		name := filepath.Clean(strings.TrimPrefix(strings.TrimSpace(hdr.Name), "/"))
		if name == "." || name == "" {
			continue
		}
		first := strings.Split(filepath.ToSlash(name), "/")[0]
		if rootComponent == "" && first != "." {
			rootComponent = first
		}

		target := filepath.Join(destClean, name)
		targetClean := filepath.Clean(target)
		if targetClean != destClean && !strings.HasPrefix(targetClean+string(os.PathSeparator), destPrefix) {
			return "", fmt.Errorf("bundle path traversal blocked: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetClean, os.FileMode(hdr.Mode)); err != nil {
				return "", err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetClean), 0o755); err != nil {
				return "", err
			}
			f, err := os.OpenFile(targetClean, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return "", err
			}
			if err := f.Close(); err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("unsupported bundle entry type %d for %s", hdr.Typeflag, hdr.Name)
		}
	}

	if rootComponent == "" {
		return destClean, nil
	}
	return filepath.Join(destClean, rootComponent), nil
}

func findGeneratedAppDir(root string) (string, error) {
	appsDir := filepath.Join(root, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return "", fmt.Errorf("apps directory missing in bundle: %w", err)
	}

	candidates := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidates = append(candidates, entry.Name())
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no generated app directory found in %s", appsDir)
	}
	sort.Strings(candidates)
	return filepath.Join(appsDir, candidates[0]), nil
}

func startGeneratedAPI(serviceDir, jobID string, apiPort int) (*launchedAPI, error) {
	if _, err := os.Stat(filepath.Join(serviceDir, "cmd", "server", "main.go")); err != nil {
		return nil, fmt.Errorf("generated api source missing: %w", err)
	}

	if _, err := exec.LookPath("go"); err == nil {
		cmd := exec.Command("go", "run", "./cmd/server")
		cmd.Dir = serviceDir
		cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", apiPort))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
		return &launchedAPI{runtime: "go", cmd: cmd, done: done}, nil
	}

	if _, err := exec.LookPath("docker"); err == nil {
		tag := "vda-generated-" + sanitizePathFragment(jobID)
		build := exec.Command("docker", "build", "-t", tag, ".")
		build.Dir = serviceDir
		build.Stdout = os.Stdout
		build.Stderr = os.Stderr
		if err := build.Run(); err != nil {
			return nil, fmt.Errorf("docker build generated api failed: %w", err)
		}
		run := exec.Command("docker", "run", "--rm", "-p", fmt.Sprintf("%d:8090", apiPort), tag)
		run.Stdout = os.Stdout
		run.Stderr = os.Stderr
		if err := run.Start(); err != nil {
			return nil, err
		}
		done := make(chan error, 1)
		go func() {
			done <- run.Wait()
		}()
		return &launchedAPI{runtime: "docker", cmd: run, done: done}, nil
	}

	return nil, fmt.Errorf("neither go nor docker is available to run generated api")
}

func stopLaunchedAPI(proc *launchedAPI) error {
	if proc == nil || proc.cmd == nil || proc.cmd.Process == nil {
		return nil
	}
	_ = proc.cmd.Process.Signal(os.Interrupt)
	select {
	case <-proc.done:
		return nil
	case <-time.After(2 * time.Second):
		_ = proc.cmd.Process.Kill()
		select {
		case <-proc.done:
		case <-time.After(2 * time.Second):
		}
	}
	return nil
}

func startStaticServer(root string, port int) (*http.Server, chan error, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, nil, err
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("not a directory: %s", root)
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.FileServer(http.Dir(root)),
	}
	errs := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
	return srv, errs, nil
}

func waitForEndpoint(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("endpoint did not become healthy in %s: %s", timeout, url)
}

func sanitizePathFragment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "bundle"
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "bundle"
	}
	return out
}

func trimString(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max < 4 {
		return value[:max]
	}
	return value[:max-3] + "..."
}
