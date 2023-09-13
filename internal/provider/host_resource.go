package provider

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &HostResource{}

func NewHostResource() resource.Resource {
	return &HostResource{}
}

// HostResource defines the resource implementation.
type HostResource struct{}

type HostResourceModel struct {
	Name               types.String `tfsdk:"name"`
	Port               types.Number `tfsdk:"port"`
	Groups             types.List   `tfsdk:"groups"`
	Variables          types.Map    `tfsdk:"variables"`
	InventoryPath      types.String `tfsdk:"inventory_path"`
	InventorySha256Sum types.String `tfsdk:"inventory_sha256_sum"`
}

func (r *HostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *HostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Host resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				// TODO: VALIDATION
				MarkdownDescription: "The host",
				Required:            true,
			},
			"port": schema.NumberAttribute{
				MarkdownDescription: "The SSH port that will be used",
				Optional:            true,
				Computed:            true,
				Default:             numberdefault.StaticBigFloat(big.NewFloat(22)),
			},
			"groups": schema.ListAttribute{
				// TODO: VALIDATION
				MarkdownDescription: "List of groups where the host will be added",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"variables": schema.MapAttribute{
				// TODO: VALIDATION
				MarkdownDescription: "Map of variables that will be set to the host",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"inventory_path": schema.StringAttribute{
				MarkdownDescription: "The path of the inventory file that will be generated",
				Computed:            true,
			},
			"inventory_sha256_sum": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *HostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

func (r *HostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.createInventory(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inventoryPath := data.InventoryPath.ValueString()
	if inventoryPath != "" {
		b, err := os.ReadFile(inventoryPath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				resp.Diagnostics.AddError(
					"can't read inventory file",
					"unable to read the inventory file: unexpected error: "+err.Error(),
				)

				return
			}

			data.InventoryPath = types.StringUnknown()
			data.InventorySha256Sum = types.StringUnknown()

			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

			return
		}

		data.InventorySha256Sum = types.StringValue(sha256Sum(b))

	} else {
		// Clear sha256 sum
		data.InventorySha256Sum = types.StringUnknown()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.createInventory(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := os.Remove(data.InventoryPath.ValueString()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			resp.Diagnostics.AddError(
				"can't remove the inventory file",
				"unable to remove the inventory file: unexpected error: "+err.Error(),
			)
		}
	}
}

func (r *HostResource) createInventory(ctx context.Context, data *HostResourceModel) diag.Diagnostics {
	inventoryPath := data.InventoryPath.ValueString()
	diags := diag.Diagnostics{}

	var f *os.File
	var err error
	if inventoryPath == "" {
		f, err = os.CreateTemp("", ".inventory-*.yml")
		if err != nil {
			diags.AddError(
				"can't create file",
				"unable to create the inventory file: unexpected error: "+err.Error(),
			)

			return diags
		}

		data.InventoryPath = types.StringValue(f.Name())

	} else {
		f, err = os.Create(inventoryPath)
		if err != nil {
			diags.AddError(
				"can't open file",
				"unable to open the inventory file: unexpected error: "+err.Error(),
			)

			return diags
		}
	}
	defer f.Close()

	port, _ := data.Port.ValueBigFloat().Float64()

	dataVariables := make(map[string]types.String, len(data.Variables.Elements()))
	d := data.Variables.ElementsAs(ctx, &dataVariables, false)
	if d.HasError() {
		diags.Append(d...)
		return diags
	}
	variables := make(map[string]string, len(dataVariables))
	for k, v := range dataVariables {
		variables[k] = v.ValueString()
	}

	hosts := map[string]interface{}{
		"hosts": map[string]interface{}{
			data.Name.ValueString() + ":" + strconv.Itoa(int(port)): variables,
		},
	}

	result := map[string]interface{}{
		"ungrouped": hosts,
	}

	groups := make([]types.String, 0, len(data.Groups.Elements()))
	d = data.Groups.ElementsAs(ctx, &groups, false)
	if d.HasError() {
		diags.Append(d...)
		return diags
	}

	for _, g := range groups {
		result[g.ValueString()] = hosts
	}

	buf := bytes.Buffer{}

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(result); err != nil {
		diags.AddError(
			"can't encode inventory file",
			"unable to encode the inventory file: unexpected error: "+err.Error(),
		)

		return diags
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		diags.AddError(
			"can't write inventory file",
			"unable to write the inventory file: unexpected error: "+err.Error(),
		)

		return diags
	}

	data.InventorySha256Sum = types.StringValue(sha256Sum(buf.Bytes()))

	return diags
}
