package provider

import (
	"gopkg.in/ini.v1"
	"log"
	"os"
	"strconv"
	"strings"
)

/*
	CREATE OPTIONS
*/

func interfaceToString(arr []interface{}) []string {
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

func createVerboseSwitch(verbosity int) string {
	verbose := ""

	if verbosity == 0 {
		return verbose
	}

	verbose += "-"
	verbose += strings.Repeat("v", verbosity)

	return verbose
}

// Build inventory.ini (NOT YAML)
// -- building inventory.ini is easier
func buildProvisionInventory(inventoryTemplatePath string, inventoryDest string, hostname string, port int, hostgroup string) {
	/*
		TODO: If the ini file doesn't exist, create one!
		note: might not need inventory template file.
	*/

	inventory, err := ini.Load(inventoryTemplatePath)
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	if hostgroup != "" {
		if !inventory.HasSection(hostgroup) {
			_, err := inventory.NewSection(hostgroup)
			if err != nil {
				log.Fatalf("Fail to create a hostgroup: %v", err)
			}
		}
	}

	if !inventory.Section(hostgroup).HasKey(hostname) {
		_, err := inventory.Section(hostgroup).NewKey(hostname, "")
		if err != nil {
			log.Fatalf("Fail to create a host: %v", err)
		}
		if port != -1 {
			portString := strconv.Itoa(port)
			err = inventory.Section(hostgroup).Key(hostname).AddShadow("ansible_port=" + portString)
			if err != nil {
				log.Fatalf("Fail to create port: %v", err)
			}
		}
	}

	err = inventory.SaveTo(inventoryDest)
	if err != nil {
		log.Fatalf("Fail to create inventory: %v", err)
	}
}

func getCurrentDir() string {
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Fail to get current working directory: %v", err)
	}

	return cwd + "/"
}
