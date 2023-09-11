// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os/exec"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"ansible": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	_, err := exec.LookPath("ansible-playbook")
	if err != nil {
		t.Fatal(err)
	}

	_, err = exec.LookPath("docker")
	if err != nil {
		t.Fatal(err)
	}

	// exec.Command("docker run -it ", arg ...string)
}

// func testAccPost(t *testing.T) {

// }
