package framework

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = (*InventoryDataSource)(nil)
)

type InventoryDataSource struct{}

func NewInventoryDataSource() datasource.DataSource {
	return &InventoryDataSource{}
}

// Metadata implements datasource.Resource.
func (i *InventoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inventory"
}

type InventoryDataSourceModel struct {
	Groups types.List   `tfsdk:"group"`
	Json   types.String `tfsdk:"json"`
}

type SharedGroupModel struct {
	Name  types.String `tfsdk:"name"`
	Vars  types.Map    `tfsdk:"vars"`
	Hosts types.List   `tfsdk:"host"`
}

type NestedGroupModel struct {
	SharedGroupModel
	Groups types.List `tfsdk:"group"`
}

type FinalGroupModel struct {
	SharedGroupModel
}

// root plus two levels of nesting
const groupNestingLevel = 2

func inventoryToJson(irm *InventoryDataSourceModel) ([]byte, diag.Diagnostics) {
	jsonValue, diags := groupsToJson(irm.Groups, 0)
	if diags.HasError() {
		return nil, diags
	}
	ret, err := json.Marshal(jsonValue)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Could not marshal inventory to JSON", err.Error()))
		return nil, diags
	}
	return ret, nil
}

func groupsToJson(list types.List, level int) (map[string]json.RawMessage, diag.Diagnostics) {
	jsonValue := map[string]json.RawMessage{}
	var diags diag.Diagnostics

	if level < groupNestingLevel {
		// There is a deeper nesting level
		var nestedGroups []NestedGroupModel
		diags := list.ElementsAs(context.Background(), &nestedGroups, false)
		if diags.HasError() {
			return nil, diags
		}
		for _, group := range nestedGroups {
			groupJson, diags := nestedGroupToJson(group, level)
			if diags.HasError() {
				return nil, diags
			}
			b, err := json.Marshal(groupJson)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Could not marshal group to JSON", err.Error()))
			}

			jsonValue[group.Name.ValueString()] = b
		}
	} else {
		// This is the final nesting level
		var finalGroups []FinalGroupModel
		diags := list.ElementsAs(context.Background(), &finalGroups, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, group := range finalGroups {
			groupJson, diags := sharedGroupToJson(group.SharedGroupModel)
			if diags.HasError() {
				return nil, diags
			}
			b, err := json.Marshal(groupJson)
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic("Could not marshal group to JSON", err.Error()))
			}

			jsonValue[group.Name.ValueString()] = b
		}
	}

	return jsonValue, diags
}
func nestedGroupToJson(group NestedGroupModel, level int) (map[string]json.RawMessage, diag.Diagnostics) {
	jsonValue, diags := sharedGroupToJson(group.SharedGroupModel)
	if diags.HasError() {
		return nil, diags
	}

	groupsJson, diags := groupsToJson(group.Groups, level+1)
	if diags.HasError() {
		return nil, diags
	}
	b, err := json.Marshal(groupsJson)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Could not marshal nested groups to JSON", err.Error()))
		return nil, diags
	}
	jsonValue["children"] = b

	return jsonValue, diags
}

func sharedGroupToJson(group SharedGroupModel) (map[string]json.RawMessage, diag.Diagnostics) {
	jsonValue := map[string]json.RawMessage{}

	var hosts []HostModel
	diags := group.Hosts.ElementsAs(context.Background(), &hosts, false)
	if diags.HasError() {
		return nil, diags
	}

	hostsJson := map[string]json.RawMessage{}
	for _, host := range hosts {
		hostJson, diags := hostToJson(&host)
		if diags.HasError() {
			return nil, diags
		}
		hostsJson[host.Name.ValueString()] = hostJson
	}
	if len(hostsJson) > 0 {
		b, err := json.Marshal(hostsJson)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic("Could not marshal hosts to JSON", err.Error()))
			return nil, diags
		}
		jsonValue["hosts"] = b
	}

	var vars map[string]string
	diags = group.Vars.ElementsAs(context.Background(), &vars, false)
	if diags.HasError() {
		return nil, diags
	}
	if len(vars) > 0 {
		b, err := json.Marshal(vars)
		if err != nil {
			diags.Append(diag.NewErrorDiagnostic("Could not marshal vars to JSON", err.Error()))
			return nil, diags
		}
		jsonValue["vars"] = b
	}

	return jsonValue, diags
}

func hostToJson(hm *HostModel) (json.RawMessage, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	ret, err := json.Marshal(JsonHostModel{
		AnsibleConnection:        hm.AnsibleConnection.ValueString(),
		AnsibleHost:              hm.AnsibleHost.ValueString(),
		AnsiblePort:              hm.AnsiblePort.ValueInt64(),
		AnsibleUser:              hm.AnsibleUser.ValueString(),
		AnsiblePassword:          hm.AnsiblePassword.ValueString(),
		AnsiblePrivateKeyFile:    hm.AnsiblePrivateKeyFile.ValueString(),
		AnsibleSSHCommonArgs:     hm.AnsibleSSHCommonArgs.ValueString(),
		AnsibleSftpExtraArgs:     hm.AnsibleSftpExtraArgs.ValueString(),
		AnsibleScpExtraArgs:      hm.AnsibleScpExtraArgs.ValueString(),
		AnsibleSSHExtraArgs:      hm.AnsibleSSHExtraArgs.ValueString(),
		AnsibleSSHPipelining:     hm.AnsibleSSHPipelining.ValueBool(),
		AnsibleSSHExecutable:     hm.AnsibleSSHExecutable.ValueString(),
		AnsibleBecome:            hm.AnsibleBecome.ValueBool(),
		AnsibleBecomeMethod:      hm.AnsibleBecomeMethod.ValueString(),
		AnsibleBecomeUser:        hm.AnsibleBecomeUser.ValueString(),
		AnsibleBecomePassword:    hm.AnsibleBecomePassword.ValueString(),
		AnsibleBecomeExe:         hm.AnsibleBecomeExe.ValueString(),
		AnsibleBecomeFlags:       hm.AnsibleBecomeFlags.ValueString(),
		AnsibleShellType:         hm.AnsibleShellType.ValueString(),
		AnsiblePythonInterpreter: hm.AnsiblePythonInterpreter.ValueString(),
		AnsibleShellExecutable:   hm.AnsibleShellExecutable.ValueString(),
	})

	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Could not marshal host when marshalling inventory to JSON", err.Error()))
		return nil, diags
	}

	return ret, diags
}

type HostModel struct {
	Name                     types.String `tfsdk:"name"`
	AnsibleConnection        types.String `tfsdk:"ansible_connection"`
	AnsibleHost              types.String `tfsdk:"ansible_host"`
	AnsiblePort              types.Int64  `tfsdk:"ansible_port"`
	AnsibleUser              types.String `tfsdk:"ansible_user"`
	AnsiblePassword          types.String `tfsdk:"ansible_password"`
	AnsiblePrivateKeyFile    types.String `tfsdk:"ansible_private_key_file"`
	AnsibleSSHCommonArgs     types.String `tfsdk:"ansible_ssh_common_args"`
	AnsibleSftpExtraArgs     types.String `tfsdk:"ansible_sftp_extra_args"`
	AnsibleScpExtraArgs      types.String `tfsdk:"ansible_scp_extra_args"`
	AnsibleSSHExtraArgs      types.String `tfsdk:"ansible_ssh_extra_args"`
	AnsibleSSHPipelining     types.Bool   `tfsdk:"ansible_ssh_pipelining"`
	AnsibleSSHExecutable     types.String `tfsdk:"ansible_ssh_executable"`
	AnsibleBecome            types.Bool   `tfsdk:"ansible_become"`
	AnsibleBecomeMethod      types.String `tfsdk:"ansible_become_method"`
	AnsibleBecomeUser        types.String `tfsdk:"ansible_become_user"`
	AnsibleBecomePassword    types.String `tfsdk:"ansible_become_password"`
	AnsibleBecomeExe         types.String `tfsdk:"ansible_become_exe"`
	AnsibleBecomeFlags       types.String `tfsdk:"ansible_become_flags"`
	AnsibleShellType         types.String `tfsdk:"ansible_shell_type"`
	AnsiblePythonInterpreter types.String `tfsdk:"ansible_python_interpreter"`
	AnsibleShellExecutable   types.String `tfsdk:"ansible_shell_executable"`
}

type JsonHostModel struct {
	AnsibleConnection        string `json:"ansible_connection,omitempty"`
	AnsibleHost              string `json:"ansible_host,omitempty"`
	AnsiblePort              int64  `json:"ansible_port,omitempty"`
	AnsibleUser              string `json:"ansible_user,omitempty"`
	AnsiblePassword          string `json:"ansible_password,omitempty"`
	AnsiblePrivateKeyFile    string `json:"ansible_private_key_file,omitempty"`
	AnsibleSSHCommonArgs     string `json:"ansible_ssh_common_args,omitempty"`
	AnsibleSftpExtraArgs     string `json:"ansible_sftp_extra_args,omitempty"`
	AnsibleScpExtraArgs      string `json:"ansible_scp_extra_args,omitempty"`
	AnsibleSSHExtraArgs      string `json:"ansible_ssh_extra_args,omitempty"`
	AnsibleSSHPipelining     bool   `json:"ansible_ssh_pipelining,omitempty"`
	AnsibleSSHExecutable     string `json:"ansible_ssh_executable,omitempty"`
	AnsibleBecome            bool   `json:"ansible_become,omitempty"`
	AnsibleBecomeMethod      string `json:"ansible_become_method,omitempty"`
	AnsibleBecomeUser        string `json:"ansible_become_user,omitempty"`
	AnsibleBecomePassword    string `json:"ansible_become_password,omitempty"`
	AnsibleBecomeExe         string `json:"ansible_become_exe,omitempty"`
	AnsibleBecomeFlags       string `json:"ansible_become_flags,omitempty"`
	AnsibleShellType         string `json:"ansible_shell_type,omitempty"`
	AnsiblePythonInterpreter string `json:"ansible_python_interpreter,omitempty"`
	AnsibleShellExecutable   string `json:"ansible_shell_executable,omitempty"`
}

// Schema implements datasource.Resource.
func (i *InventoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {

	createGroupBlock := func(blockOverrides map[string]schema.Block) schema.ListNestedBlock {
		blocks := map[string]schema.Block{
			"host": schema.ListNestedBlock{
				MarkdownDescription: "Describes an ansible host.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the host.",
							Required:            true,
							Optional:            false,
						},
						// See https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html#behavioral-parameters
						"ansible_connection": schema.StringAttribute{
							MarkdownDescription: "Specifies the connection type to the host. This can be the name of any Ansible connection plugin. SSH protocol types are ssh or paramiko. The default is ssh.",
							Required:            false,
							Optional:            true,
						},
						"ansible_host": schema.StringAttribute{
							MarkdownDescription: "Specifies the resolvable name or IP of the host to connect to, if it is different from the alias (name) you wish to give to it.",
							Required:            false,
							Optional:            true,
						},
						"ansible_port": schema.Int64Attribute{
							MarkdownDescription: "The connection port number, if not the default (22 for ssh).",
							Required:            false,
							Optional:            true,
						},
						"ansible_user": schema.StringAttribute{
							MarkdownDescription: "The username to use when connecting (logging in) to the host.",
							Required:            false,
							Optional:            true,
						},
						"ansible_password": schema.StringAttribute{
							MarkdownDescription: "The password to use when connecting (logging in) to the host.",
							Required:            false,
							Optional:            true,
							Sensitive:           true,
						},
						"ansible_private_key_file": schema.StringAttribute{
							MarkdownDescription: "Private key file used by SSH. This is useful if you use multiple keys and you do not want to use SSH agent.",
							Required:            false,
							Optional:            true,
						},
						"ansible_ssh_common_args": schema.StringAttribute{
							MarkdownDescription: "Ansible always appends this setting to the default command line for sftp, scp, and ssh. This is useful for configuring a ``ProxyCommand` for a certain host or group.",
							Required:            false,
							Optional:            true,
						},
						"ansible_sftp_extra_args": schema.StringAttribute{
							MarkdownDescription: "Extra arguments to pass to the sftp command.",
							Required:            false,
							Optional:            true,
						},
						"ansible_scp_extra_args": schema.StringAttribute{
							MarkdownDescription: "Extra arguments to pass to the scp command.",
							Required:            false,
							Optional:            true,
						},
						"ansible_ssh_extra_args": schema.StringAttribute{
							MarkdownDescription: "Extra arguments to pass to the ssh command.",
							Required:            false,
							Optional:            true,
						},
						"ansible_ssh_pipelining": schema.BoolAttribute{
							MarkdownDescription: "Enable pipelining for SSH connections.",
							Required:            false,
							Optional:            true,
						},
						"ansible_ssh_executable": schema.StringAttribute{
							MarkdownDescription: "Specify the SSH executable to use.",
							Required:            false,
							Optional:            true,
						},
						"ansible_become": schema.BoolAttribute{
							MarkdownDescription: "Allows you to force privilege escalation.",
							Required:            false,
							Optional:            true,
						},
						"ansible_become_method": schema.StringAttribute{
							MarkdownDescription: "Allows you to set the privilege escalation method to a matching become plugin.",
							Required:            false,
							Optional:            true,
						},
						"ansible_become_user": schema.StringAttribute{
							MarkdownDescription: "Allows you to set the user you become through privilege escalation.",
							Required:            false,
							Optional:            true,
						},
						"ansible_become_password": schema.StringAttribute{
							MarkdownDescription: "Allows you to set the privilege escalation password.",
							Required:            false,
							Optional:            true,
							Sensitive:           true,
						},
						"ansible_become_exe": schema.StringAttribute{
							MarkdownDescription: "Allows you to set the executable for the escalation method you selected.",
							Required:            false,
							Optional:            true,
						},
						"ansible_become_flags": schema.StringAttribute{
							MarkdownDescription: "Allows you to set the flags passed to the selected escalation method",
							Required:            false,
							Optional:            true,
						},
						"ansible_shell_type": schema.StringAttribute{
							MarkdownDescription: "Specifies the shell type of the target system. You should not use this setting unless you have set the `ansible_shell_executable` to a non-Bourne (sh) compatible shell.  By default, Ansible formats commands using sh-style syntax.  If you set this to csh or fish, commands that Ansible executes on target systems follow those shell’s syntax instead.",
							Required:            false,
							Optional:            true,
						},
						"ansible_python_interpreter": schema.StringAttribute{
							MarkdownDescription: "Allows you to set the Python interpreter to use for the target system.",
							Required:            false,
							Optional:            true,
						},
						"ansible_shell_executable": schema.StringAttribute{
							MarkdownDescription: "Allows you to set the shell executable to use for the target system.",
							Required:            false,
							Optional:            true,
						},
					},
				},
			},
		}
		maps.Copy(blocks, blockOverrides)
		return schema.ListNestedBlock{
			MarkdownDescription: "Describes an ansible group.",
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the group.",
						Required:            true,
					},
					"vars": schema.MapAttribute{
						MarkdownDescription: "Variables to be set for the group.",
						Required:            false,
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
				Blocks: blocks,
			},
		}
	}

	// This schema is nested, we will support 3 levels of nesting.
	thirdLevelGroupBlock := createGroupBlock(map[string]schema.Block{})
	secondLevelGroupBlock := createGroupBlock(map[string]schema.Block{
		"group": thirdLevelGroupBlock,
	})
	firstLevelGroupBlock := createGroupBlock(map[string]schema.Block{
		"group": secondLevelGroupBlock,
	})

	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source represents an ansible inventory. It has a json attribute containing the JSON representation of the inventory.",
		Attributes: map[string]schema.Attribute{
			"json": schema.StringAttribute{
				MarkdownDescription: "The JSON content of the inventory file.",
				Required:            false,
				Optional:            false,
				Computed:            true,
				Sensitive:           true, // Might contain sensitive info
			},
		},
		Blocks: map[string]schema.Block{
			"group": firstLevelGroupBlock,
		},
	}
}

// Read implements datasource.Resource.
func (i *InventoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan InventoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileContent, toJsonDiags := inventoryToJson(&plan)
	resp.Diagnostics.Append(toJsonDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Json = types.StringValue(string(fileContent))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
