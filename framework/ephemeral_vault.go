package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = (*VaultEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*VaultEphemeralResource)(nil)
)

type VaultEphemeralResource struct {
	runner VaultRunner
}

func NewVaultEphemeralResource() ephemeral.EphemeralResource {
	return &VaultEphemeralResource{runner: DefaultVaultRunner}
}

func (e *VaultEphemeralResource) Configure(
	_ context.Context,
	req ephemeral.ConfigureRequest,
	resp *ephemeral.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	runner, ok := req.ProviderData.(VaultRunner)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected ProviderData type",
			fmt.Sprintf("expected VaultRunner, got %T", req.ProviderData),
		)
		return
	}
	e.runner = runner
}

func (e *VaultEphemeralResource) Metadata(
	_ context.Context,
	req ephemeral.MetadataRequest,
	resp *ephemeral.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_vault"
}

type vaultEphemeralModel struct {
	VaultFile         types.String `tfsdk:"vault_file"`
	VaultPassword     types.String `tfsdk:"vault_password"`
	VaultPasswordFile types.String `tfsdk:"vault_password_file"`
	VaultID           types.String `tfsdk:"vault_id"`
	Yaml              types.String `tfsdk:"yaml"`
}

func (e *VaultEphemeralResource) Schema(
	_ context.Context,
	_ ephemeral.SchemaRequest,
	resp *ephemeral.SchemaResponse,
) {
	resp.Schema = ephemeralschema.Schema{
		MarkdownDescription: "Decrypts an ansible-vault encrypted file and exposes its content. " +
			"Unlike the data source variant, this ephemeral resource writes nothing to state.",
		Attributes: map[string]ephemeralschema.Attribute{
			"vault_file": ephemeralschema.StringAttribute{
				MarkdownDescription: "Path to the ansible-vault encrypted file.",
				Required:            true,
			},
			"vault_password": ephemeralschema.StringAttribute{
				MarkdownDescription: "Vault password as a plain string.", //nolint:goconst
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("vault_password_file")),
					stringvalidator.AtLeastOneOf(path.MatchRoot("vault_password"), path.MatchRoot("vault_password_file")),
				},
			},
			"vault_password_file": ephemeralschema.StringAttribute{
				MarkdownDescription: "Path to the file containing the vault password.", //nolint:goconst
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("vault_password")),
					stringvalidator.AtLeastOneOf(path.MatchRoot("vault_password"), path.MatchRoot("vault_password_file")),
				},
			},
			"vault_id": ephemeralschema.StringAttribute{
				MarkdownDescription: "Vault ID label used with `--vault-id <id>@<vault_password_file>`.", //nolint:goconst
				Optional:            true,
			},
			"yaml": ephemeralschema.StringAttribute{
				MarkdownDescription: "Decrypted content of the vault file.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (e *VaultEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var config vaultEphemeralModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	passwordFile, cleanup, diags := ResolvePasswordFile(config.VaultPassword.ValueString(), config.VaultPasswordFile.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	defer cleanup()

	content, diags := e.runner.View(ctx, passwordFile, config.VaultID.ValueString(), config.VaultFile.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Yaml = types.StringValue(content)
	resp.Diagnostics.Append(resp.Result.Set(ctx, &config)...)
}
