package framework_test

import (
	"os"
	"testing"

	"github.com/ansible/terraform-provider-ansible/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- buildVaultViewArgs ---

func TestBuildVaultViewArgs_withoutVaultID(t *testing.T) {
	t.Parallel()
	args := framework.BuildVaultViewArgs("/tmp/pass", "", "/tmp/secret.yml")
	assert.Equal(t, []string{"view", "--vault-password-file", "/tmp/pass", "/tmp/secret.yml"}, args)
}

func TestBuildVaultViewArgs_withVaultID(t *testing.T) {
	t.Parallel()
	args := framework.BuildVaultViewArgs("/tmp/pass", "myid", "/tmp/secret.yml")
	assert.Equal(t, []string{"view", "--vault-id", "myid@/tmp/pass", "/tmp/secret.yml"}, args)
}

// --- buildVaultDecryptArgs ---

func TestBuildVaultDecryptArgs_withoutVaultID(t *testing.T) {
	t.Parallel()
	args := framework.BuildVaultDecryptArgs("/tmp/pass", "")
	assert.Equal(t, []string{"decrypt", "--vault-password-file", "/tmp/pass", "--output=-", "-"}, args)
}

func TestBuildVaultDecryptArgs_withVaultID(t *testing.T) {
	t.Parallel()
	args := framework.BuildVaultDecryptArgs("/tmp/pass", "myid")
	assert.Equal(t, []string{"decrypt", "--vault-id", "myid@/tmp/pass", "--output=-", "-"}, args)
}

// --- resolvePasswordFile ---

func TestResolvePasswordFile_withVaultPasswordFile_returnsPathDirectly(t *testing.T) {
	t.Parallel()
	path, cleanup, diags := framework.ResolvePasswordFile("", "/my/pass")
	defer cleanup()
	assert.False(t, diags.HasError())
	assert.Equal(t, "/my/pass", path)
}

func TestResolvePasswordFile_withVaultPassword_writesTempFile(t *testing.T) {
	t.Parallel()
	path, cleanup, diags := framework.ResolvePasswordFile("mypassword", "")
	defer cleanup()
	require.False(t, diags.HasError())

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "mypassword", string(content))
}

func TestResolvePasswordFile_withVaultPassword_cleanupRemovesTempFile(t *testing.T) {
	t.Parallel()
	path, cleanup, diags := framework.ResolvePasswordFile("mypassword", "")
	require.False(t, diags.HasError())

	cleanup()

	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), "temp file %q should have been removed", path)
}

func TestResolvePasswordFile_bothSet_vaultPasswordTakesPrecedence(t *testing.T) {
	t.Parallel()
	path, cleanup, diags := framework.ResolvePasswordFile("inlinepass", "/file/pass")
	defer cleanup()
	require.False(t, diags.HasError())

	// Should have written a temp file, not returned the file path directly.
	assert.NotEqual(t, "/file/pass", path)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "inlinepass", string(content))
}
