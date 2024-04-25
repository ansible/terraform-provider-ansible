package provider_test

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const ansibleVaultContent = `
some: ansible_variable
another: testing_variable_23@
`

func TestAccResourceVault(t *testing.T) {
	// Create temporary directory
	vaultDirName, err := os.MkdirTemp("", "*-vault")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(vaultDirName)
	vaultPasswordFile, vaultFile, vaultID := initTestConfiguration(vaultDirName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "ansible_vault" "secrets" {
					vault_password_file = "%s"
					vault_file          = "%s"
					vault_id            = "%s"
				}`, vaultPasswordFile, vaultFile, vaultID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_vault.secrets", "id", vaultFile),
					resource.TestCheckResourceAttr("ansible_vault.secrets", "yaml", ansibleVaultContent),
					resource.TestCheckResourceAttr("ansible_vault.secrets", "args.0", "view"),
					resource.TestCheckResourceAttr("ansible_vault.secrets", "args.1", "--vault-id"),
					resource.TestCheckResourceAttr("ansible_vault.secrets", "args.2", fmt.Sprintf("%s@%s", vaultID, vaultPasswordFile)),
					resource.TestCheckResourceAttr("ansible_vault.secrets", "args.3", vaultFile),
				),
			},
		},
	})
}

func createFile(path string, content string) {
	fileHandler, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	_, err = fileHandler.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
	if err = fileHandler.Close(); err != nil {
		log.Fatal(err)
	}
}

// This function creates the following resources
// - A vault password file used to encrypt the vault file
// - A vault encrypted file
// - A vault id use to encrypt the file.
func initTestConfiguration(vaultDirName string) (string, string, string) {
	vaultPasswordFile := filepath.Join(vaultDirName, "vault_password")
	vaultFile := filepath.Join(vaultDirName, "vault.yaml")

	// Create vault file
	createFile(vaultFile, ansibleVaultContent)

	// Create vault password file
	createFile(vaultPasswordFile, acctest.RandString(30))

	// Encrypt the vault file
	vaultId := acctest.RandomWithPrefix(acctest.RandString(10))
	args := []string{
		"encrypt",
		vaultFile,
		"--vault-password-file",
		vaultPasswordFile,
		"--vault-id",
		vaultId,
	}
	cmd := exec.Command("ansible-vault", args...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to encrypt vault file. Command = [%s] Error = [%v]", cmd.String(), err)
	}
	return vaultPasswordFile, vaultFile, vaultId
}
