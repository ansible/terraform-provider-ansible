package providerutils

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

func (p *ResourceDataParser) HasError() bool {
	return p.Diags.HasError()
}

func (p *ResourceDataParser) Get(key string) interface{} {
	result := p.Data.Get(key)
	if result == nil {
		p.Diags = append(p.Diags, diag.Errorf("couldn't get key '%s'!", key)...)
	}

	return result
}

func assertType[T int | string | bool](key string, data interface{}, value *T) diag.Diagnostics {
	var diags diag.Diagnostics
	if data == nil {
		return diags
	}

	tmpValue, ok := data.(T)
	if !ok {
		diags = diag.Errorf("couldn't assert %T: for key [%s]", reflect.TypeOf(*value), key)

		return diags
	}

	*value = tmpValue

	return diags
}

func (p *ResourceDataParser) ReadString(key string, value *string) {
	p.Diags = append(p.Diags, assertType(key, p.Get(key), value)...)
}

func (p *ResourceDataParser) ReadInt(key string, value *int) {
	p.Diags = append(p.Diags, assertType(key, p.Get(key), value)...)
}

func (p *ResourceDataParser) ReadBool(key string, value *bool) {
	p.Diags = append(p.Diags, assertType(key, p.Get(key), value)...)
}

func (p *ResourceDataParser) ReadStringList(key string, value *[]string) {
	data := p.Get(key)
	if data != nil {
		dataList, ok := data.([]interface{})
		if !ok {
			p.Diags = append(p.Diags, diag.Errorf("Could not assert type []string for key [%s]", key)...)
		} else if len(dataList) > 0 {
			tmpResult := []string{}
			for idx, item := range dataList {
				var itemStr string
				p.Diags = append(p.Diags, assertType(fmt.Sprintf("%s.%d", key, idx), item, &itemStr)...)
				if len(itemStr) > 0 {
					tmpResult = append(tmpResult, itemStr)
				}
			}
			*value = tmpResult
		}
	}
}

func (p *ResourceDataParser) ReadMapString(key string, value *map[string]string) {
	data := p.Get(key)
	if data != nil {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			p.Diags = append(p.Diags, diag.Errorf("Could not assert type map[string]interface{} for key [%s]", key)...)
		} else {
			tmpMap := make(map[string]string)
			for keyItem, valItem := range dataMap {
				var itemStr string
				p.Diags = append(p.Diags, assertType(fmt.Sprintf("%s.%s", key, keyItem), valItem, &itemStr)...)
				if len(itemStr) > 0 {
					tmpMap[keyItem] = itemStr
				}
			}
			*value = tmpMap
		}
	}
}
