package provider

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccHostResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccHostResourceConfig("0.0.0.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						return testAccHostResourceAssertConfig(s, "0.0.0.0")
					},
				),
			},
			// Update and Read Testing
			{
				Config: testAccHostResourceConfig("127.0.0.1"),
				Check: func(s *terraform.State) error {
					return testAccHostResourceAssertConfig(s, "127.0.0.1")
				},
			},
		},
		CheckDestroy: func(s *terraform.State) error {
			host := s.RootModule().Resources["ansible_host.test"].Primary
			path := host.Attributes["inventory_path"]

			_, err := os.Stat(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return nil
				}

				return fmt.Errorf("unexpected errror: %w", err)
			}

			return errors.New("the file still exists")
		},
	})
}

func testAccHostResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
	resource "ansible_host" "test" {
		name = "%s"
		port = 1312
		groups = ["my-custom-group"]
		variables = {
			key = "value"
		}
	}
	`, configurableAttribute)
}

func testAccHostResourceAssertConfig(s *terraform.State, configurableAttribute string) error {
	host := s.RootModule().Resources["ansible_host.test"].Primary
	path := host.Attributes["inventory_path"]

	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read inventory file: %w", err)
	}

	expected := fmt.Sprintf(`my-custom-group:
  hosts:
    %s:1312:
      key: value
ungrouped:
  hosts:
    %s:1312:
      key: value
`, configurableAttribute, configurableAttribute)

	if expected != string(b) {
		return fmt.Errorf("expected: '%s', actual: '%s'", expected, string(b))
	}

	return nil
}
