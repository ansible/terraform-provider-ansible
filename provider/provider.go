package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Provider exported function.
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"ansible_vault": resourceVault(),
			"ansible_host":  resourceHost(),
			"ansible_group": resourceGroup(),
		},
	}
}
