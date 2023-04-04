package provider_utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"strconv"
	"strings"
)

/*
	CREATE OPTIONS
*/

func InterfaceToString(arr []interface{}) []string {
	result := []string{}

	for _, val := range arr {
		tmpVal, ok := val.(string)
		if !ok {
			log.Fatal("Error: couldn't parse value to string!")
		}

		result = append(result, tmpVal)
	}

	return result
}

// Create a "verbpse" switch
// example: verbosity = 2 --> verbose_switch = "-vv"
func CreateVerboseSwitch(verbosity int) string {
	verbose := ""

	if verbosity == 0 {
		return verbose
	}

	verbose += "-"
	verbose += strings.Repeat("v", verbosity)

	return verbose
}

/*
	Build inventory.ini (NOT YAML)
	-- building inventory.ini is easier
*/
func BuildPlaybookInventory(inventoryDest string, hostname string, port int, hostgroups []interface{}) {
	// Check if inventory file is already present
	// if not, create one
	if _, err := os.Stat(inventoryDest); err != nil {
		log.Printf("Inventory %s doesn't exist. Creating one.%v", inventoryDest, err)
		f, err := os.Create(inventoryDest)
		if err != nil {
			log.Fatalf("Fail to create inventory file: %v", err)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Fatalf("Fail to close inventory file: %v", err)
			}
		}(f)
		log.Printf("Inventory %s was created", f.Name())
	}

	// Then, read inventory and add desired settings to it
	inventory, err := ini.Load(inventoryDest)
	if err != nil {
		log.Printf("Fail to read inventory: %v", err)
	}

	if len(hostgroups) > 0 { // if there is a list of groups specified for the desired host
		for _, hostgroup := range hostgroups {
			hostgroupStr := hostgroup.(string)

			if !inventory.HasSection(hostgroupStr) {
				_, err := inventory.NewRawSection(hostgroupStr, "")
				if err != nil {
					log.Fatalf("Fail to create a hostgroup: %v", err)
				}
			}

			if !inventory.Section(hostgroupStr).HasKey(hostname) {
				body := hostname
				if port != -1 {
					body += " ansible_port=" + strconv.Itoa(port)
				}

				inventory.Section(hostgroupStr).SetBody(body)
			}
		}
	} else { // if there are no groups specified, Section("") means empty group or no group
		if !inventory.Section("").HasKey(hostname) {
			body := hostname
			if port != -1 {
				body += " ansible_port=" + strconv.Itoa(port)
			}

			inventory.Section("").SetBody(body)
		}
	}

	err = inventory.SaveTo(inventoryDest)
	if err != nil {
		log.Fatalf("Fail to create inventory: %v", err)
	}
}

func RemoveFile(filename string) {
	err := os.Remove(filename)
	if err != nil {
		log.Fatalf("Fail to remove file %s: %v", filename, err)
	}
}

// Get current working directory --- cwd
func GetCurrentDir() string {
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Fail to get current working directory: %v", err)
	}

	log.Printf("[MY CWD]: %s", cwd)
	return cwd + "/"
}

func GetParameterValue(data *schema.ResourceData, parameterKey string, resourceName string) interface{} {
	val, okay := data.Get(parameterKey).(interface{})
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get '%s'!", resourceName, parameterKey)
	}
	return val
}
