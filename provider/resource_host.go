package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostCreate,
		ReadContext:   resourceHostRead,
		UpdateContext: resourceHostUpdate,
		DeleteContext: resourceHostDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Name of the host.",
			},
			"groups": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of group names.",
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

func resourceHostCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	hostName, ok := data.Get("name").(string)

	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "ERROR [ansible-group]: couldn't get 'name'!",
		})
	}

	data.SetId(hostName)

	resourceHostRead(ctx, data, meta)

	return diags
}

func resourceHostRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceHostUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceHostRead(ctx, data, meta)
}

func resourceHostDelete(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	data.SetId("")

	return nil
}
