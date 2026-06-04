//go:build integration

package framework_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ansible/terraform-provider-ansible/framework"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	integrationPassword  = "integration-test-password"
	integrationVaultID   = "inttest"
	integrationPlaintext = "hello: from vault!\na_number: 42\n"
)

// integrationSetup creates a temp dir with a vault password file, writes a
// minimal ansible.cfg into it, and points ANSIBLE_CONFIG there for the
// duration of the test. This prevents any system-level vault IDs (from
// ~/.ansible.cfg or /etc/ansible/ansible.cfg) from leaking into the
// ansible-vault subprocesses and causing ambiguous-vault-id errors.
// The test is skipped when ansible-vault is absent.
func integrationSetup(t *testing.T) (dir, passFile string) {
	t.Helper()
	if _, err := exec.LookPath("ansible-vault"); err != nil {
		t.Skip("ansible-vault not found in PATH — skipping integration tests")
	}
	dir = t.TempDir()
	passFile = filepath.Join(dir, "vault_pass")
	if err := os.WriteFile(passFile, []byte(integrationPassword), 0o600); err != nil {
		t.Fatalf("write pass file: %v", err)
	}
	cfgFile := filepath.Join(dir, "ansible.cfg")
	if err := os.WriteFile(cfgFile, []byte("[defaults]\n"), 0o600); err != nil {
		t.Fatalf("write ansible.cfg: %v", err)
	}
	t.Setenv("ANSIBLE_CONFIG", cfgFile)
	return
}

// encryptFile writes plaintext to a file inside dir and encrypts it in-place.
func encryptFile(t *testing.T, dir, passFile, filename, plaintext string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(plaintext), 0o600); err != nil {
		t.Fatalf("write plaintext file: %v", err)
	}
	out, err := exec.Command("ansible-vault", "encrypt", "--vault-password-file", passFile, path).CombinedOutput()
	if err != nil {
		t.Fatalf("ansible-vault encrypt failed: %s", string(out))
	}
	return path
}

// encryptFileWithID encrypts a file using a vault ID label.
func encryptFileWithID(t *testing.T, dir, passFile, vaultID, filename, plaintext string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(plaintext), 0o600); err != nil {
		t.Fatalf("write plaintext file: %v", err)
	}
	out, err := exec.Command("ansible-vault", "encrypt", "--vault-id", vaultID+"@"+passFile, path).CombinedOutput()
	if err != nil {
		t.Fatalf("ansible-vault encrypt with vault-id failed: %s", string(out))
	}
	return path
}

// encryptString writes plaintext to a temp file inside dir, encrypts it in-place,
// and returns the encrypted content trimmed of any trailing newline.
func encryptString(t *testing.T, dir, passFile, plaintext string) string {
	t.Helper()
	path := filepath.Join(dir, "str.txt")
	if err := os.WriteFile(path, []byte(plaintext), 0o600); err != nil {
		t.Fatalf("write string file: %v", err)
	}
	out, err := exec.Command("ansible-vault", "encrypt", "--vault-password-file", passFile, path).CombinedOutput()
	if err != nil {
		t.Fatalf("ansible-vault encrypt failed: %s", string(out))
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read encrypted file: %v", err)
	}
	return strings.TrimRight(string(raw), "\n")
}

// encryptStringWithID encrypts a string using a vault ID label.
func encryptStringWithID(t *testing.T, dir, passFile, vaultID, plaintext string) string {
	t.Helper()
	path := filepath.Join(dir, "str_id.txt")
	if err := os.WriteFile(path, []byte(plaintext), 0o600); err != nil {
		t.Fatalf("write string file: %v", err)
	}
	out, err := exec.Command("ansible-vault", "encrypt", "--vault-id", vaultID+"@"+passFile, path).CombinedOutput()
	if err != nil {
		t.Fatalf("ansible-vault encrypt with vault-id failed: %s", string(out))
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read encrypted file: %v", err)
	}
	return strings.TrimRight(string(raw), "\n")
}

// TestAccVaultDataSource_basic decrypts a real vault-encrypted file via
// vault_password_file and verifies the yaml attribute in state.
func TestAccVaultDataSource_basic(t *testing.T) {
	dir, passFile := integrationSetup(t)
	vaultFile := encryptFile(t, dir, passFile, "secret.yml", integrationPlaintext)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(framework.DefaultVaultRunner),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ansible_vault" "test" {
  vault_file          = %q
  vault_password_file = %q
}`, vaultFile, passFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ansible_vault.test", "yaml", integrationPlaintext),
				),
			},
		},
	})
}

// TestAccVaultDataSource_withVaultID decrypts using a vault ID label.
func TestAccVaultDataSource_withVaultID(t *testing.T) {
	dir, passFile := integrationSetup(t)
	vaultFile := encryptFileWithID(t, dir, passFile, integrationVaultID, "secret_id.yml", integrationPlaintext)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(framework.DefaultVaultRunner),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ansible_vault" "test" {
  vault_file          = %q
  vault_password_file = %q
  vault_id            = %q
}`, vaultFile, passFile, integrationVaultID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ansible_vault.test", "yaml", integrationPlaintext),
				),
			},
		},
	})
}

// TestAccVaultDataSource_wrongPassword verifies that a bad password yields
// a diagnostic error and not a panic or silent failure.
func TestAccVaultDataSource_wrongPassword(t *testing.T) {
	dir, passFile := integrationSetup(t)
	vaultFile := encryptFile(t, dir, passFile, "secret_bad.yml", integrationPlaintext)
	wrongPass := filepath.Join(dir, "wrong_pass")
	if err := os.WriteFile(wrongPass, []byte("not-the-password"), 0o600); err != nil {
		t.Fatalf("write wrong pass file: %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(framework.DefaultVaultRunner),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ansible_vault" "test" {
  vault_file          = %q
  vault_password_file = %q
}`, vaultFile, wrongPass),
				ExpectError: regexp.MustCompile(`ansible-vault view failed`),
			},
		},
	})
}

// TestAccVaultStringDataSource_basic decrypts a real vault-encrypted string.
func TestAccVaultStringDataSource_basic(t *testing.T) {
	dir, passFile := integrationSetup(t)
	encrypted := encryptString(t, dir, passFile, integrationPlaintext)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(framework.DefaultVaultRunner),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ansible_vault_string" "test" {
  content             = %q
  vault_password_file = %q
}`, encrypted, passFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ansible_vault_string.test", "plaintext", integrationPlaintext),
				),
			},
		},
	})
}

// TestAccVaultStringDataSource_withVaultID decrypts using a vault ID label.
func TestAccVaultStringDataSource_withVaultID(t *testing.T) {
	dir, passFile := integrationSetup(t)
	encrypted := encryptStringWithID(t, dir, passFile, integrationVaultID, integrationPlaintext)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(framework.DefaultVaultRunner),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ansible_vault_string" "test" {
  content             = %q
  vault_password_file = %q
  vault_id            = %q
}`, encrypted, passFile, integrationVaultID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ansible_vault_string.test", "plaintext", integrationPlaintext),
				),
			},
		},
	})
}

// TestAccVaultStringDataSource_wrongPassword verifies that a bad password yields
// a diagnostic error.
func TestAccVaultStringDataSource_wrongPassword(t *testing.T) {
	dir, passFile := integrationSetup(t)
	encrypted := encryptString(t, dir, passFile, integrationPlaintext)
	wrongPass := filepath.Join(dir, "wrong_pass")
	if err := os.WriteFile(wrongPass, []byte("not-the-password"), 0o600); err != nil {
		t.Fatalf("write wrong pass file: %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(framework.DefaultVaultRunner),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "ansible_vault_string" "test" {
  content             = %q
  vault_password_file = %q
}`, encrypted, wrongPass),
				ExpectError: regexp.MustCompile(`ansible-vault decrypt failed`),
			},
		},
	})
}

// TestAccVaultFileEphemeralResource_basic decrypts a real vault-encrypted file
// through the ephemeral resource and validates the content via precondition.
func TestAccVaultFileEphemeralResource_basic(t *testing.T) {
	dir, passFile := integrationSetup(t)
	vaultFile := encryptFile(t, dir, passFile, "eph_secret.yml", integrationPlaintext)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(framework.DefaultVaultRunner),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
ephemeral "ansible_vault" "test" {
  vault_file          = %q
  vault_password_file = %q
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault.test.yaml == %q
      error_message = "Unexpected content from vault file ephemeral resource"
    }
  }
}`, vaultFile, passFile, integrationPlaintext),
				Check: resource.TestCheckResourceAttrSet("terraform_data.check", "id"),
			},
		},
	})
}

// TestAccVaultStringEphemeralResource_basic decrypts a real vault-encrypted
// string through the ephemeral resource and validates the plaintext via precondition.
func TestAccVaultStringEphemeralResource_basic(t *testing.T) {
	dir, passFile := integrationSetup(t)
	encrypted := encryptString(t, dir, passFile, integrationPlaintext)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(framework.DefaultVaultRunner),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
ephemeral "ansible_vault_string" "test" {
  content             = %q
  vault_password_file = %q
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault_string.test.plaintext == %q
      error_message = "Unexpected plaintext from vault string ephemeral resource"
    }
  }
}`, encrypted, passFile, integrationPlaintext),
				Check: resource.TestCheckResourceAttrSet("terraform_data.check", "id"),
			},
		},
	})
}

// --- Direct runner tests ---
// These test the runner functions in isolation so failures point directly at
// the subprocess layer, without going through the full Terraform provider stack.

func TestAccRunAnsibleVaultDecrypt_basic(t *testing.T) {
	dir, passFile := integrationSetup(t)
	encrypted := encryptString(t, dir, passFile, integrationPlaintext)

	plaintext, diags := framework.DefaultVaultRunner.Decrypt(context.Background(), passFile, "", encrypted)
	require.False(t, diags.HasError())
	assert.Equal(t, integrationPlaintext, plaintext)
	assert.NotContains(t, plaintext, "Decryption successful")
}

func TestAccRunAnsibleVaultDecrypt_withVaultID(t *testing.T) {
	dir, passFile := integrationSetup(t)
	encrypted := encryptStringWithID(t, dir, passFile, integrationVaultID, integrationPlaintext)

	plaintext, diags := framework.DefaultVaultRunner.Decrypt(context.Background(), passFile, integrationVaultID, encrypted)
	require.False(t, diags.HasError())
	assert.Equal(t, integrationPlaintext, plaintext)
	assert.NotContains(t, plaintext, "Decryption successful")
}

func TestAccRunAnsibleVaultDecrypt_wrongPassword(t *testing.T) {
	dir, passFile := integrationSetup(t)
	encrypted := encryptString(t, dir, passFile, integrationPlaintext)
	wrongPass := filepath.Join(dir, "wrong_pass")
	require.NoError(t, os.WriteFile(wrongPass, []byte("not-the-password"), 0o600))

	_, diags := framework.DefaultVaultRunner.Decrypt(context.Background(), wrongPass, "", encrypted)
	require.True(t, diags.HasError())
	assert.Equal(t, "ansible-vault decrypt failed", diags[0].Summary())
}

func TestAccRunAnsibleVaultView_basic(t *testing.T) {
	dir, passFile := integrationSetup(t)
	vaultFile := encryptFile(t, dir, passFile, "secret.yml", integrationPlaintext)

	plaintext, diags := framework.DefaultVaultRunner.View(context.Background(), passFile, "", vaultFile)
	require.False(t, diags.HasError())
	assert.Equal(t, integrationPlaintext, plaintext)
}

func TestAccRunAnsibleVaultView_withVaultID(t *testing.T) {
	dir, passFile := integrationSetup(t)
	vaultFile := encryptFileWithID(t, dir, passFile, integrationVaultID, "secret_id.yml", integrationPlaintext)

	plaintext, diags := framework.DefaultVaultRunner.View(context.Background(), passFile, integrationVaultID, vaultFile)
	require.False(t, diags.HasError())
	assert.Equal(t, integrationPlaintext, plaintext)
}

func TestAccRunAnsibleVaultView_wrongPassword(t *testing.T) {
	dir, passFile := integrationSetup(t)
	vaultFile := encryptFile(t, dir, passFile, "secret_bad.yml", integrationPlaintext)
	wrongPass := filepath.Join(dir, "wrong_pass")
	require.NoError(t, os.WriteFile(wrongPass, []byte("not-the-password"), 0o600))

	_, diags := framework.DefaultVaultRunner.View(context.Background(), wrongPass, "", vaultFile)
	require.True(t, diags.HasError())
	assert.Equal(t, "ansible-vault view failed", diags[0].Summary())
}
