package provider

import (
	"log"
	"os/exec"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceVault() *schema.Resource {
	return &schema.Resource{
		Create: resourceVaultCreate,
		Read:   resourceVaultRead,
		Update: resourceVaultUpdate,
		Delete: resourceVaultDelete,

		Schema: map[string]*schema.Schema{
			"vault_file": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Path to encrypted vault file.",
			},
			"vault_password_file": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Path to vault password file.",
			},

			"vault_id": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "",
				Description: "ID of the encrypted vault file.",
			},

			// computed
			"yaml": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// computed - for debug
			"args": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVaultCreate(data *schema.ResourceData, meta interface{}) error {
	vaultFile, okay := data.Get("vault_file").(string)
	if !okay {
		log.Print("WARNING [ansible-vault]: couldn't get 'vault_file'!")
	}

	vaultPasswordFile, okay := data.Get("vault_password_file").(string)
	if !okay {
		log.Print("WARNING [ansible-vault]: couldn't get 'vault_password_file'!")
	}

	vaultID, okay := data.Get("vault_id").(string)
	if !okay {
		log.Print("WARNING [ansible-vault]: couldn't get 'vault_id'!")
	}

	data.SetId(vaultFile)

	var args interface{}

	// Compute arguments (args)
	if vaultID != "" {
		args = []string{
			"view",
			"--vault-id",
			vaultID + "@" + vaultPasswordFile,
			vaultFile,
		}
	} else {
		args = []string{
			"view",
			"--vault-password-file",
			vaultPasswordFile,
			vaultFile,
		}
	}

	log.Print("LOG [ansible-vault]: ARGS")
	log.Print(args)

	if err := data.Set("args", args); err != nil {
		log.Fatalf("ERROR [ansible-vault]: couldn't calculate 'args' variable! %s", err)
	}

	return resourceVaultRead(data, meta)
}

func resourceVaultRead(data *schema.ResourceData, meta interface{}) error {
	vaultFile, okay := data.Get("vault_file").(string)
	if !okay {
		log.Print("WARNING [ansible-vault]: couldn't get 'vault_file'!")
	}

	vaultPasswordFile, okay := data.Get("vault_password_file").(string)
	if !okay {
		log.Print("WARNING [ansible-vault]: couldn't get 'vault_password_file'!")
	}

	argsTerraform, okay := data.Get("args").([]interface{})
	if !okay {
		log.Print("WARNING [ansible-vault]: couldn't get 'args'!")
	}

	log.Printf("LOG [ansible-vault]: vault_file = %s, vault_password_file = %s\n", vaultFile, vaultPasswordFile)

	args := providerutils.InterfaceToString(argsTerraform)

	cmd := exec.Command("ansible-vault", args...)

	yamlString, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("ERROR [ansible-vault]: couldn't access ansible vault file%s with "+
			"password file %s! %v", vaultFile, vaultPasswordFile, err)
	}

	if err := data.Set("yaml", string(yamlString)); err != nil {
		log.Fatalf("ERROR [ansible-vault]: couldn't calculate 'yaml' variable! %s", err)
	}

	return nil
}

func resourceVaultUpdate(data *schema.ResourceData, meta interface{}) error {
	return resourceVaultRead(data, meta)
}

func resourceVaultDelete(data *schema.ResourceData, meta interface{}) error {
	data.SetId("")

	return nil
}
