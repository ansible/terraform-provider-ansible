package provider

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceHostCreate,
		Read:   resourceHostRead,
		Update: resourceHostUpdate,
		Delete: resourceHostDelete,

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

func resourceHostCreate(data *schema.ResourceData, meta interface{}) error {
	hostName, okay := data.Get("name").(string)
	if !okay {
		log.Print("WARNING [ansible-group]: couldn't get 'name'!")
	}

	data.SetId(hostName)

	return resourceHostRead(data, meta)
}

func resourceHostRead(data *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceHostUpdate(data *schema.ResourceData, meta interface{}) error {
	return resourceHostRead(data, meta)
}

func resourceHostDelete(data *schema.ResourceData, meta interface{}) error {
	data.SetId("")

	return nil
}
