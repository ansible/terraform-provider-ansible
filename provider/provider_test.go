package provider_test

import (
	"os/exec"
	"testing"

	"github.com/ansible/terraform-provider-ansible/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var TestAccProviders = map[string]*schema.Provider{
	"ansible": provider.Provider(),
}

func testAccPreCheck(t *testing.T) {
	// Ensure the required executable are present
	requiredExecutables := []string{"ansible-playbook", "ansible-vault"}
	for _, binFile := range requiredExecutables {
		if _, validateBinPath := exec.LookPath(binFile); validateBinPath != nil {
			t.Fatalf("couldn't find executable %s: %v", binFile, validateBinPath)
		}
	}
}
