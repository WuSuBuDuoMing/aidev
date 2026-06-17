// Package cli implements the self-upgrade command.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	githubRepo = "user/neocode"
	upgradeTimeout = 5 * time.Minute
)

// GitHubRelease represents a GitHub release.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckForUpdate checks if a newer version is available.
func CheckForUpdate(currentVersion string) (*GitHubRelease, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("check update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, false, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("decode release: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if latest == current {
		return &release, false, nil // already up to date
	}

	return &release, true, nil
}

// Upgrade downloads and replaces the current binary.
func Upgrade(currentVersion string) error {
	release, hasUpdate, err := CheckForUpdate(currentVersion)
	if err != nil {
		return fmt.Errorf("check for update: %w", err)
	}
	if !hasUpdate {
		fmt.Printf("  Already up to date: %s\n", currentVersion)
		return nil
	}

	fmt.Printf("  New version available: %s → %s\n", currentVersion, release.TagName)
	fmt.Printf("  Downloading...\n")

	// Find the correct asset
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}

	assetName := fmt.Sprintf("neocode-%s-%s%s", goos, goarch, ext)
	downloadURL := ""

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
		// Also check for tar.gz/zip
		archiveName := fmt.Sprintf("neocode_%s_%s.tar.gz", strings.Title(goos), goarch)
		if goos == "windows" {
			archiveName = fmt.Sprintf("neocode_%s_%s.zip", strings.Title(goos), goarch)
		}
		if asset.Name == archiveName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s", goos, goarch)
	}

	// Download
	client := &http.Client{Timeout: upgradeTimeout}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	// Read into temp file
	tmpFile, err := os.CreateTemp("", "neocode-upgrade-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	hasher := sha256.New()
	writer := io.MultiWriter(tmpFile, hasher)

	written, err := io.Copy(writer, resp.Body)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("download: %w", err)
	}
	tmpFile.Close()

	checksum := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("  Downloaded: %d bytes (SHA256: %s)\n", written, checksum[:16]+"...")

	// Replace current binary
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}

	// On Windows, rename old binary first
	if goos == "windows" {
		oldPath := execPath + ".old"
		os.Remove(oldPath) // remove any previous .old
		if err := os.Rename(execPath, oldPath); err != nil {
			return fmt.Errorf("rename old binary: %w", err)
		}
	}

	// Copy new binary
	newData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("read temp: %w", err)
	}

	if err := os.WriteFile(execPath, newData, 0o755); err != nil {
		return fmt.Errorf("write binary: %w", err)
	}

	fmt.Printf("  ✓ Updated to %s\n", release.TagName)
	fmt.Printf("  Restart neocode to use the new version.\n")

	return nil
}
