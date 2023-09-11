package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type numberBetweenValidator struct {
	Min float64
	Max float64
}

func (v numberBetweenValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	var num float64
	if req.ConfigValue.ValueBigFloat() != nil {
		num, _ = req.ConfigValue.ValueBigFloat().Float64()
	} else {
		num = 0
	}

	if num < v.Min || num > v.Max {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid number",
			fmt.Sprintf("It must be between %d and %d", int(v.Min), int(v.Max)),
		)
		return
	}
}

func (v numberBetweenValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("number must be between %d and %d", int(v.Min), int(v.Max))
}

func (v numberBetweenValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("number must be between `%d` and `%d`", int(v.Min), int(v.Max))
}

type jsonValidator struct{}

func (v jsonValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if !json.Valid([]byte(req.ConfigValue.ValueString())) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid JSON",
			"string must be a valid JSON string, hint: use jsonencode",
		)
		return
	}
}

func (v jsonValidator) Description(ctx context.Context) string {
	return "string must be a valid JSON string"
}

func (v jsonValidator) MarkdownDescription(ctx context.Context) string {
	return "string must be a valid JSON string"
}
