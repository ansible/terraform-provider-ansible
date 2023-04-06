package provider

import (
	"github.com/ansible/terraform-provider-ansible/provider_utils"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"os/exec"
	"strings"
)

const ap = "ansible-playbook"

func resourcePlaybook() *schema.Resource {
	return &schema.Resource{
		Create: resourcePlaybookCreate,
		Read:   resourcePlaybookRead,
		Update: resourcePlaybookUpdate,
		Delete: resourcePlaybookDelete,

		Schema: map[string]*schema.Schema{
			// Required settings
			"playbook": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Path to ansible playbook.",
			},

			// Optional settings
			"ansible_playbook_binary": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "ansible-playbook",
				Description: "Path to ansible-playbook executable (binary)",
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Name of the desired host on which the playbook will be executed.",
			},

			"groups": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of desired groups of hosts no which the playbook will be executed.",
			},

			"replayable": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  true,
				Description: "" +
					"If 'true', the playbook will be executed on every 'terraform apply'." +
					"If 'false', the playbook will be executed only on the first 'terraform apply'.",
			},

			// ansible execution commands
			"verbosity": { // verbosity is between = (0, 6)
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				Default:  0,
				Description: "A verbosity level between 0 and 6." +
					"Set ansible 'verbose' parameter, which causes Ansible to print more debug messages." +
					"The higher the 'verbosity', the more debug details will be printed.",
			},

			"tags": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of tags of plays and tasks to run.",
			},

			"limit": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of hosts to exclude from the playbook execution.",
			},

			"check_mode": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
				Description: "If 'true', playbook execution won't make any changes but " +
					"only change predictions will be made.",
			},

			"diff_mode": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
				Description: "" +
					"If 'true', when changing (small) files and templates, differences in those files will be shown." +
					"Recommended usage with 'check_mode'.",
			},

			// connection configs are handled with extra_vars
			"force_handlers": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Default:     false,
				Description: "If 'true', run handlers even if a task fails.",
			},

			// become configs are handled with extra_vars --> these are also connection configs
			"extra_vars": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "A map of additional variables as: { key-1 = value-1, key-2 = value-2, ... }.",
			},

			"var_files": { // adds @ at the beginning of filename
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of variable files.",
			},

			// Ansible Vault
			"vault_files": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of vault files.",
			},

			"vault_password_file": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "",
				Description: "Path to a vault password file.",
			},

			"vault_id": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Default:     "",
				Description: "ID of the desired vault(s)",
			},

			// computed
			"play_first_time": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Used to check if the playbook is being played for the first time (first 'terraform apply'.",
			},
			// debug output
			"args": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "Used to build arguments to run Ansible playbook with.",
			},
			// envs
			"env_vars": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Description: "A list of environment variables passed through Terraform." +
					"All environment variables for this resource, must have a prefix string 'ANSIBLE'.",
			},

			"temp_inventory_file": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Path to created temporary inventory file.",
			},
		},
	}
}

func resourcePlaybookCreate(data *schema.ResourceData, meta interface{}) error {
	// required settings
	playbook := provider_utils.GetParameterValue(data, "playbook", ap).(string)

	// optional settings
	name := provider_utils.GetParameterValue(data, "name", ap).(string)
	verbosity := provider_utils.GetParameterValue(data, "verbosity", ap).(int)
	tags := provider_utils.GetParameterValue(data, "tags", ap).([]interface{})
	limit := provider_utils.GetParameterValue(data, "limit", ap).([]interface{})
	checkMode := provider_utils.GetParameterValue(data, "check_mode", ap).(bool)
	diffMode := provider_utils.GetParameterValue(data, "diff_mode", ap).(bool)
	forceHandlers := provider_utils.GetParameterValue(data, "force_handlers", ap).(bool)
	extraVars := provider_utils.GetParameterValue(data, "extra_vars", ap).(map[string]interface{})

	varFiles := provider_utils.GetParameterValue(data, "var_files", ap).([]interface{})
	vaultFiles := provider_utils.GetParameterValue(data, "vault_files", ap).([]interface{})
	vaultPasswordFile := provider_utils.GetParameterValue(data, "vault_password_file", ap).(string)
	vaultID := provider_utils.GetParameterValue(data, "vault_id", ap).(string)

	data.SetId(playbook)

	if err := data.Set("play_first_time", true); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'playbook'! %s", err)
	}

	// Get environment vars: All environment variables MUST have a prefix "ANSIBLE"
	envVars := provider_utils.GetAnsibleEnvironmentVars()
	log.Print("[ENV VARS]:")
	log.Print(envVars)

	/********************
	* 	PREP THE OPTIONS (ARGS)
	 */
	args := []string{}

	verbose := provider_utils.CreateVerboseSwitch(verbosity)
	if verbose != "" {
		args = append(args, verbose)
	}

	if forceHandlers {
		args = append(args, "--force-handlers")
	}

	args = append(args, "-e", "hostname="+name)

	if len(tags) > 0 {
		tmpTags := []string{}
		for _, tag := range tags {
			tagStr := tag.(string)
			tmpTags = append(tmpTags, tagStr)
		}
		tagsStr := strings.Join(tmpTags, ",")
		args = append(args, "--tags", tagsStr)
	}

	if len(limit) > 0 {
		tmpLimit := []string{}
		for _, l := range limit {
			limitStr := l.(string)
			tmpLimit = append(tmpLimit, limitStr)
		}
		limitStr := strings.Join(tmpLimit, ",")
		args = append(args, "--limit", limitStr)
	}

	if checkMode {
		args = append(args, "--check")
	}

	if diffMode {
		args = append(args, "--diff")
	}

	// Pass environment variables to extra vars
	for _, envVar := range envVars {
		args = append(args, "-e", envVar)
	}

	if len(varFiles) != 0 {
		for _, varFile := range varFiles {
			varFileString := varFile.(string)
			args = append(args, "-e", "@"+varFileString)
		}
	}

	// Ansible vault
	if len(vaultFiles) != 0 {
		for _, vaultFile := range vaultFiles {
			vaultFileString := vaultFile.(string)
			args = append(args, "-e", "@"+vaultFileString)
		}
		args = append(args, "--vault-id")

		vaultIdArg := ""
		if vaultID != "" {
			vaultIdArg += vaultID
		}

		if vaultPasswordFile != "" {
			vaultIdArg += "@" + vaultPasswordFile
		} else {
			log.Fatal("ERROR [ansible-playbook]: can't access vault file(s)! Missing 'vault_password_file'!")
		}
		args = append(args, vaultIdArg)
	}

	if len(extraVars) != 0 {
		for key, val := range extraVars {
			tmpVal, okay := val.(string)
			if !okay {
				log.Fatal("ERROR [ansible-playbook]: couldn't assert type: string")
			}

			args = append(args, "-e", key+"="+tmpVal)
		}
	}
	args = append(args, playbook)

	// set up the args
	log.Print("[ANSIBLE ARGS]:")
	log.Print(args)

	if err := data.Set("args", args); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'args'! %v", err)
	}
	if err := data.Set("env_vars", envVars); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'env_vars'! %v", err)
	}

	return resourcePlaybookUpdate(data, meta)
}

func resourcePlaybookRead(data *schema.ResourceData, meta interface{}) error {
	ansiblePlaybookBinary := provider_utils.GetParameterValue(data, "ansible_playbook_binary", ap).(string)

	playbook := provider_utils.GetParameterValue(data, "playbook", ap).(string)
	log.Printf("LOG [ansible-playbook]: playbook = %s", playbook)

	argsTf := provider_utils.GetParameterValue(data, "args", ap).([]interface{})
	replayable := provider_utils.GetParameterValue(data, "replayable", ap).(bool)
	playFirstTime := provider_utils.GetParameterValue(data, "play_first_time", ap).(bool)

	if playFirstTime || replayable {
		args := []string{}

		// Get the rest of args
		for _, arg := range argsTf {
			tmpArg, okay := arg.(string)
			if !okay {
				log.Fatal("ERROR [ansible-playbook]: couldn't assert type: string")
			}

			args = append(args, tmpArg)
		}

		runAnsiblePlay := exec.Command(ansiblePlaybookBinary, args...)

		runAnsiblePlayOut, runAnsiblePlayErr := runAnsiblePlay.CombinedOutput()
		if runAnsiblePlayErr != nil {
			log.Fatalf("ERROR [ansible-playbook]: couldn't run ansible-playbook\n%s! There may be an error within your playbook.\n%v", playbook, runAnsiblePlayErr)
		}

		log.Printf("LOG [ansible-playbook]: %s", runAnsiblePlayOut)

		if err := data.Set("play_first_time", false); err != nil {
			log.Fatal("ERROR [ansible-playbook]: couldn't set 'play_first_time'!")
		}
	}

	return nil
}

func resourcePlaybookUpdate(data *schema.ResourceData, meta interface{}) error {
	playbook := provider_utils.GetParameterValue(data, "playbook", ap).(string)
	data.SetId(playbook + "-taint")

	name := provider_utils.GetParameterValue(data, "name", ap).(string)
	groups := provider_utils.GetParameterValue(data, "groups", ap).([]interface{})
	argsTf := provider_utils.GetParameterValue(data, "args", ap).([]interface{})

	args := []string{}

	for _, arg := range argsTf {
		tmpArg, okay := arg.(string)
		if !okay {
			log.Fatal("ERROR [ansible-playbook]: couldn't assert type: string")
		}

		args = append(args, tmpArg)
	}

	inventoryFileName := ".inventory-*" + ".ini" // playbook --> resource ID

	createdTempInventory := provider_utils.BuildPlaybookInventory(inventoryFileName, name, -1, groups)
	if err := data.Set("temp_inventory_file", createdTempInventory); err != nil {
		log.Fatal("ERROR [ansible-playbook]: couldn't set 'temp_inventory_file'!")
	}

	// Get all available temp inventories and pass them as args
	inventories := provider_utils.GetAllInventories()
	log.Print("[INVENTORIES]:")
	log.Print(inventories)
	for _, inventory := range inventories {
		args = append(args, "-i", inventory)
	}

	if err := data.Set("args", args); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'args'! %s", err)
	}

	data.SetId(playbook)
	return resourcePlaybookRead(data, meta)
}

// On "terraform destroy", every resource removes its temporary inventory file
func resourcePlaybookDelete(data *schema.ResourceData, meta interface{}) error {
	tempInventoryFile := provider_utils.GetParameterValue(data, "temp_inventory_file", ap).(string)
	log.Printf("Removing file %s.", tempInventoryFile)

	provider_utils.RemoveFile(tempInventoryFile)

	return nil
}
