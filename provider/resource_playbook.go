package provider

import (
	"fmt"
	"log"
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
		Exists: resourcePlaybookExists,

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
				Description: "Path to ansible-playbook executable (binary).",
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Name of the desired host on which the playbook will be executed.",
			},

			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				Optional:    false,
				Default:     22,
				Description: "The port used to access a desired host. The default value is 22.",
			},

			"groups": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    false,
				Optional:    true,
				Description: "List of desired groups of hosts on which the playbook will be executed.",
			},

			"replayable": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  true,
				Description: "" +
					"If 'true', the playbook will be executed on every 'terraform apply' and with that, the resource" +
					" will be recreated. " +
					"If 'false', the playbook will be executed only on the first 'terraform apply'. " +
					"Note, that if set to 'true', when doing 'terraform destroy', it might not show in the destroy " +
					"output, even though the resource still gets destroyed.",
			},

			"ignore_playbook_failure": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
				Description: "This parameter is good for testing. " +
					"Set to 'true' if the desired playbook is meant to fail, " +
					"but still want the resource to run successfully.",
			},

			// ansible execution commands
			"verbosity": { // verbosity is between = (0, 6)
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				Default:  0,
				Description: "A verbosity level between 0 and 6. " +
					"Set ansible 'verbose' parameter, which causes Ansible to print more debug messages. " +
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
					"If 'true', when changing (small) files and templates, differences in those files will be shown. " +
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
				Description: "ID of the desired vault(s).",
			},

			// computed
			// debug output
			"args": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "Used to build arguments to run Ansible playbook with.",
			},

			"temp_inventory_file": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Path to created temporary inventory file.",
			},

			"ansible_playbook_stdout": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An ansible-playbook CLI stdout output.",
			},

			"ansible_playbook_stderr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An ansible-playbook CLI stderr output.",
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
	data.SetId(time.Now().String())

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

	if err := data.Set("temp_inventory_file", ""); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'temp_inventory_file'! %v", err)
	}

	return resourcePlaybookUpdate(data, meta)
}

func resourcePlaybookRead(data *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourcePlaybookExists(data *schema.ResourceData, meta interface{}) (bool, error) {
	replayable, okay := data.Get("replayable").(bool)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'replayable'!", ansiblePlaybook)
	}

	// if (replayable == true) --> then we want to recreate (reapply) this resource: exits == false
	// if (replayable == false) --> we don't want to recreate (reapply) this resource: exists == true
	if replayable {
		// return false, and make sure to do destroy of this resource.
		return false, resourcePlaybookDelete(data, meta)
	}

	return true, nil
}

func resourcePlaybookUpdate(data *schema.ResourceData, meta interface{}) error {
	name, okay := data.Get("name").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'name'!", ansiblePlaybook)
	}

	port, okay := data.Get("port").(int)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'port'!", ansiblePlaybook)
	}

	groups, okay := data.Get("groups").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'groups'!", ansiblePlaybook)
	}

	ansiblePlaybookBinary, okay := data.Get("ansible_playbook_binary").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'ansible_playbook_binary'!", ansiblePlaybook)
	}

	playbook, okay := data.Get("playbook").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'playbook'!", ansiblePlaybook)
	}

	log.Printf("LOG [ansible-playbook]: playbook = %s", playbook)

	ignorePlaybookFailure, okay := data.Get("ignore_playbook_failure").(bool)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'ignore_playbook_failure'!", ansiblePlaybook)
	}

	argsTf, okay := data.Get("args").([]interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'args'!", ansiblePlaybook)
	}

	tempInventoryFile, okay := data.Get("temp_inventory_file").(string)
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get 'temp_inventory_file'!", ansiblePlaybook)
	}

	inventoryFileNamePrefix := ".inventory-"

	if tempInventoryFile == "" {
		tempInventoryFile = providerutils.BuildPlaybookInventory(inventoryFileNamePrefix+"*.ini", name, port, groups)
		if err := data.Set("temp_inventory_file", tempInventoryFile); err != nil {
			log.Fatal("ERROR [ansible-playbook]: couldn't set 'temp_inventory_file'!")
		}
	}

	log.Printf("Temp Inventory File: %s", tempInventoryFile)

	// ********************************* RUN PLAYBOOK ********************************

	args := []string{}

	args = append(args, "-i", tempInventoryFile)

	for _, arg := range argsTf {
		tmpArg, okay := arg.(string)
		if !okay {
			log.Fatal("ERROR [ansible-playbook]: couldn't assert type: string")
		}

		args = append(args, tmpArg)
	}

	runAnsiblePlay := exec.Command(ansiblePlaybookBinary, args...)

	runAnsiblePlayOut, runAnsiblePlayErr := runAnsiblePlay.CombinedOutput()
	ansiblePlayStderrString := ""

	if runAnsiblePlayErr != nil {
		playbookFailMsg := fmt.Sprintf("ERROR [ansible-playbook]: couldn't run ansible-playbook\n%s! "+
			"There may be an error within your playbook.\n%v",
			playbook,
			runAnsiblePlayErr,
		)
		if !ignorePlaybookFailure {
			log.Fatal(playbookFailMsg)
		} else {
			log.Print(playbookFailMsg)
		}

		ansiblePlayStderrString = runAnsiblePlayErr.Error()
	}

	// Set the ansible_playbook_stdout to the CLI stdout of call "ansible-playbook" command above
	if err := data.Set("ansible_playbook_stdout", string(runAnsiblePlayOut)); err != nil {
		log.Fatalf("ERROR [%s]: couldn't set 'ansible_playbook_stdout' ", ansiblePlaybook)
	}

	// Set the ansible_playbook_stderr to the CLI stderr of call "ansible-playbook" command above
	if err := data.Set("ansible_playbook_stderr", ansiblePlayStderrString); err != nil {
		log.Fatalf("ERROR [%s]: couldn't set 'ansible_playbook_stderr' ", ansiblePlaybook)
	}

	log.Printf("LOG [ansible-playbook]: %s", runAnsiblePlayOut)

	// Wait for playbook execution to finish, then remove the temporary file
	err := runAnsiblePlay.Wait()
	if err != nil {
		log.Printf("LOG [ansible-playbook]: didn't wait for playbook to execute: %v", err)
	}

	providerutils.RemoveFile(tempInventoryFile)

	if err := data.Set("temp_inventory_file", ""); err != nil {
		log.Fatalf("ERROR [ansible-playbook]: couldn't set 'temp_inventory_file'!")
	}

	// *******************************************************************************

	return resourcePlaybookRead(data, meta)
}

// On "terraform destroy", every resource removes its temporary inventory file.
func resourcePlaybookDelete(data *schema.ResourceData, meta interface{}) error {
	data.SetId("")

	return nil
}
