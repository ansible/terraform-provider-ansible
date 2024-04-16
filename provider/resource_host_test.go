package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceHost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "ansible_host" "main" {
					name      = "localhost"
					groups    = ["some_group", "another_group"]
					variables = {
						greetings   = "from host!"
						some        = "variable"
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_host.main", "id", "localhost"),
				),
			},
		},
	})
}
