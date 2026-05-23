package framework_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Ephemeral resources are never stored in state. Plaintext is validated via a
// terraform_data lifecycle precondition, which is the correct Terraform 1.10+
// mechanism for asserting on ephemeral values inside a managed resource.

func TestVaultStringEphemeralResource_decryptsPlaintext(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("supersecret")),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault_string" "test" {
  content             = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password_file = "/fake/pass"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault_string.test.plaintext == "supersecret"
      error_message = "Unexpected plaintext from vault string ephemeral resource"
    }
  }
}`,
				Check: resource.TestCheckResourceAttrSet("terraform_data.check", "id"),
			},
		},
	})
}

func TestVaultStringEphemeralResource_withVaultID(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("db_password")),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault_string" "test" {
  content             = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password_file = "/fake/pass"
  vault_id            = "prod"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault_string.test.plaintext == "db_password"
      error_message = "Unexpected plaintext from vault string ephemeral resource with vault_id"
    }
  }
}`,
				Check: resource.TestCheckResourceAttrSet("terraform_data.check", "id"),
			},
		},
	})
}

func TestVaultStringEphemeralResource_withVaultPassword(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("supersecret")),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault_string" "test" {
  content        = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password = "mypassword"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault_string.test.plaintext == "supersecret"
      error_message = "Unexpected plaintext from vault string ephemeral resource with vault_password"
    }
  }
}`,
				Check: resource.TestCheckResourceAttrSet("terraform_data.check", "id"),
			},
		},
	})
}

func TestVaultStringEphemeralResource_missingBothPasswordOptions(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("")),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault_string" "test" {
  content = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault_string.test.plaintext != ""
      error_message = "Should not reach here"
    }
  }
}`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

func TestVaultStringEphemeralResource_propagatesDecryptError(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(errRunner()),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault_string" "test" {
  content             = "$ANSIBLE_VAULT;1.1;AES256\nfakedata"
  vault_password_file = "/fake/wrong-pass"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault_string.test.plaintext != ""
      error_message = "Should not reach here"
    }
  }
}`,
				ExpectError: regexp.MustCompile(`ansible-vault view failed`),
			},
		},
	})
}
