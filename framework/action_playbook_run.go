package framework

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ action.ActionWithValidateConfig = (*runPlaybookRunAction)(nil)
)

func NewRunPlaybookRunAction() action.Action {
	return &runPlaybookRunAction{}
}

type runPlaybookRunAction struct{}

func (a *runPlaybookRunAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "ansible_playbook_run"
}

func (a *runPlaybookRunAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This action runs the ansible-playbook CLI command.",
		Attributes: map[string]schema.Attribute{
			// Positional arguments
			"playbooks": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Optional:    false,
				Description: "Paths to ansible playbooks.",
			},

			// Flag arguments
			"become_password_file": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Path to file containing password for privilege escalation.",
			},

			"connection_password_file": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Path to file containing password for connection.",
			},

			"force_handlers": schema.BoolAttribute{
				Required:    false,
				Optional:    true,
				Description: "Force handlers to run even if a task fails.",
			},

			"flush_cache": schema.BoolAttribute{
				Required:    false,
				Optional:    true,
				Description: "Flush the cache before running the playbook.",
			},

			"skip_tags": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "List of tags to skip during playbook execution.",
			},

			"start_at_task": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Name of task to start execution at.",
			},

			"vault_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "The vault identities to use",
			},

			"vault_password_file": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "The vault password file to use",
			},

			"check_mode": schema.BoolAttribute{
				Required:    false,
				Optional:    true,
				Description: "Run in check mode",
			},

			"diff_mode": schema.BoolAttribute{
				Required:    false,
				Optional:    true,
				Description: "Run in diff mode",
			},

			"module_paths": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "Prepend path(s) to module library",
			},

			"extra_vars": schema.MapAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "Extra variables to pass to the playbook",
			},

			"extra_vars_files": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "List of variable files with extra variables",
			},

			"forks": schema.Int64Attribute{
				Required:    false,
				Optional:    true,
				Description: "Number of parallel forks to use",
			},

			"inventories": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "List of inventories in JSON format (use ansible_inventory to generate)",
			},

			"inventory_files": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "Specify inventory host path or comma separated host list",
			},

			"limit": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Limit the execution to hosts matching a pattern",
			},

			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    false,
				Optional:    true,
				Description: "Limit the execution to tasks matching a tag",
			},

			"verbosity": schema.Int32Attribute{
				Required:    false,
				Optional:    true,
				Description: "Verbosity level",
			},

			"private_key_file": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Path to private key file",
			},

			"scp_extra_args": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Extra arguments to pass to scp",
			},

			"sftp_extra_args": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Extra arguments to pass to sftp",
			},

			"ssh_common_args": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Specify common arguments to pass to sftp/scp/ssh (e.g. ProxyCommand)",
			},

			"ssh_extra_args": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Extra arguments to pass to ssh",
			},

			"timeout": schema.Int32Attribute{
				Required:    false,
				Optional:    true,
				Description: "Override the connection timeout in seconds",
			},

			"connection_type": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Connection type to use (default=ssh)",
			},

			"user": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Connect as this user (default=None)",
			},

			"become_user": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Become this user (default=root)",
			},

			"become_method": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Privilege escalation method to use (default=sudo), use `ansible-doc -t become -l` to list valid choices.",
			},

			"become": schema.BoolAttribute{
				Required:    false,
				Optional:    true,
				Description: "Run operations with become",
			},

			// Terraform Only options
			"quiet": schema.BoolAttribute{
				Required:    false,
				Optional:    true,
				Description: "Suppress output completely",
			},

			"ansible_playbook_binary": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Path to ansible-playbook executable (binary).",
			},
		},
	}
}

type runPlaybookActionModel struct {
	Playbooks              types.List   `tfsdk:"playbooks"`
	AnsiblePlaybookBinary  types.String `tfsdk:"ansible_playbook_binary"`
	BecomePasswordFile     types.String `tfsdk:"become_password_file"`
	ConnectionPasswordFile types.String `tfsdk:"connection_password_file"`
	SkipTags               types.List   `tfsdk:"skip_tags"`
	StartAtTask            types.String `tfsdk:"start_at_task"`
	VaultIds               types.List   `tfsdk:"vault_ids"`
	VaultPasswordFile      types.String `tfsdk:"vault_password_file"`
	CheckMode              types.Bool   `tfsdk:"check_mode"`
	DiffMode               types.Bool   `tfsdk:"diff_mode"`
	ModulePaths            types.List   `tfsdk:"module_paths"`
	ExtraVars              types.Map    `tfsdk:"extra_vars"`
	ExtraVarsFiles         types.List   `tfsdk:"extra_vars_files"`
	Forks                  types.Int64  `tfsdk:"forks"`
	Inventories            types.List   `tfsdk:"inventories"`
	InventoryFiles         types.List   `tfsdk:"inventory_files"`
	Limit                  types.String `tfsdk:"limit"`
	Tags                   types.List   `tfsdk:"tags"`
	Verbosity              types.Int32  `tfsdk:"verbosity"`
	Quiet                  types.Bool   `tfsdk:"quiet"`
	PrivateKeyFile         types.String `tfsdk:"private_key_file"`
	ScpExtraArgs           types.String `tfsdk:"scp_extra_args"`
	SftpExtraArgs          types.String `tfsdk:"sftp_extra_args"`
	SshCommonArgs          types.String `tfsdk:"ssh_common_args"`
	SshExtraArgs           types.String `tfsdk:"ssh_extra_args"`
	Timeout                types.Int32  `tfsdk:"timeout"`
	ConnectionType         types.String `tfsdk:"connection_type"`
	User                   types.String `tfsdk:"user"`
	BecomeUser             types.String `tfsdk:"become_user"`
	BecomeMethod           types.String `tfsdk:"become_method"`
	Become                 types.Bool   `tfsdk:"become"`
	FlushCache             types.Bool   `tfsdk:"flush_cache"`
	ForceHandlers          types.Bool   `tfsdk:"force_handlers"`
}

func (a *runPlaybookRunAction) ValidateConfig(ctx context.Context, req action.ValidateConfigRequest, resp *action.ValidateConfigResponse) {
	var config runPlaybookActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var playbooks []types.String
	resp.Diagnostics.Append(config.Playbooks.ElementsAs(ctx, &playbooks, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(playbooks) == 0 {
		resp.Diagnostics.AddError("No playbooks specified", "At least one playbook must be specified")
		return
	}

	for i, playbook := range playbooks {
		if _, err := os.Stat(playbook.ValueString()); os.IsNotExist(err) {
			resp.Diagnostics.AddAttributeError(path.Root("playbooks").AtListIndex(i), "playbook not found", fmt.Sprintf("The playbook file %q does not exist: %s", playbook.ValueString(), err.Error()))
		}
	}

	if !config.Inventories.IsUnknown() {
		var inventories []types.String
		resp.Diagnostics.Append(config.Inventories.ElementsAs(ctx, &inventories, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for i, inventory := range inventories {
			// Validate all inventories are valid JSON
			if !inventory.IsUnknown() && !isJSON(inventory.ValueString()) {
				resp.Diagnostics.AddAttributeError(path.Root("inventories").AtListIndex(i), "Invalid JSON", fmt.Sprintf("Expected the inventory to contain valid JSON, got %q", inventory.ValueString()))
			}
		}
	}

	if !config.VaultIds.IsUnknown() && !config.VaultPasswordFile.IsUnknown() {
		var vaultFiles []types.String
		resp.Diagnostics.Append(config.VaultIds.ElementsAs(ctx, &vaultFiles, false)...)
		// We can already do some validations here during plan
		if len(vaultFiles) != 0 && config.VaultPasswordFile.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(path.Root("vault_password_file"), "vault_password_file is not found", "Can not access vault_files without passing the vault_password_file")
		}
	}

	if config.BecomePasswordFile.ValueString() != "" {
		if _, err := os.Stat(config.BecomePasswordFile.ValueString()); os.IsNotExist(err) {
			resp.Diagnostics.AddAttributeError(path.Root("become_password_file"), "become_password_file not found", fmt.Sprintf("The become password file %q does not exist: %s", config.BecomePasswordFile.ValueString(), err.Error()))
		}
	}

	if config.ConnectionPasswordFile.ValueString() != "" {
		if _, err := os.Stat(config.ConnectionPasswordFile.ValueString()); os.IsNotExist(err) {
			resp.Diagnostics.AddAttributeError(path.Root("connection_password_file"), "connection_password_file not found", fmt.Sprintf("The connection password file %q does not exist: %s", config.ConnectionPasswordFile.ValueString(), err.Error()))
		}
	}

	if config.VaultPasswordFile.ValueString() != "" {
		if _, err := os.Stat(config.VaultPasswordFile.ValueString()); os.IsNotExist(err) {
			resp.Diagnostics.AddAttributeError(path.Root("vault_password_file"), "vault_password_file not found", fmt.Sprintf("The vault password file %q does not exist: %s", config.VaultPasswordFile.ValueString(), err.Error()))
		}
	}

	if config.PrivateKeyFile.ValueString() != "" {
		if _, err := os.Stat(config.PrivateKeyFile.ValueString()); os.IsNotExist(err) {
			resp.Diagnostics.AddAttributeError(path.Root("private_key_file"), "private_key_file not found", fmt.Sprintf("The private key file %q does not exist: %s", config.PrivateKeyFile.ValueString(), err.Error()))
		}
	}

	if !config.ExtraVarsFiles.IsUnknown() {
		var extraVarsFiles []types.String
		resp.Diagnostics.Append(config.ExtraVarsFiles.ElementsAs(ctx, &extraVarsFiles, false)...)
		for i, extraVarsFile := range extraVarsFiles {
			if extraVarsFile.ValueString() != "" {
				if _, err := os.Stat(extraVarsFile.ValueString()); os.IsNotExist(err) {
					resp.Diagnostics.AddAttributeError(
						path.Root("extra_vars_files").AtListIndex(i),
						fmt.Sprintf("extra_vars_files[%d] not found", i),
						fmt.Sprintf(
							"The extra vars file %q does not exist: %s",
							extraVarsFile.ValueString(),
							err.Error(),
						),
					)
				}
			}
		}
	}
}

func (a *runPlaybookRunAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config runPlaybookActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ansiblePlaybookBinary := "ansible-playbook"
	if config.AnsiblePlaybookBinary.ValueString() != "" {
		ansiblePlaybookBinary = config.AnsiblePlaybookBinary.ValueString()
	}

	// Validate ansible-playbook binary
	if _, validateBinPath := exec.LookPath(ansiblePlaybookBinary); validateBinPath != nil {
		resp.Diagnostics.AddAttributeError(path.Root("ansible_playbook_binary"), "ansible_playbook_binary is not found", fmt.Sprintf("The ansible-playbook binary is not found: %s", validateBinPath))
		return
	}
	/********************
	* 	PREP THE OPTIONS (ARGS)
	 */
	positionalArgs := []string{}

	var playbooks []types.String
	resp.Diagnostics.Append(config.Playbooks.ElementsAs(ctx, &playbooks, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(playbooks) == 0 {
		resp.Diagnostics.AddError("No playbooks specified", "At least one playbook must be specified")
		return
	}

	for _, playbook := range playbooks {
		positionalArgs = append(positionalArgs, playbook.ValueString())
	}

	flags := []string{}

	verbosityLevel := int(config.Verbosity.ValueInt32())
	verbose := providerutils.CreateVerboseSwitch(verbosityLevel)
	if verbose != "" {
		flags = append(flags, verbose)
	}

	becomePasswordFile := config.BecomePasswordFile.ValueString()
	if becomePasswordFile != "" {
		flags = append(flags, "--become-password-file", becomePasswordFile)
	}

	connectionPasswordFile := config.ConnectionPasswordFile.ValueString()
	if connectionPasswordFile != "" {
		flags = append(flags, "--connection-password-file", connectionPasswordFile)
	}

	if config.ForceHandlers.ValueBool() {
		flags = append(flags, "--force-handlers")
	}

	if config.FlushCache.ValueBool() {
		flags = append(flags, "--flush-cache")
	}

	var skipTags []types.String
	resp.Diagnostics.Append(config.SkipTags.ElementsAs(ctx, &skipTags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, tag := range skipTags {
		flags = append(flags, "--skip-tags", tag.ValueString())
	}

	startAtTask := config.StartAtTask.ValueString()
	if startAtTask != "" {
		flags = append(flags, "--start-at-task", startAtTask)
	}

	var vaultIds []types.String
	resp.Diagnostics.Append(config.VaultIds.ElementsAs(ctx, &vaultIds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(vaultIds) > 0 {
		for _, vaultId := range vaultIds {
			flags = append(flags, "--vault-id", vaultId.ValueString())
		}
	}

	vaultPasswordFile := config.VaultPasswordFile.ValueString()
	if vaultPasswordFile != "" {
		flags = append(flags, "--vault-password-file", vaultPasswordFile)
	}

	if config.CheckMode.ValueBool() {
		flags = append(flags, "--check")
	}

	if config.DiffMode.ValueBool() {
		flags = append(flags, "--diff")
	}

	var modulePaths []types.String
	resp.Diagnostics.Append(config.ModulePaths.ElementsAs(ctx, &modulePaths, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, modulePath := range modulePaths {
		flags = append(flags, "--module-path", modulePath.ValueString())
	}

	var extraVars map[string]string
	resp.Diagnostics.Append(config.ExtraVars.ElementsAs(ctx, &extraVars, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for key, value := range extraVars {
		flags = append(flags, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	var extraVarsFiles []types.String
	resp.Diagnostics.Append(config.ExtraVarsFiles.ElementsAs(ctx, &extraVarsFiles, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, extraVarsFile := range extraVarsFiles {
		flags = append(flags, "-e", "@"+extraVarsFile.ValueString())
	}

	forks := config.Forks.ValueInt64()
	if forks != 0 {
		flags = append(flags, "--forks", fmt.Sprintf("%d", forks))
	}

	var inventoryFiles []types.String
	resp.Diagnostics.Append(config.InventoryFiles.ElementsAs(ctx, &inventoryFiles, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, inventory := range inventoryFiles {
		flags = append(flags, "--inventory", inventory.ValueString())
	}

	var inventories []types.String
	resp.Diagnostics.Append(config.Inventories.ElementsAs(ctx, &inventories, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for i, inventory := range inventories {
		tflog.Warn(ctx, fmt.Sprintf("inventory --> %#v", inventory))
		if !isJSON(inventory.ValueString()) {
			resp.Diagnostics.AddAttributeError(path.Root("inventories").AtListIndex(i), "Invalid JSON", fmt.Sprintf("Expected the inventory to contain valid JSON, got %q", inventory.ValueString()))
			return
		}

		tmpInventoryFile, err := os.CreateTemp("", "action_ansible_playbook_run_inventory_*.json")
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("inventories").AtListIndex(i), "Failed to create temporary inventory file", err.Error())
			return
		}
		defer os.Remove(tmpInventoryFile.Name())

		_, err = tmpInventoryFile.WriteString(inventory.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("inventories").AtListIndex(i), "Failed to write temporary inventory file", err.Error())
			return
		}

		flags = append(flags, "--inventory", tmpInventoryFile.Name())
	}

	limit := config.Limit.ValueString()
	if limit != "" {
		flags = append(flags, "--limit", limit)
	}

	var tags []types.String
	resp.Diagnostics.Append(config.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, tag := range tags {
		flags = append(flags, "--tags", tag.ValueString())
	}

	privateKeyFile := config.PrivateKeyFile.ValueString()
	if privateKeyFile != "" {
		flags = append(flags, "--private-key", privateKeyFile)
	}

	scpExtraArgs := config.ScpExtraArgs.ValueString()
	if scpExtraArgs != "" {
		flags = append(flags, "--scp-extra-args", scpExtraArgs)
	}

	sftpExtraArgs := config.SftpExtraArgs.ValueString()
	if sftpExtraArgs != "" {
		flags = append(flags, "--sftp-extra-args", sftpExtraArgs)
	}

	sshCommonArgs := config.SshCommonArgs.ValueString()
	if sshCommonArgs != "" {
		flags = append(flags, "--ssh-common-args", sshCommonArgs)
	}

	sshExtraArgs := config.SshExtraArgs.ValueString()
	if sshExtraArgs != "" {
		flags = append(flags, "--ssh-extra-args", sshExtraArgs)
	}

	timeout := config.Timeout.ValueInt32()
	if timeout != 0 {
		flags = append(flags, "--timeout", fmt.Sprintf("%d", timeout))
	}

	connection := config.ConnectionType.ValueString()
	if connection != "" {
		flags = append(flags, "--connection", connection)
	}

	user := config.User.ValueString()
	if user != "" {
		flags = append(flags, "--user", user)
	}

	becomeUser := config.BecomeUser.ValueString()
	if becomeUser != "" {
		flags = append(flags, "--become-user", becomeUser)
	}

	becomeMethod := config.BecomeMethod.ValueString()
	if becomeMethod != "" {
		flags = append(flags, "--become-method", becomeMethod)
	}

	if config.Become.ValueBool() {
		flags = append(flags, "--become")
	}

	args := append(flags, positionalArgs...)

	tflog.Info(ctx, fmt.Sprintf("Running Command <%s %s>", ansiblePlaybookBinary, strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, ansiblePlaybookBinary, args...)

	var stderr strings.Builder
	cmd.Stderr = &stderr
	cmd.Stdout = &TerraformUiWriter{
		send: func(s string) {
			if !config.Quiet.ValueBool() {
				resp.SendProgress(action.InvokeProgressEvent{
					Message: fmt.Sprintf("ansible-playbook: %s", s),
				})
			}
		},
	}

	if !config.Quiet.ValueBool() {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Running %s", cmd.String()),
		})
	}

	err := cmd.Run()

	stderrStr := stderr.String()
	if err != nil {
		if len(stderrStr) > 0 {
			resp.Diagnostics.AddError(
				"ansible-playbook failed",
				stderrStr,
			)
			return
		}

		resp.Diagnostics.AddError(
			"Failed to execute ansible-playbook",
			err.Error(),
		)
		return
	}
}

type TerraformUiWriter struct {
	send      func(s string)
	buffer    string
	closed    bool
	lastFlush time.Time
}

func (t *TerraformUiWriter) Write(p []byte) (n int, err error) {
	if t.closed {
		return 0, errors.New("Writing on closed writer")
	}
	t.buffer += string(p)

	now := time.Now()
	shouldFlush := false

	if t.lastFlush.IsZero() || now.Sub(t.lastFlush) >= time.Second {
		shouldFlush = true
	}

	if shouldFlush && len(t.buffer) > 0 {
		t.send(t.buffer)
		t.buffer = ""
		t.lastFlush = now
	}

	return len(p), nil
}

func (t *TerraformUiWriter) Close() error {
	if t.closed {
		return errors.New("Closing closed writer")
	}
	t.closed = true
	if t.buffer != "" {
		t.send(t.buffer)
		t.buffer = ""
	}
	return nil
}

func isJSON(str string) bool {
	var j json.RawMessage
	return json.Unmarshal([]byte(str), &j) == nil
}
