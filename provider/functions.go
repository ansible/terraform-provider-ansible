package provider

import (
	"log"
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
