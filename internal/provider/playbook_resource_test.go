package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.#", "17"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.0", "-vvvv"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.1", "--force-handlers"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.2", "-i"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.3", "localhost,"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.4", "-i"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.5", "./.inventory.yml"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.6", "--tags"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.7", "my_tag,my_second_tag"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.8", "--limit"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.9", "my_host,other_host"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.10", "--check"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.11", "--diff"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.12", "-e"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.13", "@./.vars"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.14", "-e"),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.15", `{"my_variable":true}`),
					resource.TestCheckResourceAttr("ansible_playbook.test", "args.16", "./.test.yml"),
				),
			},
			// Ensure idempotency
			{
				Config: testAccPlaybookResourceConfig(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:      testAccPlaybookResourceConfigTimeout(),
				ExpectError: regexp.MustCompile(".*context deadline exceeded.*"),
			},
		},
	})
}

// TODO: Test Vaults
func testAccPlaybookResourceConfig() string {
	return `
	resource "ansible_playbook" "test" {
		playbook = "${path.module}/.test.yml"
		on_destroy_playbook = "${path.module}/.destruction.yml"
		ansible_playbook_binary = "echo"
		name = "localhost"
		extra_inventory_files = [ "${path.module}/.inventory.yml" ]
		verbosity = 4
		tags = [ "my_tag", "my_second_tag" ]
		limit = [ "my_host", "other_host" ]
		check_mode = true
		diff_mode = true
		force_handlers = true
		extra_vars = jsonencode({
			my_variable = true
		})
		var_files = [ "${path.module}/.vars" ]
	}
	`
}

func testAccPlaybookResourceConfigTimeout() string {
	return `
	resource "ansible_playbook" "test" {
		playbook = "${path.module}/.test.yml"
		timeout = 0.0000000000000001
		ansible_playbook_binary = "echo"
		name = "localhost"
	}
	`
}
