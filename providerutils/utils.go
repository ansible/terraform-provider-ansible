package providerutils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	var diags diag.Diagnostics
	// Check if inventory file is already present
	// if not, create one
	fileInfo, err := os.CreateTemp("", inventoryDest)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Fail to create inventory file: %v", err),
		})
	}

	tempFileName := fileInfo.Name()
	log.Printf("Inventory %s was created", fileInfo.Name())

	// Then, read inventory and add desired settings to it
	inventory, err := ini.Load(tempFileName)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Fail to read inventory: %v", err),
		})
	}

	tempHostgroups := hostgroups

	if len(tempHostgroups) == 0 {
		tempHostgroups = append(tempHostgroups, DefaultHostGroup)
	}

	if len(tempHostgroups) > 0 { // if there is a list of groups specified for the desired host
		for _, hostgroup := range tempHostgroups {
			hostgroupStr, okay := hostgroup.(string)
			if !okay {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Couldn't assert type: string",
				})
			}

			if !inventory.HasSection(hostgroupStr) {
				_, err := inventory.NewRawSection(hostgroupStr, "")
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("Fail to create a hostgroup: %v", err),
					})
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

	err = inventory.SaveTo(tempFileName)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Fail to create inventory: %v", err),
		})
	}

	return tempFileName
}

func RemoveFile(filename string) {
	var diags diag.Diagnostics

	err := os.Remove(filename)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Fail to remove file %s: %v", filename, err),
		})
	}
}

func GetAllInventories(inventoryPrefix string) []string {
	var diags diag.Diagnostics
	tempDir := os.TempDir()
	log.Printf("[TEMP DIR]: %s", tempDir)

	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Fail to read dir %s: %v", tempDir, err),
		})
	}

	inventories := []string{}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), inventoryPrefix) {
			inventoryAbsPath := filepath.Join(tempDir, file.Name())
			inventories = append(inventories, inventoryAbsPath)
		}
	}

	return inventories
}
