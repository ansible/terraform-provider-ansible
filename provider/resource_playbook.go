package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ansiblePlaybookBinary = "ansible-playbook"

func resourcePlaybook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePlaybookCreate,
		ReadContext:   resourcePlaybookRead,
		UpdateContext: resourcePlaybookUpdate,
		DeleteContext: resourcePlaybookDelete,

		Schema: map[string]*schema.Schema{
			// Required settings
			"playbook": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Path to ansible playbook.",
			},

			// Optional settings
			"ansible_playbook_binary": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     ansiblePlaybookBinary,
				Description: "Path to ansible-playbook executable (binary).",
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Name of the desired host on which the playbook will be executed.",
			},

			"groups": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of desired groups of hosts on which the playbook will be executed.",
			},

			"replayable": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  true,
				Description: "" +
					"If 'true', the playbook will be executed on every 'terraform apply' and with that, the resource" +
					" will be recreated. " +
					"If 'false', the playbook will be executed only on the first 'terraform apply'. " +
					"Note, that if set to 'true', when doing 'terraform destroy', it might not show in the destroy " +
					"output, even though the resource still gets destroyed.",
			},

			"ignore_playbook_failure": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
				Description: "This parameter is good for testing. " +
					"Set to 'true' if the desired playbook is meant to fail, " +
					"but still want the resource to run successfully.",
			},

			// ansible execution commands
			"verbosity": { // verbosity is between = (0, 6)
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				Default:  0,
				Description: "A verbosity level between 0 and 6. " +
					"Set ansible 'verbose' parameter, which causes Ansible to print more debug messages. " +
					"The higher the 'verbosity', the more debug details will be printed.",
			},

			"tags": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of tags of plays and tasks to run.",
			},

			"limit": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of hosts to exclude from the playbook execution.",
			},

			"check_mode": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
				Description: "If 'true', playbook execution won't make any changes but " +
					"only change predictions will be made.",
			},

			"diff_mode": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
				Description: "" +
					"If 'true', when changing (small) files and templates, differences in those files will be shown. " +
					"Recommended usage with 'check_mode'.",
			},

			// connection configs are handled with extra_vars
			"force_handlers": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Default:     false,
				Description: "If 'true', run handlers even if a task fails.",
			},

			// become configs are handled with extra_vars --> these are also connection configs
			"extra_vars": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "A map of additional variables as: { key-1 = value-1, key-2 = value-2, ... }.",
			},

			"var_files": { // adds @ at the beginning of filename
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of variable files.",
			},

			// Ansible Vault
			"vault_files": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of vault files.",
			},

			"vault_password_file": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "",
				Description: "Path to a vault password file.",
			},

			"vault_id": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "",
				Description: "ID of the desired vault(s).",
			},

			// computed
			// debug output
			"args": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "Used to build arguments to run Ansible playbook with.",
			},

			"temp_inventory_file": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Path to created temporary inventory file.",
			},

			"ansible_playbook_stdout": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An ansible-playbook CLI stdout output.",
			},

			"ansible_playbook_stderr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An ansible-playbook CLI stderr output.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute), //nolint:gomnd
		},
	}
}

type PlaybookModel struct {
	Playbook              string
	AnsiblePlaybookBinary string
	Name                  string
	Groups                []string
	Replayable            bool
	IgnorePlaybookFailure bool
	Verbosity             int
	Tags                  []string
	Limit                 []string
	CheckMode             bool
	DiffMode              bool
	ForceHandlers         bool
	ExtraVars             map[string]string
	VarFiles              []string
	VaultFiles            []string
	VaultPasswordFile     string
	VaultID               string
}

func (d *PlaybookModel) ReadTerraformResourceData(data *schema.ResourceData) diag.Diagnostics {
	dataParser := providerutils.ResourceDataParser{
		Data:   data,
		Detail: "ansible_playbook",
	}

	dataParser.ReadString("ansible_playbook_binary", &d.AnsiblePlaybookBinary)
	dataParser.ReadString("playbook", &d.Playbook)
	dataParser.ReadString("name", &d.Name)
	dataParser.ReadString("vault_password_file", &d.VaultPasswordFile)
	dataParser.ReadString("vault_id", &d.VaultID)
	dataParser.ReadInt("verbosity", &d.Verbosity)
	dataParser.ReadBool("check_mode", &d.CheckMode)
	dataParser.ReadBool("diff_mode", &d.DiffMode)
	dataParser.ReadBool("force_handlers", &d.ForceHandlers)
	dataParser.ReadBool("replayable", &d.Replayable)
	dataParser.ReadBool("ignore_playbook_failure", &d.IgnorePlaybookFailure)

	dataParser.ReadStringList("tags", &d.Tags)
	dataParser.ReadStringList("groups", &d.Groups)
	dataParser.ReadStringList("limit", &d.Limit)
	dataParser.ReadStringList("var_files", &d.VarFiles)
	dataParser.ReadStringList("vault_files", &d.VaultFiles)

	dataParser.ReadMapString("extra_vars", &d.ExtraVars)

	return dataParser.Diags
}

func appendArg[T bool | string](args []string, argKey string, data T) []string {
	result := args
	switch t := any(data).(type) {
	case bool:
		if t {
			result = append(result, argKey)
		}
	case string:
		if t != "" {
			result = append(result, fmt.Sprintf("%s '%s'", argKey, t))
		}
	}
	return result
}

func appendListArg(args []string, argKey string, data []string) []string {
	result := args
	if len(data) > 0 {
		result = append(result, fmt.Sprintf("%s %s", argKey, strings.Join(data, ",")))
	}
	return result
}

func appendFilesListArg(args []string, argKey string, dataFiles []string) []string {
	result := args
	if len(dataFiles) > 0 {
		for _, iFile := range dataFiles {
			result = append(result, fmt.Sprintf("%s @%s", argKey, iFile))
		}
	}
	return result
}

func (d *PlaybookModel) BuildArgs() ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	args := []string{fmt.Sprintf("-e hostname=%s", d.Name)}

	// Verbosity
	verbose := providerutils.CreateVerboseSwitch(d.Verbosity)
	if verbose != "" {
		args = append(args, verbose)
	}
	// Force handlers
	args = appendArg(args, "--force-handlers", d.ForceHandlers)
	// Tags
	args = appendListArg(args, "--tags", d.Tags)
	// Limit
	args = appendListArg(args, "--limit", d.Limit)
	// Check mode
	args = appendArg(args, "--check", d.CheckMode)
	// Diff mode
	args = appendArg(args, "--diff", d.DiffMode)
	// Var Files
	args = appendFilesListArg(args, "-e", d.VarFiles)
	// Ansible vault
	args = appendFilesListArg(args, "-e", d.VaultFiles)
	if len(d.VaultFiles) > 0 {
		if d.VaultPasswordFile == "" {
			diags = append(diags, diag.Errorf("can't access vault file(s)! Missing 'vault_password_file'!")...)
			return nil, diags
		}
		args = append(args, fmt.Sprintf("--vault-id %s@%s", d.VaultID, d.VaultPasswordFile))
	}

	// Extra Vars
	if len(d.ExtraVars) > 0 {
		for key, val := range d.ExtraVars {
			args = append(args, fmt.Sprintf("-e %s='%s'", key, val))
		}
	}

	args = append(args, d.Playbook)

	return args, diags
}

func resourcePlaybookCreate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Generate ID
	data.SetId(time.Now().String())
	return runPlaybook(ctx, data, true)
}

func resourcePlaybookRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourcePlaybookUpdate(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return runPlaybook(ctx, data, false)
}

func runPlaybook(ctx context.Context, data *schema.ResourceData, fromCreate bool) diag.Diagnostics {
	var resourceInfo PlaybookModel
	var diags diag.Diagnostics

	diags = append(diags, resourceInfo.ReadTerraformResourceData(data)...)
	if diags.HasError() {
		return diags
	}

	// replayable=true, the playbook is always executed
	// replayable=false, the playbook is executed only in the first apply (fromCreate=true)
	if !resourceInfo.Replayable && !fromCreate {
		return diags
	}

	// Build command args
	args, diagsArgs := resourceInfo.BuildArgs()
	diags = append(diags, diagsArgs...)
	if diags.HasError() {
		return diags
	}

	tflog.Info(ctx, fmt.Sprintf("Ansible ARGS = %v", args))

	inventoryFileNamePrefix := ".inventory-"
	tempInventoryFile, diagsFromUtils := providerutils.BuildPlaybookInventory(
		inventoryFileNamePrefix+"*.ini",
		resourceInfo.Name,
		-1,
		resourceInfo.Groups,
	)
	diags = append(diags, diagsFromUtils...)

	tflog.Debug(ctx, fmt.Sprintf("Temp Inventory File: %s", tempInventoryFile))

	// ********************************* RUN PLAYBOOK ********************************

	// Validate ansible-playbook binary
	tflog.Info(ctx, fmt.Sprintf("Look ansible-playbook binary path [%s]", resourceInfo.AnsiblePlaybookBinary))
	_, validateBinPath := exec.LookPath(resourceInfo.AnsiblePlaybookBinary)
	if validateBinPath != nil {
		errorDiags := diag.Errorf("couldn't find executable %s: %v", resourceInfo.AnsiblePlaybookBinary, validateBinPath)
		diags = append(diags, errorDiags...)

		return diags
	}

	args = append(args, "-i", tempInventoryFile)

	runAnsiblePlay := exec.Command(ansiblePlaybookBinary, args...)

	tflog.Info(ctx, fmt.Sprintf("Running command <%s> and waiting for it to finish...", runAnsiblePlay.String()))
	runAnsiblePlayOut, runAnsiblePlayErr := runAnsiblePlay.CombinedOutput()
	tflog.Info(ctx, fmt.Sprintf("Command stdout = %s", runAnsiblePlayOut))

	var ansiblePlayStderrString string

	if runAnsiblePlayErr != nil {
		playbookFailMsg := string(runAnsiblePlayOut)
		if !resourceInfo.IgnorePlaybookFailure {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  playbookFailMsg,
				Detail:   "ansible-playbook",
			})
			return diags
		}

		tflog.Info(ctx, playbookFailMsg)

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  playbookFailMsg,
			Detail:   "ansible-playbook",
		})

		ansiblePlayStderrString = runAnsiblePlayErr.Error()
	}

	tflog.Info(ctx, fmt.Sprintf("Command stderr = %s", ansiblePlayStderrString))

	// Remove temporary file
	diags = append(diags, providerutils.RemoveFile(tempInventoryFile)...)

	if err := data.Set("args", args); err != nil {
		diags = append(diags, diag.Errorf("couldn't set 'args'! %v", err)...)
	}

	// Set the ansible_playbook_stdout to the CLI stdout of call "ansible-playbook" command above
	if err := data.Set("ansible_playbook_stdout", string(runAnsiblePlayOut)); err != nil {
		diags = append(diags, diag.Errorf("couldn't set 'ansible_playbook_stdout'")...)
	}

	// Set the ansible_playbook_stderr to the CLI stderr of call "ansible-playbook" command above
	if err := data.Set("ansible_playbook_stderr", ansiblePlayStderrString); err != nil {
		diags = append(diags, diag.Errorf("couldn't set 'ansible_playbook_stderr' ")...)
	}

	if err := data.Set("temp_inventory_file", ""); err != nil {
		diags = append(diags, diag.Errorf("couldn't set 'temp_inventory_file'!")...)
	}

	// *******************************************************************************

	// NOTE: Calling `resourcePlaybookRead` will make a call to `resourcePlaybookDelete` which sets
	//		 data.SetId(""), so when replayable is true, the resource gets created and then immediately deleted.
	//		 This causes provider to fail, therefore we essentially can't call data.SetId("") during a create task

	return diags
}

func resourcePlaybookDelete(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	data.SetId("")

	return nil
}
