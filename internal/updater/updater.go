package updater

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultOwner          = "naodEthiop"
	defaultRepo           = "lalibela-cli"
	defaultProjectName    = "lalibela"
	checksumFileName      = "checksums.txt"
	checksumSignatureName = "checksums.txt.sig"
)

var ErrAlreadyLatest = errors.New("already up to date")

type Options struct {
	CurrentVersion string
	Owner          string
	Repo           string
	ExecutablePath string
	HTTPClient     *http.Client
}

type releaseResponse struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func SelfUpdate(opts Options) error {
	if strings.TrimSpace(opts.CurrentVersion) == "" {
		return errors.New("current version is required")
	}

	if strings.TrimSpace(opts.ExecutablePath) == "" {
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolving executable path: %w", err)
		}
		opts.ExecutablePath = exePath
	}
	if strings.TrimSpace(opts.Owner) == "" {
		opts.Owner = defaultOwner
	}
	if strings.TrimSpace(opts.Repo) == "" {
		opts.Repo = defaultRepo
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	latest, err := fetchLatestRelease(opts)
	if err != nil {
		return err
	}

	cmp, err := compareSemVer(opts.CurrentVersion, latest.TagName)
	if err != nil {
		return fmt.Errorf("comparing versions: %w", err)
	}
	if cmp >= 0 {
		return ErrAlreadyLatest
	}

	archiveAsset, err := findArchiveAsset(latest.Assets, latest.TagName, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	checksumsAsset, err := findAssetByName(latest.Assets, checksumFileName)
	if err != nil {
		return err
	}
	signatureAsset, err := findAssetByName(latest.Assets, checksumSignatureName)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "lalibela-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archiveAsset.Name)
	checksumPath := filepath.Join(tmpDir, checksumsAsset.Name)
	signaturePath := filepath.Join(tmpDir, signatureAsset.Name)

	if err := downloadFile(opts.HTTPClient, archiveAsset.BrowserDownloadURL, archivePath); err != nil {
		return err
	}
	if err := downloadFile(opts.HTTPClient, checksumsAsset.BrowserDownloadURL, checksumPath); err != nil {
		return err
	}
	if err := downloadFile(opts.HTTPClient, signatureAsset.BrowserDownloadURL, signaturePath); err != nil {
		return err
	}

	if err := verifyChecksumSignature(checksumPath, signaturePath); err != nil {
		return err
	}
	if err := verifyFileChecksum(archivePath, archiveAsset.Name, checksumPath); err != nil {
		return err
	}

	extractedBinary, err := extractBinary(archivePath, tmpDir)
	if err != nil {
		return err
	}

	if err := replaceExecutableAtomically(extractedBinary, opts.ExecutablePath); err != nil {
		return err
	}
	return nil
}

func fetchLatestRelease(opts Options) (releaseResponse, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", opts.Owner, opts.Repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return releaseResponse{}, fmt.Errorf("building latest release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "lalibela-updater")

	res, err := opts.HTTPClient.Do(req)
	if err != nil {
		return releaseResponse{}, fmt.Errorf("fetching latest release: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4*1024))
		return releaseResponse{}, fmt.Errorf("latest release request failed: status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var release releaseResponse
	if err := json.NewDecoder(res.Body).Decode(&release); err != nil {
		return releaseResponse{}, fmt.Errorf("decoding latest release response: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return releaseResponse{}, errors.New("latest release tag is empty")
	}
	return release, nil
}

func findArchiveAsset(assets []releaseAsset, tag, goos, goarch string) (releaseAsset, error) {
	candidates := archiveCandidates(tag, goos, goarch)
	for _, candidate := range candidates {
		asset, err := findAssetByName(assets, candidate)
		if err == nil {
			return asset, nil
		}
	}
	return releaseAsset{}, fmt.Errorf("release asset not found for %s/%s", goos, goarch)
}

func archiveCandidates(tag, goos, goarch string) []string {
	trimmed := strings.TrimSpace(tag)
	withoutV := strings.TrimPrefix(trimmed, "v")
	withV := trimmed
	if !strings.HasPrefix(withV, "v") {
		withV = "v" + withV
	}

	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}

	return []string{
		fmt.Sprintf("%s_%s_%s_%s%s", defaultProjectName, withV, goos, goarch, ext),
		fmt.Sprintf("%s_%s_%s_%s%s", defaultProjectName, withoutV, goos, goarch, ext),
	}
}

func findAssetByName(assets []releaseAsset, name string) (releaseAsset, error) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, nil
		}
	}
	return releaseAsset{}, fmt.Errorf("asset %q not found", name)
}

func downloadFile(client *http.Client, sourceURL, destinationPath string) error {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(sourceURL)), "https://") {
		return fmt.Errorf("insecure download URL: %s", sourceURL)
	}

	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("building download request for %s: %w", sourceURL, err)
	}
	req.Header.Set("User-Agent", "lalibela-updater")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", sourceURL, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed for %s: status %d", sourceURL, res.StatusCode)
	}

	out, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", destinationPath, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, res.Body); err != nil {
		return fmt.Errorf("writing %s: %w", destinationPath, err)
	}
	return nil
}

func verifyFileChecksum(filePath, assetName, checksumPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("hashing %s: %w", filePath, err)
	}
	actual := hex.EncodeToString(hasher.Sum(nil))

	expected, err := expectedChecksum(assetName, checksumPath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, expected) {
		return errors.New("checksum verification failed")
	}
	return nil
}

func expectedChecksum(assetName, checksumPath string) (string, error) {
	checksumFile, err := os.Open(checksumPath)
	if err != nil {
		return "", fmt.Errorf("opening checksum file: %w", err)
	}
	defer checksumFile.Close()

	scanner := bufio.NewScanner(checksumFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		candidate := strings.TrimPrefix(fields[len(fields)-1], "*")
		if candidate == assetName {
			return strings.TrimSpace(fields[0]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading checksum file: %w", err)
	}
	return "", fmt.Errorf("checksum not found for asset %s", assetName)
}

func verifyChecksumSignature(checksumPath, signaturePath string) error {
	keyPath, cleanup, err := resolvePublicKeyPath()
	if err != nil {
		return err
	}
	defer cleanup()

	gnupgHome, err := os.MkdirTemp("", "lalibela-gpg-*")
	if err != nil {
		return fmt.Errorf("creating gpg home: %w", err)
	}
	defer os.RemoveAll(gnupgHome)

	if err := runGPG("--batch", "--yes", "--homedir", gnupgHome, "--import", keyPath); err != nil {
		return fmt.Errorf("importing public key: %w", err)
	}
	if err := runGPG("--batch", "--yes", "--homedir", gnupgHome, "--verify", signaturePath, checksumPath); err != nil {
		return errors.New("signature verification failed")
	}
	return nil
}

func resolvePublicKeyPath() (string, func(), error) {
	if path := strings.TrimSpace(os.Getenv("LALIBELA_UPDATE_PUBLIC_KEY_PATH")); path != "" {
		return path, func() {}, nil
	}

	home, _ := os.UserHomeDir()
	defaultPath := filepath.Join(home, ".lalibela", "public.gpg")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, func() {}, nil
	}

	armored := strings.TrimSpace(os.Getenv("LALIBELA_UPDATE_PUBLIC_KEY"))
	if armored == "" {
		return "", func() {}, errors.New("missing update public key (set LALIBELA_UPDATE_PUBLIC_KEY_PATH or LALIBELA_UPDATE_PUBLIC_KEY)")
	}

	tmpDir, err := os.MkdirTemp("", "lalibela-pubkey-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("creating public key temp directory: %w", err)
	}
	path := filepath.Join(tmpDir, "public.gpg")
	if err := os.WriteFile(path, []byte(armored), 0o600); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", func() {}, fmt.Errorf("writing public key file: %w", err)
	}
	return path, func() { _ = os.RemoveAll(tmpDir) }, nil
}

func runGPG(args ...string) error {
	cmd := exec.Command("gpg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gpg %v failed: %s", args, strings.TrimSpace(string(out)))
	}
	return nil
}

func extractBinary(archivePath, destinationDir string) (string, error) {
	lower := strings.ToLower(archivePath)
	if strings.HasSuffix(lower, ".zip") {
		return extractBinaryFromZip(archivePath, destinationDir)
	}
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return extractBinaryFromTarGz(archivePath, destinationDir)
	}
	return "", fmt.Errorf("unsupported archive format: %s", archivePath)
}

func extractBinaryFromZip(archivePath, destinationDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("opening zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		base := filepath.Base(file.Name)
		if base != "lalibela" && base != "lalibela.exe" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("opening zip member %s: %w", file.Name, err)
		}
		defer rc.Close()

		outPath := filepath.Join(destinationDir, base)
		out, err := os.Create(outPath)
		if err != nil {
			return "", fmt.Errorf("creating extracted binary %s: %w", outPath, err)
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			return "", fmt.Errorf("extracting binary from zip: %w", err)
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return outPath, nil
	}
	return "", errors.New("binary not found in zip archive")
}

func extractBinaryFromTarGz(archivePath, destinationDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("opening tar.gz archive: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("reading gzip archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("reading tar archive: %w", err)
		}

		base := filepath.Base(header.Name)
		if base != "lalibela" && base != "lalibela.exe" {
			continue
		}
		outPath := filepath.Join(destinationDir, base)
		out, err := os.Create(outPath)
		if err != nil {
			return "", fmt.Errorf("creating extracted binary %s: %w", outPath, err)
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return "", fmt.Errorf("extracting binary from tar.gz: %w", err)
		}
		if err := out.Close(); err != nil {
			return "", err
		}
		return outPath, nil
	}
	return "", errors.New("binary not found in tar.gz archive")
}

func replaceExecutableAtomically(extractedPath, executablePath string) error {
	stat, err := os.Stat(executablePath)
	if err != nil {
		return fmt.Errorf("stat existing executable: %w", err)
	}
	mode := stat.Mode().Perm()
	if mode == 0 {
		mode = 0o755
	}

	if runtime.GOOS == "windows" {
		return replaceWindowsExecutable(extractedPath, executablePath)
	}

	tmp := executablePath + ".new"
	if err := copyFile(extractedPath, tmp, mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, executablePath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replacing executable: %w", err)
	}
	return nil
}

func replaceWindowsExecutable(extractedPath, executablePath string) error {
	stagedPath := executablePath + ".new"
	if err := copyFile(extractedPath, stagedPath, 0o755); err != nil {
		return err
	}

	scriptPath := filepath.Join(os.TempDir(), "lalibela-update-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".cmd")
	body := "@echo off\r\n" +
		"ping 127.0.0.1 -n 2 > nul\r\n" +
		"move /Y \"" + stagedPath + "\" \"" + executablePath + "\" > nul\r\n" +
		"del \"%~f0\"\r\n"
	if err := os.WriteFile(scriptPath, []byte(body), 0o600); err != nil {
		return fmt.Errorf("writing windows update script: %w", err)
	}

	cmd := exec.Command("cmd", "/C", "start", "", "/b", "cmd", "/c", scriptPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting windows update script: %w", err)
	}
	return nil
}

func copyFile(sourcePath, destinationPath string, mode os.FileMode) error {
	in, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", sourcePath, err)
	}
	defer in.Close()

	out, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("creating %s: %w", destinationPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying %s to %s: %w", sourcePath, destinationPath, err)
	}
	return nil
}

func compareSemVer(current, latest string) (int, error) {
	cv, err := parseSemVer(current)
	if err != nil {
		return 0, err
	}
	lv, err := parseSemVer(latest)
	if err != nil {
		return 0, err
	}

	if cv.major != lv.major {
		if cv.major < lv.major {
			return -1, nil
		}
		return 1, nil
	}
	if cv.minor != lv.minor {
		if cv.minor < lv.minor {
			return -1, nil
		}
		return 1, nil
	}
	if cv.patch != lv.patch {
		if cv.patch < lv.patch {
			return -1, nil
		}
		return 1, nil
	}

	if cv.pre == lv.pre {
		return 0, nil
	}
	if cv.pre == "" && lv.pre != "" {
		return 1, nil
	}
	if cv.pre != "" && lv.pre == "" {
		return -1, nil
	}
	if cv.pre < lv.pre {
		return -1, nil
	}
	return 1, nil
}

type semVer struct {
	major int
	minor int
	patch int
	pre   string
}

func parseSemVer(raw string) (semVer, error) {
	value := strings.TrimSpace(strings.TrimPrefix(raw, "v"))
	if value == "" {
		return semVer{}, fmt.Errorf("invalid version %q", raw)
	}

	core := value
	pre := ""
	if parts := strings.SplitN(value, "-", 2); len(parts) == 2 {
		core = parts[0]
		pre = parts[1]
	}

	parts := strings.Split(core, ".")
	if len(parts) < 3 {
		return semVer{}, fmt.Errorf("invalid semantic version %q", raw)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semVer{}, fmt.Errorf("invalid major version in %q", raw)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semVer{}, fmt.Errorf("invalid minor version in %q", raw)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semVer{}, fmt.Errorf("invalid patch version in %q", raw)
	}

	return semVer{major: major, minor: minor, patch: patch, pre: pre}, nil
}
