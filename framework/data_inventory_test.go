package framework_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	errDataSourceNotFound = errors.New("data.ansible_inventory.test not found in state")
	errInvalidInventory   = errors.New("invalid inventory JSON structure")
	errValueMismatch      = errors.New("attribute value mismatch")
)

func TestInventoryDataSource_basic(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: ansibleProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
    data "ansible_inventory" "test" {
      group {
        name = "web"
        host {
          name         = "server01"
          ansible_host = "1.2.3.4"
          vars = {
            custom_var = "custom_value"
            port       = "8080"
          }
        }
        host {
          name		   = "server02"
          ansible_host = "5.6.7.8"
        }
      }
    }
    `,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ansible_inventory.test", "json"),
					func(s *terraform.State) error {
						rs, exists := s.RootModule().Resources["data.ansible_inventory.test"]
						if !exists {
							return errDataSourceNotFound
						}

						jsonVal := rs.Primary.Attributes["json"]
						var inventory map[string]any
						if err := json.Unmarshal([]byte(jsonVal), &inventory); err != nil {
							return fmt.Errorf("%w: failed to unmarshal: %w", errInvalidInventory, err)
						}

						webGroup, foundGroup := inventory["web"].(map[string]any)
						if !foundGroup {
							return fmt.Errorf("%w: web group missing or not a map", errInvalidInventory)
						}

						hosts, foundHosts := webGroup["hosts"].(map[string]any)
						if !foundHosts {
							return fmt.Errorf("%w: hosts map missing under web group", errInvalidInventory)
						}

						server1, foundServer1 := hosts["server01"].(map[string]any)
						if !foundServer1 {
							return fmt.Errorf("%w: server01 missing under hosts", errInvalidInventory)
						}

						if server1["ansible_host"] != "1.2.3.4" {
							return fmt.Errorf("%w: expected ansible_host to be 1.2.3.4, got %v", errValueMismatch, server1["ansible_host"])
						}
						if server1["custom_var"] != "custom_value" {
							return fmt.Errorf("%w: expected custom_var to be custom_value, got %v", errValueMismatch, server1["custom_var"])
						}
						if server1["port"] != "8080" {
							return fmt.Errorf("%w: expected port to be 8080, got %v", errValueMismatch, server1["port"])
						}

						server2, foundServer2 := hosts["server02"].(map[string]any)
						if !foundServer2 {
							return fmt.Errorf("%w: server02 missing under hosts", errInvalidInventory)
						}

						if server2["ansible_host"] != "5.6.7.8" {
							return fmt.Errorf("%w: expected ansible_host to be 5.6.7.8, got %v", errValueMismatch, server2["ansible_host"])
						}
						if _, hasCustomVar := server2["custom_var"]; hasCustomVar {
							return fmt.Errorf("%w: server02 should not have custom_var", errValueMismatch)
						}

						return nil
					},
				),
			},
		},
	})
}
