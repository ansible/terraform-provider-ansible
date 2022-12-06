package provider

import (
	"log"
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
