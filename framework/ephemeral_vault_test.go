package framework_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Ephemeral resources are never stored in state. Content is validated via a
// terraform_data lifecycle precondition, which is the correct Terraform 1.10+
// mechanism for asserting on ephemeral values inside a managed resource.

func TestVaultEphemeralResource_decryptsContent(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("hello: world\n")),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault" "test" {
  vault_file          = "/fake/vault.yml"
  vault_password_file = "/fake/pass"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault.test.yaml == "hello: world\n"
      error_message = "Unexpected content from vault file ephemeral resource"
    }
  }
}`,
				Check: resource.TestCheckResourceAttrSet("terraform_data.check", "id"),
			},
		},
	})
}

func TestVaultEphemeralResource_withVaultID(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("secret: value\n")),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault" "test" {
  vault_file          = "/fake/vault.yml"
  vault_password_file = "/fake/pass"
  vault_id            = "prod"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault.test.yaml == "secret: value\n"
      error_message = "Unexpected content from vault file ephemeral resource with vault_id"
    }
  }
}`,
				Check: resource.TestCheckResourceAttrSet("terraform_data.check", "id"),
			},
		},
	})
}

func TestVaultEphemeralResource_withVaultPassword(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("hello: world\n")),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault" "test" {
  vault_file     = "/fake/vault.yml"
  vault_password = "mypassword"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault.test.yaml == "hello: world\n"
      error_message = "Unexpected content from vault file ephemeral resource with vault_password"
    }
  }
}`,
				Check: resource.TestCheckResourceAttrSet("terraform_data.check", "id"),
			},
		},
	})
}

func TestVaultEphemeralResource_missingBothPasswordOptions(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(okRunner("")),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault" "test" {
  vault_file = "/fake/vault.yml"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault.test.yaml != ""
      error_message = "Should not reach here"
    }
  }
}`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

func TestVaultEphemeralResource_propagatesDecryptError(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleVaultProviderFactories(errRunner()),
		Steps: []resource.TestStep{
			{
				Config: `
ephemeral "ansible_vault" "test" {
  vault_file          = "/fake/vault.yml"
  vault_password_file = "/fake/wrong-pass"
}

resource "terraform_data" "check" {
  lifecycle {
    precondition {
      condition     = ephemeral.ansible_vault.test.yaml != ""
      error_message = "Should not reach here"
    }
  }
}`,
				ExpectError: regexp.MustCompile(`ansible-vault view failed`),
			},
		},
	})
}
