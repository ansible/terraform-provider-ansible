package framework_test

import (
	"context"

	"github.com/ansible/terraform-provider-ansible/framework"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfprotov6 "github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testProvider is a minimal provider that registers all framework resources.
// Used by unit and acceptance tests in this package without requiring the full
// muxed provider or any external configuration.
type testProvider struct {
	runner framework.VaultRunner
}

var (
	_ provider.Provider                       = (*testProvider)(nil)
	_ provider.ProviderWithEphemeralResources = (*testProvider)(nil)
)

func (p *testProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ansible"
}

func (p *testProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *testProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.DataSourceData = p.runner
	resp.EphemeralResourceData = p.runner
}

func (p *testProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		framework.NewInventoryDataSource,
		framework.NewVaultDataSource,
		framework.NewVaultStringDataSource,
	}
}

func (p *testProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		framework.NewHostResource,
		framework.NewGroupResource,
	}
}

func (p *testProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		framework.NewVaultEphemeralResource,
		framework.NewVaultStringEphemeralResource,
	}
}

// ansibleProviderFactories returns a protocol v6 provider factory backed by
// testProvider, suitable for use in ProtoV6ProviderFactories for both unit and
// acceptance tests.
func ansibleProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ansible": providerserver.NewProtocol6WithError(&testProvider{}),
	}
}

// ansibleVaultProviderFactories returns a protocol v6 provider factory backed by
// testProvider, suitable for use in ProtoV6ProviderFactories for unit. The given
// runner is injected into all vault resources so tests control what ansible-vault
// "returns" without executing any process.
func ansibleVaultProviderFactories(runner framework.VaultRunner) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ansible": providerserver.NewProtocol6WithError(&testProvider{runner: runner}),
	}
}

type vaultRunnerFunc func(ctx context.Context, passwordFile, vaultID, vaultFile string) (string, diag.Diagnostics)

func (f vaultRunnerFunc) View(ctx context.Context, passwordFile, vaultID, vaultFile string) (string, diag.Diagnostics) {
	return f(ctx, passwordFile, vaultID, vaultFile)
}

func (f vaultRunnerFunc) Decrypt(ctx context.Context, passwordFile, vaultID, content string) (string, diag.Diagnostics) {
	return f(ctx, passwordFile, vaultID, content)
}

// okRunner returns a VaultRunner that always succeeds with the given plaintext.
func okRunner(plaintext string) framework.VaultRunner {
	return vaultRunnerFunc(func(_ context.Context, _, _, _ string) (string, diag.Diagnostics) {
		return plaintext, nil
	})
}

// errRunner returns a VaultRunner that always fails with the given error.
func errRunner() framework.VaultRunner {
	return vaultRunnerFunc(func(_ context.Context, _, _, _ string) (string, diag.Diagnostics) {
		var d diag.Diagnostics
		d.AddError("ansible-vault view failed", "ERROR! Decryption failed (no vault secrets would decrypt)")
		return "", d
	})
}
