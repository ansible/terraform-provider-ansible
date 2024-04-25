package provider

import (
	"context"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Name of the group.",
			},
			"children": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of group children.",
			},
			"variables": {
				Type:        schema.TypeMap,
				Required:    false,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Map of variables.",
			},
		},
	}
}

func resourceGroupCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	dataParser := providerutils.ResourceDataParser{
		Data:   data,
		Detail: "ansible_group",
	}
	// required settings
	var groupName string

	dataParser.ReadString("name", &groupName)

	if dataParser.HasError() {
		return dataParser.Diags
	}

	data.SetId(groupName)

	resourceGroupRead(ctx, data, meta)

	return diags
}

func resourceGroupRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceGroupUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceGroupRead(ctx, data, meta)
}

func resourceGroupDelete(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	data.SetId("")

	return nil
}
