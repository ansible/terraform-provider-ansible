// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure AnsibleProvider satisfies various provider interfaces.
var _ provider.Provider = &AnsibleProvider{}

// AnsibleProvider defines the provider implementation.
type AnsibleProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AnsibleProviderModel describes the provider data model.
type AnsibleProviderModel struct{}

func (p *AnsibleProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ansible"
	resp.Version = p.version
}

func (p *AnsibleProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
}

func (p *AnsibleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AnsibleProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *AnsibleProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGalaxyResource,
		NewHostResource,
		NewPlaybookResource,
	}
}

func (p *AnsibleProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AnsibleProvider{
			version: version,
		}
	}
}
