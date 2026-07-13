package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostToJson_WarningOnConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	varsMap, mapDiags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"ansible_connection": "local", // Conflicts with explicit connection below
		"custom_var":         "custom_value",
	})
	require.False(t, mapDiags.HasError())

	hostModel := &HostModel{
		AnsibleConnection: types.StringValue("ssh"),
		Vars:              varsMap,
	}

	_, diags := hostToJson(ctx, hostModel)

	assert.False(t, diags.HasError(), "Should not produce errors")

	warnings := diags.Warnings()
	require.Len(t, warnings, 1, "Should produce exactly 1 warning diagnostic")
	assert.Contains(t, warnings[0].Summary(), "Ansible Host Variable Conflict")
	assert.Contains(t, warnings[0].Detail(), "ansible_connection")
}

func TestHostToJson_NoWarningOnNoConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	varsMap, mapDiags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"custom_var": "custom_value",
	})
	require.False(t, mapDiags.HasError())

	hostModel := &HostModel{
		AnsibleConnection: types.StringValue("ssh"),
		Vars:              varsMap,
	}

	_, diags := hostToJson(ctx, hostModel)

	assert.False(t, diags.HasError())
	assert.Empty(t, diags.Warnings(), "Should not produce any warning diagnostics")
}
