package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// githubRelease holds the fields we care about from the GitHub releases API.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// runUpdate self-updates the binary to the latest GitHub release.
func runUpdate() {
	localVersion := Version

	release, err := fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching latest release: %v\n", err)
		os.Exit(1)
	}

	remoteVersion := strings.TrimPrefix(release.TagName, "v")
	localVersionTrimmed := strings.TrimPrefix(localVersion, "v")

	cmp, err := compareVersions(localVersionTrimmed, remoteVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error comparing versions: %v\n", err)
		os.Exit(1)
	}
	if cmp == 0 {
		fmt.Printf("Already up to date (%s)\n", localVersion)
		checkAndReInit(".")
		return
	}
	if cmp > 0 {
		fmt.Printf("Local version (%s) is newer than latest release (%s). Nothing to do.\n", localVersion, release.TagName)
		checkAndReInit(".")
		return
	}

	// Build the asset name for the current platform and architecture.
	assetName := fmt.Sprintf("smaqit_%s_%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		fmt.Fprintf(os.Stderr, "No asset named %q found in release %s\n", assetName, release.TagName)
		os.Exit(1)
	}

	tmpFile := filepath.Join(os.TempDir(), "smaqit.new")
	if err := downloadBinary(downloadURL, tmpFile); err != nil {
		_ = os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "Error downloading binary: %v\n", err)
		os.Exit(1)
	}

	if err := os.Chmod(tmpFile, 0755); err != nil {
		_ = os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "Error setting executable bit: %v\n", err)
		os.Exit(1)
	}

	// Resolve current binary path using os.Executable (cross-platform).
	currentBin, err := os.Executable()
	if err != nil {
		_ = os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "Error detecting binary path: %v\n", err)
		os.Exit(1)
	}
	currentBin, err = filepath.EvalSymlinks(currentBin)
	if err != nil {
		_ = os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "Error resolving binary path: %v\n", err)
		os.Exit(1)
	}

	if err := replaceBinary(tmpFile, currentBin); err != nil {
		_ = os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "Error replacing binary: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated from %s to %s\n", localVersion, release.TagName)

	checkAndReInit(".")
}

// fetchLatestRelease queries the GitHub API and returns release metadata.
func fetchLatestRelease() (*githubRelease, error) {
	const apiURL = "https://api.github.com/repos/ruifrvaz/smaqit/releases/latest"

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "smaqit/"+Version)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("GitHub API rate limit reached. Try again in a few minutes, or download manually from https://github.com/ruifrvaz/smaqit/releases")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API request forbidden (status 403). Check your network or try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing GitHub API response: %w", err)
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("no tag_name in GitHub API response")
	}
	return &release, nil
}

// compareVersions compares two semver strings (without leading "v").
// Returns -1 if a < b, 0 if equal, 1 if a > b, and a non-nil error if
// either version string cannot be parsed.
func compareVersions(a, b string) (int, error) {
	parse := func(v string) ([]int, error) {
		parts := strings.Split(v, ".")
		nums := make([]int, len(parts))
		for i, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("invalid version segment %q in %q", p, v)
			}
			nums[i] = n
		}
		return nums, nil
	}

	av, aerr := parse(a)
	bv, berr := parse(b)
	if aerr != nil {
		return 0, aerr
	}
	if berr != nil {
		return 0, berr
	}

	// Pad shorter slice
	for len(av) < len(bv) {
		av = append(av, 0)
	}
	for len(bv) < len(av) {
		bv = append(bv, 0)
	}

	for i := range av {
		if av[i] < bv[i] {
			return -1, nil
		}
		if av[i] > bv[i] {
			return 1, nil
		}
	}
	return 0, nil
}

// downloadBinary downloads url into destPath.
func downloadBinary(url, destPath string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "smaqit/"+Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	return nil
}

// replaceBinary atomically replaces currentPath with tmpPath.
// Falls back to a same-filesystem copy + rename when /tmp is on a different
// filesystem than the binary (e.g., tmpfs vs ext4).
func replaceBinary(tmpPath, currentPath string) error {
	// Fast path: rename is atomic if src and dst are on the same filesystem.
	if err := os.Rename(tmpPath, currentPath); err == nil {
		return nil
	}

	// Fallback: write to a temp file in the same directory, then rename.
	sameDir := filepath.Dir(currentPath)
	tmpSameFS := filepath.Join(sameDir, fmt.Sprintf(".smaqit-%d.new", os.Getpid()))

	if err := copyFile(tmpPath, tmpSameFS); err != nil {
		_ = os.Remove(tmpSameFS)
		return fmt.Errorf("copy to same filesystem: %w", err)
	}
	if err := os.Chmod(tmpSameFS, 0755); err != nil {
		_ = os.Remove(tmpSameFS)
		return err
	}
	if err := os.Rename(tmpSameFS, currentPath); err != nil {
		_ = os.Remove(tmpSameFS)
		return fmt.Errorf("rename on same filesystem: %w", err)
	}
	_ = os.Remove(tmpPath)
	return nil
}

// copyFile copies src to dst, creating dst if necessary.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}
	return nil
}

// checkAndReInit checks whether dir contains a .smaqit/ directory. If so it
// re-runs the init command to deploy updated agents, skills, and templates.
func checkAndReInit(dir string) {
	smaqitPath := filepath.Join(dir, ".smaqit")
	if _, err := os.Stat(smaqitPath); err != nil {
		// .smaqit/ not present — skip auto-init
		fmt.Println("Run `smaqit init` to update your project assets")
		return
	}

	fmt.Println("Detected .smaqit/ — re-initializing project assets...")
	cmdInit(dir)
	fmt.Println("Re-initialized project assets")
}
