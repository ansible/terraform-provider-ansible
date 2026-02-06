package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// variablesMustBeObjectValidator ensures the dynamic value is an object/map at the top level.
type variablesMustBeObjectValidator struct{}

var _ validator.Dynamic = variablesMustBeObjectValidator{}

func (v variablesMustBeObjectValidator) Description(_ context.Context) string {
	return "value must be an object (map of key-value pairs)"
}

func (v variablesMustBeObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v variablesMustBeObjectValidator) ValidateDynamic(
	_ context.Context,
	req validator.DynamicRequest,
	resp *validator.DynamicResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	underlying := req.ConfigValue.UnderlyingValue()

	switch underlying.(type) {
	case types.Object, types.Map:
		return
	default:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid variables type",
			fmt.Sprintf(
				"The variables attribute must be an object (map), got %T. "+
					"Use an object value like { key = \"value\" }.",
				underlying,
			),
		)
	}
}

// variablesMustBeObject returns a validator ensuring the dynamic value is a map/object.
func variablesMustBeObject() validator.Dynamic {
	return variablesMustBeObjectValidator{}
}

// priorHostSchema returns the v0 schema for ansible_host (matching the old SDK resource).
func priorHostSchema() *schema.Schema {
	return &schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"groups": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"variables": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// priorGroupSchema returns the v0 schema for ansible_group (matching the old SDK resource).
func priorGroupSchema() *schema.Schema {
	return &schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"children": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"variables": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}
