package provider

import (
	"github.com/ansible/terraform-provider-ansible/provider_utils"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"os"
	"os/exec"
	"strings"
)

const MainInventory = ".inventory.ini"
const ap = "ansible-playbook"

func resourcePlaybook() *schema.Resource {
	return &schema.Resource{
		Create: resourcePlaybookCreate,
		Read:   resourcePlaybookRead,
		Update: resourcePlaybookUpdate,
		Delete: resourcePlaybookDelete,

		Schema: map[string]*schema.Schema{
			// Required settings
			"ansible_playbook_binary": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},

			"playbook": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},

			// Optional settings
			"name": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "localhost",
			},

			"groups": {
				Type:     schema.TypeList,
				Elem:     schema.TypeString,
				Required: false,
				Optional: true,
			},

			"replayable": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  true,
			},

			// ansible execution commands
			"verbosity": { // verbosity is between = (0, 6)
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				Default:  0,
			},

			"tags": {
				Type:     schema.TypeList,
				Elem:     schema.TypeString,
				Required: false,
				Optional: true,
			},

			"limit": {
				Type:     schema.TypeList,
				Elem:     schema.TypeString,
				Required: false,
				Optional: true,
			},

			"check_mode": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
			},

			"diff_mode": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
			},

			// keys
			"private_key": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
			},
			// connection configs are handled with extra_vars
			"force_handlers": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
			},

			// become configs are handled with extra_vars --> these are also connection configs
			"inventory": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
				// Default: "cloud.terraform.terraform_provider",
			},
			"extra_vars": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: false,
				Optional: true,
			},

			"var_files": { // add @ at the beginning of filename
				Type:     schema.TypeList,
				Elem:     schema.TypeString,
				Required: false,
				Optional: true,
			},

			// Ansible Vault
			"vault_files": {
				Type:     schema.TypeList,
				Elem:     schema.TypeString,
				Required: false,
				Optional: true,
			},

			"vault_password_file": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
			},

			"vault_id": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
			},

			// envs
			"ansible_config": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
			},
			"environment_vars": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: false,
				Optional: true,
			},

			// computed
			"play_first_time": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			// debug output
			"args": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
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
	privateKey := provider_utils.GetParameterValue(data, "private_key", ap).(string)
	forceHandlers := provider_utils.GetParameterValue(data, "force_handlers", ap).(bool)
	inventory := provider_utils.GetParameterValue(data, "inventory", ap).(string)
	extraVars := provider_utils.GetParameterValue(data, "extra_vars", ap).(map[string]interface{})

	varFiles := provider_utils.GetParameterValue(data, "var_files", ap).([]interface{})
	vaultFiles := provider_utils.GetParameterValue(data, "vault_files", ap).([]interface{})
	vaultPasswordFile := provider_utils.GetParameterValue(data, "vault_password_file", ap).(string)
	vaultID := provider_utils.GetParameterValue(data, "vault_id", ap).(string)

	data.SetId(playbook)

	if err := data.Set("play_first_time", true); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'playbook'! %s", err)
	}

	/********************
	* 	PREP THE OPTIONS (ARGS)
	 */
	args := []string{}

	verbose := provider_utils.CreateVerboseSwitch(verbosity)
	if verbose != "" {
		args = append(args, verbose)
	}

	if privateKey != "" {
		args = append(args, "--private-key", privateKey)
	}

	if forceHandlers {
		args = append(args, "--force-handlers")
	}

	if inventory != "" {
		args = append(args, "-i", inventory)
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
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'args'! %s", err)
	}

	return resourcePlaybookUpdate(data, meta)
}

func resourcePlaybookRead(data *schema.ResourceData, meta interface{}) error {
	ansiblePlaybookBinary := provider_utils.GetParameterValue(data, "ansible_playbook_binary", ap).(string)

	ansibleConfig := provider_utils.GetParameterValue(data, "ansible_config", ap).(string)
	environmentVars := provider_utils.GetParameterValue(data, "environment_vars", ap).(map[string]interface{})
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

		// args = append(args, "-i", cwd+MainInventory)
		runAnsiblePlay := exec.Command(ansiblePlaybookBinary, args...)

		if ansibleConfig != "" {
			runAnsiblePlay.Env = os.Environ()
			runAnsiblePlay.Env = append(runAnsiblePlay.Env, "ANSIBLE_CONFIG="+ansibleConfig)
		}

		if len(environmentVars) != 0 {
			runAnsiblePlay.Env = os.Environ()

			for key, env := range environmentVars {
				tmpEnv, okay := env.(string)
				if !okay {
					log.Fatal("ERROR [ansible-playbook]: couldn't assert type: string")
				}

				environ := key + "=" + tmpEnv
				runAnsiblePlay.Env = append(runAnsiblePlay.Env, environ)
			}
		}

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

	cwd := provider_utils.GetCurrentDir()
	provider_utils.BuildPlaybookInventory(cwd+MainInventory, name, -1, groups)
	args = append(args, "-i", cwd+MainInventory)

	if err := data.Set("args", args); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'args'! %s", err)
	}

	data.SetId(playbook)
	return resourcePlaybookRead(data, meta)
}

func resourcePlaybookDelete(data *schema.ResourceData, meta interface{}) error {
	data.SetId("")

	cwd := provider_utils.GetCurrentDir()
	provider_utils.RemoveFile(cwd + MainInventory)

	return nil
}
