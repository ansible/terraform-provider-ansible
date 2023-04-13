package providerutils

import (
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"gopkg.in/ini.v1"
)

/*
	CREATE OPTIONS
*/

const DefaultHostGroup = "default"

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

// Build inventory.ini (NOT YAML)
//  -- building inventory.ini is easier

func BuildPlaybookInventory(inventoryDest string, hostname string, port int, hostgroups []interface{}) string {
	// Check if inventory file is already present
	// if not, create one
	tempDir := os.TempDir() + "/"
	inventoryTempPath := tempDir + inventoryDest //nolint:all

	var tempFileName string

	fileInfo, err := os.CreateTemp("", inventoryDest)
	if err != nil {
		log.Fatalf("Fail to create inventory file: %v", err)
	}

	tempFileName = fileInfo.Name()
	log.Printf("Inventory %s was created", fileInfo.Name())

	inventoryTempPath = tempFileName

	// Then, read inventory and add desired settings to it
	inventory, err := ini.Load(inventoryTempPath)
	if err != nil {
		log.Printf("Fail to read inventory: %v", err)
	}

	tempHostgroups := hostgroups

	if len(tempHostgroups) == 0 {
		tempHostgroups = append(tempHostgroups, DefaultHostGroup)
	}

	if len(tempHostgroups) > 0 { // if there is a list of groups specified for the desired host
		for _, hostgroup := range tempHostgroups {
			hostgroupStr, okay := hostgroup.(string)
			if !okay {
				log.Fatalf("Couldn't assert type: string")
			}

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
	}

	err = inventory.SaveTo(inventoryTempPath)
	if err != nil {
		log.Fatalf("Fail to create inventory: %v", err)
	}

	return inventoryTempPath
}

func RemoveFile(filename string) {
	err := os.Remove(filename)
	if err != nil {
		log.Fatalf("Fail to remove file %s: %v", filename, err)
	}
}

func GetAllInventories() []string {
	tempDir := os.TempDir()
	log.Printf("[TEMP DIR]: %s", tempDir)

	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		log.Fatalf("Fail to read dir %s: %v", tempDir, err)
	}

	inventories := []string{}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".inventory-") {
			inventoryAbsPath := tempDir + "/" + file.Name()
			inventories = append(inventories, inventoryAbsPath)
		}
	}

	return inventories
}

// Get current working directory --- cwd.
func GetCurrentDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Fail to get current working directory: %v", err)
	}

	log.Printf("[MY CWD]: %s", cwd)

	return cwd + "/"
}

func GetParameterValue(data *schema.ResourceData, parameterKey string, resourceName string) interface{} {
	val, okay := data.Get(parameterKey).(interface{}) //nolint:all
	if !okay {
		log.Fatalf("ERROR [%s]: couldn't get '%s'!", resourceName, parameterKey)
	}

	return val
}

func GetAnsibleEnvironmentVars() []string {
	return os.Environ()
}

func GeneratedHashString(str string) string {
	hash := fnv.New32a()

	if _, err := hash.Write([]byte(str)); err != nil {
		log.Fatalf("Fail to generate a hash: %v", err)
	}

	hashedUint32 := hash.Sum32()

	return strconv.Itoa(int(hashedUint32))
}
