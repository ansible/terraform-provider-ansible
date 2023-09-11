package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPlaybookResouce(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPlaybookResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.#", "3"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.0", "-i"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.1", "localhost,"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.2", "./.test.yml"),
				),
			},
		},
	})
}

func testAccPlaybookResourceConfig() string {
	return `
	resource "ansible_playbook" "test" {
		playbook = "${path.module}/.test.yml"
		ansible_playbook_binary = "echo"
		name = "localhost"
	}
	`
}
