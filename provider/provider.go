package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider exported function.
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"ansible_playbook": resourcePlaybook(),
			"ansible_vault":    resourceVault(),
			"ansible_host":     resourceHost(),
			// Disabled: below use V1, not compatible with V2 Provider
			// "ansible_group":    resourceGroup(),
		},
	}
}
