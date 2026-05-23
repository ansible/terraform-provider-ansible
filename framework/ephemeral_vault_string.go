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
	_ ephemeral.EphemeralResource              = (*VaultStringEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*VaultStringEphemeralResource)(nil)
)

type VaultStringEphemeralResource struct {
	runner VaultRunner
}

func NewVaultStringEphemeralResource() ephemeral.EphemeralResource {
	return &VaultStringEphemeralResource{runner: DefaultVaultRunner}
}

func (e *VaultStringEphemeralResource) Configure(
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

func (e *VaultStringEphemeralResource) Metadata(
	_ context.Context,
	req ephemeral.MetadataRequest,
	resp *ephemeral.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_vault_string"
}

type vaultStringEphemeralModel struct {
	Content           types.String `tfsdk:"content"`
	VaultPassword     types.String `tfsdk:"vault_password"`
	VaultPasswordFile types.String `tfsdk:"vault_password_file"`
	VaultID           types.String `tfsdk:"vault_id"`
	Plaintext         types.String `tfsdk:"plaintext"`
}

func (e *VaultStringEphemeralResource) Schema(
	_ context.Context,
	_ ephemeral.SchemaRequest,
	resp *ephemeral.SchemaResponse,
) {
	resp.Schema = ephemeralschema.Schema{
		MarkdownDescription: "Decrypts an ansible-vault encrypted string and exposes its plaintext. " +
			"Unlike the data source variant, this ephemeral resource writes nothing to state.",
		Attributes: map[string]ephemeralschema.Attribute{
			"content": ephemeralschema.StringAttribute{
				MarkdownDescription: "The ansible-vault encrypted string (begins with `$ANSIBLE_VAULT;...`).",
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
			"plaintext": ephemeralschema.StringAttribute{
				MarkdownDescription: "Decrypted plaintext of the vault string.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (e *VaultStringEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var config vaultStringEphemeralModel
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

	plaintext, diags := e.runner.Decrypt(ctx, passwordFile, config.VaultID.ValueString(), config.Content.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Plaintext = types.StringValue(plaintext)
	resp.Diagnostics.Append(resp.Result.Set(ctx, &config)...)
}
