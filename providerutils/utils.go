package providerutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/ini.v1"
)

/*
	CREATE OPTIONS
*/

const DefaultHostGroup = "default"

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

func BuildPlaybookInventory(
	inventoryDest string,
	hostname string,
	port int,
	hostgroups []string,
) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	// Check if inventory file is already present
	// if not, create one
	fileInfo, err := os.CreateTemp("", inventoryDest)
	if err != nil {
		diags = append(diags, diag.Errorf("Fail to create inventory file: %v", err)...)
	}

	tempFileName := fileInfo.Name()

	// Then, read inventory and add desired settings to it
	inventory, err := ini.Load(tempFileName)
	if err != nil {
		diags = append(diags, diag.Errorf("Fail to read inventory: %v", err)...)
	}

	tempHostgroups := hostgroups

	if len(tempHostgroups) == 0 {
		tempHostgroups = append(tempHostgroups, DefaultHostGroup)
	}

	if len(tempHostgroups) > 0 { // if there is a list of groups specified for the desired host
		for _, hostgroup := range tempHostgroups {
			if !inventory.HasSection(hostgroup) {
				_, err = inventory.NewRawSection(hostgroup, "")
				if err != nil {
					diags = append(diags, diag.Errorf("Fail to create a hostgroup: %v", err)...)
				}
			}

			if !inventory.Section(hostgroup).HasKey(hostname) {
				body := hostname
				if port != -1 {
					body += " ansible_port=" + strconv.Itoa(port)
				}

				inventory.Section(hostgroup).SetBody(body)
			}
		}
	}

	err = inventory.SaveTo(tempFileName)

	if err != nil {
		diags = append(diags, diag.Errorf("Fail to create inventory: %v", err)...)
	}

	return tempFileName, diags
}

func RemoveFile(filename string) diag.Diagnostics {
	var diags diag.Diagnostics

	err := os.Remove(filename)
	if err != nil {
		diags = append(diags, diag.Errorf("Fail to remove file %s: %v", filename, err)...)
	}

	return diags
}

func GetAllInventories(inventoryPrefix string) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	tempDir := os.TempDir()

	files, err := os.ReadDir(tempDir)
	if err != nil {
		diags = append(diags, diag.Errorf("Fail to read dir %s: %v", tempDir, err)...)
	}

	inventories := []string{}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), inventoryPrefix) {
			inventoryAbsPath := filepath.Join(tempDir, file.Name())
			inventories = append(inventories, inventoryAbsPath)
		}
	}

	return inventories, diags
}

type ResourceDataParser struct {
	Data   *schema.ResourceData
	Detail string
	Diags  diag.Diagnostics
}

func (p *ResourceDataParser) AddError(diagError diag.Diagnostic) {
	p.Diags = append(p.Diags, diagError)
}

func (p *ResourceDataParser) HasError() bool {
	return p.Diags.HasError()
}

func (p *ResourceDataParser) Get(key string) interface{} {
	result := p.Data.Get(key)
	if result == nil {
		p.AddError(diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("ERROR: couldn't get '%s'!", key),
			Detail:   p.Detail,
		})
	}
	return result
}

func (p *ResourceDataParser) ReadString(key string, value *string) {
	data := p.Get(key)
	if data != nil {
		*value = data.(string)
	}
}

func (p *ResourceDataParser) ReadInt(key string, value *int) {
	data := p.Get(key)
	if data != nil {
		*value = data.(int)
	}
}

func (p *ResourceDataParser) ReadBool(key string, value *bool) {
	data := p.Get(key)
	if data != nil {
		*value = data.(bool)
	}
}

func (p *ResourceDataParser) ReadStringList(key string, value *[]string) {
	data := p.Get(key)
	if data != nil {
		tmpValue := data.([]interface{})
		if len(tmpValue) > 0 {
			tmpResult := []string{}
			for _, item := range tmpValue {
				itemStr, ok := item.(string)
				if !ok {
					p.AddError(diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "ERROR: couldn't assert type: string",
						Detail:   p.Detail,
					})
				}
				tmpResult = append(tmpResult, itemStr)
			}
			*value = tmpResult
		}
	}
}

func (p *ResourceDataParser) ReadMapString(key string, value *map[string]string) {
	data := p.Get(key)
	if data != nil {
		var tmpMap = make(map[string]string)
		for key, val := range data.(map[string]interface{}) {
			tmpVal, okay := val.(string)
			if !okay {
				p.AddError(diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "ERROR: couldn't assert type: string",
					Detail:   p.Detail,
				})
			}
			tmpMap[key] = tmpVal
		}
		*value = tmpMap
	}
}
