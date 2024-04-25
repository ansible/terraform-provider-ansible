package provider

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ansibleVault       = "ansible_vault"
	ansibleVaultBinary = "ansible-vault"
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
	dataParser := providerutils.ResourceDataParser{
		Data:   data,
		Detail: ansibleVault,
	}
	// required settings
	var vaultFile, vaultPasswordFile, vaultID string

	dataParser.ReadString("vault_file", &vaultFile)
	dataParser.ReadString("vault_password_file", &vaultPasswordFile)
	dataParser.ReadString("vault_id", &vaultID)

	if dataParser.HasError() {
		return dataParser.Diags
	}

	data.SetId(vaultFile)

	args := []string{"view"}

	// Compute arguments (args)
	if vaultID != "" {
		args = append(args, []string{"--vault-id", fmt.Sprintf("%s@%s", vaultID, vaultPasswordFile), vaultFile}...)
	} else {
		args = append(args, []string{"--vault-password-file", vaultPasswordFile, vaultFile}...)
	}

	tflog.Info(ctx, fmt.Sprintf("ARGS = %v", args))

	var diags diag.Diagnostics
	if err := data.Set("args", args); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("couldn't calculate 'args' variable! %s", err),
			Detail:   ansibleVault,
		})

		return diags
	}

	diags = append(diags, resourceVaultRead(ctx, data, meta)...)

	return diags
}

func resourceVaultRead(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	dataParser := providerutils.ResourceDataParser{
		Data:   data,
		Detail: ansibleVault,
	}

	// required settings
	var vaultFile, vaultPasswordFile string
	var args []string

	dataParser.ReadString("vault_file", &vaultFile)
	dataParser.ReadString("vault_password_file", &vaultPasswordFile)
	dataParser.ReadStringList("args", &args)

	if dataParser.HasError() {
		return dataParser.Diags
	}

	tflog.Info(ctx, fmt.Sprintf("vault_file = %s, vault_password_file = %s\n", vaultFile, vaultPasswordFile))

	// Validate ansible-vault binary
	_, validateBinPath := exec.LookPath(ansibleVaultBinary)
	if validateBinPath != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ansible-vault]: couldn't find executable %s: %v", ansibleVaultBinary, validateBinPath),
		})

		return diags
	}

	cmd := exec.Command(ansibleVaultBinary, args...)

	yamlString, err := cmd.CombinedOutput()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  string(yamlString),
			Detail:   ansibleVault,
		})
	}

	if err := data.Set("yaml", string(yamlString)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("[ansible-vault]: couldn't calculate 'yaml' variable! %s", err),
			Detail:   ansibleVault,
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
