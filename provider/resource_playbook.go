package provider

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ansible/terraform-provider-ansible/providerutils"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const ansiblePlaybook = "ansible-playbook"

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
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute), //nolint:gomnd
		},
	}
}

//nolint:maintidx
func resourcePlaybookCreate(data *schema.ResourceData, meta interface{}) error {
	// required settings
	playbook, okay := data.Get("playbook").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'playbook'!", ansiblePlaybook)
	}

	// optional settings
	name, okay := data.Get("name").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'name'!", ansiblePlaybook)
	}

	verbosity, okay := data.Get("verbosity").(int)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'verbosity'!", ansiblePlaybook)
	}

	tags, okay := data.Get("tags").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'tags'!", ansiblePlaybook)
	}

	limit, okay := data.Get("limit").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'limit'!", ansiblePlaybook)
	}

	checkMode, okay := data.Get("check_mode").(bool)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'check_mode'!", ansiblePlaybook)
	}

	diffMode, okay := data.Get("diff_mode").(bool)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'diff_mode'!", ansiblePlaybook)
	}

	forceHandlers, okay := data.Get("force_handlers").(bool)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'force_handlers'!", ansiblePlaybook)
	}

	extraVars, okay := data.Get("extra_vars").(map[string]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'extra_vars'!", ansiblePlaybook)
	}

	varFiles, okay := data.Get("var_files").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'var_files'!", ansiblePlaybook)
	}

	vaultFiles, okay := data.Get("vault_files").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'vault_files'!", ansiblePlaybook)
	}

	vaultPasswordFile, okay := data.Get("vault_password_file").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'vault_password_file'!", ansiblePlaybook)
	}

	vaultID, okay := data.Get("vault_id").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'vault_id'!", ansiblePlaybook)
	}

	// Generate ID
	resourceHash := providerutils.GeneratedHashString(time.Now().String())
	data.SetId(playbook + "-" + resourceHash)

	if err := data.Set("play_first_time", true); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'playbook'! %s", err)
	}

	// Get environment vars: All environment variables MUST have a prefix "ANSIBLE"
	envVars := providerutils.GetAnsibleEnvironmentVars()

	log.Print("[ENV VARS]:")
	log.Print(envVars)

	/********************
	* 	PREP THE OPTIONS (ARGS)
	 */
	args := []string{}

	verbose := providerutils.CreateVerboseSwitch(verbosity)
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
			tagStr, okay := tag.(string)
			if !okay {
				log.Fatalf("ERROR [%s]: couldn't assert type: string", ansiblePlaybook)
			}

			tmpTags = append(tmpTags, tagStr)
		}

		tagsStr := strings.Join(tmpTags, ",")
		args = append(args, "--tags", tagsStr)
	}

	if len(limit) > 0 {
		tmpLimit := []string{}

		for _, l := range limit {
			limitStr, okay := l.(string)
			if !okay {
				log.Fatalf("ERROR [%s]: couldn't assert type: string", ansiblePlaybook)
			}

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
			varFileString, okay := varFile.(string)
			if !okay {
				log.Fatalf("ERROR [%s]: couldn't assert type: string", ansiblePlaybook)
			}

			args = append(args, "-e", "@"+varFileString)
		}
	}

	// Ansible vault
	if len(vaultFiles) != 0 {
		for _, vaultFile := range vaultFiles {
			vaultFileString, okay := vaultFile.(string)
			if !okay {
				log.Fatalf("ERROR [%s]: couldn't assert type: string", ansiblePlaybook)
			}

			args = append(args, "-e", "@"+vaultFileString)
		}

		args = append(args, "--vault-id")

		vaultIDArg := ""
		if vaultID != "" {
			vaultIDArg += vaultID
		}

		if vaultPasswordFile != "" {
			vaultIDArg += "@" + vaultPasswordFile
		} else {
			log.Fatal("ERROR [ansible-playbook]: can't access vault file(s)! Missing 'vault_password_file'!")
		}

		args = append(args, vaultIDArg)
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
	// Make sure an inventory exists
	tempInventoryFile, okay := data.Get("temp_inventory_file").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'temp_inventory_file'!", ansiblePlaybook)
	}

	_, err := os.Stat(tempInventoryFile)
	if os.IsNotExist(err) {
		return resourcePlaybookUpdate(data, meta)
	}

	/* ================================= */

	ansiblePlaybookBinary, okay := data.Get("ansible_playbook_binary").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'ansible_playbook_binary'!", ansiblePlaybook)
	}

	playbook, okay := data.Get("playbook").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'playbook'!", ansiblePlaybook)
	}

	log.Printf("LOG [ansible-playbook]: playbook = %s", playbook)

	argsTf, okay := data.Get("args").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'args'!", ansiblePlaybook)
	}

	replayable, okay := data.Get("replayable").(bool)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'replayable'!", ansiblePlaybook)
	}

	playFirstTime, okay := data.Get("play_first_time").(bool)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'play_first_time'!", ansiblePlaybook)
	}

	log.Printf("[MY CURRENT ID IS]: %s", data.Id())

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
			log.Fatalf("ERROR [ansible-playbook]: couldn't run ansible-playbook\n%s! "+
				"There may be an error within your playbook.\n%v",
				playbook,
				runAnsiblePlayErr,
			)
		}

		log.Printf("LOG [ansible-playbook]: %s", runAnsiblePlayOut)

		if err := data.Set("play_first_time", false); err != nil {
			log.Fatal("ERROR [ansible-playbook]: couldn't set 'play_first_time'!")
		}

		// Wait for playbook execution to finish, then remove the temporary file
		err := runAnsiblePlay.Wait()
		if err != nil {
			log.Printf("ERROR [ansible-playbook]: couldn't wait for playbook to execute.")
		}

		providerutils.RemoveFile(tempInventoryFile)
	}

	return nil
}

func resourcePlaybookUpdate(data *schema.ResourceData, meta interface{}) error {
	originalID := data.Id()

	data.SetId(originalID + "-taint")

	name, okay := data.Get("name").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'name'!", ansiblePlaybook)
	}

	groups, okay := data.Get("groups").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'groups'!", ansiblePlaybook)
	}

	argsTf, okay := data.Get("args").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'args'!", ansiblePlaybook)
	}

	args := []string{}

	for _, arg := range argsTf {
		tmpArg, okay := arg.(string)
		if !okay {
			log.Fatal("ERROR [ansible-playbook]: couldn't assert type: string")
		}

		args = append(args, tmpArg)
	}

	inventoryFileName := ".inventory-*" + ".ini" // playbook --> resource ID

	createdTempInventory := providerutils.BuildPlaybookInventory(inventoryFileName, name, -1, groups)
	if err := data.Set("temp_inventory_file", createdTempInventory); err != nil {
		log.Fatal("ERROR [ansible-playbook]: couldn't set 'temp_inventory_file'!")
	}

	// Get all available temp inventories and pass them as args
	inventories := providerutils.GetAllInventories()

	log.Print("[INVENTORIES]:")
	log.Print(inventories)

	for _, inventory := range inventories {
		args = append(args, "-i", inventory)
	}

	if err := data.Set("args", args); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'args'! %s", err)
	}

	data.SetId(originalID)

	return resourcePlaybookRead(data, meta)
}

// On "terraform destroy", every resource removes its temporary inventory file.
func resourcePlaybookDelete(data *schema.ResourceData, meta interface{}) error {
	data.SetId("")

	return nil
}
