package framework_test

import (
	"context"

	"github.com/ansible/terraform-provider-ansible/framework"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tfprotov6 "github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testProvider is a minimal provider that registers all framework resources.
// Used by unit and acceptance tests in this package without requiring the full
// muxed provider or any external configuration.
type testProvider struct{}

func (p *testProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ansible"
}

func (p *testProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *testProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *testProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *testProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		framework.NewHostResource,
		framework.NewGroupResource,
	}
}

// ansibleProviderFactories returns a protocol v6 provider factory backed by
// testProvider, suitable for use in ProtoV6ProviderFactories for both unit and
// acceptance tests. Named distinctly from PR #156's protoV6ProviderFactories(runner)
// to avoid conflicts when the vault test branch merges.
func ansibleProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"ansible": providerserver.NewProtocol6WithError(&testProvider{}),
	}
}
