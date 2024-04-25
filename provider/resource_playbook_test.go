package provider_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/ansible/terraform-provider-ansible/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testBooleanArg(t *testing.T, args []string, argKey string, argValue bool) {
	if slices.Contains(args, argKey) && !argValue {
		t.Errorf("arg (%s) found while it should not.", argKey)
	}

	if !slices.Contains(args, argKey) && argValue {
		t.Errorf("missing arg (%s).", argKey)
	}
}

func TestResourcePlaybookBuildArgs(t *testing.T) {
	testTable := []struct {
		name     string
		data     provider.PlaybookModel
		expected []string
	}{
		{
			name: "Verbose_v",
			data: provider.PlaybookModel{
				Playbook:      "playbook.yaml",
				Name:          acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
				Verbosity:     1,
				DiffMode:      false,
				CheckMode:     true,
				ForceHandlers: true,
				VarFiles: []string{
					"secret_files.txt", "password_files.txt",
				},
				VaultFiles: []string{
					"variables_files.yaml",
					"configuration.yml",
				},
				VaultPasswordFile: "vault_password.txt",
			},
			expected: []string{
				"playbook.yaml",
				"-v",
				"-e @secret_files.txt",
				"-e @password_files.txt",
				"-e @variables_files.yaml",
				"-e @configuration.yml",
			},
		},
		{
			name: "Verbose_vvv",
			data: provider.PlaybookModel{
				Playbook:      "another.yml",
				Name:          acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
				Verbosity:     3,
				DiffMode:      true,
				CheckMode:     false,
				ForceHandlers: false,
				Tags: []string{
					"some_tag", "another_tag",
				},
				Limit: []string{
					"terraform", "redhat", "fedora",
				},
				ExtraVars: map[string]string{
					"first":  "variable",
					"second": "another_variable",
				},
				VaultPasswordFile: "vault_password.txt",
				VaultID:           "ansible-test-vault-id",
			},
			expected: []string{
				"another.yml",
				"-vvv",
				"--tags some_tag,another_tag",
				"--limit terraform,redhat,fedora",
				"-e first='variable'",
				"-e second='another_variable'",
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			args, diags := test.data.BuildArgs()
			if diags.HasError() {
				for i := range diags {
					if diags[i].Severity == diag.Error {
						t.Fatalf("Summary = %s - Detail = %s", diags[i].Summary, diags[i].Detail)
					}
				}
			}

			// Force handlers
			testBooleanArg(t, args, "--force-handlers", test.data.ForceHandlers)
			// Check mode
			testBooleanArg(t, args, "--check", test.data.CheckMode)
			// Diff mode
			testBooleanArg(t, args, "--diff", test.data.DiffMode)

			// Ansible vault
			vault_arg := fmt.Sprintf("--vault-id %s@%s", test.data.VaultID, test.data.VaultPasswordFile)
			if len(test.data.VaultFiles) > 0 {
				if !slices.Contains(args, vault_arg) {
					t.Errorf("Arg (%s) is missing from arguments list %v", vault_arg, args)
				}
			} else if slices.Contains(args, vault_arg) {
				t.Errorf("arg (%s) is present from arguments list %v while it should not", vault_arg, args)
			}

			expected := test.expected
			expected = append(expected, []string{fmt.Sprintf("-e hostname=%s", test.data.Name)}...)
			for _, arg := range expected {
				if !slices.Contains(args, arg) {
					t.Errorf("Arg (%s) is missing from arguments list %v", arg, args)
				}
			}
		})
	}
}

type TestResourcePlaybookConfig struct {
	PlaybookFile          string
	Name                  string
	Groups                []string
	Replayable            bool
	IgnorePlaybookFailure bool
	Verbosity             int
	Tags                  []string
	Limit                 []string
	CheckMode             bool
	DiffMode              bool
	ForceHandlers         bool
	ExtraVars             map[string]string
	VarFiles              []string
	VaultFiles            []string
	VaultPasswordFile     string
	VaultID               string
}

func TestAccResourcePlaybook_Minimal(t *testing.T) {
	dirName, err := os.MkdirTemp("", "terraform_resource_playbook_minimal_*")

	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dirName)

	playbookFile := filepath.Join(dirName, "playbook.yaml")
	testFile := filepath.Join(dirName, "iteration.txt")

	resourceConfig := TestResourcePlaybookConfig{
		PlaybookFile: playbookFile,
		Name:         acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Replayable:   true,
		Verbosity:    1,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig.createTestConfig(testFile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "name", resourceConfig.Name),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "ansible_playbook_binary", "ansible-playbook"),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "replayable", "true"),
					testAccCheckIncrement(testFile, "1"),
				),
			},
		},
	})
}

func TestAccResourcePlaybook_NotReplayable(t *testing.T) {
	dirName, err := os.MkdirTemp("", "terraform_resource_playbook_not_replayable_*")

	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dirName)

	playbookFile := filepath.Join(dirName, "playbook.yaml")
	testFile := filepath.Join(dirName, "iteration.txt")

	resourceInitial := TestResourcePlaybookConfig{
		PlaybookFile: playbookFile,
		Name:         acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Replayable:   false,
		Verbosity:    1,
	}

	resourceUpdate := TestResourcePlaybookConfig{
		PlaybookFile: playbookFile,
		Name:         acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Replayable:   false,
		Verbosity:    2,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				// Run playbook
				Config: resourceInitial.createTestConfig(testFile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "name", resourceInitial.Name),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "ansible_playbook_binary", "ansible-playbook"),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "replayable", "false"),
					testAccCheckIncrement(testFile, "1"),
				),
			},
			{
				// Run once again and ensure playbook was not replayed
				Config: resourceUpdate.createTestConfig(testFile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "name", resourceUpdate.Name),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "ansible_playbook_binary", "ansible-playbook"),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "replayable", "false"),
					testAccCheckIncrement(testFile, "1"),
				),
			},
		},
	})
}

func TestAccResourcePlaybook_Replayable(t *testing.T) {
	dirName, err := os.MkdirTemp("", "terraform_resource_playbook_replayable_*")

	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dirName)

	playbookFile := filepath.Join(dirName, "playbook.yaml")
	testFile := filepath.Join(dirName, "iteration.txt")

	resourceInitial := TestResourcePlaybookConfig{
		PlaybookFile: playbookFile,
		Name:         acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Replayable:   true,
		Verbosity:    1,
	}

	resourceUpdate := TestResourcePlaybookConfig{
		PlaybookFile: playbookFile,
		Name:         acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Replayable:   true,
		Verbosity:    2,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: TestAccProviders,
		Steps: []resource.TestStep{
			{
				// Run playbook
				Config: resourceInitial.createTestConfig(testFile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "name", resourceInitial.Name),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "ansible_playbook_binary", "ansible-playbook"),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "replayable", "true"),
					testAccCheckIncrement(testFile, "1"),
				),
			},
			{
				// Run once again and ensure playbook was not replayed
				Config: resourceUpdate.createTestConfig(testFile),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "name", resourceUpdate.Name),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "ansible_playbook_binary", "ansible-playbook"),
					resource.TestCheckResourceAttr("ansible_playbook.minimal", "replayable", "true"),
					testAccCheckIncrement(testFile, "2"),
				),
			},
		},
	})
}

func writeFile(filePath string, content string) error {
	// create file in case it does not already exists
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		fileHandler, err := os.Create(filePath)
		if err != nil {
			return err
		}

		_, err = fileHandler.WriteString(content)

		if err != nil {
			return err
		}
		if err = fileHandler.Close(); err != nil {
			return err
		}
	}

	return nil
}

// createTestConfig returns a configuration for an ansible_playbook resource.
func (d *TestResourcePlaybookConfig) createTestConfig(testFile string) string {
	// create test file
	if err := writeFile(testFile, "0"); err != nil {
		log.Fatal(err)
	}

	// create playbook file
	playbookContent := fmt.Sprintf(`
- hosts: localhost
  gather_facts: false

  tasks:
  - set_fact:
      content: "{{ lookup('file', '%s') }}"

  - name: Increment content
    copy:
      dest: '%s'
      content: "{{ content | int + 1 }}"
`, testFile, testFile)
	if err := writeFile(d.PlaybookFile, playbookContent); err != nil {
		log.Fatal(err)
	}

	// Create terraform configuration
	return fmt.Sprintf(`
resource "ansible_playbook" "minimal" {
  name = "%s"
  playbook = "%s"
  replayable = "%s"
  verbosity = %d
}`, d.Name, d.PlaybookFile, strconv.FormatBool(d.Replayable), d.Verbosity)
}

func testAccCheckIncrement(testFile string, expectedValue string) func(s *terraform.State) error {
	return func(_ *terraform.State) error {
		var content []byte
		var err error

		if content, err = os.ReadFile(testFile); err != nil {
			return err
		}
		if string(content) != expectedValue {
			return fmt.Errorf("Data differ, expected [%s] found [%s]", expectedValue, string(content))
		}

		return nil
	}
}
