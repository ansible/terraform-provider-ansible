package provider

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	var diags diag.Diagnostics

	vaultFile, okay := data.Get("vault_file").(string)

	if !okay {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "WARNING [ansible-vault]: couldn't get 'vault_file'!",
		})
	}

	vaultPasswordFile, okay := data.Get("vault_password_file").(string)
	if !okay {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "WARNING [ansible-vault]: couldn't get 'vault_password_file'!",
		})
	}

	vaultID, okay := data.Get("vault_id").(string)
	if !okay {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "WARNING [ansible-vault]: couldn't get 'vault_id'!",
		})
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
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("ERROR [ansible-vault]: couldn't calculate 'args' variable! %s", err),
			Detail:   ansiblePlaybook,
		})
	}

	diagsFromRead := resourceVaultRead(ctx, data, meta)
	diags = append(diags, diagsFromRead...)

	return diags
}

func resourceVaultRead(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	vaultFile, okay := data.Get("vault_file").(string)

	if !okay {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "WARNING [ansible-vault]: couldn't get 'vault_file'!",
		})
	}

	vaultPasswordFile, okay := data.Get("vault_password_file").(string)
	if !okay {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "WARNING [ansible-vault]: couldn't get 'vault_password_file'!",
		})
	}

	argsTerraform, okay := data.Get("args").([]interface{})
	if !okay {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "WARNING [ansible-vault]: couldn't get 'args'!",
		})
	}

	log.Printf("LOG [ansible-vault]: vault_file = %s, vault_password_file = %s\n", vaultFile, vaultPasswordFile)

	args, diagsFromUtils := providerutils.InterfaceToString(argsTerraform)

	diags = append(diags, diagsFromUtils...)

	cmd := exec.Command("ansible-vault", args...)

	yamlString, err := cmd.CombinedOutput()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  string(yamlString),
			Detail:   ansiblePlaybook,
		})
	}

	if err := data.Set("yaml", string(yamlString)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("ERROR [ansible-vault]: couldn't calculate 'yaml' variable! %s", err),
			Detail:   ansiblePlaybook,
		})
	}

	return diags
}

func resourceVaultUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVaultRead(ctx, data, meta)
}

func resourceVaultDelete(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	data.SetId("")

	return nil
}
