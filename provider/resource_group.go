package provider

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,

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

func resourceGroupCreate(data *schema.ResourceData, meta interface{}) error {
	groupName, okay := data.Get("name").(string)
	if !okay {
		log.Print("WARNING [ansible-group]: couldn't get 'name'!")
	}

	data.SetId(groupName)

	return resourceGroupRead(data, meta)
}

func resourceGroupRead(data *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceGroupUpdate(data *schema.ResourceData, meta interface{}) error {
	return resourceGroupRead(data, meta)
}

func resourceGroupDelete(data *schema.ResourceData, meta interface{}) error {
	data.SetId("")

	return nil
}
