package provider

import (
	"context"
	"math/big"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PlaybookResource{}

func NewPlaybookResource() resource.Resource {
	return &PlaybookResource{}
}

// TODO: On destroy
// TODO: Copy other hashicorp APIs (ssh host, port, bastion, etc)
// TODO: Options for strict host checking, known hosts file, etc, timeout
// TODO: Finish the test

// PlaybookResource defines the resource implementation.
type PlaybookResource struct{}

type PlaybookResourceModel struct {
	Playbook              types.String `tfsdk:"playbook"`
	PlaybookSha256Sum     types.String `tfsdk:"playbook_sha256_sum"`
	AnsiblePlaybookBinary types.String `tfsdk:"ansible_playbook_binary"`
	Name                  types.String `tfsdk:"name"`
	Groups                types.List   `tfsdk:"groups"`
	ExtraInventoryFiles   types.List   `tfsdk:"extra_inventory_files"`
	Replayable            types.Bool   `tfsdk:"replayable"`
	IgnorePlaybookFailure types.Bool   `tfsdk:"ignore_playbook_failure"`
	Verbosity             types.Number `tfsdk:"verbosity"`
	Tags                  types.List   `tfsdk:"tags"`
	Limit                 types.List   `tfsdk:"limit"`
	CheckMode             types.Bool   `tfsdk:"check_mode"`
	DiffMode              types.Bool   `tfsdk:"diff_mode"`
	ForceHandlers         types.Bool   `tfsdk:"force_handlers"`
	ExtraVars             types.String `tfsdk:"extra_vars"`
	VarFiles              types.List   `tfsdk:"var_files"`
	VaultFiles            types.List   `tfsdk:"vault_files"`
	VaultPasswordFile     types.String `tfsdk:"vault_password_file"`
	VaultID               types.String `tfsdk:"vault_id"`
	Args                  types.List   `tfsdk:"args"`
	AnsiblePlaybookOutput types.String `tfsdk:"ansible_playbook_output"`
	AnsiblePlaybookErr    types.String `tfsdk:"ansible_playbook_err"`
}

func (r *PlaybookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_playbook"
}

func (r *PlaybookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Playbook resource",

		Attributes: map[string]schema.Attribute{
			"playbook": schema.StringAttribute{
				MarkdownDescription: "Path to ansible playbook.",
				Required:            true,
			},
			"playbook_sha256_sum": schema.StringAttribute{
				Computed: true,
			},
			"ansible_playbook_binary": schema.StringAttribute{
				MarkdownDescription: "Path to ansible-playbook executable (binary).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ansible-playbook"),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the desired host on which the playbook will be executed.",
				Optional:            true,
			},
			"groups": schema.ListAttribute{
				MarkdownDescription: "List of desired groups of hosts on which the playbook will be executed.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"extra_inventory_files": schema.ListAttribute{
				MarkdownDescription: "List of extra inventory files that the playbook will include, hint: use together with ansible_host.inventory_path",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"replayable": schema.BoolAttribute{
				MarkdownDescription: "" +
					"If 'true', the playbook will be executed on every 'terraform apply' and with that, the resource" +
					" will be recreated. " +
					"If 'false', the playbook will be executed only on the first 'terraform apply'. " +
					"Note, that if set to 'true', when doing 'terraform destroy', it might not show in the destroy " +
					"output, even though the resource still gets destroyed.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"ignore_playbook_failure": schema.BoolAttribute{
				MarkdownDescription: "This parameter is good for testing. " +
					"Set to 'true' if the desired playbook is meant to fail, " +
					"but still want the resource to run successfully.",
				Optional: true,
			},

			// ansible execution commands
			"verbosity": schema.NumberAttribute{
				MarkdownDescription: "A verbosity level between 0 and 6. " +
					"Set ansible 'verbose' parameter, which causes Ansible to print more debug messages. " +
					"The higher the 'verbosity', the more debug details will be printed.",
				Optional: true,
				Computed: true,
				Validators: []validator.Number{
					numberBetweenValidator{0, 6},
				},
				Default: numberdefault.StaticBigFloat(big.NewFloat(0)),
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "List of tags of plays and tasks to run.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"limit": schema.ListAttribute{
				MarkdownDescription: "List of hosts to exclude from the playbook execution.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"check_mode": schema.BoolAttribute{
				MarkdownDescription: "If 'true', playbook execution won't make any changes but " +
					"only change predictions will be made.",
				Optional: true,
			},
			"diff_mode": schema.BoolAttribute{
				MarkdownDescription: "" +
					"If 'true', when changing (small) files and templates, differences in those files will be shown. " +
					"Recommended usage with 'check_mode'.",
				Optional: true,
			},

			// connection configs are handled with extra_vars
			"force_handlers": schema.BoolAttribute{
				MarkdownDescription: "If 'true', run handlers even if a task fails.",
				Optional:            true,
			},

			// become configs are handled with extra_vars --> these are also connection configs
			"extra_vars": schema.StringAttribute{
				MarkdownDescription: "A JSON dict of additional variables as: { key-1 = value-1, key-2 = value-2, ... }. Hint: use jsonencode()",
				Optional:            true,
				Validators: []validator.String{
					jsonValidator{},
				},
			},
			"var_files": schema.ListAttribute{ // adds @ at the beginning of filename
				MarkdownDescription: "List of variable files.",
				ElementType:         types.StringType,
				Optional:            true,
			},

			// Ansible Vault
			"vault_files": schema.ListAttribute{
				Description: "List of vault files.",
				ElementType: types.StringType,
				Optional:    true,
			},

			"vault_password_file": schema.StringAttribute{
				MarkdownDescription: "Path to a vault password file.",
				Optional:            true,
			},

			"vault_id": schema.StringAttribute{
				MarkdownDescription: "ID of the desired vault(s).",
				Optional:            true,
			},

			// computed
			// debug output
			"args": schema.ListAttribute{
				MarkdownDescription: "Used to build arguments to run Ansible playbook with.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"ansible_playbook_output": schema.StringAttribute{
				MarkdownDescription: "An ansible-playbook CLI stdout output.",
				Computed:            true,
			},

			"ansible_playbook_err": schema.StringAttribute{
				MarkdownDescription: "An ansible-playbook CLI stderr output.",
				Computed:            true,
			},
		},
	}
}

func (r *PlaybookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *PlaybookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PlaybookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, err := os.ReadFile(data.Playbook.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"can't read playbook",
			"Unable to read the playbook: unexpected error: "+err.Error(),
		)
		return
	}

	data.PlaybookSha256Sum = types.StringValue(sha256Sum(b))

	diags := r.runPlaybook(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlaybookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PlaybookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, err := os.ReadFile(data.Playbook.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"can't read playbook",
			"Unable to read the playbook: unexpected error: "+err.Error(),
		)
		return
	}

	data.PlaybookSha256Sum = types.StringValue(sha256Sum(b))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlaybookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PlaybookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, err := os.ReadFile(data.Playbook.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"can't read playbook",
			"Unable to read the playbook: unexpected error: "+err.Error(),
		)
		return
	}

	data.PlaybookSha256Sum = types.StringValue(sha256Sum(b))

	diags := r.runPlaybook(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// TODO: EXISTS (REPLAYABLE)

func (r *PlaybookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *PlaybookResource) computeArgs(ctx context.Context, data *PlaybookResourceModel) ([]string, diag.Diagnostics) {
	args := []string{}
	diags := diag.Diagnostics{}

	if verbosity, _ := data.Verbosity.ValueBigFloat().Float64(); verbosity != 0 {
		args = append(args, "-"+strings.Repeat("v", int(verbosity)))
	}

	if data.ForceHandlers.ValueBool() {
		args = append(args, "--force-handlers")
	}

	if data.Name.ValueString() == "" && len(data.ExtraInventoryFiles.Elements()) == 0 {
		diags.AddError(
			"either name or extra_inventory_files need to be set",
			"you need to set either name or extra_inventory_files to specify the inventory",
		)
		return args, diags
	}

	if data.Name.ValueString() != "" {
		args = append(args, "-i", data.Name.ValueString()+",")
	}

	extraInventoryFiles := make([]types.String, 0, len(data.ExtraInventoryFiles.Elements()))
	d := data.ExtraInventoryFiles.ElementsAs(ctx, &extraInventoryFiles, false)
	if d.HasError() {
		diags.Append(d...)
		return args, diags
	}
	for _, i := range extraInventoryFiles {
		args = append(args, "-i", i.ValueString())
	}

	// TODO: Groups

	tags := make([]types.String, 0, len(data.Tags.Elements()))
	d = data.Tags.ElementsAs(ctx, &tags, false)
	if d.HasError() {
		diags.Append(d...)
		return args, diags
	}
	if len(tags) != 0 {
		tagsStr := make([]string, 0, len(tags))
		for i, l := range tags {
			tagsStr[i] = l.ValueString()
		}

		args = append(args, "--tags", strings.Join(tagsStr, ","))
	}

	limits := make([]types.String, 0, len(data.Limit.Elements()))
	d = data.Limit.ElementsAs(ctx, &limits, false)
	if d.HasError() {
		diags.Append(d...)
		return args, diags
	}
	if len(limits) != 0 {
		limitsStr := make([]string, 0, len(limits))
		for _, l := range limits {
			limitsStr = append(limitsStr, l.ValueString())
		}

		args = append(args, "--limit", strings.Join(limitsStr, ","))
	}

	if data.CheckMode.ValueBool() {
		args = append(args, "--check")
	}

	if data.DiffMode.ValueBool() {
		args = append(args, "--diff")
	}

	varFiles := make([]types.String, 0, len(data.VarFiles.Elements()))
	d = data.VarFiles.ElementsAs(ctx, &varFiles, false)
	if d.HasError() {
		diags.Append(d...)
		return args, diags
	}
	for _, v := range varFiles {
		args = append(args, "-e", "@"+v.ValueString())
	}

	vaultFiles := make([]types.String, 0, len(data.VaultFiles.Elements()))
	d = data.VaultFiles.ElementsAs(ctx, &vaultFiles, false)
	if d.HasError() {
		diags.Append(d...)
		return args, diags
	}
	if len(vaultFiles) != 0 {
		for _, v := range vaultFiles {
			args = append(args, "-e", "@"+v.ValueString())
		}

		args = append(args, "--vault-id")

		vaultID := data.VaultID.ValueString()

		passwordFile := data.VaultPasswordFile.ValueString()
		if passwordFile != "" {
			vaultID += "@" + passwordFile
		} else {
			diags.AddError(
				"vault_password_file missing",
				"can't access vault file(s): 'vault_password_file' missing",
			)
			return args, diags
		}

		args = append(args, vaultID)
	}

	if data.ExtraVars.ValueString() != "" {
		args = append(args, "-e", data.ExtraVars.ValueString())
	}

	args = append(args, data.Playbook.ValueString())

	data.Args, d = types.ListValueFrom(ctx, types.StringType, args)
	if d.HasError() {
		diags.Append(d...)
		return args, diags
	}

	return args, diags
}

func (r *PlaybookResource) runPlaybook(ctx context.Context, data *PlaybookResourceModel) diag.Diagnostics {
	args, diags := r.computeArgs(ctx, data)
	if diags.HasError() {
		return diags
	}

	cmd := exec.CommandContext(ctx, data.AnsiblePlaybookBinary.ValueString(), args...)
	out, err := cmd.CombinedOutput()
	outStr := string(out)
	if err != nil {
		if !data.IgnorePlaybookFailure.ValueBool() {
			diags.AddError(
				"playbook execution failed",
				"args: "+strings.Join(args, " ")+", output: "+
					err.Error()+": "+outStr,
			)
			return diags
		}

		tflog.Warn(ctx, "playbook execution failed", map[string]interface{}{
			"err":    err.Error(),
			"output": outStr,
		})
	}

	data.AnsiblePlaybookOutput = types.StringValue(outStr)
	if err != nil {
		data.AnsiblePlaybookErr = types.StringValue(err.Error())
	} else {
		data.AnsiblePlaybookErr = types.StringValue("")
	}

	return diags
}
