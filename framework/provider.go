package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &fwprovider{}
var _ provider.ProviderWithEphemeralResources = &fwprovider{}

// New returns a new, initialized Terraform Plugin Framework-style provider instance.
// The provider instance is fully configured once the `Configure` method has been called.
func New(primary interface{ Meta() interface{} }) provider.Provider {
	return &fwprovider{
		Primary: primary,
	}
}

type fwprovider struct {
	Primary interface{ Meta() interface{} }
}

func (f *fwprovider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "ansible"
}

func (f *fwprovider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
		Blocks:     map[string]schema.Block{},
	}
}

func (f *fwprovider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	// Provider's parsed configuration (its instance state) is available through the primary provider's Meta() method.
	v := f.Primary.Meta()
	response.DataSourceData = v
	response.ResourceData = v
	response.EphemeralResourceData = v
	response.ActionData = v
}

func (f *fwprovider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (f *fwprovider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (f *fwprovider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return nil
}
