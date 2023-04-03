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
func buildProvisionInventory(inventoryDest string, hostname string, port int, hostgroup string) {
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

	if hostgroup != "" {
		if !inventory.HasSection(hostgroup) {
			_, err := inventory.NewRawSection(hostgroup, "")
			if err != nil {
				log.Fatalf("Fail to create a hostgroup: %v", err)
			}
		}
	}

	if !inventory.Section(hostgroup).HasKey(hostname) {
		body := hostname
		if port != -1 {
			body += " ansible_port=" + strconv.Itoa(port)
		}

		inventory.Section(hostgroup).SetBody(body)
		//if err != nil {
		//	log.Fatalf("Fail to create a host: %v", err)
		//}
		//if port != -1 {
		//	portString := strconv.Itoa(port)
		//	err = inventory.Section(hostgroup).Key(hostname).AddShadow("ansible_port=" + portString)
		//	if err != nil {
		//		log.Fatalf("Fail to create port: %v", err)
		//	}
		//}
	}

	err = inventory.SaveTo(inventoryDest)
	if err != nil {
		log.Fatalf("Fail to create inventory: %v", err)
	}
}

func removeFile(filename string) {
	err := os.Remove(filename)
	if err != nil {
		log.Fatalf("Fail to remove file %s: %v", filename, err)
	}
}

func getCurrentDir() string {
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Fail to get current working directory: %v", err)
	}

	log.Printf("[MY CWD]: %s", cwd)
	return cwd + "/"
}
