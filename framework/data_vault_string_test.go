package framework_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestVaultStringDataSource_decryptsPlaintext(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("supersecret")),
		Steps: []resource.TestStep{
			{
				Config: `
data "ansible_vault_string" "test" {
  content             = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password_file = "/fake/pass"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ansible_vault_string.test", "plaintext", "supersecret"),
				),
			},
		},
	})
}

func TestVaultStringDataSource_withVaultID(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("db_password")),
		Steps: []resource.TestStep{
			{
				Config: `
data "ansible_vault_string" "test" {
  content             = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password_file = "/fake/pass"
  vault_id            = "prod"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ansible_vault_string.test", "plaintext", "db_password"),
				),
			},
		},
	})
}

func TestVaultStringDataSource_propagatesDecryptError(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(errRunner()),
		Steps: []resource.TestStep{
			{
				Config: `
data "ansible_vault_string" "test" {
  content             = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password_file = "/fake/wrong-pass"
}`,
				ExpectError: regexp.MustCompile(`ansible-vault view failed`),
			},
		},
	})
}

func TestVaultStringDataSource_withVaultPassword(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("supersecret")),
		Steps: []resource.TestStep{
			{
				Config: `
data "ansible_vault_string" "test" {
  content        = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password = "mypassword"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ansible_vault_string.test", "plaintext", "supersecret"),
				),
			},
		},
	})
}

func TestVaultStringDataSource_missingBothPasswordOptions(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("")),
		Steps: []resource.TestStep{
			{
				Config: `
data "ansible_vault_string" "test" {
  content = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
}`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

func TestVaultStringDataSource_computedPlaintextWithInputsInState(t *testing.T) {
	t.Parallel()
	// Terraform always writes config attributes into data source state (it uses them
	// to detect when a re-read is needed). Only `plaintext` is provider-computed; the
	// other attributes are echoed from config. For portability across machines use
	// the ansible_vault_string ephemeral resource instead, which stores nothing.
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("value")),
		Steps: []resource.TestStep{
			{
				Config: `
data "ansible_vault_string" "test" {
  content             = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password_file = "/fake/pass"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ansible_vault_string.test", "plaintext", "value"),
				),
			},
		},
	})
}
