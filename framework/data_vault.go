package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*VaultDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*VaultDataSource)(nil)
)

type VaultDataSource struct {
	runner VaultRunner
}

func NewVaultDataSource() datasource.DataSource {
	return &VaultDataSource{runner: DefaultVaultRunner}
}

func (d *VaultDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
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
	d.runner = runner
}

func (d *VaultDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_vault"
}

type vaultConfigModel struct {
	VaultFile         types.String `tfsdk:"vault_file"`
	VaultPassword     types.String `tfsdk:"vault_password"`
	VaultPasswordFile types.String `tfsdk:"vault_password_file"`
	VaultID           types.String `tfsdk:"vault_id"`
	Yaml              types.String `tfsdk:"yaml"`
}

func (d *VaultDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Decrypts an ansible-vault encrypted file and exposes its content.",
		Attributes: map[string]schema.Attribute{
			"vault_file": schema.StringAttribute{
				MarkdownDescription: "Path to the ansible-vault encrypted file.",
				Required:            true,
			},
			"vault_password": schema.StringAttribute{
				MarkdownDescription: "Vault password as a plain string.", //nolint:goconst
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("vault_password_file")),
					stringvalidator.AtLeastOneOf(path.MatchRoot("vault_password"), path.MatchRoot("vault_password_file")),
				},
			},
			"vault_password_file": schema.StringAttribute{
				MarkdownDescription: "Path to the file containing the vault password.", //nolint:goconst
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("vault_password")),
					stringvalidator.AtLeastOneOf(path.MatchRoot("vault_password"), path.MatchRoot("vault_password_file")),
				},
			},
			"vault_id": schema.StringAttribute{
				MarkdownDescription: "Vault ID label used with `--vault-id <id>@<vault_password_file>`.", //nolint:goconst
				Optional:            true,
			},
			"yaml": schema.StringAttribute{
				MarkdownDescription: "Decrypted content of the vault file.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *VaultDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config vaultConfigModel
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

	yaml, diags := d.runner.View(ctx, passwordFile, config.VaultID.ValueString(), config.VaultFile.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Yaml = types.StringValue(yaml)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
