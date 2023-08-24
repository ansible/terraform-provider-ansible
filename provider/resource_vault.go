package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"os/exec"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVault() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVaultCreate,
		ReadContext:   resourceVaultRead,
		UpdateContext: resourceVaultUpdate,
		DeleteContext: resourceVaultDelete,

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
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
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

func resourceVaultCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	return resourceVaultRead(ctx, data, meta)
}

func resourceVaultRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		log.Fatalf("ERROR [ansible-vault]: couldn't access ansible vault file %s with "+
			"password file %s! %v", vaultFile, vaultPasswordFile, err)
	}

	if err := data.Set("yaml", string(yamlString)); err != nil {
		log.Fatalf("ERROR [ansible-vault]: couldn't calculate 'yaml' variable! %s", err)
	}

	return nil
}

func resourceVaultUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVaultRead(ctx, data, meta)
}

func resourceVaultDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId("")

	return nil
}
