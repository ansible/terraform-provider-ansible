package provider

import (
	"log"
	"os"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceProvision() *schema.Resource {
	return &schema.Resource{
		Create: resourceProvisionCreate,
		Read:   resourceProvisionRead,
		Update: resourceProvisionUpdate,
		Delete: resourceProvisionDelete,

		Schema: map[string]*schema.Schema{
			// Required settings
			"playbook": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},

			// Optional settings
			"hostname": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "localhost",
			},

			"replayable": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
			},

			// ansible execution commands
			"verbosity": { // verbosity is between = (0, 6)
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
				Default:  0,
			},
			// keys
			"private_key": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
			},
			"remote_user": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "None",
			},
			// connections
			"ansible_connection": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "smart",
			},
			"ansible_connection_password_file": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
			},
			"force_handlers": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
			},
			// becomes
			"become": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
			},
			"become_method": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "sudo",
			},
			"become_user": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "root",
			},
			"become_password_file": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
			},
			"inventory": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  "",
				// Default: "cloud.terraform.terraform_provider",
			},
			"extra_vars": {
				Type:     schema.TypeMap,
				Required: false,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Ansible Vault
			"vault_secrets": { // vault_secrets = { [vault_1, vault_1_passwd], [vault_2, vault_2_passwd], ... }
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
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
				Required: false,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// computed
			"play_first_time": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			// debug output
			"args": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceProvisionCreate(data *schema.ResourceData, meta interface{}) error {
	// required settings
	playbook, okay := data.Get("playbook").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'playbook'!")
	}

	// optional settings
	hostName, okay := data.Get("hostname").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'hostname'!")
	}

	verbosity, okay := data.Get("verbosity").(int)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'verbosity'!")
	}

	privateKey, okay := data.Get("private_key").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'private_key'!")
	}

	remoteUser, okay := data.Get("remote_user").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'remote_user'!")
	}

	ansibleConnection, okay := data.Get("ansible_connection").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'ansible_connection'!")
	}

	ansibleConnectionPasswordFile, okay := data.Get("ansible_connection_password_file").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'ansible_connection_password_file'!")
	}

	forceHandlers, okay := data.Get("force_handlers").(bool)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'force_handlers'!")
	}

	become, okay := data.Get("become").(bool)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'become'!")
	}

	becomeMethod, okay := data.Get("become_method").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'become_method'!")
	}

	becomeUser, okay := data.Get("become_user").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'become_user'!")
	}

	becomePasswordFile, okay := data.Get("become_password_file").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'become_password_file'!")
	}

	inventory, okay := data.Get("inventory").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'inventory'!")
	}

	extraVars, okay := data.Get("extra_vars").(map[string]interface{})
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'extra_vars'!")
	}

	vaultSecrets, okay := data.Get("vault_secrets").([]interface{})
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'vault_secrets'!")
	}

	data.SetId(playbook)

	if err := data.Set("play_first_time", true); err != nil {
		log.Fatalf("ERROR [ansible-provision]: couldn't set 'playbook'! %s", err)
	}

	/********************
	* 	PREP THE OPTIONS (ARGS)
	 */
	args := []string{}

	verbose := createVerboseSwitch(verbosity)
	if verbose != "" {
		args = append(args, verbose)
	}

	if privateKey != "" {
		args = append(args, "--private-key", privateKey)
	}

	if remoteUser != "None" {
		args = append(args, "-u", remoteUser)
	}

	args = append(args, "-c", ansibleConnection)
	if ansibleConnectionPasswordFile != "" {
		args = append(args, "--connection-password-file", ansibleConnectionPasswordFile)
	}

	log.Print(len(args), args)

	if forceHandlers {
		args = append(args, "--force-handlers")
	}

	if become {
		args = append(args, "-b")
	}

	args = append(args, "--become-method", becomeMethod)
	args = append(args, "--become-user", becomeUser)

	if becomePasswordFile != "" {
		args = append(args, "--connection-password-file", ansibleConnectionPasswordFile)
	}

	if inventory != "" {
		args = append(args, "-i", inventory)
	}

	args = append(args, "-e", "hostname="+hostName)

	if len(vaultSecrets) != 0 {
		for _, val := range vaultSecrets {
			vaultData, okay := val.(map[string]interface{})
			if !okay {
				log.Fatal("ERROR [ansible-provision]: couldn't assert type: map[string]interface{}")
			}

			secretName, okay := vaultData["secret_name"].(string)
			if !okay {
				log.Fatal("ERROR [ansible-provision]: couldn't assert type: string")
			}

			vaultFile, okay := vaultData["vault_file"].(string)
			if !okay {
				log.Fatal("ERROR [ansible-provision]: couldn't assert type: string")
			}

			vaultPasswordFile, okay := vaultData["vault_password_file"].(string)
			if !okay {
				log.Fatal("ERROR [ansible-provision]: couldn't assert type: string")
			}

			// -e secret=@vault.yml --vault-password-file vault_password_file
			args = append(args, "-e", secretName+"=@"+vaultFile, "--vault-password-file", vaultPasswordFile)
		}
	}

	if len(extraVars) != 0 {
		for key, val := range extraVars {
			tmpVal, okay := val.(string)
			if !okay {
				log.Fatal("ERROR [ansible-provision]: couldn't assert type: string")
			}

			args = append(args, "-e", key+"="+tmpVal)
		}
	}

	args = append(args, playbook)

	// set up the args
	log.Print("[ANSIBLE ARGS]:")
	log.Print(args)

	if err := data.Set("args", args); err != nil {
		log.Fatalf("ERROR [ansible-provision]: couldn't set 'args'! %s", err)
	}

	//setupHost := exec.Command("ansible-playbook", ANSIBLE_HELPERS_PATH+"add_provision_host.yml", "-e", "hostname="+hostName)
	//runSetupHostOut, runSetupHostErr := setupHost.CombinedOutput()
	//if runSetupHostErr != nil {
	//	log.Fatalf("ERROR [ansible-playbook]: couldn't set up your host\n%s! The error is:\n%s", hostName, runSetupHostErr)
	//}
	//
	//log.Printf("LOG [ansible-provision]: host %s has been made reachable.", hostName)
	//log.Printf("LOG [ansible-provision]: %s", runSetupHostOut)

	return resourceProvisionRead(data, meta)
}

func resourceProvisionRead(data *schema.ResourceData, meta interface{}) error {
	ansibleConfig, okay := data.Get("ansible_config").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'ansible_config'!")
	}

	environmentVars, okay := data.Get("environment_vars").(map[string]interface{})
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'environment_vars'!")
	}

	playbook, okay := data.Get("playbook").(string)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'playbook'!")
	}

	log.Printf("LOG [ansible-provision]: playbook = %s", playbook)

	argsTf, okay := data.Get("args").([]interface{})
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'args'!")
	}

	replayable, okay := data.Get("replayable").(bool)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'replayable'!")
	}

	playFirstTime, okay := data.Get("play_first_time").(bool)
	if !okay {
		log.Fatal("ERROR [ansible-provision]: couldn't get 'replayable'!")
	}

	if playFirstTime || replayable {
		args := []string{}

		for _, arg := range argsTf {
			tmpArg, okay := arg.(string)
			if !okay {
				log.Fatal("ERROR [ansible-provision]: couldn't assert type: string")
			}

			args = append(args, tmpArg)
		}

		runAnsiblePlay := exec.Command("ansible-playbook", args...)

		if ansibleConfig != "" {
			runAnsiblePlay.Env = os.Environ()
			runAnsiblePlay.Env = append(runAnsiblePlay.Env, "ANSIBLE_CONFIG="+ansibleConfig)
		}

		if len(environmentVars) != 0 {
			runAnsiblePlay.Env = os.Environ()

			for key, env := range environmentVars {
				tmpEnv, okay := env.(string)
				if !okay {
					log.Fatal("ERROR [ansible-provision]: couldn't assert type: string")
				}

				environ := key + "=" + tmpEnv
				runAnsiblePlay.Env = append(runAnsiblePlay.Env, environ)
			}
		}

		runAnsiblePlayOut, runAnsiblePlayErr := runAnsiblePlay.CombinedOutput()
		if runAnsiblePlayErr != nil {
			log.Fatalf("ERROR [ansible-playbook]: couldn't run ansible-playbook\n%s! There may be an error within your playbook.\n%s", playbook, runAnsiblePlayErr)
		}

		log.Printf("LOG [ansible-provision]: %s", runAnsiblePlayOut)

		if err := data.Set("play_first_time", false); err != nil {
			log.Fatal("ERROR [ansible-provision]: couldn't set 'play_first_time'!")
		}
	}

	return nil
}

func resourceProvisionUpdate(data *schema.ResourceData, meta interface{}) error {
	return resourceProvisionRead(data, meta)
}

func resourceProvisionDelete(data *schema.ResourceData, meta interface{}) error {
	data.SetId("")

	return nil
}
