package provider

import (
	"context"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &GalaxyResource{}

func NewGalaxyResource() resource.Resource {
	return &GalaxyResource{}
}

// PlaybookResource defines the resource implementation.
type GalaxyResource struct{}

type GalaxyResourceModel struct {
	Role                types.String `tfsdk:"role"`
	Version             types.String `tfsdk:"version"`
	Name                types.String `tfsdk:"name"`
	AnsibleGalaxyBinary types.String `tfsdk:"ansible_galaxy_binary"`
	Path                types.String `tfsdk:"path"`
}

func (r *GalaxyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_galaxy"
}

func (r *GalaxyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Galaxy resource",

		Attributes: map[string]schema.Attribute{
			"role": schema.StringAttribute{
				MarkdownDescription: "The user/role combination. This can also be a URL to a Git repository",
				Required:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The version of the role",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name that will be given to the role",
				Optional:            true,
			},
			"ansible_galaxy_binary": schema.StringAttribute{
				MarkdownDescription: "Path to the ansible-galaxy executable (binary)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ansible-galaxy"),
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path where the role is installed. To use with `ANSIBLE_ROLES_PATH`",
				Computed:            true,
			},
		},
	}
}

func (r *GalaxyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *GalaxyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GalaxyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dir, err := os.MkdirTemp("", ".role*")
	if err != nil {
		resp.Diagnostics.AddError(
			"can't create the role directory",
			"Unable to create the role directory: unexpected error: "+err.Error(),
		)
		return
	}
	data.Path = types.StringValue(dir)

	diags := r.installGalaxy(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GalaxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GalaxyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := os.Stat(data.Path.ValueString()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			resp.Diagnostics.AddError(
				"can't read the role directory",
				"Unable to read the role directory: unexpected error: "+err.Error(),
			)
			return
		}

		data.Path = types.StringUnknown()
	}

	found := false
	if data.Path.ValueString() != "" {
		out, err := exec.CommandContext(
			ctx,
			data.AnsibleGalaxyBinary.ValueString(),
			"list",
			"--roles-path",
			data.Path.ValueString(),
		).CombinedOutput()
		outStr := string(out)
		if err != nil {
			resp.Diagnostics.AddError(
				"can't list the roles",
				"Unable to list the roles: unexpected error: "+
					err.Error()+": "+outStr,
			)
			return
		}

		name := r.localGalaxyName(data)
		for _, l := range strings.Split(outStr, "\n") {
			if strings.Contains(l, name) {
				found = true

				if data.Version.ValueString() != "" {
					if !strings.Contains(l, data.Version.ValueString()) {
						data.Version = types.StringUnknown()
					}
				}
			}
		}
	}

	if !found {
		data.Role = types.StringUnknown()

		if data.Version.ValueString() != "" {
			data.Version = types.StringUnknown()
		}

		if data.Name.ValueString() != "" {
			data.Name = types.StringUnknown()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GalaxyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GalaxyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Path.ValueString() == "" {
		dir, err := os.MkdirTemp("", ".role*")
		if err != nil {
			resp.Diagnostics.AddError(
				"can't create the role directory",
				"Unable to create the role directory: unexpected error: "+err.Error(),
			)
			return
		}
		data.Path = types.StringValue(dir)
	}

	diags := r.installGalaxy(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GalaxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GalaxyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := r.localGalaxyName(data)

	out, err := exec.CommandContext(
		ctx,
		data.AnsibleGalaxyBinary.ValueString(),
		"remove",
		name,
	).CombinedOutput()
	if err != nil {
		resp.Diagnostics.AddError(
			"can't remove the role",
			"Unable to remove the role: unexpected error: "+
				err.Error()+": "+string(out),
		)
		return
	}

	if err := os.RemoveAll(data.Path.ValueString()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			resp.Diagnostics.AddError(
				"can't remove the role directory",
				"Unable to remove the role directory: unexpected error: "+err.Error(),
			)
			return
		}
	}
}

func (r *GalaxyResource) localGalaxyName(data GalaxyResourceModel) string {
	if data.Name.ValueString() != "" {
		return data.Name.ValueString()
	}

	u, err := url.Parse(data.Role.ValueString())
	// It's not an URL, we can assume the role name will be verbatim
	if err != nil {
		return data.Name.ValueString()
	}

	// If it's an URL, extract the last item in the path
	base := filepath.Base(u.Path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (r *GalaxyResource) installGalaxy(ctx context.Context, data *GalaxyResourceModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	role := data.Role.ValueString()
	version := data.Version.ValueString()
	if version != "" {
		version = "," + version
	}
	name := data.Name.ValueString()
	if name != "" {
		name = "," + name
	}

	out, err := exec.CommandContext(
		ctx,
		data.AnsibleGalaxyBinary.ValueString(),
		"install",
		"--roles-path",
		data.Path.ValueString(),
		"--force",
		role+version+name,
	).CombinedOutput()
	if err != nil {
		diags.AddError(
			"can't install the role",
			"Unable to install the role: unexpected error: "+
				err.Error()+": "+string(out),
		)
		return diags
	}

	return diags
}
