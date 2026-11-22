//go:build integration

package provider_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ansible/terraform-provider-ansible/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const playbookWritingMessage = `- hosts: all
  gather_facts: false
  tasks:
    - ansible.builtin.copy:
        dest: "{{ result }}"
        content: "{{ message }}"
`

// TestAccPlaybook_extraVarsUpdate checks that changing extra_vars runs the playbook
// with the new values (https://github.com/ansible/terraform-provider-ansible/issues/31).
func TestAccPlaybook_extraVarsUpdate(t *testing.T) {
	if _, err := exec.LookPath("ansible-playbook"); err != nil {
		t.Skip("ansible-playbook not found in PATH — skipping")
	}

	dir := t.TempDir()
	playbook := filepath.Join(dir, "playbook.yml")
	result := filepath.Join(dir, "result.txt")

	if err := os.WriteFile(playbook, []byte(playbookWritingMessage), 0o600); err != nil {
		t.Fatalf("write playbook: %v", err)
	}

	config := func(message string) string {
		return fmt.Sprintf(`
resource "ansible_playbook" "test" {
  name       = "localhost"
  playbook   = %q
  replayable = false

  extra_vars = {
    ansible_connection = "local"
    result             = %q
    message            = %q
  }
}
`, playbook, result, message)
	}

	resultIs := func(want string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			got, err := os.ReadFile(result)
			if err != nil {
				return fmt.Errorf("read playbook result: %w", err)
			}

			if string(got) != want {
				return fmt.Errorf("playbook wrote %q, want %q", got, want)
			}

			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"ansible": func() (*schema.Provider, error) { return provider.Provider(), nil },
		},
		Steps: []resource.TestStep{
			{
				Config: config("first"),
				Check:  resultIs("first"),
			},
			{
				Config: config("second"),
				Check:  resultIs("second"),
			},
		},
	})
}
