package framework_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	hostConfigCreate = `
resource "ansible_host" "test" {
  name   = "server01"
  groups = ["web", "production"]
  variables = {
    ansible_user = "admin"
    port         = 22
  }
}`

	hostConfigUpdate = `
resource "ansible_host" "test" {
  name   = "server01"
  groups = ["db"]
}`
)

// TestHostResource_lifecycle exercises Create and Update in a single unit test
// run, covering groups (types.List) and dynamic variables (types.Dynamic).
func TestHostResource_lifecycle(t *testing.T) {
	t.Parallel()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hostConfigCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_host.test", "id", "server01"),
					resource.TestCheckResourceAttr("ansible_host.test", "name", "server01"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.#", "2"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.0", "web"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.1", "production"),
					resource.TestCheckResourceAttr("ansible_host.test", "variables.%", "2"),
					resource.TestCheckResourceAttr("ansible_host.test", "variables.ansible_user", "admin"),
					resource.TestCheckResourceAttr("ansible_host.test", "variables.port", "22"),
				),
			},
			{
				Config: hostConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_host.test", "id", "server01"),
					resource.TestCheckResourceAttr("ansible_host.test", "name", "server01"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.0", "db"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.#", "1"),
					resource.TestCheckNoResourceAttr("ansible_host.test", "variables"),
				),
			},
		},
	})
}

// TestAccHostResource_basic runs the same two-step lifecycle through the full
// Terraform engine, verifying plan+apply+refresh produces no spurious diffs.
func TestAccHostResource_basic(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hostConfigCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_host.test", "id", "server01"),
					resource.TestCheckResourceAttr("ansible_host.test", "name", "server01"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.#", "2"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.0", "web"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.1", "production"),
					resource.TestCheckResourceAttr("ansible_host.test", "variables.%", "2"),
					resource.TestCheckResourceAttr("ansible_host.test", "variables.ansible_user", "admin"),
					resource.TestCheckResourceAttr("ansible_host.test", "variables.port", "22"),
				),
			},
			{
				Config: hostConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_host.test", "id", "server01"),
					resource.TestCheckResourceAttr("ansible_host.test", "name", "server01"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.0", "db"),
					resource.TestCheckResourceAttr("ansible_host.test", "groups.#", "1"),
					resource.TestCheckNoResourceAttr("ansible_host.test", "variables"),
				),
			},
		},
	})
}
