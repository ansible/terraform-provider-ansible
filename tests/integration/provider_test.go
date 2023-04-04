// integration tests package
package integration_tests_test

import (
	"log"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/stretchr/testify/assert"
)

const (
	actualJSON   = "../terraform_tests/actual_tfstate.json"
	expectedJSON = "../expected_tfstate.json"
)

func TestAnsibleProviderOutputs(t *testing.T) {
	t.Parallel()

	actual, errAct := gabs.ParseJSONFile(actualJSON)
	expected, errExp := gabs.ParseJSONFile(expectedJSON)

	// "serial" is a changing variable (it changes after
	//	every 'terraform destroy'), so we're not testing that.
	if _, err := expected.Set(actual.Path("serial").Data(), "serial"); err != nil {
		log.Fatalf("Error: couldn't ignore 'serial' field! %s", err)
	}

	// "lineage" is a changing variable (it is dependent on the
	//	terraform working directory), so we're not testing that.
	if _, err := expected.Set(actual.Path("lineage").Data(), "lineage"); err != nil {
		log.Fatalf("Error: couldn't ignore 'lineage' field! %s", err)
	}

	if errAct != nil {
		log.Fatal("Error in " + actualJSON + "!")
	}

	if errExp != nil {
		log.Fatal("Error in " + expectedJSON + "!")
	}

	assert.JSONEq(t, expected.String(), actual.String(), "Actual and Expected JSON files don't match!")
}
