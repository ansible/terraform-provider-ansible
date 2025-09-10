package framework

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ action.Action = (*runPlaybookAction)(nil)
)

func NewRunPlaybookAction() action.Action {
	return &runPlaybookAction{}
}

type runPlaybookAction struct{}

func (a *runPlaybookAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName
}

func (a *runPlaybookAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.UnlinkedSchema{
		Description: "",
		Attributes: map[string]schema.Attribute{
			// Required settings
			"playbook": schema.StringAttribute{
				Required:    true,
				Optional:    false,
				Description: "Path to ansible playbook.",
			},

			// Optional settings
			"ansible_playbook_binary": schema.StringAttribute{
				Required: false,
				Optional: true,
				// Default:     "ansible-playbook",
				Description: "Path to ansible-playbook executable (binary).",
			},

			"name": schema.StringAttribute{
				Required:    true,
				Optional:    false,
				Description: "Name of the desired host on which the playbook will be executed.",
			},

			"groups": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "List of desired groups of hosts on which the playbook will be executed.",
			},

			"ignore_playbook_failure": schema.BoolAttribute{
				Required: false,
				Optional: true,
				Description: "This parameter is good for testing. " +
					"Set to 'true' if the desired playbook is meant to fail, " +
					"but still want the resource to run successfully.",
			},

			// ansible execution commands
			"verbosity": schema.NumberAttribute{ // verbosity is between = (0, 6)
				Required: false,
				Optional: true,
				Description: "A verbosity level between 0 and 6. " +
					"Set ansible 'verbose' parameter, which causes Ansible to print more debug messages. " +
					"The higher the 'verbosity', the more debug details will be printed.",
			},

			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "List of tags of plays and tasks to run.",
			},

			"limit": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "List of hosts to include in playbook execution.",
			},

			"check_mode": schema.BoolAttribute{
				Required: false,
				Optional: true,
				Description: "If 'true', playbook execution won't make any changes but " +
					"only change predictions will be made.",
			},

			"diff_mode": schema.BoolAttribute{
				Required: false,
				Optional: true,
				Description: "" +
					"If 'true', when changing (small) files and templates, differences in those files will be shown. " +
					"Recommended usage with 'check_mode'.",
			},

			// connection configs are handled with extra_vars
			"force_handlers": schema.BoolAttribute{
				Required:    false,
				Optional:    true,
				Description: "If 'true', run handlers even if a task fails.",
			},

			// become configs are handled with extra_vars --> these are also connection configs
			"extra_vars": schema.MapAttribute{
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
				Required:    false,
				Optional:    true,
				Description: "A map of additional variables as: { key-1 = value-1, key-2 = value-2, ... }.",
			},

			"var_files": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "List of variable files.",
			},

			// Ansible Vault
			"vault_files": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "List of vault files.",
			},

			"vault_password_file": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Path to a vault password file.",
			},

			"vault_id": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "ID of the desired vault(s).",
			},
		},
	}
}

type runPlaybookActionModel struct {
	Playbook              types.String `tfsdk:"playbook"`
	AnsiblePlaybookBinary types.String `tfsdk:"ansible_playbook_binary"`
	Name                  types.String `tfsdk:"name"`
	Groups                types.List   `tfsdk:"groups"`
	IgnorePlaybookFailure types.Bool   `tfsdk:"ignore_playbook_failure"`
	Verbosity             types.Int64  `tfsdk:"verbosity"`
	Tags                  types.List   `tfsdk:"tags"`
	Limit                 types.List   `tfsdk:"limit"`
	CheckMode             types.Bool   `tfsdk:"check_mode"`
	DiffMode              types.Bool   `tfsdk:"diff_mode"`
	ForceHandlers         types.Bool   `tfsdk:"force_handlers"`
	ExtraVars             types.Map    `tfsdk:"extra_vars"`
	VarsFiles             types.List   `tfsdk:"vars_files"`
	VaultFiles            types.List   `tfsdk:"vault_files"`
	VaultPasswordFile     types.String `tfsdk:"vault_password_file"`
	VaultID               types.String `tfsdk:"vault_id"`
}

func (a *runPlaybookAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config runPlaybookActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate every part of the config is known
	if config.AnsiblePlaybookBinary.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("ansible_playbook_binary"), "ansible_playbook_binary is unknown", "The ansible-playbook binary is unknown, but must be known to invoke the action.")
		return
	}
	if config.Verbosity.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("verbosity"), "verbosity is unknown", "The verbosity is unknown, but must be known to invoke the action.")
		return
	}
	if config.ForceHandlers.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("force_handlers"), "force_handlers is unknown", "The force_handlers is unknown, but must be known to invoke the action.")
		return
	}
	if config.Name.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("name"), "name is unknown", "The name is unknown, but must be known to invoke the action.")
		return
	}
	if config.Limit.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("limit"), "limit is unknown", "The limit is unknown, but must be known to invoke the action.")
		return
	}
	if config.Tags.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("tags"), "tags is unknown", "The tags are unknown, but must be known to invoke the action.")
		return
	}
	if config.CheckMode.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("check_mode"), "check_mode is unknown", "The check_mode is unknown, but must be known to invoke the action.")
		return
	}
	if config.DiffMode.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("diff_mode"), "diff_mode is unknown", "The diff_mode is unknown, but must be known to invoke the action.")
		return
	}
	if config.VarsFiles.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("vars_files"), "vars_files is unknown", "The vars_files are unknown, but must be known to invoke the action.")
		return
	}
	if config.Playbook.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("playbook"), "playbook is unknown", "The playbook is unknown, but must be known to invoke the action.")
		return
	}

	// Make sure complex types are of the correct type
	var groups []types.String
	resp.Diagnostics.Append(config.Groups.ElementsAs(ctx, &groups, false)...)
	var tags []types.String
	resp.Diagnostics.Append(config.Tags.ElementsAs(ctx, &tags, false)...)
	var extraVars map[string]string
	resp.Diagnostics.Append(config.ExtraVars.ElementsAs(ctx, &extraVars, false)...)
	var varsFiles []types.String
	resp.Diagnostics.Append(config.VarsFiles.ElementsAs(ctx, &varsFiles, false)...)
	var vaultFiles []types.String
	resp.Diagnostics.Append(config.VaultFiles.ElementsAs(ctx, &vaultFiles, false)...)
	var limit []types.String
	resp.Diagnostics.Append(config.Limit.ElementsAs(ctx, &limit, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate ansible-playbook binary
	if _, validateBinPath := exec.LookPath(config.AnsiblePlaybookBinary.ValueString()); validateBinPath != nil {
		resp.Diagnostics.AddAttributeError(path.Root("ansible_playbook_binary"), "ansible_playbook_binary is not found", fmt.Sprintf("The ansible-playbook binary is not found: %s", validateBinPath))
		return
	}

	/********************
	* 	PREP THE OPTIONS (ARGS)
	 */
	args := []string{}

	verbosityLevel := int(config.Verbosity.ValueInt64())
	verbose := providerutils.CreateVerboseSwitch(verbosityLevel)
	if verbose != "" {
		args = append(args, verbose)
	}

	if config.ForceHandlers.ValueBool() {
		args = append(args, "--force-handlers")
	}
	args = append(args, "-e", "hostname="+config.Name.ValueString())

	if len(tags) > 0 {
		tmpTags := []string{}

		for _, tag := range tags {
			tmpTags = append(tmpTags, tag.ValueString())
		}

		args = append(args, "--tags", strings.Join(tmpTags, ","))
	}

	if len(limit) > 0 {
		tmpLimit := []string{}

		for _, l := range limit {
			tmpLimit = append(tmpLimit, l.ValueString())
		}

		limitStr := strings.Join(tmpLimit, ",")
		args = append(args, "--limit", limitStr)
	}

	if config.CheckMode.ValueBool() {
		args = append(args, "--check")
	}

	if config.DiffMode.ValueBool() {
		args = append(args, "--diff")
	}

	if len(varsFiles) != 0 {
		for _, varFile := range varsFiles {
			args = append(args, "-e", "@"+varFile.ValueString())
		}
	}

	// Ansible vault
	if len(vaultFiles) != 0 {
		for _, vaultFile := range vaultFiles {
			args = append(args, "-e", "@"+vaultFile.ValueString())
		}

		args = append(args, "--vault-id")

		vaultIDArg := ""
		if config.VaultID.ValueString() != "" {
			vaultIDArg += config.VaultID.ValueString()
		}

		if config.VaultPasswordFile.ValueString() != "" {
			vaultIDArg += "@" + config.VaultPasswordFile.ValueString()
		} else {
			resp.Diagnostics.AddAttributeError(path.Root("vault_password_file"), "vault_password_file is not found", "Can not access vault_files without passing the vault_password_file")
		}
		args = append(args, vaultIDArg)
	}

	if len(extraVars) != 0 {
		for key, val := range extraVars {
			args = append(args, "-e", fmt.Sprintf("%s='%s'", key, val))
		}
	}

	args = append(args, config.Playbook.ValueString())

	groupStrings := []interface{}{}
	for _, g := range groups {
		groupStrings = append(groupStrings, g.ValueString())
	}

	// Build temporary inventory file
	inventoryFileNamePrefix := ".inventory-"

	tempInventoryFile, diagsFromUtils := providerutils.BuildPlaybookInventory(
		inventoryFileNamePrefix+"*.ini",
		config.Name.ValueString(),
		-1,
		groupStrings,
	)
	// TODO: Is there an elegant way to move from sdk to framework errors?
	// resp.Diagnostics.Append(diagsFromUtils)
	if diagsFromUtils.HasError() {
		panic(diagsFromUtils)
	}
	tflog.Debug(ctx, "Temp Inventory File created", map[string]interface{}{
		"file": tempInventoryFile,
	})

	args = append(args, "-i", tempInventoryFile)
	tflog.Debug(ctx, "constructed args", map[string]interface{}{"args": args})

	tflog.Info(ctx, fmt.Sprintf("Running Command <%s %s>", config.AnsiblePlaybookBinary.ValueString(), strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, config.AnsiblePlaybookBinary.ValueString(), args...)

	// TODO: Stdout and stderr should possibly be streamed

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if verbosityLevel > 0 {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Running %s", cmd.String()),
		})
	}

	stdout, err := cmd.Output()

	if verbosityLevel > 0 {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("ansible-playbook: %s", string(stdout)),
		})
	}

	stderrStr := stderr.String()
	if err != nil {
		if len(stderrStr) > 0 {
			resp.Diagnostics.AddError(
				"ansible-playbook failed",
				stderrStr,
			)
			return
		}

		resp.Diagnostics.AddAttributeError(
			path.Root("program"),
			"Failed to execute ansible-playbook",
			err.Error(),
		)
		return
	}

	// TODO: Add diags from sdk to framework outlet
	removeFileDiags := providerutils.RemoveFile(tempInventoryFile)
	if removeFileDiags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("LOG [ansible-playbook]: failed to remove temporary inventory file: %v", removeFileDiags))
	}

}
