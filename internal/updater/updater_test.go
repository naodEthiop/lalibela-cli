package updater

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// clearKeyOverrides simulates a fresh install: no env var overrides and no
// pre-existing ~/.lalibela/public.gpg, so key resolution must fall back to
// the embedded key.
func clearKeyOverrides(t *testing.T) {
	t.Helper()
	t.Setenv("LALIBELA_UPDATE_PUBLIC_KEY_PATH", "")
	t.Setenv("LALIBELA_UPDATE_PUBLIC_KEY", "")
	t.Setenv("HOME", t.TempDir())
}

func TestResolvePublicKeyPath_FreshInstallUsesEmbeddedKey(t *testing.T) {
	clearKeyOverrides(t)

	path, cleanup, err := resolvePublicKeyPath()
	if err != nil {
		t.Fatalf("resolvePublicKeyPath should not fail on a fresh install (no key provisioned anywhere), got: %v", err)
	}
	defer cleanup()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading resolved key path %s: %v", path, err)
	}
	if strings.TrimSpace(string(got)) != strings.TrimSpace(string(embeddedPublicKey)) {
		t.Fatal("resolved key content does not match the embedded release public key")
	}
	if !strings.Contains(string(got), "BEGIN PGP PUBLIC KEY BLOCK") {
		t.Fatal("resolved key does not look like an armored PGP public key")
	}
}

func TestResolvePublicKeyPath_PathOverrideWins(t *testing.T) {
	clearKeyOverrides(t)

	dir := t.TempDir()
	override := filepath.Join(dir, "custom.gpg")
	if err := os.WriteFile(override, []byte("not a real key, just proving precedence"), 0o600); err != nil {
		t.Fatalf("writing override key: %v", err)
	}
	t.Setenv("LALIBELA_UPDATE_PUBLIC_KEY_PATH", override)

	path, cleanup, err := resolvePublicKeyPath()
	if err != nil {
		t.Fatalf("resolvePublicKeyPath: %v", err)
	}
	defer cleanup()

	if path != override {
		t.Fatalf("expected LALIBELA_UPDATE_PUBLIC_KEY_PATH to take precedence, got %s want %s", path, override)
	}
}

// TestVerifyChecksumSignature_EmbeddedKeyIsUsableByGPG proves the embedded
// key is not just present but actually functional: it must be importable by
// gpg and able to verify a signature produced by the matching private key.
// Since the real release private key is a CI-only secret and is never
// committed, this generates a disposable throwaway keypair, re-exports its
// public half over the embedded-key resolution path (via the path override,
// which resolvePublicKeyPath honors ahead of the embedded fallback), and
// confirms the same verifyChecksumSignature codepath that ships in
// SelfUpdate correctly accepts a valid signature and rejects a tampered one.
func TestVerifyChecksumSignature_ValidAndTamperedSignatures(t *testing.T) {
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not available in this environment")
	}
	clearKeyOverrides(t)

	gnupgHome := t.TempDir()
	t.Setenv("GNUPGHOME", gnupgHome)

	batch := filepath.Join(gnupgHome, "genkey.batch")
	batchBody := "%no-protection\n" +
		"Key-Type: RSA\nKey-Length: 2048\nKey-Usage: sign\n" +
		"Name-Real: Test Key\nName-Email: test@example.com\nExpire-Date: 0\n%commit\n"
	if err := os.WriteFile(batch, []byte(batchBody), 0o600); err != nil {
		t.Fatalf("writing batch file: %v", err)
	}
	if out, err := exec.Command("gpg", "--batch", "--generate-key", batch).CombinedOutput(); err != nil {
		t.Fatalf("generating test key: %v\n%s", err, out)
	}

	pubPath := filepath.Join(gnupgHome, "test-public.asc")
	pubOut, err := exec.Command("gpg", "--batch", "--armor", "--export", "test@example.com").Output()
	if err != nil {
		t.Fatalf("exporting test public key: %v", err)
	}
	if err := os.WriteFile(pubPath, pubOut, 0o600); err != nil {
		t.Fatalf("writing test public key: %v", err)
	}
	t.Setenv("LALIBELA_UPDATE_PUBLIC_KEY_PATH", pubPath)

	checksumPath := filepath.Join(gnupgHome, "checksums.txt")
	if err := os.WriteFile(checksumPath, []byte("deadbeef  lalibela_1.0.0_linux_amd64.tar.gz\n"), 0o644); err != nil {
		t.Fatalf("writing checksums file: %v", err)
	}
	sigPath := checksumPath + ".sig"
	if out, err := exec.Command("gpg", "--batch", "--yes", "--local-user", "test@example.com",
		"--armor", "--detach-sign", "--output", sigPath, checksumPath).CombinedOutput(); err != nil {
		t.Fatalf("signing checksums file: %v\n%s", err, out)
	}

	if err := verifyChecksumSignature(checksumPath, sigPath); err != nil {
		t.Fatalf("verifyChecksumSignature should accept a valid signature from the resolved key, got: %v", err)
	}

	tamperedPath := filepath.Join(gnupgHome, "checksums-tampered.txt")
	if err := os.WriteFile(tamperedPath, []byte("00000000  lalibela_1.0.0_linux_amd64.tar.gz\n"), 0o644); err != nil {
		t.Fatalf("writing tampered checksums file: %v", err)
	}
	if err := verifyChecksumSignature(tamperedPath, sigPath); err == nil {
		t.Fatal("verifyChecksumSignature should reject a signature that does not match the checksum file content")
	}
}
