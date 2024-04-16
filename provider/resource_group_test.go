package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "ansible_group" "main" {
					name      = "somegroup"
					children  = ["somechild"]
					variables = {
					  hello = "from group!"
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_group.main", "name", "somegroup"),
					resource.TestCheckResourceAttr("ansible_group.main", "id", "somegroup"),
				),
			},
		},
	})
}
