package framework

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// VaultRunner abstracts the ansible-vault invocation so tests can inject a mock.
type VaultRunner interface {
	View(ctx context.Context, passwordFile, vaultID, vaultFile string) (string, diag.Diagnostics)
	Decrypt(ctx context.Context, passwordFile, vaultID, encryptedContent string) (string, diag.Diagnostics)
}

type ansibleVaultRunner struct{}

// DefaultVaultRunner is the production VaultRunner that shells out to ansible-vault.
var DefaultVaultRunner VaultRunner = &ansibleVaultRunner{} //nolint:gochecknoglobals

func (r *ansibleVaultRunner) View(ctx context.Context, passwordFile, vaultID, vaultFile string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var stderr strings.Builder

	args := BuildVaultViewArgs(passwordFile, vaultID, vaultFile)
	cmd := exec.CommandContext(ctx, "ansible-vault", args...)
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if err != nil {
		diags.AddError("ansible-vault view failed", stderr.String())
		return "", diags
	}

	return string(stdout), diags
}

func (r *ansibleVaultRunner) Decrypt(ctx context.Context, passwordFile, vaultID, encryptedContent string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var stderr strings.Builder

	args := BuildVaultDecryptArgs(passwordFile, vaultID)
	cmd := exec.CommandContext(ctx, "ansible-vault", args...)
	cmd.Stdin = strings.NewReader(encryptedContent)
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if err != nil {
		diags.AddError("ansible-vault decrypt failed", stderr.String())
		return "", diags
	}

	return string(stdout), diags
}

// BuildVaultViewArgs returns the ansible-vault view arguments for the given inputs.
// When vaultID is non-empty it uses --vault-id <id>@<passwordFile>; otherwise --vault-password-file.
func BuildVaultViewArgs(passwordFile, vaultID, vaultFile string) []string {
	if vaultID != "" {
		return []string{"view", "--vault-id", vaultID + "@" + passwordFile, vaultFile}
	}
	return []string{"view", "--vault-password-file", passwordFile, vaultFile}
}

// BuildVaultDecryptArgs returns the ansible-vault decrypt arguments that read from stdin and write to stdout.
func BuildVaultDecryptArgs(passwordFile, vaultID string) []string {
	if vaultID != "" {
		return []string{"decrypt", "--vault-id", vaultID + "@" + passwordFile, "--output=-", "-"}
	}
	return []string{"decrypt", "--vault-password-file", passwordFile, "--output=-", "-"}
}

// ResolvePasswordFile returns the effective password file path from either an inline vault_password
// string or an explicit vault_password_file path.
// When vault_password is used, it is written to a temp file; the caller must invoke the returned
// cleanup func when done.
func ResolvePasswordFile(vaultPassword, vaultPasswordFile string) (string, func(), diag.Diagnostics) {
	var diags diag.Diagnostics
	noop := func() {}

	if vaultPassword != "" {
		tmpFile, err := os.CreateTemp("", "ansible-vault-pass-*")
		if err != nil {
			diags.AddError("Failed to create temp password file", err.Error())
			return "", noop, diags
		}
		if _, err := tmpFile.WriteString(vaultPassword); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			diags.AddError("Failed to write vault password to temp file", err.Error())
			return "", noop, diags
		}
		tmpFile.Close()
		path := tmpFile.Name()
		return path, func() { os.Remove(path) }, diags
	}

	return vaultPasswordFile, noop, diags
}
